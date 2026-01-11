package envoy

type Route struct {
	Host        string
	LocalPort   int
	ClusterName string
	Type        string // "http" or "tcp"
	Protocol    string // "http" | "http2" | "grpc" (HTTPプロトコルバージョン選択用)
	ListenPort  int    // TCP用のリスンポート（Type="tcp"の場合のみ使用）
}

func BuildConfig(listenerPort int, routes []Route) map[string]any {
	var clusters []any
	var vhosts []any
	var listeners []any

	// HTTPルートとTCPルートを分離
	var httpRoutes []Route
	var tcpRoutes []Route

	for _, r := range routes {
		if r.Type == "tcp" {
			tcpRoutes = append(tcpRoutes, r)
		} else {
			// デフォルトはHTTP（既存の動作を維持）
			httpRoutes = append(httpRoutes, r)
		}
	}

	// すべてのルート用のクラスタを生成
	for _, r := range routes {
		cluster := map[string]any{
			"name":            r.ClusterName,
			"type":            "STATIC",
			"connect_timeout": "1s",
			"load_assignment": map[string]any{
				"cluster_name": r.ClusterName,
				"endpoints": []any{
					map[string]any{
						"lb_endpoints": []any{
							map[string]any{
								"endpoint": map[string]any{
									"address": map[string]any{
										"socket_address": map[string]any{
											"address":    "127.0.0.1",
											"port_value": r.LocalPort,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		// HTTP/gRPCの場合はプロトコルオプションを追加
		if r.Type != "tcp" {
			var httpConfig map[string]any

			if r.Protocol == "grpc" || r.Protocol == "http2" {
				// gRPC / HTTP/2: HTTP/2（h2c）
				httpConfig = map[string]any{
					"http2_protocol_options": map[string]any{},
				}
			} else {
				// HTTP/1.1（デフォルト、protocol: http または未指定）
				httpConfig = map[string]any{
					"http_protocol_options": map[string]any{},
				}
			}

			cluster["typed_extension_protocol_options"] = map[string]any{
				"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": map[string]any{
					"@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions",
					"explicit_http_config": httpConfig,
				},
			}
		}

		clusters = append(clusters, cluster)
	}

	// HTTPリスナーの生成（HTTPルートが存在する場合）
	if len(httpRoutes) > 0 {
		for _, r := range httpRoutes {
			vhosts = append(vhosts, map[string]any{
				"name":    r.ClusterName,
				"domains": []any{r.Host},
				"routes": []any{
					map[string]any{
						"match": map[string]any{"prefix": "/"},
						"route": map[string]any{
							"cluster": r.ClusterName,
							"timeout": "0s",
						},
					},
				},
			})
		}

		httpListener := map[string]any{
			"name": "listener_http",
			"address": map[string]any{
				"socket_address": map[string]any{
					"address":    "0.0.0.0",
					"port_value": listenerPort,
				},
			},
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
									"virtual_hosts": vhosts,
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

	// TCPリスナーの生成（TCPルートごとに独立したリスナー）
	for _, r := range tcpRoutes {
		tcpListener := map[string]any{
			"name": "listener_tcp_" + r.ClusterName,
			"address": map[string]any{
				"socket_address": map[string]any{
					"address":    "0.0.0.0",
					"port_value": r.ListenPort,
				},
			},
			"filter_chains": []any{
				map[string]any{
					"filters": []any{
						map[string]any{
							"name": "envoy.filters.network.tcp_proxy",
							"typed_config": map[string]any{
								"@type":       "type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy",
								"stat_prefix": "tcp_" + r.ClusterName,
								"cluster":     r.ClusterName,
							},
						},
					},
				},
			},
		}
		listeners = append(listeners, tcpListener)
	}

	return map[string]any{
		"static_resources": map[string]any{
			"listeners": listeners,
			"clusters":  clusters,
		},
	}
}
