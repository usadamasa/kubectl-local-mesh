#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNAPSHOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PROJECT_ROOT="$(cd "$SNAPSHOT_DIR/../.." && pwd)"

TESTDATA_DIR="$SNAPSHOT_DIR/testdata"
CONFIGS_DIR="$TESTDATA_DIR/configs"
MOCKS_DIR="$TESTDATA_DIR/mocks"
SNAPSHOTS_DIR="$TESTDATA_DIR/snapshots"

KUBECTL_LOCALMESH="$PROJECT_ROOT/bin/kubectl-localmesh"

# ビルド確認
if [ ! -f "$KUBECTL_LOCALMESH" ]; then
    echo "Error: kubectl-localmesh binary not found at $KUBECTL_LOCALMESH"
    echo "Please run 'task build' first"
    exit 1
fi

# テストケース自動検出（configs/*.yamlから）
TEST_CASES=()
for config_file in "$CONFIGS_DIR"/*.yaml; do
    if [ -f "$config_file" ]; then
        test_case=$(basename "$config_file" .yaml)
        TEST_CASES+=("$test_case")
    fi
done

if [ ${#TEST_CASES[@]} -eq 0 ]; then
    echo "Error: No test cases found in $CONFIGS_DIR"
    exit 1
fi

FAILED_TESTS=()
PASSED_TESTS=()

echo "🧪 Running snapshot tests..."
echo

for test_case in "${TEST_CASES[@]}"; do
    config="$CONFIGS_DIR/${test_case}.yaml"
    mock="$MOCKS_DIR/${test_case}-mocks.yaml"
    snapshot="$SNAPSHOTS_DIR/${test_case}.yaml"

    if [ ! -f "$config" ]; then
        echo "❌ SKIP: $test_case (config not found)"
        FAILED_TESTS+=("$test_case (config not found)")
        continue
    fi

    if [ ! -f "$mock" ]; then
        echo "❌ SKIP: $test_case (mock not found)"
        FAILED_TESTS+=("$test_case (mock not found)")
        continue
    fi

    if [ ! -f "$snapshot" ]; then
        echo "❌ SKIP: $test_case (snapshot not found - run update-snapshots.sh first)"
        FAILED_TESTS+=("$test_case (snapshot not found)")
        continue
    fi

    # 実行して差分チェック
    if "$SCRIPT_DIR/diff-snapshot.sh" "$test_case"; then
        echo "✅ PASS: $test_case"
        PASSED_TESTS+=("$test_case")
    else
        echo "❌ FAIL: $test_case"
        FAILED_TESTS+=("$test_case")
    fi
    echo
done

# 結果サマリー
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Test Results:"
echo "   Passed: ${#PASSED_TESTS[@]}"
echo "   Failed: ${#FAILED_TESTS[@]}"
echo "   Total:  ${#TEST_CASES[@]}"

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo
    echo "Failed tests:"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
    exit 1
fi

echo
echo "✅ All snapshot tests passed!"
exit 0
