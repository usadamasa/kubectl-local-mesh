package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func resetValidateOpts() {
	validateOpts.configFile = ""
	validateOpts.strict = false
}

func TestValidateCmd_ValidConfig(t *testing.T) {
	resetValidateOpts()
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

	cmd := rootCmd
	cmd.SetArgs([]string{"validate", "-f", configPath})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateCmd_ValidConfigWithPositionalArg(t *testing.T) {
	resetValidateOpts()
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

	cmd := rootCmd
	cmd.SetArgs([]string{"validate", configPath})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateCmd_MissingConfigFile(t *testing.T) {
	resetValidateOpts()
	cmd := rootCmd
	cmd.SetArgs([]string{"validate"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestValidateCmd_NonexistentFile(t *testing.T) {
	resetValidateOpts()
	cmd := rootCmd
	cmd.SetArgs([]string{"validate", "-f", "/nonexistent/path.yaml"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestValidateCmd_InvalidConfig(t *testing.T) {
	resetValidateOpts()
	// kindフィールドが欠落
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

	cmd := rootCmd
	cmd.SetArgs([]string{"validate", "-f", configPath})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid config (missing kind)")
	}
}

func TestValidateCmd_InvalidKindValue(t *testing.T) {
	resetValidateOpts()
	content := `
listener_port: 80
services:
  - kind: invalid
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

	cmd := rootCmd
	cmd.SetArgs([]string{"validate", "-f", configPath})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid kind value")
	}
}

func TestValidateCmd_StrictFlag_ValidConfig(t *testing.T) {
	resetValidateOpts()
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

	cmd := rootCmd
	cmd.SetArgs([]string{"validate", "-f", configPath, "--strict"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected no error with --strict flag, got: %v", err)
	}
}

func TestValidateCmd_StrictFlag_InvalidConfig(t *testing.T) {
	resetValidateOpts()
	// additionalPropertiesに引っかかるフィールド
	// config.Load()はunknown fieldをエラーにしないので、--strictフラグで検出される
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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := rootCmd
	cmd.SetArgs([]string{"validate", "-f", configPath, "--strict"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error with --strict flag for config with unknown field")
	}
}

func TestValidateCmd_NoServices(t *testing.T) {
	resetValidateOpts()
	content := `
listener_port: 80
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := rootCmd
	cmd.SetArgs([]string{"validate", "-f", configPath})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for config with no services")
	}
}
