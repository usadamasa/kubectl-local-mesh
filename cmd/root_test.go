package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/usadamasa/kubectl-localmesh/internal/version"
)

func TestRootCommand(t *testing.T) {
	t.Run("Execute returns no error", func(t *testing.T) {
		err := Execute()
		if err != nil {
			t.Errorf("Execute() returned error: %v", err)
		}
	})
}

func TestSetVersion(t *testing.T) {
	tests := []struct {
		name            string
		info            version.Info
		wantVersion     string
		wantInTemplate  string
		wantNotTemplate string
	}{
		{
			name:            "全フィールド設定",
			info:            version.Info{Version: "v1.2.3", Commit: "abc1234", Date: "2024-01-01T00:00:00Z"},
			wantVersion:     "v1.2.3",
			wantInTemplate:  "v1.2.3 (commit: abc1234, built: 2024-01-01T00:00:00Z)",
			wantNotTemplate: "",
		},
		{
			name:            "デフォルト値で括弧なし",
			info:            version.Info{Version: "dev", Commit: "none", Date: "unknown"},
			wantVersion:     "dev",
			wantInTemplate:  "dev",
			wantNotTemplate: "(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersion(tt.info)

			if rootCmd.Version != tt.wantVersion {
				t.Errorf("rootCmd.Version = %q, want %q", rootCmd.Version, tt.wantVersion)
			}

			// VersionTemplateの内容を検証
			tmpl := rootCmd.VersionTemplate()
			if !strings.Contains(tmpl, tt.wantInTemplate) {
				t.Errorf("VersionTemplate() = %q, want to contain %q", tmpl, tt.wantInTemplate)
			}
			if tt.wantNotTemplate != "" && strings.Contains(tmpl, tt.wantNotTemplate) {
				t.Errorf("VersionTemplate() = %q, should not contain %q", tmpl, tt.wantNotTemplate)
			}
		})
	}
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantLogLevel string
	}{
		{
			name:         "デフォルトログレベル",
			args:         []string{},
			wantLogLevel: "info",
		},
		{
			name:         "グローバルフラグでdebug指定",
			args:         []string{"--log-level", "debug"},
			wantLogLevel: "debug",
		},
		{
			name:         "グローバルフラグでwarn指定",
			args:         []string{"--log-level", "warn"},
			wantLogLevel: "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalLogLevel = "info"

			cmd := &cobra.Command{Use: "test"}
			cmd.PersistentFlags().StringVar(&globalLogLevel, "log-level", "info", "log level")
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("Execute() error = %v", err)
			}

			if globalLogLevel != tt.wantLogLevel {
				t.Errorf("globalLogLevel = %v, want %v", globalLogLevel, tt.wantLogLevel)
			}
		})
	}
}
