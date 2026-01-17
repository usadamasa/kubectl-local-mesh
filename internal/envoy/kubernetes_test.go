package envoy

import (
	"testing"
)

func TestKubernetesServiceBuilder_Build_HTTPRoute(t *testing.T) {
	// OverwriteListenPortsなし → HTTPComponents（従来動作）
	builder := NewKubernetesServiceBuilder(
		"api.localhost", "http",
		"default", "api", "http", 8080,
		nil, // OverwriteListenPorts
	)

	result := builder.Build("api_cluster", 10001)

	// HTTPComponentsを返すことを確認
	httpComponents, ok := result.(HTTPComponents)
	if !ok {
		t.Fatalf("expected HTTPComponents, got %T", result)
	}

	// クラスタ設定の確認
	if httpComponents.Cluster["name"] != "api_cluster" {
		t.Errorf("expected cluster name 'api_cluster', got %v", httpComponents.Cluster["name"])
	}

	// ルート設定の確認
	if httpComponents.Route["name"] != "api_cluster" {
		t.Errorf("expected route name 'api_cluster', got %v", httpComponents.Route["name"])
	}
}

func TestKubernetesServiceBuilder_Build_WithOverwriteListenPorts(t *testing.T) {
	// OverwriteListenPortsあり → IndividualListenerComponents（個別リスナー）
	builder := NewKubernetesServiceBuilder(
		"grpc.localhost", "grpc",
		"default", "grpc-service", "grpc", 50051,
		[]int{50051, 50052},
	)

	result := builder.Build("grpc_cluster", 10001)

	// IndividualListenerComponentsを返すことを確認
	listenerComponents, ok := result.(IndividualListenerComponents)
	if !ok {
		t.Fatalf("expected IndividualListenerComponents, got %T", result)
	}

	// クラスタ設定の確認
	if listenerComponents.Cluster["name"] != "grpc_cluster" {
		t.Errorf("expected cluster name 'grpc_cluster', got %v", listenerComponents.Cluster["name"])
	}

	// リスナーが2つあることを確認
	if len(listenerComponents.Listeners) != 2 {
		t.Errorf("expected 2 listeners, got %d", len(listenerComponents.Listeners))
	}

	// 各リスナーのポート確認
	expectedPorts := []int{50051, 50052}
	for i, listener := range listenerComponents.Listeners {
		address := listener["address"].(map[string]any)
		socketAddr := address["socket_address"].(map[string]any)
		port := socketAddr["port_value"].(int)
		if port != expectedPorts[i] {
			t.Errorf("expected listener %d port %d, got %d", i, expectedPorts[i], port)
		}
	}
}

func TestKubernetesServiceBuilder_Build_SingleListenPort(t *testing.T) {
	// OverwriteListenPorts 1つ → IndividualListenerComponents
	builder := NewKubernetesServiceBuilder(
		"grpc.localhost", "grpc",
		"default", "grpc-service", "grpc", 50051,
		[]int{50051},
	)

	result := builder.Build("grpc_cluster", 10001)

	listenerComponents, ok := result.(IndividualListenerComponents)
	if !ok {
		t.Fatalf("expected IndividualListenerComponents, got %T", result)
	}

	// リスナーが1つあることを確認
	if len(listenerComponents.Listeners) != 1 {
		t.Errorf("expected 1 listener, got %d", len(listenerComponents.Listeners))
	}

	// リスナーのポート確認
	listener := listenerComponents.Listeners[0]
	address := listener["address"].(map[string]any)
	socketAddr := address["socket_address"].(map[string]any)
	port := socketAddr["port_value"].(int)
	if port != 50051 {
		t.Errorf("expected listener port 50051, got %d", port)
	}
}

func TestKubernetesServiceBuilder_Build_HTTP_WithOverwriteListenPorts(t *testing.T) {
	// HTTPプロトコルでOverwriteListenPortsあり → HTTP/1.1リスナー
	builder := NewKubernetesServiceBuilder(
		"http.localhost", "http",
		"default", "http-service", "http", 8080,
		[]int{8080},
	)

	result := builder.Build("http_cluster", 10001)

	listenerComponents, ok := result.(IndividualListenerComponents)
	if !ok {
		t.Fatalf("expected IndividualListenerComponents, got %T", result)
	}

	// リスナーが1つあることを確認
	if len(listenerComponents.Listeners) != 1 {
		t.Errorf("expected 1 listener, got %d", len(listenerComponents.Listeners))
	}

	// HTTP/1.1の設定確認（クラスタ側）
	protocolOpts := listenerComponents.Cluster["typed_extension_protocol_options"].(map[string]any)
	httpOpts := protocolOpts["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"].(map[string]any)
	explicitConfig := httpOpts["explicit_http_config"].(map[string]any)

	if _, ok := explicitConfig["http_protocol_options"]; !ok {
		t.Error("expected http_protocol_options for HTTP/1.1 protocol")
	}
}

func TestKubernetesServiceBuilder_Build_gRPC_WithOverwriteListenPorts(t *testing.T) {
	// gRPCプロトコルでOverwriteListenPortsあり → HTTP/2リスナー
	builder := NewKubernetesServiceBuilder(
		"grpc.localhost", "grpc",
		"default", "grpc-service", "grpc", 50051,
		[]int{50051},
	)

	result := builder.Build("grpc_cluster", 10001)

	listenerComponents, ok := result.(IndividualListenerComponents)
	if !ok {
		t.Fatalf("expected IndividualListenerComponents, got %T", result)
	}

	// HTTP/2の設定確認（クラスタ側）
	protocolOpts := listenerComponents.Cluster["typed_extension_protocol_options"].(map[string]any)
	httpOpts := protocolOpts["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"].(map[string]any)
	explicitConfig := httpOpts["explicit_http_config"].(map[string]any)

	if _, ok := explicitConfig["http2_protocol_options"]; !ok {
		t.Error("expected http2_protocol_options for gRPC protocol")
	}
}

func TestKubernetesServiceBuilder_GetHost(t *testing.T) {
	builder := NewKubernetesServiceBuilder(
		"test.localhost", "http",
		"default", "test", "http", 8080,
		nil,
	)

	if builder.GetHost() != "test.localhost" {
		t.Errorf("expected host 'test.localhost', got '%s'", builder.GetHost())
	}
}
