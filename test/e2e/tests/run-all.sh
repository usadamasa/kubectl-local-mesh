#!/bin/bash
set -euo pipefail

LOCALMESH_HOST="${LOCALMESH_HOST:-localmesh}"
LOCALMESH_PORT="${LOCALMESH_PORT:-18080}"

echo "=== E2E Test Suite ==="
echo "Target: ${LOCALMESH_HOST}:${LOCALMESH_PORT}"
echo ""

# HTTP テスト
echo "--- Test: HTTP Routing ---"

# Store both response and exit code
http_code=$(curl -s -o /tmp/response.txt -w "%{http_code}" -H "Host: http-test.localdomain" "http://${LOCALMESH_HOST}:${LOCALMESH_PORT}/" || true)
response=$(cat /tmp/response.txt 2>/dev/null || echo "")

if [ "$http_code" != "200" ]; then
    echo "FAILED: HTTP routing (HTTP status: ${http_code})"
    echo "Response body: ${response}"
    exit 1
fi

if echo "$response" | grep -q "Hello from E2E test"; then
    echo "PASSED: HTTP routing"
else
    echo "FAILED: HTTP routing (unexpected response)"
    echo "HTTP status: ${http_code}"
    echo "Response: ${response}"
    exit 1
fi

echo ""

# gRPC テスト
echo "--- Test: gRPC Routing ---"

# grpcurlで -authority フラグを使用（HTTP Host headerと同等）
# || true で exit code をキャプチャ（既存HTTPテストと同じパターン）
grpc_response=$(grpcurl -plaintext \
  -authority grpc-test.localdomain \
  -d '{"name":"e2e-test"}' \
  "${LOCALMESH_HOST}:${LOCALMESH_PORT}" \
  helloworld.Greeter/SayHello 2>&1 || true)

# grpcurlはエラー時にstderrに出力するので、"ERROR"を含むかチェック
if echo "$grpc_response" | grep -qi "error\|failed\|refused"; then
    echo "FAILED: gRPC routing"
    echo "Response: ${grpc_response}"
    exit 1
fi

# レスポンスに "Hello e2e-test" が含まれることを確認
if echo "$grpc_response" | grep -q "Hello e2e-test"; then
    echo "PASSED: gRPC routing"
else
    echo "FAILED: gRPC routing (unexpected response)"
    echo "Response: ${grpc_response}"
    exit 1
fi

echo ""
echo "=== All tests passed ==="
