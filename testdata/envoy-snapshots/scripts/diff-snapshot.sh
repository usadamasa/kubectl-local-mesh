#!/usr/bin/env bash
set -euo pipefail

TEST_CASE="$1"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNAPSHOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PROJECT_ROOT="$(cd "$SNAPSHOT_DIR/../.." && pwd)"

TESTDATA_DIR="$SNAPSHOT_DIR/testdata"
CONFIG="$TESTDATA_DIR/configs/${TEST_CASE}.yaml"
MOCK="$TESTDATA_DIR/mocks/${TEST_CASE}-mocks.yaml"
SNAPSHOT="$TESTDATA_DIR/snapshots/${TEST_CASE}.yaml"

KUBECTL_LOCALMESH="$PROJECT_ROOT/bin/kubectl-localmesh"

# 一時ファイルに出力
ACTUAL=$(mktemp)
trap "rm -f $ACTUAL" EXIT

# dump-envoy-configで実際の出力を生成
"$KUBECTL_LOCALMESH" dump-envoy-config -f "$CONFIG" --mock-config "$MOCK" > "$ACTUAL" 2>&1

# 差分チェック
if diff -u "$SNAPSHOT" "$ACTUAL" > /dev/null; then
    exit 0
else
    echo "Snapshot mismatch detected:"
    diff -u "$SNAPSHOT" "$ACTUAL" || true
    exit 1
fi
