package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultListenerPort(t *testing.T) {
	// listener_portを指定しない設定ファイル
	content := `
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    port: 8080
    protocol: http
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expectedPort := 80
	if cfg.ListenerPort != expectedPort {
		t.Errorf("expected default listener_port %d, got %d", expectedPort, cfg.ListenerPort)
	}
}

func TestLoad_ExplicitListenerPort(t *testing.T) {
	// listener_portを明示的に指定
	content := `
listener_port: 8080
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    port: 8080
    protocol: http
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expectedPort := 8080
	if cfg.ListenerPort != expectedPort {
		t.Errorf("expected listener_port %d, got %d", expectedPort, cfg.ListenerPort)
	}
}

func TestLoad_SSHBastionWithTCPService(t *testing.T) {
	// SSH Bastion経由のTCP接続設定
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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// SSH Bastionの確認
	if len(cfg.SSHBastions) != 1 {
		t.Fatalf("expected 1 ssh_bastion, got %d", len(cfg.SSHBastions))
	}
	bastion, ok := cfg.SSHBastions["primary"]
	if !ok {
		t.Fatal("expected bastion 'primary' not found")
	}
	if bastion.Instance != "bastion-1" {
		t.Errorf("expected instance 'bastion-1', got '%s'", bastion.Instance)
	}
	if bastion.Zone != "asia-northeast1-a" {
		t.Errorf("expected zone 'asia-northeast1-a', got '%s'", bastion.Zone)
	}
	if bastion.Project != "test-project" {
		t.Errorf("expected project 'test-project', got '%s'", bastion.Project)
	}

	// Serviceの確認
	if len(cfg.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(cfg.Services))
	}
	svcDef := cfg.Services[0]
	tcpSvc, ok := svcDef.AsTCP()
	if !ok {
		t.Fatal("expected TCP service, got different type")
	}
	if tcpSvc.Host != "db.localhost" {
		t.Errorf("expected host 'db.localhost', got '%s'", tcpSvc.Host)
	}
	if tcpSvc.SSHBastion != "primary" {
		t.Errorf("expected ssh_bastion 'primary', got '%s'", tcpSvc.SSHBastion)
	}
	if tcpSvc.TargetHost != "10.0.0.1" {
		t.Errorf("expected target_host '10.0.0.1', got '%s'", tcpSvc.TargetHost)
	}
	if tcpSvc.TargetPort != 5432 {
		t.Errorf("expected target_port 5432, got %d", tcpSvc.TargetPort)
	}
}

func TestLoad_SSHBastionReferenceNotFound(t *testing.T) {
	// 存在しないbastion名を参照
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
    ssh_bastion: nonexistent
    target_host: 10.0.0.1
    target_port: 5432
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for nonexistent bastion reference, got nil")
	}
	if !containsString(err.Error(), "ssh_bastion 'nonexistent' not found") {
		t.Errorf("expected error containing 'ssh_bastion 'nonexistent' not found', got '%s'", err.Error())
	}
}

func TestLoad_TCPServiceWithoutTargetHost(t *testing.T) {
	// kind=tcpだがtarget_hostが未指定
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
    target_port: 5432
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for missing target_host, got nil")
	}
	if !containsString(err.Error(), "target_host is required") {
		t.Errorf("expected error containing 'target_host is required', got '%s'", err.Error())
	}
}

func TestLoad_TCPServiceWithoutTargetPort(t *testing.T) {
	// kind=tcpだがtarget_portが未指定
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
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for missing target_port, got nil")
	}
	if !containsString(err.Error(), "target_port is required") {
		t.Errorf("expected error containing 'target_port is required', got '%s'", err.Error())
	}
}

func TestLoad_K8sServiceWithoutNamespace(t *testing.T) {
	// kind=kubernetesだがnamespaceが未指定
	content := `
listener_port: 80
services:
  - kind: kubernetes
    host: api.localhost
    service: api-svc
    protocol: http
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for missing namespace, got nil")
	}
	if !containsString(err.Error(), "namespace is required") {
		t.Errorf("expected error containing 'namespace is required', got '%s'", err.Error())
	}
}

func TestLoad_MixedK8sAndTCPServices(t *testing.T) {
	// Kubernetes ServiceとTCP Serviceの混在
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
    service: api-svc
    protocol: grpc
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cfg.Services))
	}

	// Kubernetes Service
	k8sSvcDef := cfg.Services[0]
	k8sSvc, ok := k8sSvcDef.AsKubernetes()
	if !ok {
		t.Fatal("expected first service to be Kubernetes service")
	}
	if k8sSvc.Protocol != "grpc" {
		t.Errorf("expected protocol 'grpc', got '%s'", k8sSvc.Protocol)
	}
	if k8sSvc.Namespace != "default" {
		t.Errorf("expected namespace 'default', got '%s'", k8sSvc.Namespace)
	}
	if k8sSvc.Service != "api-svc" {
		t.Errorf("expected service 'api-svc', got '%s'", k8sSvc.Service)
	}

	// TCP Service
	tcpSvcDef := cfg.Services[1]
	tcpSvc, ok := tcpSvcDef.AsTCP()
	if !ok {
		t.Fatal("expected second service to be TCP service")
	}
	if tcpSvc.SSHBastion != "primary" {
		t.Errorf("expected ssh_bastion 'primary', got '%s'", tcpSvc.SSHBastion)
	}
	if tcpSvc.TargetHost != "10.0.0.1" {
		t.Errorf("expected target_host '10.0.0.1', got '%s'", tcpSvc.TargetHost)
	}
}

// ========== 新形式（kind: kubernetes / kind: tcp）のテスト ==========

func TestLoad_MissingKindField(t *testing.T) {
	// kindフィールドが存在しない場合
	content := `
listener_port: 80
services:
  - host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for missing kind field, got nil")
	}
	if !containsString(err.Error(), "kind") {
		t.Errorf("expected error to contain 'kind', got '%s'", err.Error())
	}
}

func TestLoad_InvalidKindValue(t *testing.T) {
	// 不正なkind値
	content := `
listener_port: 80
services:
  - kind: invalid_kind
    host: test.localhost
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid kind value, got nil")
	}
	if !containsString(err.Error(), "unknown service kind") && !containsString(err.Error(), "invalid_kind") {
		t.Errorf("expected error to contain 'unknown service kind' or 'invalid_kind', got '%s'", err.Error())
	}
}

func TestLoad_NewFormat_KubernetesServiceValid(t *testing.T) {
	// 新形式のKubernetesサービス（正常系）
	content := `
listener_port: 80
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(cfg.Services))
	}

	svcDef := cfg.Services[0]
	k8sSvc, ok := svcDef.AsKubernetes()
	if !ok {
		t.Fatal("expected KubernetesService, got different type")
	}

	if k8sSvc.Host != "test.localhost" {
		t.Errorf("expected host 'test.localhost', got '%s'", k8sSvc.Host)
	}
	if k8sSvc.Namespace != "test" {
		t.Errorf("expected namespace 'test', got '%s'", k8sSvc.Namespace)
	}
	if k8sSvc.Service != "test-svc" {
		t.Errorf("expected service 'test-svc', got '%s'", k8sSvc.Service)
	}
	if k8sSvc.Protocol != "http" {
		t.Errorf("expected protocol 'http', got '%s'", k8sSvc.Protocol)
	}
}

func TestLoad_NewFormat_KubernetesServiceMissingNamespace(t *testing.T) {
	// 新形式のKubernetesサービス（namespace不足）
	content := `
listener_port: 80
services:
  - kind: kubernetes
    host: test.localhost
    service: test-svc
    protocol: http
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for missing namespace, got nil")
	}
	if !containsString(err.Error(), "namespace") && !containsString(err.Error(), "required") {
		t.Errorf("expected error to contain 'namespace' and 'required', got '%s'", err.Error())
	}
}

func TestLoad_NewFormat_KubernetesServiceInvalidProtocol(t *testing.T) {
	// 新形式のKubernetesサービス（不正なprotocol）
	content := `
listener_port: 80
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: invalid
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid protocol, got nil")
	}
	if !containsString(err.Error(), "protocol") {
		t.Errorf("expected error to contain 'protocol', got '%s'", err.Error())
	}
}

func TestLoad_NewFormat_TCPServiceValid(t *testing.T) {
	// 新形式のTCPサービス（正常系）
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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(cfg.Services))
	}

	svcDef := cfg.Services[0]
	tcpSvc, ok := svcDef.AsTCP()
	if !ok {
		t.Fatal("expected TCPService, got different type")
	}

	if tcpSvc.Host != "db.localhost" {
		t.Errorf("expected host 'db.localhost', got '%s'", tcpSvc.Host)
	}
	if tcpSvc.SSHBastion != "primary" {
		t.Errorf("expected ssh_bastion 'primary', got '%s'", tcpSvc.SSHBastion)
	}
	if tcpSvc.TargetHost != "10.0.0.1" {
		t.Errorf("expected target_host '10.0.0.1', got '%s'", tcpSvc.TargetHost)
	}
	if tcpSvc.TargetPort != 5432 {
		t.Errorf("expected target_port 5432, got %d", tcpSvc.TargetPort)
	}
}

func TestLoad_NewFormat_TCPServiceMissingTargetHost(t *testing.T) {
	// 新形式のTCPサービス（target_host不足）
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
    target_port: 5432
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for missing target_host, got nil")
	}
	if !containsString(err.Error(), "target_host") && !containsString(err.Error(), "required") {
		t.Errorf("expected error to contain 'target_host' and 'required', got '%s'", err.Error())
	}
}

func TestLoad_NewFormat_MixedServicesValid(t *testing.T) {
	// 新形式の混在設定（正常系）
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
    service: api-svc
    protocol: grpc
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cfg.Services))
	}

	// Kubernetes Service
	k8sSvcDef := cfg.Services[0]
	k8sSvc, ok := k8sSvcDef.AsKubernetes()
	if !ok {
		t.Fatal("expected first service to be KubernetesService")
	}
	if k8sSvc.Protocol != "grpc" {
		t.Errorf("expected protocol 'grpc', got '%s'", k8sSvc.Protocol)
	}
	if k8sSvc.Namespace != "default" {
		t.Errorf("expected namespace 'default', got '%s'", k8sSvc.Namespace)
	}

	// TCP Service
	tcpSvcDef := cfg.Services[1]
	tcpSvc, ok := tcpSvcDef.AsTCP()
	if !ok {
		t.Fatal("expected second service to be TCPService")
	}
	if tcpSvc.SSHBastion != "primary" {
		t.Errorf("expected ssh_bastion 'primary', got '%s'", tcpSvc.SSHBastion)
	}
	if tcpSvc.TargetHost != "10.0.0.1" {
		t.Errorf("expected target_host '10.0.0.1', got '%s'", tcpSvc.TargetHost)
	}
}

func TestServiceDefinition_AsKubernetes(t *testing.T) {
	// KubernetesServiceの型アサーションテスト
	content := `
listener_port: 80
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// SSH bastionsを追加
	contentWithBastion := `
listener_port: 80
ssh_bastions:
  primary:
    instance: bastion-1
    zone: asia-northeast1-a
services:
  - kind: kubernetes
    host: test.localhost
    namespace: test
    service: test-svc
    protocol: http
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`
	if err := os.WriteFile(configPath, []byte(contentWithBastion), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// KubernetesServiceはAsKubernetes()がtrueを返す
	k8sSvcDef := cfg.Services[0]
	if _, ok := k8sSvcDef.AsKubernetes(); !ok {
		t.Error("expected AsKubernetes() to return true for KubernetesService")
	}
	if _, ok := k8sSvcDef.AsTCP(); ok {
		t.Error("expected AsTCP() to return false for KubernetesService")
	}

	// TCPServiceはAsTCP()がtrueを返す
	tcpSvcDef := cfg.Services[1]
	if _, ok := tcpSvcDef.AsTCP(); !ok {
		t.Error("expected AsTCP() to return true for TCPService")
	}
	if _, ok := tcpSvcDef.AsKubernetes(); ok {
		t.Error("expected AsKubernetes() to return false for TCPService")
	}
}

// ヘルパー関数
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// カバレッジ向上のための追加テスト

func TestServiceInterface_GetHostAndKind(t *testing.T) {
	// GetHost()とGetKind()メソッドのテスト
	content := `
listener_port: 80
ssh_bastions:
  primary:
    instance: bastion-1
    zone: asia-northeast1-a
    project: test-project
services:
  - kind: kubernetes
    host: k8s.localhost
    namespace: test
    service: test-svc
    protocol: http
  - kind: tcp
    host: db.localhost
    ssh_bastion: primary
    target_host: 10.0.0.1
    target_port: 5432
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cfg.Services))
	}

	// Kubernetes Serviceのテスト
	k8sSvc := cfg.Services[0].Get()
	if k8sSvc.GetHost() != "k8s.localhost" {
		t.Errorf("expected host 'k8s.localhost', got '%s'", k8sSvc.GetHost())
	}
	if k8sSvc.GetKind() != "kubernetes" {
		t.Errorf("expected kind 'kubernetes', got '%s'", k8sSvc.GetKind())
	}

	// TCP Serviceのテスト
	tcpSvc := cfg.Services[1].Get()
	if tcpSvc.GetHost() != "db.localhost" {
		t.Errorf("expected host 'db.localhost', got '%s'", tcpSvc.GetHost())
	}
	if tcpSvc.GetKind() != "tcp" {
		t.Errorf("expected kind 'tcp', got '%s'", tcpSvc.GetKind())
	}
}

func TestServiceDefinition_MarshalYAML(t *testing.T) {
	// MarshalYAMLのテスト（KubernetesService）
	k8sSvc := &KubernetesService{
		Host:      "test.localhost",
		Namespace: "test",
		Service:   "test-svc",
		Protocol:  "http",
	}
	svcDef := ServiceDefinition{service: k8sSvc}

	data, err := svcDef.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML failed for KubernetesService: %v", err)
	}
	if data == nil {
		t.Error("expected non-nil marshaled data")
	}

	// MarshalYAMLのテスト（TCPService）
	tcpSvc := &TCPService{
		Host:       "db.localhost",
		SSHBastion: "primary",
		TargetHost: "10.0.0.1",
		TargetPort: 5432,
	}
	tcpSvcDef := ServiceDefinition{service: tcpSvc}

	tcpData, err := tcpSvcDef.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML failed for TCPService: %v", err)
	}
	if tcpData == nil {
		t.Error("expected non-nil marshaled data for TCP service")
	}
}

func TestLoadMockConfig_Valid(t *testing.T) {
	// LoadMockConfigのテスト
	content := `
mocks:
  - namespace: test
    service: test-svc
    port_name: http
    resolved_port: 8080
  - namespace: another
    service: another-svc
    port_name: grpc
    resolved_port: 50051
`
	tmpDir := t.TempDir()
	mockPath := filepath.Join(tmpDir, "mocks.yaml")
	if err := os.WriteFile(mockPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	mockCfg, err := LoadMockConfig(mockPath)
	if err != nil {
		t.Fatalf("LoadMockConfig failed: %v", err)
	}

	if mockCfg == nil {
		t.Fatal("expected non-nil MockConfig")
	}

	if len(mockCfg.Mocks) != 2 {
		t.Fatalf("expected 2 mocks, got %d", len(mockCfg.Mocks))
	}

	// 最初のモック
	if mockCfg.Mocks[0].Namespace != "test" {
		t.Errorf("expected namespace 'test', got '%s'", mockCfg.Mocks[0].Namespace)
	}
	if mockCfg.Mocks[0].Service != "test-svc" {
		t.Errorf("expected service 'test-svc', got '%s'", mockCfg.Mocks[0].Service)
	}
	if mockCfg.Mocks[0].ResolvedPort != 8080 {
		t.Errorf("expected resolved_port 8080, got %d", mockCfg.Mocks[0].ResolvedPort)
	}
}

func TestLoadMockConfig_EmptyPath(t *testing.T) {
	// 空パスの場合はnilを返す
	mockCfg, err := LoadMockConfig("")
	if err != nil {
		t.Fatalf("expected no error for empty path, got %v", err)
	}
	if mockCfg != nil {
		t.Error("expected nil MockConfig for empty path")
	}
}

func TestLoadMockConfig_InvalidFile(t *testing.T) {
	// 存在しないファイル
	_, err := LoadMockConfig("/nonexistent/path/to/mocks.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestLoadMockConfig_InvalidYAML(t *testing.T) {
	// 不正なYAMLファイル
	content := `
mocks:
  - namespace: test
    invalid yaml here
`
	tmpDir := t.TempDir()
	mockPath := filepath.Join(tmpDir, "mocks.yaml")
	if err := os.WriteFile(mockPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadMockConfig(mockPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestServiceDefinition_MarshalYAML_UnknownType(t *testing.T) {
	// 未知の型でMarshalYAMLを呼ぶ（nilサービス）
	svcDef := ServiceDefinition{service: nil}
	_, err := svcDef.MarshalYAML()
	if err == nil {
		t.Fatal("expected error for nil service, got nil")
	}
}

func TestKubernetesService_ValidateEdgeCases(t *testing.T) {
	// KubernetesServiceの全バリデーションパスをカバー
	cfg := &Config{}

	tests := []struct {
		name    string
		svc     *KubernetesService
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing host",
			svc:     &KubernetesService{Namespace: "test", Service: "svc", Protocol: "http"},
			wantErr: true,
			errMsg:  "host is required",
		},
		{
			name:    "missing service",
			svc:     &KubernetesService{Host: "test.localhost", Namespace: "test", Protocol: "http"},
			wantErr: true,
			errMsg:  "service is required",
		},
		{
			name:    "invalid protocol",
			svc:     &KubernetesService{Host: "test.localhost", Namespace: "test", Service: "svc", Protocol: "invalid"},
			wantErr: true,
			errMsg:  "protocol must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.svc.Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !containsString(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestTCPService_ValidateEdgeCases(t *testing.T) {
	// TCPServiceの全バリデーションパスをカバー
	cfg := &Config{
		SSHBastions: map[string]*SSHBastion{
			"primary": {Instance: "test", Zone: "zone", Project: "proj"},
		},
	}

	tests := []struct {
		name    string
		svc     *TCPService
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing ssh_bastion",
			svc:     &TCPService{Host: "db.localhost", TargetHost: "10.0.0.1", TargetPort: 5432},
			wantErr: true,
			errMsg:  "ssh_bastion is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.svc.Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !containsString(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestLoad_InvalidFile(t *testing.T) {
	// 存在しないファイル
	_, err := Load("/nonexistent/path/to/config.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestLoad_NoServices(t *testing.T) {
	// servicesが空の場合
	content := `
listener_port: 80
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for no services, got nil")
	}
	if !containsString(err.Error(), "no services configured") {
		t.Errorf("expected error containing 'no services configured', got '%s'", err.Error())
	}
}
