package envoy

import (
	"testing"
)

func TestBuildConfig_HTTPOnly(t *testing.T) {
	// HTTP/gRPCのみの設定（既存の動作確認）
	routes := []Route{
		{
			Host:        "api.localhost",
			LocalPort:   10001,
			ClusterName: "api_cluster",
			Type:        "http",
			Protocol:    "http",
		},
	}

	cfg := BuildConfig(80, routes)

	// static_resourcesの存在確認
	staticRes, ok := cfg["static_resources"].(map[string]any)
	if !ok {
		t.Fatal("static_resources not found")
	}

	// listenersの確認
	listeners, ok := staticRes["listeners"].([]any)
	if !ok {
		t.Fatal("listeners not found")
	}

	// HTTPリスナーが1つ存在することを確認
	if len(listeners) != 1 {
		t.Errorf("expected 1 listener (HTTP), got %d", len(listeners))
	}

	// clustersの確認
	clusters, ok := staticRes["clusters"].([]any)
	if !ok {
		t.Fatal("clusters not found")
	}
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(clusters))
	}
}

func TestBuildConfig_TCPOnly(t *testing.T) {
	// TCPのみの設定
	routes := []Route{
		{
			Host:        "db.localhost",
			LocalPort:   10002,
			ClusterName: "db_cluster",
			Type:        "tcp",
			ListenPort:  5432, // DBポート
		},
	}

	cfg := BuildConfig(80, routes)

	staticRes, ok := cfg["static_resources"].(map[string]any)
	if !ok {
		t.Fatal("static_resources not found")
	}

	listeners, ok := staticRes["listeners"].([]any)
	if !ok {
		t.Fatal("listeners not found")
	}

	// TCPリスナーが1つ存在することを確認
	if len(listeners) != 1 {
		t.Errorf("expected 1 listener (TCP), got %d", len(listeners))
	}

	// TCPリスナーの詳細確認
	listener := listeners[0].(map[string]any)
	address := listener["address"].(map[string]any)
	socketAddr := address["socket_address"].(map[string]any)

	if socketAddr["port_value"] != 5432 {
		t.Errorf("expected TCP listener port 5432, got %v", socketAddr["port_value"])
	}

	// clustersの確認
	clusters, ok := staticRes["clusters"].([]any)
	if !ok {
		t.Fatal("clusters not found")
	}
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(clusters))
	}
}

func TestBuildConfig_MixedHTTPAndTCP(t *testing.T) {
	// HTTP/gRPCとTCPの混在
	routes := []Route{
		{
			Host:        "api.localhost",
			LocalPort:   10001,
			ClusterName: "api_cluster",
			Type:        "http",
			Protocol:    "http",
		},
		{
			Host:        "db.localhost",
			LocalPort:   10002,
			ClusterName: "db_cluster",
			Type:        "tcp",
			ListenPort:  5432,
		},
		{
			Host:        "cache.localhost",
			LocalPort:   10003,
			ClusterName: "cache_cluster",
			Type:        "tcp",
			ListenPort:  6379,
		},
	}

	cfg := BuildConfig(80, routes)

	staticRes, ok := cfg["static_resources"].(map[string]any)
	if !ok {
		t.Fatal("static_resources not found")
	}

	listeners, ok := staticRes["listeners"].([]any)
	if !ok {
		t.Fatal("listeners not found")
	}

	// HTTPリスナー1つ + TCPリスナー2つ = 計3つ
	if len(listeners) != 3 {
		t.Errorf("expected 3 listeners (1 HTTP + 2 TCP), got %d", len(listeners))
	}

	// clustersの確認（3つ）
	clusters, ok := staticRes["clusters"].([]any)
	if !ok {
		t.Fatal("clusters not found")
	}
	if len(clusters) != 3 {
		t.Errorf("expected 3 clusters, got %d", len(clusters))
	}
}

func TestBuildConfig_MultipleTCPSamePort(t *testing.T) {
	// 同じポート番号を持つ複数のTCPサービス（エラーケース）
	routes := []Route{
		{
			Host:        "db1.localhost",
			LocalPort:   10002,
			ClusterName: "db1_cluster",
			Type:        "tcp",
			ListenPort:  5432,
		},
		{
			Host:        "db2.localhost",
			LocalPort:   10003,
			ClusterName: "db2_cluster",
			Type:        "tcp",
			ListenPort:  5432, // 重複
		},
	}

	// 現時点ではエラーチェックなし（将来的に追加する可能性）
	cfg := BuildConfig(80, routes)

	staticRes, ok := cfg["static_resources"].(map[string]any)
	if !ok {
		t.Fatal("static_resources not found")
	}

	listeners, ok := staticRes["listeners"].([]any)
	if !ok {
		t.Fatal("listeners not found")
	}

	// 2つのリスナーが作成される（ポート重複は警告すべきだが、現時点では許容）
	if len(listeners) != 2 {
		t.Errorf("expected 2 listeners, got %d", len(listeners))
	}
}

func TestBuildConfig_HTTPProtocol(t *testing.T) {
	// protocol: http → HTTP/1.1設定確認
	routes := []Route{
		{
			Host:        "api.localhost",
			LocalPort:   10001,
			ClusterName: "api_cluster",
			Type:        "http",
			Protocol:    "http",
		},
	}

	cfg := BuildConfig(80, routes)

	staticRes, ok := cfg["static_resources"].(map[string]any)
	if !ok {
		t.Fatal("static_resources not found")
	}

	clusters, ok := staticRes["clusters"].([]any)
	if !ok {
		t.Fatal("clusters not found")
	}

	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}

	cluster := clusters[0].(map[string]any)
	protocolOpts, ok := cluster["typed_extension_protocol_options"].(map[string]any)
	if !ok {
		t.Fatal("typed_extension_protocol_options not found")
	}

	httpOpts, ok := protocolOpts["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"].(map[string]any)
	if !ok {
		t.Fatal("HttpProtocolOptions not found")
	}

	explicitConfig, ok := httpOpts["explicit_http_config"].(map[string]any)
	if !ok {
		t.Fatal("explicit_http_config not found")
	}

	// HTTP/1.1の設定を確認
	if _, ok := explicitConfig["http_protocol_options"]; !ok {
		t.Error("expected http1_protocol_options for protocol: http")
	}

	// HTTP/2の設定がないことを確認
	if _, ok := explicitConfig["http2_protocol_options"]; ok {
		t.Error("unexpected http2_protocol_options for protocol: http")
	}
}

func TestBuildConfig_HTTP2Protocol(t *testing.T) {
	// protocol: http2 → HTTP/2設定確認
	routes := []Route{
		{
			Host:        "api.localhost",
			LocalPort:   10001,
			ClusterName: "api_cluster",
			Type:        "http",
			Protocol:    "http2",
		},
	}

	cfg := BuildConfig(80, routes)

	staticRes, ok := cfg["static_resources"].(map[string]any)
	if !ok {
		t.Fatal("static_resources not found")
	}

	clusters, ok := staticRes["clusters"].([]any)
	if !ok {
		t.Fatal("clusters not found")
	}

	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}

	cluster := clusters[0].(map[string]any)
	protocolOpts, ok := cluster["typed_extension_protocol_options"].(map[string]any)
	if !ok {
		t.Fatal("typed_extension_protocol_options not found")
	}

	httpOpts, ok := protocolOpts["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"].(map[string]any)
	if !ok {
		t.Fatal("HttpProtocolOptions not found")
	}

	explicitConfig, ok := httpOpts["explicit_http_config"].(map[string]any)
	if !ok {
		t.Fatal("explicit_http_config not found")
	}

	// HTTP/2の設定を確認
	if _, ok := explicitConfig["http2_protocol_options"]; !ok {
		t.Error("expected http2_protocol_options for protocol: http2")
	}

	// HTTP/1.1の設定がないことを確認
	if _, ok := explicitConfig["http_protocol_options"]; ok {
		t.Error("unexpected http1_protocol_options for protocol: http2")
	}
}

func TestBuildConfig_gRPCProtocol(t *testing.T) {
	// protocol: grpc → HTTP/2設定確認
	routes := []Route{
		{
			Host:        "grpc.localhost",
			LocalPort:   10001,
			ClusterName: "grpc_cluster",
			Type:        "http",
			Protocol:    "grpc",
		},
	}

	cfg := BuildConfig(80, routes)

	staticRes, ok := cfg["static_resources"].(map[string]any)
	if !ok {
		t.Fatal("static_resources not found")
	}

	clusters, ok := staticRes["clusters"].([]any)
	if !ok {
		t.Fatal("clusters not found")
	}

	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}

	cluster := clusters[0].(map[string]any)
	protocolOpts, ok := cluster["typed_extension_protocol_options"].(map[string]any)
	if !ok {
		t.Fatal("typed_extension_protocol_options not found")
	}

	httpOpts, ok := protocolOpts["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"].(map[string]any)
	if !ok {
		t.Fatal("HttpProtocolOptions not found")
	}

	explicitConfig, ok := httpOpts["explicit_http_config"].(map[string]any)
	if !ok {
		t.Fatal("explicit_http_config not found")
	}

	// HTTP/2の設定を確認（gRPCはHTTP/2必須）
	if _, ok := explicitConfig["http2_protocol_options"]; !ok {
		t.Error("expected http2_protocol_options for protocol: grpc")
	}

	// HTTP/1.1の設定がないことを確認
	if _, ok := explicitConfig["http_protocol_options"]; ok {
		t.Error("unexpected http1_protocol_options for protocol: grpc")
	}
}

func TestBuildConfig_MixedProtocols(t *testing.T) {
	// http/http2/grpcの共存確認
	routes := []Route{
		{
			Host:        "api.localhost",
			LocalPort:   10001,
			ClusterName: "api_cluster",
			Type:        "http",
			Protocol:    "http", // HTTP/1.1
		},
		{
			Host:        "api2.localhost",
			LocalPort:   10002,
			ClusterName: "api2_cluster",
			Type:        "http",
			Protocol:    "http2", // HTTP/2
		},
		{
			Host:        "grpc.localhost",
			LocalPort:   10003,
			ClusterName: "grpc_cluster",
			Type:        "http",
			Protocol:    "grpc", // gRPC (HTTP/2)
		},
	}

	cfg := BuildConfig(80, routes)

	staticRes, ok := cfg["static_resources"].(map[string]any)
	if !ok {
		t.Fatal("static_resources not found")
	}

	clusters, ok := staticRes["clusters"].([]any)
	if !ok {
		t.Fatal("clusters not found")
	}

	if len(clusters) != 3 {
		t.Fatalf("expected 3 clusters, got %d", len(clusters))
	}

	// 各クラスタのプロトコル設定を確認
	for i, expectedProtocol := range []string{"http1", "http2", "http2"} {
		cluster := clusters[i].(map[string]any)
		protocolOpts := cluster["typed_extension_protocol_options"].(map[string]any)
		httpOpts := protocolOpts["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"].(map[string]any)
		explicitConfig := httpOpts["explicit_http_config"].(map[string]any)

		if expectedProtocol == "http1" {
			if _, ok := explicitConfig["http_protocol_options"]; !ok {
				t.Errorf("cluster %d: expected http1_protocol_options", i)
			}
		} else {
			if _, ok := explicitConfig["http2_protocol_options"]; !ok {
				t.Errorf("cluster %d: expected http2_protocol_options", i)
			}
		}
	}
}
