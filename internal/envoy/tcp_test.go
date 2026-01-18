package envoy

import (
	"testing"

	"github.com/usadamasa/kubectl-localmesh/internal/port"
)

func TestTCPServiceBuilder_Build(t *testing.T) {
	t.Run("ListenAddrが設定されている場合はそのアドレスを使用", func(t *testing.T) {
		builder := NewTCPServiceBuilder(
			"db.localhost",
			port.TCPPort(5432),
			"127.0.0.5", // ListenAddr
			"primary",
			"10.0.0.1",
			port.TCPPort(5432),
		)

		components := builder.Build("tcp_db", 12345)

		// リスナー設定を確認
		listener := components.Listener
		address, ok := listener["address"].(map[string]any)
		if !ok {
			t.Fatal("address not found in listener")
		}
		socketAddr, ok := address["socket_address"].(map[string]any)
		if !ok {
			t.Fatal("socket_address not found in address")
		}

		// ListenAddrが使用されていることを確認
		if socketAddr["address"] != "127.0.0.5" {
			t.Errorf("expected address 127.0.0.5, got %v", socketAddr["address"])
		}
		if socketAddr["port_value"] != 5432 {
			t.Errorf("expected port_value 5432, got %v", socketAddr["port_value"])
		}
	})

	t.Run("クラスタ設定が正しく生成される", func(t *testing.T) {
		builder := NewTCPServiceBuilder(
			"db.localhost",
			port.TCPPort(5432),
			"127.0.0.2",
			"primary",
			"10.0.0.1",
			port.TCPPort(5432),
		)

		components := builder.Build("tcp_db", 54321)

		// クラスタ設定を確認
		cluster := components.Cluster
		if cluster["name"] != "tcp_db" {
			t.Errorf("expected cluster name tcp_db, got %v", cluster["name"])
		}

		// ロードアサインメントを確認
		loadAssignment, ok := cluster["load_assignment"].(map[string]any)
		if !ok {
			t.Fatal("load_assignment not found")
		}
		endpoints, ok := loadAssignment["endpoints"].([]any)
		if !ok || len(endpoints) == 0 {
			t.Fatal("endpoints not found or empty")
		}
		ep := endpoints[0].(map[string]any)
		lbEndpoints := ep["lb_endpoints"].([]any)
		endpoint := lbEndpoints[0].(map[string]any)["endpoint"].(map[string]any)
		addr := endpoint["address"].(map[string]any)["socket_address"].(map[string]any)

		// ローカルポートへ接続することを確認
		if addr["address"] != "127.0.0.1" {
			t.Errorf("expected cluster endpoint 127.0.0.1, got %v", addr["address"])
		}
		if addr["port_value"] != 54321 {
			t.Errorf("expected cluster port 54321, got %v", addr["port_value"])
		}
	})
}

func TestTCPServiceBuilder_GetHost(t *testing.T) {
	builder := NewTCPServiceBuilder(
		"mydb.localhost",
		port.TCPPort(3306),
		"127.0.0.3",
		"bastion",
		"10.0.0.5",
		port.TCPPort(3306),
	)

	if builder.GetHost() != "mydb.localhost" {
		t.Errorf("expected mydb.localhost, got %s", builder.GetHost())
	}
}

func TestTCPServiceBuilder_GetListenAddr(t *testing.T) {
	builder := NewTCPServiceBuilder(
		"mydb.localhost",
		port.TCPPort(3306),
		"127.0.0.42",
		"bastion",
		"10.0.0.5",
		port.TCPPort(3306),
	)

	if builder.GetListenAddr() != "127.0.0.42" {
		t.Errorf("expected 127.0.0.42, got %s", builder.GetListenAddr())
	}
}
