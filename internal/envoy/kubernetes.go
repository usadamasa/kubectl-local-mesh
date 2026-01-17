package envoy

import (
	"fmt"

	"github.com/usadamasa/kubectl-localmesh/internal/port"
)

// KubernetesServiceBuilder はKubernetes Service用のEnvoy設定ビルダー
type KubernetesServiceBuilder struct {
	Host                 string
	Protocol             string                        // http|http2|grpc
	OverwriteListenPorts []port.IndividualListenerPort // 個別リスナーポート（省略時はHTTPリスナーに統合）
	// メタデータ（ログ・診断用、Envoy設定生成には使用しない）
	Namespace   string
	ServiceName string
	PortName    string
	Port        port.ServicePort
}

// NewKubernetesServiceBuilder はKubernetesServiceBuilderを生成
func NewKubernetesServiceBuilder(host, protocol, namespace, serviceName, portName string, p port.ServicePort, listenPorts []port.IndividualListenerPort) *KubernetesServiceBuilder {
	if protocol == "" {
		protocol = "http" // デフォルトHTTP/1.1
	}
	return &KubernetesServiceBuilder{
		Host:                 host,
		Protocol:             protocol,
		OverwriteListenPorts: listenPorts,
		Namespace:            namespace,
		ServiceName:          serviceName,
		PortName:             portName,
		Port:                 p,
	}
}

// Build はサービスの設定コンポーネントを生成
// OverwriteListenPortsが指定されている場合はIndividualListenerComponentsを返す
// 指定されていない場合はHTTPComponentsを返す
func (b *KubernetesServiceBuilder) Build(clusterName string, localPort int) any {
	// クラスタ設定
	cluster := b.buildCluster(clusterName, localPort)

	// OverwriteListenPortsがある場合は個別リスナーを生成
	if len(b.OverwriteListenPorts) > 0 {
		listeners := make([]map[string]any, len(b.OverwriteListenPorts))
		for i, listenPort := range b.OverwriteListenPorts {
			listeners[i] = b.buildIndividualListener(clusterName, listenPort, i)
		}
		return IndividualListenerComponents{
			Cluster:   cluster,
			Listeners: listeners,
		}
	}

	// HTTPルート設定（従来動作）
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

// buildCluster はクラスタ設定を生成
func (b *KubernetesServiceBuilder) buildCluster(clusterName string, localPort int) map[string]any {
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

	// protocolに応じたHTTP設定を追加
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

	return cluster
}

// buildIndividualListener は個別リスナーを生成
func (b *KubernetesServiceBuilder) buildIndividualListener(clusterName string, listenPort port.IndividualListenerPort, index int) map[string]any {
	listenerName := fmt.Sprintf("listener_%s_%d", clusterName, listenPort)

	// HTTP connection manager設定
	httpConnManager := map[string]any{
		"@type":       "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
		"stat_prefix": fmt.Sprintf("ingress_%s_%d", clusterName, listenPort),
		"codec_type":  "AUTO",
		"route_config": map[string]any{
			"name": fmt.Sprintf("route_%s_%d", clusterName, listenPort),
			"virtual_hosts": []any{
				map[string]any{
					"name":    clusterName,
					"domains": []any{b.Host, "*"},
					"routes": []any{
						map[string]any{
							"match": map[string]any{"prefix": "/"},
							"route": map[string]any{
								"cluster": clusterName,
								"timeout": "0s",
							},
						},
					},
				},
			},
		},
		"http_filters": []any{
			map[string]any{
				"name": "envoy.filters.http.router",
				"typed_config": map[string]any{
					"@type": "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router",
				},
			},
		},
	}

	// HTTP/2対応（gRPC/http2の場合）
	if b.Protocol == "grpc" || b.Protocol == "http2" {
		httpConnManager["http2_protocol_options"] = map[string]any{}
	}

	return map[string]any{
		"name": listenerName,
		"address": map[string]any{
			"socket_address": map[string]any{
				"address":    "0.0.0.0",
				"port_value": int(listenPort),
			},
		},
		"filter_chains": []any{
			map[string]any{
				"filters": []any{
					map[string]any{
						"name":         "envoy.filters.network.http_connection_manager",
						"typed_config": httpConnManager,
					},
				},
			},
		},
	}
}

// GetHost はホスト名を取得
func (b *KubernetesServiceBuilder) GetHost() string {
	return b.Host
}
