#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNAPSHOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PROJECT_ROOT="$(cd "$SNAPSHOT_DIR/../.." && pwd)"

TESTDATA_DIR="$SNAPSHOT_DIR/testdata"
CONFIGS_DIR="$TESTDATA_DIR/configs"
MOCKS_DIR="$TESTDATA_DIR/mocks"
SNAPSHOTS_DIR="$TESTDATA_DIR/snapshots"
MAPPINGS_DIR="$TESTDATA_DIR/portforward-mappings"

KUBECTL_LOCALMESH="$PROJECT_ROOT/bin/kubectl-localmesh"

# ビルド確認
if [ ! -f "$KUBECTL_LOCALMESH" ]; then
    echo "Error: kubectl-localmesh binary not found"
    echo "Please run 'task build' first"
    exit 1
fi

# スナップショットディレクトリ作成
mkdir -p "$SNAPSHOTS_DIR"
mkdir -p "$MAPPINGS_DIR"

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

echo "🔄 Updating snapshots..."
echo

for test_case in "${TEST_CASES[@]}"; do
    config="$CONFIGS_DIR/${test_case}.yaml"
    mock="$MOCKS_DIR/${test_case}-mocks.yaml"

    if [ ! -f "$config" ]; then
        echo "⚠️  SKIP: $test_case (config not found)"
        continue
    fi

    if [ ! -f "$mock" ]; then
        echo "⚠️  SKIP: $test_case (mock not found)"
        continue
    fi

    snapshot="$SNAPSHOTS_DIR/${test_case}.yaml"
    mapping="$MAPPINGS_DIR/${test_case}-mapping.yaml"

    echo "📝 Updating: $test_case"
    "$KUBECTL_LOCALMESH" dump-envoy-config -f "$config" --mock-config "$mock" > "$snapshot"
    "$KUBECTL_LOCALMESH" dump-envoy-config -f "$config" --mock-config "$mock" --output-mapping > "$mapping"
done

echo
echo "✅ Snapshots and mappings updated successfully!"
echo "⚠️  Please review changes before committing:"
echo "    git diff testdata/envoy-snapshots/testdata/snapshots/"
echo "    git diff testdata/envoy-snapshots/testdata/portforward-mappings/"
