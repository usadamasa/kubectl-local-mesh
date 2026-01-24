package run

import (
	"context"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/usadamasa/kubectl-localmesh/internal/config"
	"github.com/usadamasa/kubectl-localmesh/internal/envoy"
	"github.com/usadamasa/kubectl-localmesh/internal/gcp"
	"github.com/usadamasa/kubectl-localmesh/internal/k8s"
	"github.com/usadamasa/kubectl-localmesh/internal/log"
	"github.com/usadamasa/kubectl-localmesh/internal/loopback"
	"github.com/usadamasa/kubectl-localmesh/internal/port"
)

// RunVisitor は Run() 処理のための Visitor 実装
type RunVisitor struct {
	ctx        context.Context
	cfg        *config.Config
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
	logger     *log.Logger

	// loopback IPアロケータ（TCPサービス用）
	ipAllocator *loopback.IPAllocator

	// ポート競合チェッカー（TCPサービス用）
	portChecker *port.PortConflictChecker

	// 結果
	serviceConfigs   []envoy.ServiceConfig
	serviceSummaries []log.ServiceSummary
}

// NewRunVisitor は RunVisitor を生成
func NewRunVisitor(
	ctx context.Context,
	cfg *config.Config,
	clientset *kubernetes.Clientset,
	restConfig *rest.Config,
	logger *log.Logger,
) *RunVisitor {
	return &RunVisitor{
		ctx:              ctx,
		cfg:              cfg,
		clientset:        clientset,
		restConfig:       restConfig,
		logger:           logger,
		ipAllocator:      loopback.NewIPAllocator(),
		portChecker:      port.NewPortConflictChecker(),
		serviceConfigs:   make([]envoy.ServiceConfig, 0),
		serviceSummaries: make([]log.ServiceSummary, 0),
	}
}

// VisitKubernetes は Kubernetes Service の処理
func (v *RunVisitor) VisitKubernetes(s *config.KubernetesService) error {
	// ポート解決
	remotePort, err := k8s.ResolveServicePort(
		v.ctx,
		v.clientset,
		s.Namespace,
		s.Service,
		s.PortName,
		s.Port,
	)
	if err != nil {
		return err
	}

	// ローカルポート割り当て
	localPort, err := port.FreeLocalPort()
	if err != nil {
		return err
	}
	clusterName := sanitize(fmt.Sprintf("%s_%s_%d", s.Namespace, s.Service, remotePort))

	// ビルダー構築
	builder := envoy.NewKubernetesServiceBuilder(
		s.Host, s.Protocol, s.Namespace, s.Service, s.PortName, s.Port, s.ListenerPort,
	)

	v.logger.Debugf(
		"pf: %-30s -> %s/%s:%d via 127.0.0.1:%d",
		s.Host,
		s.Namespace,
		s.Service,
		remotePort,
		localPort,
	)

	// ServiceSummaryを追加
	var listenPort port.ListenerPort
	if s.ListenerPort != 0 {
		listenPort = s.ListenerPort
	}
	v.serviceSummaries = append(v.serviceSummaries, log.ServiceSummary{
		Host:        s.Host,
		Protocol:    s.Protocol,
		DisplayType: "HTTP/gRPC",
		Backend:     fmt.Sprintf("%s/%s:%d", s.Namespace, s.Service, remotePort),
		ListenPort:  listenPort,
	})

	// port-forwardをgoroutineで起動
	go func(ns, svc string, local port.LocalPort, remote port.ServicePort, logger *log.Logger) {
		if err := k8s.StartPortForwardLoop(
			v.ctx,
			v.restConfig,
			v.clientset,
			ns,
			svc,
			local,
			remote,
			logger,
		); err != nil {
			if v.ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "port-forward error for %s/%s: %v\n", ns, svc, err)
			}
		}
	}(s.Namespace, s.Service, localPort, remotePort, v.logger)

	// ServiceConfig を保存
	v.serviceConfigs = append(v.serviceConfigs, envoy.ServiceConfig{
		Builder:     builder,
		ClusterName: clusterName,
		LocalPort:   localPort,
	})

	return nil
}

// VisitTCP は TCP Service の処理
func (v *RunVisitor) VisitTCP(s *config.TCPService) error {
	// SSH Bastion確認
	bastion, ok := v.cfg.SSHBastions[s.SSHBastion]
	if !ok {
		return fmt.Errorf("ssh_bastion '%s' not found for service '%s'", s.SSHBastion, s.Host)
	}

	// ローカルポート割り当て
	localPort, err := port.FreeLocalPort()
	if err != nil {
		return err
	}
	clusterName := sanitize(fmt.Sprintf("tcp_%s_%s_%d", s.SSHBastion, s.TargetHost, s.TargetPort))

	// loopback IP割り当て（同一ポート重複を回避）
	listenAddr, err := v.ipAllocator.Allocate()
	if err != nil {
		return fmt.Errorf("failed to allocate loopback IP for service '%s': %w", s.Host, err)
	}

	// ポート競合チェック（IP:port の組み合わせでチェック）
	listenPort := s.ListenPort
	if listenPort == 0 {
		listenPort = s.TargetPort
	}
	v.portChecker.RegisterWithAddr(listenAddr, int(listenPort), s.Host)

	// ビルダー構築
	builder := envoy.NewTCPServiceBuilder(s.Host, s.ListenPort, listenAddr, s.SSHBastion, s.TargetHost, s.TargetPort)

	v.logger.Debugf(
		"gcp-ssh: %-30s -> %s (instance=%s, zone=%s) -> %s:%d via %s:%d",
		s.Host,
		s.SSHBastion,
		bastion.Instance,
		bastion.Zone,
		s.TargetHost,
		s.TargetPort,
		listenAddr,
		localPort,
	)

	// ServiceSummaryを追加
	v.serviceSummaries = append(v.serviceSummaries, log.ServiceSummary{
		Host:        s.Host,
		Protocol:    "tcp",
		DisplayType: "TCP",
		Backend:     fmt.Sprintf("%s @ %s:%d", s.SSHBastion, s.TargetHost, s.TargetPort),
		ListenPort:  port.ListenerPort(s.ListenPort),
	})

	// GCP SSH tunnelをgoroutineで起動
	go func(b *config.SSHBastion, local port.LocalPort, target string, targetPort port.TCPPort, logger *log.Logger) {
		if err := gcp.StartGCPSSHTunnel(
			v.ctx,
			b,
			local,
			target,
			targetPort,
			logger,
		); err != nil {
			if v.ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "gcp-ssh tunnel error for %s: %v\n", b.Instance, err)
			}
		}
	}(bastion, localPort, s.TargetHost, s.TargetPort, v.logger)

	// ServiceConfig を保存
	v.serviceConfigs = append(v.serviceConfigs, envoy.ServiceConfig{
		Builder:     builder,
		ClusterName: clusterName,
		LocalPort:   localPort,
	})

	return nil
}

// GetServiceConfigs は収集した ServiceConfig を返す
func (v *RunVisitor) GetServiceConfigs() []envoy.ServiceConfig {
	return v.serviceConfigs
}

// GetServiceSummaries は収集した ServiceSummary を返す
func (v *RunVisitor) GetServiceSummaries() []log.ServiceSummary {
	return v.serviceSummaries
}

// GetIPAllocator はIPアロケータを返す（エイリアス管理用）
func (v *RunVisitor) GetIPAllocator() *loopback.IPAllocator {
	return v.ipAllocator
}
