package log

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		wantInfo bool
		wantDbg  bool
	}{
		{"warn level", "warn", false, false},
		{"info level", "info", true, false},
		{"debug level", "debug", true, true},
		{"unknown defaults to info", "unknown", true, false},
		{"empty defaults to info", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level)
			if got := logger.ShouldLogInfo(); got != tt.wantInfo {
				t.Errorf("ShouldLogInfo() = %v, want %v", got, tt.wantInfo)
			}
			if got := logger.ShouldLogDebug(); got != tt.wantDbg {
				t.Errorf("ShouldLogDebug() = %v, want %v", got, tt.wantDbg)
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	tests := []struct {
		name       string
		level      string
		msg        string
		wantOutput bool
	}{
		{"info level logs info", "info", "test message", true},
		{"debug level logs info", "debug", "test message", true},
		{"warn level does not log info", "warn", "test message", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewWithWriter(tt.level, &buf)
			logger.Info(tt.msg)

			gotOutput := buf.Len() > 0
			if gotOutput != tt.wantOutput {
				t.Errorf("Info() output = %v, want %v (buf: %q)", gotOutput, tt.wantOutput, buf.String())
			}
			if tt.wantOutput && buf.String() != tt.msg+"\n" {
				t.Errorf("Info() = %q, want %q", buf.String(), tt.msg+"\n")
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	tests := []struct {
		name       string
		level      string
		msg        string
		wantOutput bool
	}{
		{"debug level logs debug", "debug", "debug message", true},
		{"info level does not log debug", "info", "debug message", false},
		{"warn level does not log debug", "warn", "debug message", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewWithWriter(tt.level, &buf)
			logger.Debug(tt.msg)

			gotOutput := buf.Len() > 0
			if gotOutput != tt.wantOutput {
				t.Errorf("Debug() output = %v, want %v (buf: %q)", gotOutput, tt.wantOutput, buf.String())
			}
			if tt.wantOutput && buf.String() != "[DEBUG] "+tt.msg+"\n" {
				t.Errorf("Debug() = %q, want %q", buf.String(), "[DEBUG] "+tt.msg+"\n")
			}
		})
	}
}

func TestLogger_Infof(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter("info", &buf)
	logger.Infof("hello %s %d", "world", 42)

	want := "hello world 42\n"
	if buf.String() != want {
		t.Errorf("Infof() = %q, want %q", buf.String(), want)
	}
}

func TestLogger_Debugf(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter("debug", &buf)
	logger.Debugf("hello %s %d", "world", 42)

	want := "[DEBUG] hello world 42\n"
	if buf.String() != want {
		t.Errorf("Debugf() = %q, want %q", buf.String(), want)
	}
}

func TestLogger_Level(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantLevel string
	}{
		{"warn", "warn", "warn"},
		{"info", "info", "info"},
		{"debug", "debug", "debug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level)
			if got := logger.Level(); got != tt.wantLevel {
				t.Errorf("Level() = %v, want %v", got, tt.wantLevel)
			}
		})
	}
}

func TestLogger_EnvoyLevel(t *testing.T) {
	tests := []struct {
		name           string
		level          string
		wantEnvoyLevel string
	}{
		{"warn returns warn for envoy", "warn", "warn"},
		{"info returns warn for envoy", "info", "warn"},
		{"debug returns debug for envoy", "debug", "debug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level)
			if got := logger.EnvoyLevel(); got != tt.wantEnvoyLevel {
				t.Errorf("EnvoyLevel() = %v, want %v", got, tt.wantEnvoyLevel)
			}
		})
	}
}
