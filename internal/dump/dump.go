package dump

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/usadamasa/kubectl-localmesh/internal/config"
	"github.com/usadamasa/kubectl-localmesh/internal/envoy"
	"github.com/usadamasa/kubectl-localmesh/internal/snapshot"
)

// DumpOptions はダンプコマンドのオプション
type DumpOptions struct {
	MockConfigPath string
	OutputMapping  bool
}

func DumpEnvoyConfig(ctx context.Context, cfg *config.Config, mockConfigPath string) error {
	return DumpEnvoyConfigWithOptions(ctx, cfg, DumpOptions{MockConfigPath: mockConfigPath})
}

func DumpEnvoyConfigWithOptions(ctx context.Context, cfg *config.Config, opts DumpOptions) error {
	var mockCfg *config.MockConfig
	var err error

	// モック設定の読み込み
	if opts.MockConfigPath != "" {
		mockCfg, err = config.LoadMockConfig(opts.MockConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load mock config: %w", err)
		}
	}

	// Visitor の生成（Kubernetes clientはサービスごとにlazy初期化）
	visitor := NewDumpVisitor(ctx, cfg.Cluster, mockCfg)

	// Visitorパターンで各サービスを処理
	for i, svcDef := range cfg.Services {
		visitor.SetIndex(i)
		svc := svcDef.Get()
		if err := svc.Accept(visitor); err != nil {
			return err
		}
	}

	serviceConfigs := visitor.GetServiceConfigs()

	// マッピング出力モード
	if opts.OutputMapping {
		mappings := snapshot.BuildMappings(serviceConfigs)
		b, err := yaml.Marshal(mappings)
		if err != nil {
			return err
		}
		fmt.Print(string(b)) //nolint:forbidigo // CLIダンプ出力として意図的に使用
		return nil
	}

	// Envoy設定生成（デフォルト）
	envoyCfg := envoy.BuildConfig(cfg.ListenerPort, serviceConfigs)

	b, err := yaml.Marshal(envoyCfg)
	if err != nil {
		return err
	}

	fmt.Print(string(b)) //nolint:forbidigo // CLIダンプ出力として意図的に使用
	return nil
}
