---
description: kubectl-localmeshにおけるログレベル設計とユーザーフレンドリーな出力のガイドライン
---

# kubectl-localmesh ログ設計ガイド

このskillはkubectl-localmeshにおけるログ出力のベストプラクティスを提供します。

## ログレベル階層

| レベル | 用途 | 出力内容 |
|--------|------|----------|
| `warn` | 静かモード | エラーのみ |
| `info` | デフォルト | 接続サマリー + 基本状態 |
| `debug` | 詳細調査 | 再接続ログ + Envoy詳細 + port-forward状態 |

## 使用方法

```bash
# デフォルト（info）: サマリー表示
sudo kubectl-localmesh up -f services.yaml

# debug: 詳細ログ（再接続、port-forward状態）
sudo kubectl-localmesh up -f services.yaml --log-level debug

# warn: 最小出力（スクリプト実行向け）
sudo kubectl-localmesh up -f services.yaml --log-level warn
```

## 起動サマリーの見方

起動完了時に以下のサマリーが表示されます：

```
Service Mesh is ready!

Access your services:
  HTTP/gRPC Services:
  • http://users-api.localhost:80 (gRPC) -> users/users-api:50051
  • http://admin.localhost:80 (HTTP) -> admin/admin-web:8080

  TCP Services:
  • tcp://users-db.localhost:5432 -> primary @ 10.0.0.1:5432

Press Ctrl+C to stop and cleanup.
```

**各フィールドの意味**:
- `http://users-api.localhost:80`: ローカルからアクセスするURL
- `(gRPC)`: プロトコルタイプ
- `users/users-api:50051`: 転送先（namespace/service:port）
- `primary @ 10.0.0.1:5432`: TCPの場合はSSH bastion名とターゲットIP

## 実装パターン

### Loggerの使用

```go
import "github.com/usadamasa/kubectl-localmesh/internal/log"

// Loggerの初期化
logger := log.New(logLevel) // "warn", "info", "debug"

// ログ出力
logger.Info("Service Mesh is ready!")
logger.Infof("Listening on port %d", port)
logger.Debug("port-forward connection established")
logger.Debugf("Reconnecting to %s/%s", namespace, service)
```

### ログレベル判定

```go
// 条件付き処理
if logger.ShouldLogDebug() {
    cmd.Stdout = os.Stdout  // debug時のみ詳細出力
}
```

### Envoyへのログレベル連携

```go
envoyCmd := exec.CommandContext(ctx, "envoy",
    "-c", envoyPath,
    "-l", logger.Level(),  // loggerのレベルをEnvoyに渡す
)
```

## トラブルシューティング

### 問題: 再接続が頻発している

```bash
# debugレベルで再接続ログを確認
sudo kubectl-localmesh up -f services.yaml --log-level debug
```

出力例:
```
[DEBUG] port-forward disconnected: users/users-api -> pod/users-api-xxx (reconnecting...)
[DEBUG] port-forward ready: users/users-api -> pod/users-api-yyy (127.0.0.1:50051 -> 50051)
```

### 問題: Envoyの詳細ログが必要

```bash
# debugレベルでEnvoyのログも詳細化される
sudo kubectl-localmesh up -f services.yaml --log-level debug
```

### 問題: スクリプトでの使用時に出力を抑制したい

```bash
# warnレベルでエラーのみ出力
sudo kubectl-localmesh up -f services.yaml --log-level warn 2>&1
```

## Lintルール

このプロジェクトでは`forbidigo`リンターを使用して、意図しない標準出力を検出します。

### 禁止パターン

- `fmt.Print*`: `internal/log.Logger`を使用すること
- `fmt.Fprint*(os.Stdout, ...)`: `internal/log.Logger`を使用すること

### 例外

- `_test.go`: テストファイル
- `main.go`: CLIエントリポイント
- `cmd/`: CLI初期化
- `internal/log/`: ログ実装自体
- `internal/dump/`: 設定ダンプ出力（`//nolint:forbidigo`付き）

### nolintコメントの使用

意図的に標準出力を使用する場合：

```go
fmt.Print(output) //nolint:forbidigo // CLIダンプ出力として意図的に使用
```
