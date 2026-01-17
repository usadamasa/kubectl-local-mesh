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

// IndividualListenerComponents は個別リスナーを持つサービス用のEnvoy設定コンポーネント
// OverwriteListenPortsが指定された場合に使用（HTTP/HTTP2/gRPC問わず）
type IndividualListenerComponents struct {
	Cluster   map[string]any
	Listeners []map[string]any // 各OverwriteListenPortに対応するリスナー
}
