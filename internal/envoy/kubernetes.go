package envoy

// KubernetesServiceBuilder はKubernetes Service用のEnvoy設定ビルダー
type KubernetesServiceBuilder struct {
	Host     string
	Protocol string // http|http2|grpc
	// メタデータ（ログ・診断用、Envoy設定生成には使用しない）
	Namespace   string
	ServiceName string
	PortName    string
	Port        int
}

// NewKubernetesServiceBuilder はKubernetesServiceBuilderを生成
func NewKubernetesServiceBuilder(host, protocol, namespace, serviceName, portName string, port int) *KubernetesServiceBuilder {
	if protocol == "" {
		protocol = "http" // デフォルトHTTP/1.1
	}
	return &KubernetesServiceBuilder{
		Host:        host,
		Protocol:    protocol,
		Namespace:   namespace,
		ServiceName: serviceName,
		PortName:    portName,
		Port:        port,
	}
}

// Build はHTTPサービスの設定コンポーネントを生成
func (b *KubernetesServiceBuilder) Build(clusterName string, localPort int) HTTPComponents {
	// クラスタ設定
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

	// protocolに応じたHTTP設定を追加（制御の反転）
	var httpConfig map[string]any
	if b.Protocol == "grpc" || b.Protocol == "http2" {
		httpConfig = map[string]any{
			"http2_protocol_options": map[string]any{},
		}
	} else {
		httpConfig = map[string]any{
			"http_protocol_options": map[string]any{},
		}
	}

	cluster["typed_extension_protocol_options"] = map[string]any{
		"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": map[string]any{
			"@type":                "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions",
			"explicit_http_config": httpConfig,
		},
	}

	// HTTPルート設定
	httpRoute := map[string]any{
		"name":    clusterName,
		"domains": []any{b.Host},
		"routes": []any{
			map[string]any{
				"match": map[string]any{"prefix": "/"},
				"route": map[string]any{
					"cluster": clusterName,
					"timeout": "0s",
				},
			},
		},
	}

	return HTTPComponents{
		Cluster: cluster,
		Route:   httpRoute,
	}
}

// GetHost はホスト名を取得
func (b *KubernetesServiceBuilder) GetHost() string {
	return b.Host
}
