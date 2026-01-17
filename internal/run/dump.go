package run

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"

	"github.com/usadamasa/kubectl-localmesh/internal/config"
	"github.com/usadamasa/kubectl-localmesh/internal/envoy"
	"github.com/usadamasa/kubectl-localmesh/internal/k8s"
)

func DumpEnvoyConfig(ctx context.Context, cfg *config.Config, mockConfigPath string) error {
	var mockCfg *config.MockConfig
	var err error

	// モック設定の読み込み
	if mockConfigPath != "" {
		mockCfg, err = config.LoadMockConfig(mockConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load mock config: %w", err)
		}
	}

	// モックモードでない場合はKubernetes clientを初期化
	var clientset *kubernetes.Clientset
	if mockCfg == nil {
		var k8sErr error
		clientset, _, k8sErr = k8s.NewClient()
		if k8sErr != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", k8sErr)
		}
	}

	// Visitor の生成
	visitor := NewDumpVisitor(ctx, clientset, mockCfg)

	// Visitorパターンで各サービスを処理
	for i, svcDef := range cfg.Services {
		visitor.SetIndex(i)
		svc := svcDef.Get()
		if err := svc.Accept(visitor); err != nil {
			return err
		}
	}

	// Envoy設定生成
	envoyCfg := envoy.BuildConfig(cfg.ListenerPort, visitor.GetServiceConfigs())

	b, err := yaml.Marshal(envoyCfg)
	if err != nil {
		return err
	}

	fmt.Print(string(b))
	return nil
}

func findMockPort(mockCfg *config.MockConfig, namespace, service, portName string) (int, error) {
	for _, m := range mockCfg.Mocks {
		if m.Namespace == namespace && m.Service == service && m.PortName == portName {
			return int(m.ResolvedPort), nil
		}
	}
	return 0, fmt.Errorf("mock config not found for %s/%s (port_name=%s)", namespace, service, portName)
}
