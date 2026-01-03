package gcp

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/usadamasa/kubectl-localmesh/internal/config"
)

func TestStartGCPSSHTunnel_BasicFlow(t *testing.T) {
	// 基本的なSSH tunnel起動のテスト
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	bastion := &config.SSHBastion{
		Instance: "test-instance",
		Zone:     "asia-northeast1-a",
		Project:  "test-project",
	}

	localPort := 10000
	targetHost := "10.0.0.1"
	targetPort := 5432

	// テスト用のモックを使用する予定
	// 現時点では実装がないため、関数が存在することだけを確認
	err := StartGCPSSHTunnel(ctx, bastion, localPort, targetHost, targetPort, "info")

	// contextがキャンセルされた場合はnilが返ることを期待
	if err != nil && ctx.Err() == nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStartGCPSSHTunnel_InvalidBastion(t *testing.T) {
	// Bastionパラメータのバリデーション
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	bastion := &config.SSHBastion{
		Instance: "", // 空のinstance名
		Zone:     "asia-northeast1-a",
		Project:  "test-project",
	}

	err := StartGCPSSHTunnel(ctx, bastion, 10000, "10.0.0.1", 5432, "info")

	if err == nil {
		t.Error("expected error for empty instance name, got nil")
	}
}

func TestStartGCPSSHTunnel_InvalidPorts(t *testing.T) {
	// ポート番号のバリデーション
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	bastion := &config.SSHBastion{
		Instance: "test-instance",
		Zone:     "asia-northeast1-a",
		Project:  "test-project",
	}

	// localPort が0
	err := StartGCPSSHTunnel(ctx, bastion, 0, "10.0.0.1", 5432, "info")
	if err == nil {
		t.Error("expected error for invalid local port, got nil")
	}

	// targetPort が0
	err = StartGCPSSHTunnel(ctx, bastion, 10000, "10.0.0.1", 0, "info")
	if err == nil {
		t.Error("expected error for invalid target port, got nil")
	}
}

func TestStartGCPSSHTunnel_InvalidTargetHost(t *testing.T) {
	// ターゲットホストのバリデーション
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	bastion := &config.SSHBastion{
		Instance: "test-instance",
		Zone:     "asia-northeast1-a",
		Project:  "test-project",
	}

	err := StartGCPSSHTunnel(ctx, bastion, 10000, "", 5432, "info")
	if err == nil {
		t.Error("expected error for empty target host, got nil")
	}
}

func TestBuildGcloudSSHCommand(t *testing.T) {
	// コマンド引数構築のテスト
	tests := []struct {
		name       string
		bastion    *config.SSHBastion
		localPort  int
		targetHost string
		targetPort int
		want       []string
	}{
		{
			name: "basic case",
			bastion: &config.SSHBastion{
				Instance: "bastion-1",
				Zone:     "us-central1-a",
				Project:  "my-project",
			},
			localPort:  10000,
			targetHost: "10.0.0.1",
			targetPort: 5432,
			want: []string{
				"compute", "ssh", "bastion-1",
				"--project=my-project",
				"--zone=us-central1-a",
				"--",
				"-L", "10000:10.0.0.1:5432",
				"-N",
				"-o", "ExitOnForwardFailure=yes",
				"-o", "ServerAliveInterval=30",
				"-o", "ServerAliveCountMax=3",
			},
		},
		{
			name: "different zone and port",
			bastion: &config.SSHBastion{
				Instance: "bastion-2",
				Zone:     "asia-northeast1-a",
				Project:  "test-project",
			},
			localPort:  20000,
			targetHost: "192.168.1.1",
			targetPort: 3306,
			want: []string{
				"compute", "ssh", "bastion-2",
				"--project=test-project",
				"--zone=asia-northeast1-a",
				"--",
				"-L", "20000:192.168.1.1:3306",
				"-N",
				"-o", "ExitOnForwardFailure=yes",
				"-o", "ServerAliveInterval=30",
				"-o", "ServerAliveCountMax=3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildGcloudSSHCommand(tt.bastion, tt.localPort, tt.targetHost, tt.targetPort)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildGcloudSSHCommand() =\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}
