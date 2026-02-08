package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

// multiClusterKubeconfig は複数クラスタを含むテスト用kubeconfig
const multiClusterKubeconfig = `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
    insecure-skip-tls-verify: true
  name: test-cluster
- cluster:
    server: https://10.0.0.1:6443
    insecure-skip-tls-verify: true
  name: staging-cluster
- cluster:
    server: https://10.0.0.2:6443
    insecure-skip-tls-verify: true
  name: prod-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

func TestNewClient_ValidKubeconfig(t *testing.T) {
	// 一時ディレクトリ作成（テスト終了時に自動削除）
	tmpDir := t.TempDir()

	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
	err := os.WriteFile(kubeconfigPath, []byte(multiClusterKubeconfig), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// 環境変数を設定（テスト終了時に自動復元）
	t.Setenv("KUBECONFIG", kubeconfigPath)

	// テスト実行（空文字列 = current-contextのclusterを使用）
	clientset, restConfig, err := NewClient("")

	// アサーション
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if clientset == nil {
		t.Fatal("expected clientset to be non-nil")
	}
	if restConfig == nil {
		t.Fatal("expected restConfig to be non-nil")
	}

	// restConfigの基本的な検証（current-contextはtest-cluster → 127.0.0.1:6443）
	if restConfig.Host != "https://127.0.0.1:6443" {
		t.Errorf("expected host 'https://127.0.0.1:6443', got %q", restConfig.Host)
	}
}

func TestNewClient_WithCluster(t *testing.T) {
	tmpDir := t.TempDir()

	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
	if err := os.WriteFile(kubeconfigPath, []byte(multiClusterKubeconfig), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("KUBECONFIG", kubeconfigPath)

	// staging-clusterを明示的に指定
	clientset, restConfig, err := NewClient("staging-cluster")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if clientset == nil {
		t.Fatal("expected clientset to be non-nil")
	}
	if restConfig == nil {
		t.Fatal("expected restConfig to be non-nil")
	}

	// staging-clusterのサーバーに接続すること
	if restConfig.Host != "https://10.0.0.1:6443" {
		t.Errorf("expected host 'https://10.0.0.1:6443', got %q", restConfig.Host)
	}
}

func TestNewClient_EmptyCluster(t *testing.T) {
	tmpDir := t.TempDir()

	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
	if err := os.WriteFile(kubeconfigPath, []byte(multiClusterKubeconfig), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("KUBECONFIG", kubeconfigPath)

	// 空文字列 = current-contextのcluster（test-cluster）を使用
	_, restConfig, err := NewClient("")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if restConfig.Host != "https://127.0.0.1:6443" {
		t.Errorf("expected host 'https://127.0.0.1:6443' (current-context cluster), got %q", restConfig.Host)
	}
}

func TestNewClient_InvalidKubeconfig(t *testing.T) {
	tmpDir := t.TempDir()

	// 無効なYAMLファイル
	invalidKubeconfigPath := filepath.Join(tmpDir, "invalid-kubeconfig")
	err := os.WriteFile(invalidKubeconfigPath, []byte("invalid: yaml: content:"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("KUBECONFIG", invalidKubeconfigPath)

	_, _, err = NewClient("")
	if err == nil {
		t.Fatal("expected error for invalid kubeconfig, got nil")
	}
}

func TestNewClient_NoKubeconfig(t *testing.T) {
	tmpDir := t.TempDir()

	// 存在しないパスを設定
	nonExistentPath := filepath.Join(tmpDir, "nonexistent-kubeconfig")
	t.Setenv("KUBECONFIG", nonExistentPath)
	t.Setenv("HOME", tmpDir) // ~/.kube/configも存在しないようにする

	_, _, err := NewClient("")
	if err == nil {
		t.Fatal("expected error for non-existent kubeconfig, got nil")
	}
}
