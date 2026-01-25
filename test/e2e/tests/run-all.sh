#!/bin/bash
set -euo pipefail

LOCALMESH_HOST="${LOCALMESH_HOST:-localmesh}"
LOCALMESH_PORT="${LOCALMESH_PORT:-18080}"

echo "=== E2E Test Suite ==="
echo "Target: ${LOCALMESH_HOST}:${LOCALMESH_PORT}"
echo ""

# HTTP テスト
echo "--- Test: HTTP Routing ---"
response=$(curl -sf -H "Host: http-test.localdomain" "http://${LOCALMESH_HOST}:${LOCALMESH_PORT}/")
if echo "$response" | grep -q "Hello from E2E test"; then
    echo "PASSED: HTTP routing"
else
    echo "FAILED: HTTP routing"
    echo "Response: $response"
    exit 1
fi

echo ""
echo "=== All tests passed ==="
