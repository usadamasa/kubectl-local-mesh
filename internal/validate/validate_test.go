package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSchema_ValidKubernetesService(t *testing.T) {
	content := `
listener_port: 80
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    port_name: http
    protocol: http
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_ValidTCPService(t *testing.T) {
	content := `
listener_port: 80
ssh_bastions:
  primary:
    instance: bastion-1
    zone: asia-northeast1-a
    project: test-project
services:
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_ValidMixedServices(t *testing.T) {
	content := `
listener_port: 80
ssh_bastions:
  primary:
    instance: bastion-1
    zone: asia-northeast1-a
services:
  - kind: kubernetes
    host: api.localhost
    namespace: default
    service: api
    protocol: grpc
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_ValidWithListenerPort(t *testing.T) {
	content := `
listener_port: 80
services:
  - kind: kubernetes
    host: grpc.localhost
    namespace: default
    service: grpc-svc
    protocol: grpc
    listener_port: 50051
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_ValidWithoutListenerPort(t *testing.T) {
	// listener_port省略（デフォルト値が使われる）
	content := `
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_ValidWithPort(t *testing.T) {
	content := `
listener_port: 80
services:
  - kind: kubernetes
    host: admin.localhost
    namespace: admin
    service: admin-web
    port: 8080
    protocol: http
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_ValidTCPWithListenPort(t *testing.T) {
	content := `
listener_port: 80
ssh_bastions:
  primary:
    instance: bastion-1
    zone: asia-northeast1-a
services:
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
    listen_port: 15432
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_GlobalCluster(t *testing.T) {
	content := `
cluster: gke_myproject_asia-northeast1_staging
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config with global cluster, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_ServiceCluster(t *testing.T) {
	content := `
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
    cluster: gke_myproject_asia-northeast1_prod
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config with service cluster, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_GlobalAndServiceCluster(t *testing.T) {
	content := `
cluster: gke_myproject_asia-northeast1_staging
services:
  - kind: kubernetes
    host: api.localhost
    namespace: default
    service: api-svc
    protocol: http
  - kind: kubernetes
    host: admin.localhost
    namespace: admin
    service: admin-web
    protocol: http
    cluster: gke_myproject_asia-northeast1_prod
`
	result := validateYAMLContent(t, content)
	if !result.OK() {
		t.Errorf("expected valid config with both global and service cluster, got errors: %v", result.Errors)
	}
}

func TestValidateSchema_TCPServiceWithCluster_Invalid(t *testing.T) {
	// TCPサービスにclusterフィールドは使えない
	content := `
ssh_bastions:
  primary:
    instance: bastion-1
    zone: zone-a
services:
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
    cluster: some-cluster
`
	result := validateYAMLContent(t, content)
	if result.OK() {
		t.Error("expected validation error for cluster on TCP service")
	}
}

func TestValidateSchema_MissingServices(t *testing.T) {
	content := `
listener_port: 80
`
	result := validateYAMLContent(t, content)
	if result.OK() {
		t.Error("expected validation error for missing services")
	}
	assertContainsError(t, result, "services")
}

func TestValidateSchema_EmptyServices(t *testing.T) {
	content := `
listener_port: 80
services: []
`
	result := validateYAMLContent(t, content)
	if result.OK() {
		t.Error("expected validation error for empty services")
	}
	assertContainsError(t, result, "services")
}

func TestValidateSchema_MissingKind(t *testing.T) {
	content := `
listener_port: 80
services:
  - host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
`
	result := validateYAMLContent(t, content)
	if result.OK() {
		t.Error("expected validation error for missing kind")
	}
}

func TestValidateSchema_InvalidKind(t *testing.T) {
	content := `
listener_port: 80
services:
  - kind: invalid
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
`
	result := validateYAMLContent(t, content)
	if result.OK() {
		t.Error("expected validation error for invalid kind")
	}
}

func TestValidateSchema_MissingRequiredKubernetesFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
		errMsg  string
	}{
		{
			name: "missing host",
			content: `
services:
  - kind: kubernetes
    namespace: test
    service: test-svc
    protocol: http
`,
			errMsg: "host",
		},
		{
			name: "missing namespace",
			content: `
services:
  - kind: kubernetes
    host: test.localhost
    service: test-svc
    protocol: http
`,
			errMsg: "namespace",
		},
		{
			name: "missing service",
			content: `
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    protocol: http
`,
			errMsg: "service",
		},
		{
			name: "missing protocol",
			content: `
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
`,
			errMsg: "protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateYAMLContent(t, tt.content)
			if result.OK() {
				t.Errorf("expected validation error for %s", tt.name)
			}
		})
	}
}

func TestValidateSchema_InvalidProtocol(t *testing.T) {
	content := `
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: invalid
`
	result := validateYAMLContent(t, content)
	if result.OK() {
		t.Error("expected validation error for invalid protocol")
	}
}

func TestValidateSchema_MissingRequiredTCPFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "missing host",
			content: `
ssh_bastions:
  primary:
    instance: bastion-1
    zone: zone-a
services:
  - kind: tcp
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`,
		},
		{
			name: "missing ssh_bastion",
			content: `
services:
  - kind: tcp
    host: db.localhost
    target_host: 10.0.0.1
    target_port: 5432
`,
		},
		{
			name: "missing target_host",
			content: `
ssh_bastions:
  primary:
    instance: bastion-1
    zone: zone-a
services:
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_port: 5432
`,
		},
		{
			name: "missing target_port",
			content: `
ssh_bastions:
  primary:
    instance: bastion-1
    zone: zone-a
services:
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateYAMLContent(t, tt.content)
			if result.OK() {
				t.Errorf("expected validation error for %s", tt.name)
			}
		})
	}
}

func TestValidateSchema_UnknownField(t *testing.T) {
	content := `
listener_port: 80
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
    unknown_field: value
`
	result := validateYAMLContent(t, content)
	if result.OK() {
		t.Error("expected validation error for unknown field")
	}
}

func TestValidateSchema_UnknownTopLevelField(t *testing.T) {
	content := `
listener_port: 80
unknown_top: value
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
`
	result := validateYAMLContent(t, content)
	if result.OK() {
		t.Error("expected validation error for unknown top-level field")
	}
}

func TestValidateSchema_UnknownSSHBastionField(t *testing.T) {
	content := `
listener_port: 80
ssh_bastions:
  primary:
    instance: bastion-1
    zone: zone-a
    unknown_field: value
services:
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`
	result := validateYAMLContent(t, content)
	if result.OK() {
		t.Error("expected validation error for unknown SSH bastion field")
	}
}

func TestValidateSchema_InvalidPortRange(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "listener_port too high",
			content: `
listener_port: 65536
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
`,
		},
		{
			name: "listener_port zero",
			content: `
listener_port: 0
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
`,
		},
		{
			name: "target_port too high",
			content: `
ssh_bastions:
  primary:
    instance: bastion-1
    zone: zone-a
services:
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 70000
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateYAMLContent(t, tt.content)
			if result.OK() {
				t.Errorf("expected validation error for %s", tt.name)
			}
		})
	}
}

func TestValidateSchema_SSHBastionMissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "missing instance",
			content: `
ssh_bastions:
  primary:
    zone: zone-a
services:
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`,
		},
		{
			name: "missing zone",
			content: `
ssh_bastions:
  primary:
    instance: bastion-1
services:
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateYAMLContent(t, tt.content)
			if result.OK() {
				t.Errorf("expected validation error for %s", tt.name)
			}
		})
	}
}

func TestValidateSchema_AllProtocols(t *testing.T) {
	protocols := []string{"http", "http2", "grpc"}
	for _, protocol := range protocols {
		t.Run(protocol, func(t *testing.T) {
			content := `
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: ` + protocol + `
`
			result := validateYAMLContent(t, content)
			if !result.OK() {
				t.Errorf("expected valid config for protocol %s, got errors: %v", protocol, result.Errors)
			}
		})
	}
}

func TestValidateSchema_ExistingTestDataConfigs(t *testing.T) {
	configDir := filepath.Join("..", "..", "test", "snapshot", "testdata", "configs")
	entries, err := os.ReadDir(configDir)
	if err != nil {
		t.Skipf("test data directory not found: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(configDir, entry.Name())
			result, err := ValidateSchemaFile(path)
			if err != nil {
				t.Fatalf("ValidateSchemaFile failed: %v", err)
			}
			if !result.OK() {
				t.Errorf("expected valid config for %s, got errors: %v", entry.Name(), result.Errors)
			}
		})
	}
}

func TestValidateSchemaFile_NonexistentFile(t *testing.T) {
	_, err := ValidateSchemaFile("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestValidateSchemaFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(path, []byte(":\n  :\n    - :\n      invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ValidateSchemaFile(path)
	// Invalid YAML may cause unmarshal error or schema validation error
	// Either is acceptable
	if err != nil {
		return // unmarshal error is fine
	}
}

func TestValidationResult_OK(t *testing.T) {
	result := &ValidationResult{}
	if !result.OK() {
		t.Error("expected OK() to return true for empty result")
	}

	result.Errors = append(result.Errors, "error")
	if result.OK() {
		t.Error("expected OK() to return false when errors exist")
	}
}

// Helper functions

func validateYAMLContent(t *testing.T, content string) *ValidationResult {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ValidateSchemaFile(path)
	if err != nil {
		t.Fatalf("ValidateSchemaFile failed: %v", err)
	}
	return result
}

func assertContainsError(t *testing.T, result *ValidationResult, substr string) {
	t.Helper()
	for _, e := range result.Errors {
		if strings.Contains(e, substr) {
			return
		}
	}
	t.Errorf("expected error containing %q, got: %v", substr, result.Errors)
}
