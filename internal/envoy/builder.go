package envoy

// HTTPComponents はHTTPサービス用のEnvoy設定コンポーネント
type HTTPComponents struct {
	Cluster map[string]any
	Route   map[string]any
}

// TCPComponents はTCPサービス用のEnvoy設定コンポーネント
type TCPComponents struct {
	Cluster  map[string]any
	Listener map[string]any
}
