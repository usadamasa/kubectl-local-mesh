package log

import (
	"strings"
	"testing"

	"github.com/usadamasa/kubectl-localmesh/internal/port"
)

func TestGenerateSummary_HTTPAndGRPCServices(t *testing.T) {
	services := []ServiceSummary{
		{
			Host:        "users-api.localhost",
			Protocol:    "grpc",
			DisplayType: "HTTP/gRPC",
			Backend:     "users/users-api:50051",
		},
		{
			Host:        "admin.localhost",
			Protocol:    "http",
			DisplayType: "HTTP/gRPC",
			Backend:     "admin/admin-web:8080",
		},
	}

	result := GenerateSummary(services, 80)

	// ヘッダーの確認
	if !strings.Contains(result, "Service Mesh is ready!") {
		t.Errorf("missing header in summary:\n%s", result)
	}

	// HTTP/gRPCセクションの確認
	if !strings.Contains(result, "HTTP/gRPC Services:") {
		t.Errorf("missing HTTP/gRPC section in summary:\n%s", result)
	}

	// 各サービスの確認
	if !strings.Contains(result, "http://users-api.localhost:80") {
		t.Errorf("missing users-api service in summary:\n%s", result)
	}
	if !strings.Contains(result, "(gRPC)") {
		t.Errorf("missing gRPC protocol marker in summary:\n%s", result)
	}
	if !strings.Contains(result, "users/users-api:50051") {
		t.Errorf("missing backend info for users-api in summary:\n%s", result)
	}

	if !strings.Contains(result, "http://admin.localhost:80") {
		t.Errorf("missing admin service in summary:\n%s", result)
	}
	if !strings.Contains(result, "(HTTP)") {
		t.Errorf("missing HTTP protocol marker in summary:\n%s", result)
	}

	// フッターの確認
	if !strings.Contains(result, "Press Ctrl+C to stop and cleanup.") {
		t.Errorf("missing footer in summary:\n%s", result)
	}
}

func TestGenerateSummary_TCPServices(t *testing.T) {
	services := []ServiceSummary{
		{
			Host:        "users-db.localhost",
			Protocol:    "tcp",
			DisplayType: "TCP",
			Backend:     "primary @ 10.0.0.1:5432",
			ListenPort:  5432,
		},
	}

	result := GenerateSummary(services, 80)

	// TCPセクションの確認
	if !strings.Contains(result, "TCP Services:") {
		t.Errorf("missing TCP section in summary:\n%s", result)
	}

	// TCPサービスの確認
	if !strings.Contains(result, "tcp://users-db.localhost:5432") {
		t.Errorf("missing TCP service URL in summary:\n%s", result)
	}
	if !strings.Contains(result, "primary @ 10.0.0.1:5432") {
		t.Errorf("missing backend info for TCP service in summary:\n%s", result)
	}
}

func TestGenerateSummary_MixedServices(t *testing.T) {
	services := []ServiceSummary{
		{
			Host:        "users-api.localhost",
			Protocol:    "grpc",
			DisplayType: "HTTP/gRPC",
			Backend:     "users/users-api:50051",
		},
		{
			Host:        "users-db.localhost",
			Protocol:    "tcp",
			DisplayType: "TCP",
			Backend:     "primary @ 10.0.0.1:5432",
			ListenPort:  5432,
		},
	}

	result := GenerateSummary(services, 80)

	// 両方のセクションが存在することを確認
	if !strings.Contains(result, "HTTP/gRPC Services:") {
		t.Errorf("missing HTTP/gRPC section in mixed summary:\n%s", result)
	}
	if !strings.Contains(result, "TCP Services:") {
		t.Errorf("missing TCP section in mixed summary:\n%s", result)
	}
}

func TestGenerateSummary_EmptyServices(t *testing.T) {
	result := GenerateSummary(nil, 80)

	if !strings.Contains(result, "Service Mesh is ready!") {
		t.Errorf("missing header even for empty services:\n%s", result)
	}

	// セクションヘッダーはないはず
	if strings.Contains(result, "HTTP/gRPC Services:") {
		t.Errorf("unexpected HTTP/gRPC section for empty services:\n%s", result)
	}
	if strings.Contains(result, "TCP Services:") {
		t.Errorf("unexpected TCP section for empty services:\n%s", result)
	}
}

func TestGenerateSummary_CustomListenerPort(t *testing.T) {
	services := []ServiceSummary{
		{
			Host:        "api.localhost",
			Protocol:    "http",
			DisplayType: "HTTP/gRPC",
			Backend:     "default/api:8080",
		},
	}

	result := GenerateSummary(services, 8080)

	if !strings.Contains(result, "http://api.localhost:8080") {
		t.Errorf("missing custom listener port in summary:\n%s", result)
	}
}

func TestGenerateSummary_OverwriteListenPort(t *testing.T) {
	services := []ServiceSummary{
		{
			Host:        "special-api.localhost",
			Protocol:    "grpc",
			DisplayType: "HTTP/gRPC",
			Backend:     "default/special-api:9090",
			ListenPort:  9090, // 特別なリスナーポート
		},
	}

	result := GenerateSummary(services, 80)

	// listener_portで指定されたポートが使われるべき
	if !strings.Contains(result, "http://special-api.localhost:9090") {
		t.Errorf("missing overwritten listen port in summary:\n%s", result)
	}
}

func TestServiceSummary_EffectiveListenPort(t *testing.T) {
	tests := []struct {
		name           string
		summary        ServiceSummary
		defaultPort    port.ListenerPort
		wantListenPort port.ListenerPort
	}{
		{
			name: "use default port when ListenPort is 0",
			summary: ServiceSummary{
				Host:       "api.localhost",
				ListenPort: 0,
			},
			defaultPort:    80,
			wantListenPort: 80,
		},
		{
			name: "use ListenPort when specified",
			summary: ServiceSummary{
				Host:       "api.localhost",
				ListenPort: 9090,
			},
			defaultPort:    80,
			wantListenPort: 9090,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.summary.EffectiveListenPort(tt.defaultPort)
			if got != tt.wantListenPort {
				t.Errorf("EffectiveListenPort() = %d, want %d", got, tt.wantListenPort)
			}
		})
	}
}
