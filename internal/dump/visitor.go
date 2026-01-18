package dump

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/usadamasa/kubectl-localmesh/internal/config"
	"github.com/usadamasa/kubectl-localmesh/internal/envoy"
	"github.com/usadamasa/kubectl-localmesh/internal/k8s"
	"github.com/usadamasa/kubectl-localmesh/internal/port"
)

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
	var remotePort port.ServicePort
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
	dummyLocalPort := port.LocalPort(10000 + v.idx)
	clusterName := sanitize(fmt.Sprintf("%s_%s_%d", s.Namespace, s.Service, remotePort))

	builder := envoy.NewKubernetesServiceBuilder(
		s.Host, s.Protocol, s.Namespace, s.Service, s.PortName, s.Port, s.OverwriteListenPorts,
	)

	v.serviceConfigs = append(v.serviceConfigs, envoy.ServiceConfig{
		Builder:            builder,
		ClusterName:        clusterName,
		LocalPort:          dummyLocalPort,
		ResolvedRemotePort: remotePort,
	})

	return nil
}

// VisitTCP は TCP Service の処理（ダンプ用）
func (v *DumpVisitor) VisitTCP(s *config.TCPService) error {
	// ダミーのローカルポート
	dummyLocalPort := port.LocalPort(10000 + v.idx)
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

func findMockPort(mockCfg *config.MockConfig, namespace, service, portName string) (port.ServicePort, error) {
	for _, m := range mockCfg.Mocks {
		if m.Namespace == namespace && m.Service == service && m.PortName == portName {
			return m.ResolvedPort, nil
		}
	}
	return 0, fmt.Errorf("mock config not found for %s/%s (port_name=%s)", namespace, service, portName)
}

func sanitize(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'a' && r <= 'z' ||
			r >= 'A' && r <= 'Z' ||
			r >= '0' && r <= '9' ||
			r == '_' {
			out = append(out, r)
		} else {
			out = append(out, '_')
		}
	}
	return string(out)
}
