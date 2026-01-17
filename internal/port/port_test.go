package port

import (
	"bytes"
	"strings"
	"testing"
)

func TestPort_IsValid(t *testing.T) {
	tests := []struct {
		name string
		port Port
		want bool
	}{
		{"zero is invalid", 0, false},
		{"one is valid (minimum)", 1, true},
		{"1023 is valid", 1023, true},
		{"1024 is valid", 1024, true},
		{"65535 is valid (maximum)", 65535, true},
		{"65536 is invalid", 65536, false},
		{"negative is invalid", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.port.IsValid(); got != tt.want {
				t.Errorf("Port(%d).IsValid() = %v, want %v", tt.port, got, tt.want)
			}
		})
	}
}

func TestPort_IsPrivileged(t *testing.T) {
	tests := []struct {
		name string
		port Port
		want bool
	}{
		{"zero is privileged", 0, true},
		{"one is privileged", 1, true},
		{"80 is privileged", 80, true},
		{"443 is privileged", 443, true},
		{"1023 is privileged (max)", 1023, true},
		{"1024 is not privileged", 1024, false},
		{"8080 is not privileged", 8080, false},
		{"65535 is not privileged", 65535, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.port.IsPrivileged(); got != tt.want {
				t.Errorf("Port(%d).IsPrivileged() = %v, want %v", tt.port, got, tt.want)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		fieldName   string
		serviceName string
		wantErr     bool
		errContains string
	}{
		{"valid port", 8080, "port", "my-service", false, ""},
		{"minimum valid port", 1, "port", "my-service", false, ""},
		{"maximum valid port", 65535, "port", "my-service", false, ""},
		{"zero is invalid", 0, "port", "my-service", true, "must be between 1 and 65535"},
		{"negative is invalid", -1, "port", "my-service", true, "must be between 1 and 65535"},
		{"too high is invalid", 65536, "port", "my-service", true, "must be between 1 and 65535"},
		{"error includes field name", 0, "listener_port", "config", true, "listener_port"},
		{"error includes service name", 0, "port", "users-api", true, "users-api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port, tt.fieldName, tt.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidatePort() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

func TestValidateRequiredPort(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		fieldName   string
		serviceName string
		wantErr     bool
		errContains string
	}{
		{"valid port", 8080, "port", "my-service", false, ""},
		{"zero is required error", 0, "target_port", "my-db", true, "target_port is required"},
		{"out of range error", 65536, "port", "my-service", true, "must be between 1 and 65535"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredPort(tt.port, tt.fieldName, tt.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateRequiredPort() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

func TestValidatePorts(t *testing.T) {
	tests := []struct {
		name    string
		ports   []int
		wantErr bool
	}{
		{"empty slice is valid", []int{}, false},
		{"single valid port", []int{8080}, false},
		{"multiple valid ports", []int{80, 443, 8080}, false},
		{"contains invalid port", []int{80, 0, 8080}, true},
		{"contains out of range", []int{80, 65536, 8080}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePorts(tt.ports, "ports", "my-service")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePorts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePorts_GenericType(t *testing.T) {
	ports := []IndividualListenerPort{80, 443, 8080}
	err := ValidatePorts(ports, "overwrite_listen_ports", "my-service")
	if err != nil {
		t.Errorf("ValidatePorts() with IndividualListenerPort should succeed, got error: %v", err)
	}

	invalidPorts := []IndividualListenerPort{80, 0, 8080}
	err = ValidatePorts(invalidPorts, "overwrite_listen_ports", "my-service")
	if err == nil {
		t.Error("ValidatePorts() with invalid IndividualListenerPort should fail")
	}
}

func TestWarnPrivilegedPort(t *testing.T) {
	tests := []struct {
		name           string
		port           int
		fieldName      string
		serviceName    string
		expectsWarning bool
	}{
		{"privileged port 80", 80, "listener_port", "config", true},
		{"privileged port 443", 443, "listener_port", "config", true},
		{"privileged port 1", 1, "listener_port", "config", true},
		{"privileged port 1023", 1023, "listener_port", "config", true},
		{"non-privileged port 1024", 1024, "listener_port", "config", false},
		{"non-privileged port 8080", 8080, "listener_port", "config", false},
		{"zero port no warning", 0, "listener_port", "config", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			origWriter := warnWriter
			SetWarnWriter(&buf)
			defer SetWarnWriter(origWriter)

			WarnPrivilegedPort(tt.port, tt.fieldName, tt.serviceName)

			hasOutput := buf.Len() > 0
			if hasOutput != tt.expectsWarning {
				t.Errorf("WarnPrivilegedPort(%d) hasOutput = %v, expectsWarning = %v, output = %q",
					tt.port, hasOutput, tt.expectsWarning, buf.String())
			}
			if tt.expectsWarning {
				if !strings.Contains(buf.String(), "privileged port") {
					t.Errorf("Warning should contain 'privileged port', got: %s", buf.String())
				}
				if !strings.Contains(buf.String(), tt.fieldName) {
					t.Errorf("Warning should contain field name %q, got: %s", tt.fieldName, buf.String())
				}
			}
		})
	}
}

func TestPortConflictChecker(t *testing.T) {
	t.Run("no conflict", func(t *testing.T) {
		var buf bytes.Buffer
		origWriter := warnWriter
		SetWarnWriter(&buf)
		defer SetWarnWriter(origWriter)

		checker := NewPortConflictChecker()
		checker.Register(80, "service-a")
		checker.Register(8080, "service-b")

		if buf.Len() > 0 {
			t.Errorf("Expected no warning, got: %s", buf.String())
		}
	})

	t.Run("with conflict", func(t *testing.T) {
		var buf bytes.Buffer
		origWriter := warnWriter
		SetWarnWriter(&buf)
		defer SetWarnWriter(origWriter)

		checker := NewPortConflictChecker()
		checker.Register(80, "service-a")
		checker.Register(80, "service-b")

		if buf.Len() == 0 {
			t.Error("Expected warning for port conflict")
		}
		if !strings.Contains(buf.String(), "service-a") {
			t.Errorf("Warning should contain 'service-a', got: %s", buf.String())
		}
		if !strings.Contains(buf.String(), "service-b") {
			t.Errorf("Warning should contain 'service-b', got: %s", buf.String())
		}
		if !strings.Contains(buf.String(), "80") {
			t.Errorf("Warning should contain port '80', got: %s", buf.String())
		}
	})

	t.Run("multiple conflicts", func(t *testing.T) {
		var buf bytes.Buffer
		origWriter := warnWriter
		SetWarnWriter(&buf)
		defer SetWarnWriter(origWriter)

		checker := NewPortConflictChecker()
		checker.Register(80, "service-a")
		checker.Register(443, "service-b")
		checker.Register(80, "service-c")
		checker.Register(443, "service-d")

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("Expected 2 warning lines, got %d: %s", len(lines), buf.String())
		}
	})
}

func TestFreeLocalPort(t *testing.T) {
	port, err := FreeLocalPort()
	if err != nil {
		t.Fatalf("FreeLocalPort() error = %v", err)
	}

	if !port.IsValid() {
		t.Errorf("FreeLocalPort() returned invalid port: %d", port)
	}

	if port.IsPrivileged() {
		t.Errorf("FreeLocalPort() returned privileged port: %d", port)
	}

	anotherPort, err := FreeLocalPort()
	if err != nil {
		t.Fatalf("FreeLocalPort() second call error = %v", err)
	}

	if port == anotherPort {
		t.Logf("Note: FreeLocalPort() returned same port twice (race condition, but possible)")
	}
}
