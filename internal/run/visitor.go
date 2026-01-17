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
	"github.com/usadamasa/kubectl-localmesh/internal/pf"
)

// RunVisitor は Run() 処理のための Visitor 実装
type RunVisitor struct {
	ctx        context.Context
	cfg        *config.Config
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
	logLevel   string

	// 結果
	serviceConfigs []envoy.ServiceConfig
}

// NewRunVisitor は RunVisitor を生成
func NewRunVisitor(
	ctx context.Context,
	cfg *config.Config,
	clientset *kubernetes.Clientset,
	restConfig *rest.Config,
	logLevel string,
) *RunVisitor {
	return &RunVisitor{
		ctx:            ctx,
		cfg:            cfg,
		clientset:      clientset,
		restConfig:     restConfig,
		logLevel:       logLevel,
		serviceConfigs: make([]envoy.ServiceConfig, 0),
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
	lp, err := pf.FreeLocalPort()
	if err != nil {
		return err
	}
	localPort := lp
	clusterName := sanitize(fmt.Sprintf("%s_%s_%d", s.Namespace, s.Service, remotePort))

	// ビルダー構築
	builder := envoy.NewKubernetesServiceBuilder(
		s.Host, s.Protocol, s.Namespace, s.Service, s.PortName, s.Port, s.OverwriteListenPorts,
	)

	fmt.Printf(
		"pf: %-30s -> %s/%s:%d via 127.0.0.1:%d\n",
		s.Host,
		s.Namespace,
		s.Service,
		remotePort,
		localPort,
	)

	// port-forwardをgoroutineで起動
	go func(ns, svc string, local, remote int) {
		if err := k8s.StartPortForwardLoop(
			v.ctx,
			v.restConfig,
			v.clientset,
			ns,
			svc,
			local,
			remote,
			v.logLevel,
		); err != nil {
			if v.ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "port-forward error for %s/%s: %v\n", ns, svc, err)
			}
		}
	}(s.Namespace, s.Service, localPort, remotePort)

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
	lp, err := pf.FreeLocalPort()
	if err != nil {
		return err
	}
	localPort := lp
	clusterName := sanitize(fmt.Sprintf("tcp_%s_%s_%d", s.SSHBastion, s.TargetHost, s.TargetPort))

	// ビルダー構築
	builder := envoy.NewTCPServiceBuilder(s.Host, s.ListenPort, s.SSHBastion, s.TargetHost, s.TargetPort)

	fmt.Printf(
		"gcp-ssh: %-30s -> %s (instance=%s, zone=%s) -> %s:%d via 127.0.0.1:%d\n",
		s.Host,
		s.SSHBastion,
		bastion.Instance,
		bastion.Zone,
		s.TargetHost,
		s.TargetPort,
		localPort,
	)

	// GCP SSH tunnelをgoroutineで起動
	go func(b *config.SSHBastion, local int, target string, targetPort int) {
		if err := gcp.StartGCPSSHTunnel(
			v.ctx,
			b,
			local,
			target,
			targetPort,
			v.logLevel,
		); err != nil {
			if v.ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "gcp-ssh tunnel error for %s: %v\n", b.Instance, err)
			}
		}
	}(bastion, localPort, s.TargetHost, s.TargetPort)

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

// DumpVisitor は DumpEnvoyConfig() 処理のための Visitor 実装
type DumpVisitor struct {
	ctx       context.Context
	clientset *kubernetes.Clientset
	mockCfg   *config.MockConfig
	idx       int

	// 結果
	serviceConfigs []envoy.ServiceConfig
}

// NewDumpVisitor は DumpVisitor を生成
func NewDumpVisitor(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	mockCfg *config.MockConfig,
) *DumpVisitor {
	return &DumpVisitor{
		ctx:            ctx,
		clientset:      clientset,
		mockCfg:        mockCfg,
		serviceConfigs: make([]envoy.ServiceConfig, 0),
	}
}

// VisitKubernetes は Kubernetes Service の処理（ダンプ用）
func (v *DumpVisitor) VisitKubernetes(s *config.KubernetesService) error {
	var remotePort int
	var err error

	// モック設定がある場合はモックから取得
	if v.mockCfg != nil {
		remotePort, err = findMockPort(v.mockCfg, s.Namespace, s.Service, s.PortName)
		if err != nil {
			return err
		}
	} else {
		// モック設定がない場合は通常通りclient-goで解決
		remotePort, err = k8s.ResolveServicePort(
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
	}

	// ダミーのローカルポート
	dummyLocalPort := 10000 + v.idx
	clusterName := sanitize(fmt.Sprintf("%s_%s_%d", s.Namespace, s.Service, remotePort))

	builder := envoy.NewKubernetesServiceBuilder(
		s.Host, s.Protocol, s.Namespace, s.Service, s.PortName, s.Port, s.OverwriteListenPorts,
	)

	v.serviceConfigs = append(v.serviceConfigs, envoy.ServiceConfig{
		Builder:     builder,
		ClusterName: clusterName,
		LocalPort:   dummyLocalPort,
	})

	return nil
}

// VisitTCP は TCP Service の処理（ダンプ用）
func (v *DumpVisitor) VisitTCP(s *config.TCPService) error {
	// ダミーのローカルポート
	dummyLocalPort := 10000 + v.idx
	clusterName := sanitize(fmt.Sprintf("tcp_%s_%s_%d", s.SSHBastion, s.TargetHost, s.TargetPort))

	builder := envoy.NewTCPServiceBuilder(s.Host, s.ListenPort, s.SSHBastion, s.TargetHost, s.TargetPort)

	v.serviceConfigs = append(v.serviceConfigs, envoy.ServiceConfig{
		Builder:     builder,
		ClusterName: clusterName,
		LocalPort:   dummyLocalPort,
	})

	return nil
}

// SetIndex はダンプ用のインデックスを設定
func (v *DumpVisitor) SetIndex(idx int) {
	v.idx = idx
}

// GetServiceConfigs は収集した ServiceConfig を返す
func (v *DumpVisitor) GetServiceConfigs() []envoy.ServiceConfig {
	return v.serviceConfigs
}
