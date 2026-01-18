package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/usadamasa/kubectl-localmesh/internal/config"
	"github.com/usadamasa/kubectl-localmesh/internal/envoy"
	"github.com/usadamasa/kubectl-localmesh/internal/hosts"
	"github.com/usadamasa/kubectl-localmesh/internal/k8s"
	"github.com/usadamasa/kubectl-localmesh/internal/log"
	"gopkg.in/yaml.v3"
)

func Run(ctx context.Context, cfg *config.Config, logLevel string, updateHosts bool) error {
	// Logger初期化
	logger := log.New(logLevel)
	// Kubernetes client初期化
	clientset, restConfig, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// /etc/hosts更新が必要な場合
	if updateHosts {
		// 権限チェック
		if !hosts.HasPermission() {
			return fmt.Errorf("need sudo: try 'sudo kubectl-localmesh ...'")
		}

		// ホスト名リストを収集
		var hostnames []string
		for _, svcDef := range cfg.Services {
			svc := svcDef.Get()
			hostnames = append(hostnames, svc.GetHost())
		}

		// /etc/hostsに追加
		if err := hosts.AddEntries(hostnames); err != nil {
			return fmt.Errorf("failed to update /etc/hosts: %w", err)
		}
		logger.Debug("/etc/hosts updated successfully")

		// 終了時にクリーンアップ
		defer func() {
			if err := hosts.RemoveEntries(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to clean up /etc/hosts: %v\n", err)
			} else {
				logger.Debug("/etc/hosts cleaned up")
			}
		}()
	}

	tmpDir, err := os.MkdirTemp("", "kubectl-localmesh-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Visitor の生成
	visitor := NewRunVisitor(ctx, cfg, clientset, restConfig, logger)

	// Visitorパターンで各サービスを処理
	for _, svcDef := range cfg.Services {
		svc := svcDef.Get()
		if err := svc.Accept(visitor); err != nil {
			return err
		}
	}

	// Envoy設定生成
	envoyCfg := envoy.BuildConfig(cfg.ListenerPort, visitor.GetServiceConfigs())
	envoyPath := filepath.Join(tmpDir, "envoy.yaml")

	b, err := yaml.Marshal(envoyCfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(envoyPath, b, 0644); err != nil {
		return err
	}

	logger.Debugf("envoy config: %s", envoyPath)
	logger.Debugf("listen: 0.0.0.0:%d", cfg.ListenerPort)

	// サマリー出力
	summary := log.GenerateSummary(visitor.GetServiceSummaries(), cfg.ListenerPort)
	logger.Info(summary)

	envoyCmd := exec.CommandContext(
		ctx,
		"envoy",
		"-c", envoyPath,
		"-l", logger.EnvoyLevel(),
	)
	envoyCmd.Stdout = os.Stdout
	envoyCmd.Stderr = os.Stderr

	// Envoy実行（contextキャンセル時に自動終了）
	// port-forwardのgoroutineもcontextキャンセル時に自動終了する
	return envoyCmd.Run()
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
