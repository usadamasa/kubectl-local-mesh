package envoy

// TCPServiceBuilder はTCP Service用のEnvoy設定ビルダー
type TCPServiceBuilder struct {
	Host       string
	ListenPort int // TCPリスナーの独立ポート
	// メタデータ（ログ・診断用、Envoy設定生成には使用しない）
	SSHBastion string
	TargetHost string
	TargetPort int
}

// NewTCPServiceBuilder はTCPServiceBuilderを生成
func NewTCPServiceBuilder(host string, listenPort int, sshBastion, targetHost string, targetPort int) *TCPServiceBuilder {
	return &TCPServiceBuilder{
		Host:       host,
		ListenPort: listenPort,
		SSHBastion: sshBastion,
		TargetHost: targetHost,
		TargetPort: targetPort,
	}
}

// Build はTCPサービスの設定コンポーネントを生成
func (b *TCPServiceBuilder) Build(clusterName string, localPort int) TCPComponents {
	// クラスタ設定（TCPクラスタはHTTPプロトコルオプション不要）
	cluster := map[string]any{
		"name":            clusterName,
		"type":            "STATIC",
		"connect_timeout": "1s",
		"load_assignment": map[string]any{
			"cluster_name": clusterName,
			"endpoints": []any{
				map[string]any{
					"lb_endpoints": []any{
						map[string]any{
							"endpoint": map[string]any{
								"address": map[string]any{
									"socket_address": map[string]any{
										"address":    "127.0.0.1",
										"port_value": localPort,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// TCPリスナー設定
	tcpListener := map[string]any{
		"name": "listener_tcp_" + clusterName,
		"address": map[string]any{
			"socket_address": map[string]any{
				"address":    "0.0.0.0",
				"port_value": b.ListenPort,
			},
		},
		"filter_chains": []any{
			map[string]any{
				"filters": []any{
					map[string]any{
						"name": "envoy.filters.network.tcp_proxy",
						"typed_config": map[string]any{
							"@type":       "type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy",
							"stat_prefix": "tcp_" + clusterName,
							"cluster":     clusterName,
						},
					},
				},
			},
		},
	}

	return TCPComponents{
		Cluster:  cluster,
		Listener: tcpListener,
	}
}

// GetHost はホスト名を取得
func (b *TCPServiceBuilder) GetHost() string {
	return b.Host
}
