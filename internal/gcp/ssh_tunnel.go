package gcp

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/usadamasa/kubectl-localmesh/internal/config"
	"github.com/usadamasa/kubectl-localmesh/internal/log"
	"github.com/usadamasa/kubectl-localmesh/internal/port"
)

// StartGCPSSHTunnel はGCP Compute Instance経由でSSH tunnelを確立し、
// ローカルポートからターゲットホスト:ポートへのポートフォワーディングを行います。
// contextがキャンセルされるまで自動再接続を繰り返します。
func StartGCPSSHTunnel(
	ctx context.Context,
	bastion *config.SSHBastion,
	localPort port.LocalPort,
	targetHost string,
	targetPort port.TCPPort,
	logger *log.Logger,
) error {
	// パラメータのバリデーション
	if bastion == nil {
		return fmt.Errorf("bastion is nil")
	}
	if bastion.Instance == "" {
		return fmt.Errorf("bastion instance name is empty")
	}
	if bastion.Zone == "" {
		return fmt.Errorf("bastion zone is empty")
	}
	if !port.IsValid(localPort) {
		return fmt.Errorf("invalid local port: %d", localPort)
	}
	if targetHost == "" {
		return fmt.Errorf("target host is empty")
	}
	if !port.IsValid(targetPort) {
		return fmt.Errorf("invalid target port: %d", targetPort)
	}

	// 自動再接続ループ（k8s port-forwardと同様）
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// SSH tunnel確立を試行
		err := startSingleSSHTunnel(ctx, bastion, localPort, targetHost, targetPort, logger)

		// contextキャンセル時は正常終了
		if ctx.Err() != nil {
			return nil
		}

		// エラーは自動再接続で処理されるため、ここではdebugログ出力のみ
		if err != nil {
			logger.Debugf("SSH tunnel disconnected: %s -> %s:%d (reconnecting...): %v",
				bastion.Instance, targetHost, int(targetPort), err)
		}

		// 300ms待機後に再接続
		time.Sleep(300 * time.Millisecond)
	}
}

// buildGcloudSSHCommand はgcloud compute sshコマンドの引数を構築します。
// テスト可能にするため、package private関数として定義しています。
func buildGcloudSSHCommand(
	bastion *config.SSHBastion,
	localPort port.LocalPort,
	targetHost string,
	targetPort port.TCPPort,
) []string {
	return []string{
		"compute", "ssh",
		bastion.Instance,
		fmt.Sprintf("--project=%s", bastion.Project),
		fmt.Sprintf("--zone=%s", bastion.Zone),
		"--tunnel-through-iap", // 明示的にIAPを使用（警告抑止）
		"--",
		"-L", fmt.Sprintf("%d:%s:%d", int(localPort), targetHost, int(targetPort)),
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=30",
		"-o", "ServerAliveCountMax=3",
	}
}

// startSingleSSHTunnel は1回のSSH tunnel接続を試行します。
// 接続が切断されるか、エラーが発生した場合に返ります。
func startSingleSSHTunnel(
	ctx context.Context,
	bastion *config.SSHBastion,
	localPort port.LocalPort,
	targetHost string,
	targetPort port.TCPPort,
	logger *log.Logger,
) error {
	// 1. gcloudコマンドのパスを取得
	gcloudPath, err := exec.LookPath("gcloud")
	if err != nil {
		return fmt.Errorf("gcloud command not found: %w (install: https://cloud.google.com/sdk/docs/install)", err)
	}

	// 2. コマンド引数を構築
	args := buildGcloudSSHCommand(bastion, localPort, targetHost, targetPort)

	// 3. コマンド実行
	cmd := exec.CommandContext(ctx, gcloudPath, args...)

	// 4. 標準出力/エラー出力の処理（logLevelで制御）
	if logger.ShouldLogDebug() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = os.Stderr
	}

	// 5. 実行（ブロッキング、contextキャンセル時に自動終了）
	return cmd.Run()
}
