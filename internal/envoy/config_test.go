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
