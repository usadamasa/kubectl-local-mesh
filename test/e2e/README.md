# E2E Tests

kubectl-localmesh の E2E テスト環境です。
docker-compose と k3s コンテナを使用して、完全コンテナ化された環境でテストを実行します。

## 前提条件

- Docker
- docker-compose
- Task (Taskfile.dev)

## クイックスタート

```bash
# 1. バイナリビルド
task build

# 2. E2Eテスト実行
task test:e2e
```

## デバッグ用コマンド

```bash
# 環境を起動したまま維持
task test:e2e:up

# ログを確認
task test:e2e:logs

# 環境を停止
task test:e2e:down
```

## アーキテクチャ

```
┌─────────────────────────────────────────────────────────────┐
│  docker-compose network: e2e-network                        │
│                                                             │
│  ┌─────────────┐    ┌─────────────────┐    ┌─────────────┐ │
│  │ test-client │───▶│   localmesh     │───▶│    k3s      │ │
│  │ (curl/grpc) │    │ (envoy+kubectl  │    │  (k3s API)  │ │
│  │             │    │  +localmesh)    │    │             │ │
│  └─────────────┘    └─────────────────┘    └─────────────┘ │
│        ↓                    ↓                    ↓         │
│   HTTPリクエスト      port-forward         K8s Services    │
│   Host: xxx.local     via WebSocket        (HTTP/gRPC)     │
└─────────────────────────────────────────────────────────────┘
```

## コンポーネント

| サービス | 説明 |
|---------|------|
| k3s | 軽量 Kubernetes クラスタ (rancher/k3s) |
| k8s-setup | テスト用サービスをデプロイ |
| localmesh | kubectl-localmesh + Envoy |
| test-client | テストスクリプト実行 (curl + grpcurl) |

## テストケース

### HTTP ルーティングテスト

`http-test.localdomain` へのリクエストが K8s サービスに正しくルーティングされることを確認します。

## ディレクトリ構成

```
test/e2e/
├── compose.yaml              # docker-compose 設定
├── Dockerfile.localmesh      # localmesh コンテナ
├── Dockerfile.test-client    # テストクライアント
├── fixtures/
│   ├── k8s/                  # K8s マニフェスト
│   └── configs/              # localmesh 設定
├── tests/                    # テストスクリプト
└── output/                   # k3s kubeconfig 出力先 (git 無視)
```

## トラブルシューティング

### k3s が起動しない

```bash
# k3s コンテナのログを確認
docker compose logs k3s
```

### localmesh が起動しない

```bash
# localmesh コンテナのログを確認
docker compose logs localmesh

# kubeconfig が正しく生成されているか確認
cat test/e2e/output/kubeconfig.yaml
```
