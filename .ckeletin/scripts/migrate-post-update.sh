#!/bin/bash
# Post-update migrations for task ckeletin:update
# Each migration checks its precondition before running, so this script
# is safe to run multiple times (idempotent).

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

# Migration 3: Detect missing task forwardings in project Taskfile.yml
# When: .ckeletin/Taskfile.yml has tasks that the project Taskfile doesn't forward
# The expected-forwardings.txt file lists framework tasks that should have
# project-level aliases. This runs on every update to catch new tasks.
FORWARDINGS_FILE="$SCRIPT_DIR/expected-forwardings.txt"
if [ -f "$FORWARDINGS_FILE" ] && [ -f "Taskfile.yml" ]; then
    missing=()
    while IFS= read -r task_name; do
        # Skip empty lines and comments
        [[ -z "$task_name" || "$task_name" == \#* ]] && continue
        # Check if the project Taskfile has a forwarding for this task
        if ! grep -q "^  ${task_name}:" Taskfile.yml 2>/dev/null; then
            missing+=("$task_name")
        fi
    done < "$FORWARDINGS_FILE"

    if [ ${#missing[@]} -gt 0 ]; then
        echo ""
        echo "   ⚠ Missing task forwardings in Taskfile.yml:"
        echo "     The following framework tasks have no project-level alias."
        echo "     Add these to your Taskfile.yml under '# === Convenience Aliases ===':"
        echo ""
        for task_name in "${missing[@]}"; do
            # Derive description from framework Taskfile
            desc=$(grep -A1 "^  ${task_name}:" .ckeletin/Taskfile.yml 2>/dev/null | grep "desc:" | sed 's/.*desc: //' | head -1)
            echo "  ${task_name}:"
            if [ -n "$desc" ]; then
                echo "    desc: ${desc}"
            fi
            echo "    cmds: [task: ckeletin:${task_name}]"
            echo ""
        done
    fi
fi

# Migration 4: Ensure .go-arch-lint.yml includes all framework infrastructure packages
# When: Framework adds new packages (e.g., .ckeletin/pkg/output/) that aren't
# in the downstream project's infrastructure component
if [ -f ".go-arch-lint.yml" ]; then
    FRAMEWORK_INFRA_PKGS=(
        ".ckeletin/pkg/config/**"
        ".ckeletin/pkg/logger/**"
        ".ckeletin/pkg/output/**"
        ".ckeletin/pkg/testutil/**"
    )
    missing_pkgs=()
    for pkg in "${FRAMEWORK_INFRA_PKGS[@]}"; do
        if ! grep -qF "$pkg" .go-arch-lint.yml; then
            missing_pkgs+=("$pkg")
        fi
    done
    if [ ${#missing_pkgs[@]} -gt 0 ]; then
        echo ""
        echo "   ⚠ Missing framework packages in .go-arch-lint.yml infrastructure component:"
        for pkg in "${missing_pkgs[@]}"; do
            echo "     - $pkg"
        done
        echo ""
        echo "   Add these to the 'infrastructure.in' section in .go-arch-lint.yml"
        echo "   to prevent 'not attached to any component' errors."
    fi
fi
