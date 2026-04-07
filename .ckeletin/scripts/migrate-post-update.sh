#!/bin/bash
# Post-update migrations for task ckeletin:update
# Each migration checks its precondition before running, so this script
# is safe to run multiple times (idempotent).

set -eo pipefail

# Migration 1: Remove stale public component from .go-arch-lint.yml
# When: pkg/ directory doesn't exist but .go-arch-lint.yml still references it
if [ ! -d "pkg" ] && [ -f ".go-arch-lint.yml" ] && grep -q "pkg/\*\*" .go-arch-lint.yml; then
    echo "   ✓ Removing stale pkg/** references from .go-arch-lint.yml"

    # Remove public component section, commonComponents entry, and deps entry
    # Uses a temp file for portable sed behavior across macOS and Linux
    tmpfile=$(mktemp)
    awk '
    /PUBLIC PACKAGES/      { skip=1; next }
    /pkg\/ contains/       { skip=1; next }
    /Do NOT import from/   { skip=1; next }
    /Can be imported by/   { skip=1; next }
    /Are independent of/   { skip=1; next }
    /See ADR-010/          { skip=1; next }
    /Cannot depend on any internal/ { skip=1; next }
    /Can use any external vendor/   { skip=1; next }
    /Public packages are completely/ { skip=1; next }
    /^  public:/           { skip=1; next }
    skip && /^[^ ]|^  [^ ]/ && !/^    / { skip=0 }
    skip                   { next }
    /- public/             { next }
    { print }
    ' .go-arch-lint.yml > "$tmpfile"
    mv "$tmpfile" .go-arch-lint.yml
fi

# Migration 2: JSON output types moved from internal/ui to .ckeletin/pkg/output
# When: internal/ui/json.go still exists (pre-refactoring state)
if [ -f "internal/ui/json.go" ] && grep -q "JSONEnvelope" internal/ui/json.go; then
    echo "   ✓ Removing internal/ui/json.go (moved to .ckeletin/pkg/output/)"
    rm -f internal/ui/json.go internal/ui/json_test.go
    echo "   ⚠ ACTION REQUIRED: Update imports in your code:"
    echo "     - Replace 'internal/ui' with '.ckeletin/pkg/output' for JSON types"
    echo "     - Affected: cmd/root.go, main.go, internal/check/executor.go"
    echo "     - Types moved: JSONEnvelope, JSONError, JSONResponder, IsJSONMode, SetOutputMode, etc."
fi
