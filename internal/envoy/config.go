package envoy

import "github.com/usadamasa/kubectl-localmesh/internal/port"

// ServiceConfig はビルダーとメタデータを保持
type ServiceConfig struct {
	Builder            interface{} // *KubernetesServiceBuilder または *TCPServiceBuilder
	ClusterName        string
	LocalPort          int
	ResolvedRemotePort int // Kubernetesサービスの解決済みリモートポート（マッピング出力用）
}

// BuildConfig は ServiceConfig のリストから Envoy 設定を生成
func BuildConfig(listenerPort port.ListenerPort, configs []ServiceConfig) map[string]any {
	var clusters []any
	var httpRoutes []any
	var tcpListeners []any

	var individualListeners []any

	for _, cfg := range configs {
		// type switchで各ビルダーを処理
		switch builder := cfg.Builder.(type) {
		case *KubernetesServiceBuilder:
			result := builder.Build(cfg.ClusterName, cfg.LocalPort, int(listenerPort))
			// 戻り値の型によって処理を分岐
			switch components := result.(type) {
			case HTTPComponents:
				clusters = append(clusters, components.Cluster)
				httpRoutes = append(httpRoutes, components.Route)
			case IndividualListenerComponents:
				clusters = append(clusters, components.Cluster)
				for _, listener := range components.Listeners {
					individualListeners = append(individualListeners, listener)
				}
			}

		case *TCPServiceBuilder:
			components := builder.Build(cfg.ClusterName, cfg.LocalPort)
			clusters = append(clusters, components.Cluster)
			tcpListeners = append(tcpListeners, components.Listener)
		}
	}

	var listeners []any

	// HTTPリスナー（HTTPルートがある場合のみ）
	if len(httpRoutes) > 0 {
		httpListener := map[string]any{
			"name": "listener_http",
			"address": map[string]any{
				"socket_address": map[string]any{
					"address":    "0.0.0.0",
					"port_value": int(listenerPort),
				},
			},
			"enable_reuse_port": map[string]any{"value": false},
			"filter_chains": []any{
				map[string]any{
					"filters": []any{
						map[string]any{
							"name": "envoy.filters.network.http_connection_manager",
							"typed_config": map[string]any{
								"@type":                  "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
								"stat_prefix":            "ingress_http",
								"codec_type":             "AUTO",
								"http2_protocol_options": map[string]any{},
								"route_config": map[string]any{
									"name":          "local_route",
									"virtual_hosts": httpRoutes,
								},
								"http_filters": []any{
									map[string]any{
										"name": "envoy.filters.http.router",
										"typed_config": map[string]any{
											"@type": "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		listeners = append(listeners, httpListener)
	}

	// 個別リスナーを追加（OverwriteListenPortsが指定されたサービス用）
	listeners = append(listeners, individualListeners...)

	// TCPリスナーを追加
	listeners = append(listeners, tcpListeners...)

	return map[string]any{
		"static_resources": map[string]any{
			"listeners": listeners,
			"clusters":  clusters,
		},
		"overload_manager": map[string]any{
			"refresh_interval": map[string]any{
				"seconds": 0,
				"nanos":   250000000,
			},
			"resource_monitors": []any{
				map[string]any{
					"name": "envoy.resource_monitors.global_downstream_max_connections",
					"typed_config": map[string]any{
						"@type":                             "type.googleapis.com/envoy.extensions.resource_monitors.downstream_connections.v3.DownstreamConnectionsConfig",
						"max_active_downstream_connections": 5000,
					},
				},
			},
		},
	}
}
