#!/usr/bin/env bash
# PBOM CLI integration test script
# Simulates a GitHub Actions environment and exercises all commands.
set -euo pipefail

PASS=0
FAIL=0
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN="$REPO_ROOT/bin/pbom"
OUTPUT_DIR="$SCRIPT_DIR/output"

green() { printf "\033[32m%s\033[0m\n" "$1"; }
red()   { printf "\033[31m%s\033[0m\n" "$1"; }

check() {
    local name="$1"; shift
    if "$@" >/dev/null 2>&1; then
        green "  PASS: $name"
        PASS=$((PASS + 1))
    else
        red "  FAIL: $name"
        FAIL=$((FAIL + 1))
    fi
}

check_fail() {
    local name="$1"; shift
    if ! "$@" >/dev/null 2>&1; then
        green "  PASS: $name (expected failure)"
        PASS=$((PASS + 1))
    else
        red "  FAIL: $name (expected failure but succeeded)"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== PBOM Integration Tests ==="
echo ""

# Setup
mkdir -p "$OUTPUT_DIR"

# 1. Build binary
echo "--- Building binary ---"
(cd "$REPO_ROOT" && go build -o bin/pbom ./cmd/pbom)
check "Binary builds successfully" test -x "$BIN"

# 2. Version command
echo ""
echo "--- pbom version ---"
check "Version prints output" "$BIN" version

# 3. Generate command (with fake GHA env vars)
echo ""
echo "--- pbom generate ---"
export GITHUB_SHA="a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
export GITHUB_REPOSITORY="acme-corp/payments-service"
export GITHUB_REF="refs/heads/main"
export GITHUB_REF_NAME="main"
export GITHUB_ACTOR="jane.doe"
export GITHUB_RUN_ID="7890123456"
export GITHUB_WORKFLOW="CI"
export GITHUB_EVENT_NAME="push"
export GITHUB_WORKFLOW_REF="acme-corp/payments-service/.github/workflows/ci.yml@refs/heads/main"

check "Generate to file" "$BIN" generate -o "$OUTPUT_DIR/test-pbom.json"
check "Output file exists" test -f "$OUTPUT_DIR/test-pbom.json"
check "Output is valid JSON" python3 -c "import json; json.load(open('$OUTPUT_DIR/test-pbom.json'))"

# 4. Validate command
echo ""
echo "--- pbom validate ---"
check "Validate passes on generated PBOM" "$BIN" validate "$OUTPUT_DIR/test-pbom.json"

# 5. Inspect command
echo ""
echo "--- pbom inspect ---"
check "Inspect shows formatted output" "$BIN" inspect "$OUTPUT_DIR/test-pbom.json"
check "Inspect --json shows JSON" "$BIN" inspect --json "$OUTPUT_DIR/test-pbom.json"

# 6. Validate with malformed file
echo ""
echo "--- Validation failure tests ---"
echo '{"pbom_version":"1.0.0"}' > "$OUTPUT_DIR/bad-pbom.json"
check_fail "Validate rejects incomplete PBOM" "$BIN" validate "$OUTPUT_DIR/bad-pbom.json"

echo '{"pbom_version":"2.0.0","id":"x","source":{"repository":"r","commit_sha":"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"},"build":{"workflow_run_id":"1","workflow_name":"CI","actor":"a","status":"success"}}' > "$OUTPUT_DIR/bad-version.json"
check_fail "Validate rejects wrong version" "$BIN" validate "$OUTPUT_DIR/bad-version.json"

echo "not json" > "$OUTPUT_DIR/bad-json.json"
check_fail "Validate rejects invalid JSON" "$BIN" validate "$OUTPUT_DIR/bad-json.json"

# 7. Generate to stdout
echo ""
echo "--- pbom generate to stdout ---"
STDOUT_OUTPUT=$("$BIN" generate)
check "Generate to stdout produces output" test -n "$STDOUT_OUTPUT"

# Cleanup env vars
unset GITHUB_SHA GITHUB_REPOSITORY GITHUB_REF GITHUB_REF_NAME
unset GITHUB_ACTOR GITHUB_RUN_ID GITHUB_WORKFLOW GITHUB_EVENT_NAME GITHUB_WORKFLOW_REF

# Summary
echo ""
echo "================================"
echo "  Results: $PASS passed, $FAIL failed"
echo "================================"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
