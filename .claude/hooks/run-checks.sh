#!/bin/bash
set -e

cd "$CLAUDE_PROJECT_DIR"

echo "=== Running tests ==="
task test

echo "=== Running linters ==="
task lint

echo "=== All checks passed ==="
