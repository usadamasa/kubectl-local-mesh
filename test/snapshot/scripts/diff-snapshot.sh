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
MAPPING="$TESTDATA_DIR/portforward-mappings/${TEST_CASE}-mapping.yaml"

KUBECTL_LOCALMESH="$PROJECT_ROOT/bin/kubectl-localmesh"

# 一時ファイルに出力
ACTUAL=$(mktemp)
ACTUAL_MAPPING=$(mktemp)
trap "rm -f $ACTUAL $ACTUAL_MAPPING" EXIT

# dump-envoy-configで実際の出力を生成（stdoutのみ、警告はstderrに出力される）
"$KUBECTL_LOCALMESH" dump-envoy-config -f "$CONFIG" --mock-config "$MOCK" > "$ACTUAL"
"$KUBECTL_LOCALMESH" dump-envoy-config -f "$CONFIG" --mock-config "$MOCK" --output-mapping > "$ACTUAL_MAPPING"

RESULT=0

# Envoy設定の差分チェック
if ! diff -u "$SNAPSHOT" "$ACTUAL" > /dev/null; then
    echo "Envoy config snapshot mismatch detected:"
    diff -u "$SNAPSHOT" "$ACTUAL" || true
    RESULT=1
fi

# マッピングの差分チェック（マッピングファイルが存在する場合のみ）
if [ -f "$MAPPING" ]; then
    if ! diff -u "$MAPPING" "$ACTUAL_MAPPING" > /dev/null; then
        echo "Mapping snapshot mismatch detected:"
        diff -u "$MAPPING" "$ACTUAL_MAPPING" || true
        RESULT=1
    fi
fi

exit $RESULT
