#!/bin/bash
set -euo pipefail

LOCALMESH_HOST="${LOCALMESH_HOST:-localmesh}"
LOCALMESH_PORT="${LOCALMESH_PORT:-18080}"

echo "Testing HTTP routing..."

response=$(curl -sf -H "Host: http-test.localdomain" "http://${LOCALMESH_HOST}:${LOCALMESH_PORT}/")

if echo "$response" | grep -q "Hello from E2E test"; then
    echo "PASSED: HTTP routing test"
    exit 0
else
    echo "FAILED: HTTP routing test"
    echo "Response: $response"
    exit 1
fi
