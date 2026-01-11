package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"

	"github.com/usadamasa/kubectl-localmesh/internal/config"
	"github.com/usadamasa/kubectl-localmesh/internal/envoy"
	"github.com/usadamasa/kubectl-localmesh/internal/gcp"
	"github.com/usadamasa/kubectl-localmesh/internal/hosts"
	"github.com/usadamasa/kubectl-localmesh/internal/k8s"
	"github.com/usadamasa/kubectl-localmesh/internal/pf"
)

func Run(ctx context.Context, cfg *config.Config, logLevel string, updateHosts bool) error {
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
		fmt.Println("/etc/hosts updated successfully")

		// 終了時にクリーンアップ
		defer func() {
			if err := hosts.RemoveEntries(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to clean up /etc/hosts: %v\n", err)
			} else {
				fmt.Println("/etc/hosts cleaned up")
			}
		}()
	}

	tmpDir, err := os.MkdirTemp("", "kubectl-localmesh-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	var serviceConfigs []envoy.ServiceConfig

	for _, svcDef := range cfg.Services {
		svc := svcDef.Get()
		var localPort int
		var clusterName string
		var builder interface{}

		// type switchで型判別
		switch s := svc.(type) {
		case *config.TCPService:
			// TCP + SSH Bastion経由の接続
			bastion, ok := cfg.SSHBastions[s.SSHBastion]
			if !ok {
				return fmt.Errorf("ssh_bastion '%s' not found for service '%s'", s.SSHBastion, s.Host)
			}

			lp, err := pf.FreeLocalPort()
			if err != nil {
				return err
			}
			localPort = lp
			clusterName = sanitize(fmt.Sprintf("tcp_%s_%s_%d", s.SSHBastion, s.TargetHost, s.TargetPort))

			// TCPビルダーを構築（設定生成ロジックはビルダー内に隠蔽）
			// ListenPortを使用（省略時はTargetPortと同じ）
			builder = envoy.NewTCPServiceBuilder(s.Host, s.ListenPort, s.SSHBastion, s.TargetHost, s.TargetPort)

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

			// GCP SSH tunnelをgoroutineで起動（自動再接続）
			go func(b *config.SSHBastion, local int, target string, targetPort int) {
				if err := gcp.StartGCPSSHTunnel(
					ctx,
					b,
					local,
					target,
					targetPort,
					logLevel,
				); err != nil {
					// contextキャンセル以外のエラーをログ出力
					if ctx.Err() == nil {
						fmt.Fprintf(os.Stderr, "gcp-ssh tunnel error for %s: %v\n", b.Instance, err)
					}
				}
			}(bastion, localPort, s.TargetHost, s.TargetPort)

		case *config.KubernetesService:
			// Kubernetes Service経由の接続
			remotePort, err := k8s.ResolveServicePort(
				ctx,
				clientset,
				s.Namespace,
				s.Service,
				s.PortName,
				s.Port,
			)
			if err != nil {
				return err
			}

			lp, err := pf.FreeLocalPort()
			if err != nil {
				return err
			}
			localPort = lp
			clusterName = sanitize(fmt.Sprintf("%s_%s_%d", s.Namespace, s.Service, remotePort))

			// Kubernetesビルダーを構築（protocol分岐はビルダー内に隠蔽）
			// メタデータも渡す（ログ・診断用）
			builder = envoy.NewKubernetesServiceBuilder(
				s.Host, s.Protocol, s.Namespace, s.Service, s.PortName, s.Port,
			)

			fmt.Printf(
				"pf: %-30s -> %s/%s:%d via 127.0.0.1:%d\n",
				s.Host,
				s.Namespace,
				s.Service,
				remotePort,
				localPort,
			)

			// port-forwardをgoroutineで起動（自動再接続）
			go func(ns, svc string, local, remote int) {
				if err := k8s.StartPortForwardLoop(
					ctx,
					restConfig,
					clientset,
					ns,
					svc,
					local,
					remote,
					logLevel,
				); err != nil {
					// contextキャンセル以外のエラーをログ出力
					if ctx.Err() == nil {
						fmt.Fprintf(os.Stderr, "port-forward error for %s/%s: %v\n", ns, svc, err)
					}
				}
			}(s.Namespace, s.Service, localPort, remotePort)

		default:
			return fmt.Errorf("unknown service type: %T", s)
		}

		serviceConfigs = append(serviceConfigs, envoy.ServiceConfig{
			Builder:     builder,
			ClusterName: clusterName,
			LocalPort:   localPort,
		})
	}

	// ビルダーベースのBuildConfigを呼び出し
	envoyCfg := envoy.BuildConfig(cfg.ListenerPort, serviceConfigs)
	envoyPath := filepath.Join(tmpDir, "envoy.yaml")

	b, err := yaml.Marshal(envoyCfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(envoyPath, b, 0644); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("envoy config: %s\n", envoyPath)
	fmt.Printf("listen: 0.0.0.0:%d\n\n", cfg.ListenerPort)

	envoyCmd := exec.CommandContext(
		ctx,
		"envoy",
		"-c", envoyPath,
		"-l", logLevel,
	)
	envoyCmd.Stdout = os.Stdout
	envoyCmd.Stderr = os.Stderr

	// Envoy実行（contextキャンセル時に自動終了）
	// port-forwardのgoroutineもcontextキャンセル時に自動終了する
	return envoyCmd.Run()
}

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

	var serviceConfigs []envoy.ServiceConfig

	for i, svcDef := range cfg.Services {
		svc := svcDef.Get()

		// type switchで型判別
		switch s := svc.(type) {
		case *config.KubernetesService:
			var remotePort int

			// モック設定が指定されている場合はモックから取得
			if mockCfg != nil {
				remotePort, err = findMockPort(mockCfg, s.Namespace, s.Service, s.PortName)
				if err != nil {
					return err
				}
			} else {
				// モック設定がない場合は通常通りclient-goで解決
				remotePort, err = k8s.ResolveServicePort(
					ctx,
					clientset,
					s.Namespace,
					s.Service,
					s.PortName,
					s.Port,
				)
				if err != nil {
					return err
				}
			}

			// ダミーのローカルポートを割り当て（実際にはport-forwardしない）
			dummyLocalPort := 10000 + i
			clusterName := sanitize(fmt.Sprintf("%s_%s_%d", s.Namespace, s.Service, remotePort))

			builder := envoy.NewKubernetesServiceBuilder(
				s.Host, s.Protocol, s.Namespace, s.Service, s.PortName, s.Port,
			)

			serviceConfigs = append(serviceConfigs, envoy.ServiceConfig{
				Builder:     builder,
				ClusterName: clusterName,
				LocalPort:   dummyLocalPort,
			})

		case *config.TCPService:
			// TCPサービスの場合（dump-envoy-configでは簡易処理）
			dummyLocalPort := 10000 + i
			clusterName := sanitize(fmt.Sprintf("tcp_%s_%s_%d", s.SSHBastion, s.TargetHost, s.TargetPort))

			builder := envoy.NewTCPServiceBuilder(s.Host, s.ListenPort, s.SSHBastion, s.TargetHost, s.TargetPort)

			serviceConfigs = append(serviceConfigs, envoy.ServiceConfig{
				Builder:     builder,
				ClusterName: clusterName,
				LocalPort:   dummyLocalPort,
			})

		default:
			return fmt.Errorf("unknown service type: %T", s)
		}
	}

	envoyCfg := envoy.BuildConfig(cfg.ListenerPort, serviceConfigs)

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
