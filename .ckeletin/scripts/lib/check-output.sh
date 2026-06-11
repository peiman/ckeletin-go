#!/usr/bin/env bash
# Standard output functions for check scripts
# Usage: source scripts/lib/check-output.sh
#
# Sourced library: deliberately no `set -eo pipefail` here — shell options
# would leak into (or fight with) the sourcing script, which owns strict mode

# Colors and formatting
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Standard separator line
SEPARATOR="━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Category separator line width
CATEGORY_WIDTH=48

# category_header: Display category header with separator
# Usage: category_header "Code Quality"
category_header() {
    local title="$1"
    local title_length=${#title}
    local separator_length=$((CATEGORY_WIDTH - title_length - 2))

    if [ $separator_length -lt 3 ]; then
        separator_length=3
    fi

    # Build separator string character by character (more portable)
    local separator=""
    for ((i=0; i<separator_length; i++)); do
        separator="${separator}─"
    done

    echo ""
    echo "─── ${title} ${separator}"
}

# check_header: Display check header
# Usage: check_header "Checking code formatting"
check_header() {
    local message="$1"
    echo "🔍 ${message}..."
}

# check_success: Display success message
# Usage: check_success "All files properly formatted"
check_success() {
    local message="$1"
    echo "✅ ${message}"
}

# check_failure: Display failure message with details and remediation
# Usage: check_failure "Format check failed" "$error_output" "Run: task format"
check_failure() {
    local title="$1"
    local details="$2"
    local remediation="$3"

    echo ""
    echo "❌ ${title}"

    if [ -n "$details" ]; then
        echo ""
        echo "Details:"
        echo "$details" | sed 's/^/  /'
    fi

    if [ -n "$remediation" ]; then
        echo ""
        echo "How to fix:"
        echo "$remediation" | sed 's/^/  • /'
    fi

    echo ""
}

# check_summary: Display summary box for detailed checks
# Usage: check_summary "Success" "All checks passed" "• Item 1" "• Item 2"
check_summary() {
    local status="$1"
    local title="$2"
    shift 2
    local items=("$@")

    echo ""
    echo "$SEPARATOR"

    if [ "$status" = "success" ]; then
        echo "✅ ${title}"
    else
        echo "❌ ${title}"
    fi

    if [ ${#items[@]} -gt 0 ]; then
        echo ""
        for item in "${items[@]}"; do
            echo "$item"
        done
    fi

    echo "$SEPARATOR"
}

# check_info: Display optional context information
# Usage: check_info "Tool: go-licenses" "Policy: MIT, Apache-2.0"
check_info() {
    for line in "$@"; do
        echo "   $line"
    done
}

# check_note: Display informational note (for success cases with additional context)
# Usage: check_note "This is a fast source-based check. Run 'task check:license:binary' for accuracy."
check_note() {
    local message="$1"
    echo ""
    echo "Note: ${message}"
}

# run_check: Run a check command and handle success/failure
# Usage:
#   if run_check "command to run"; then
#       check_success "Success message"
#   else
#       check_failure "Failure title" "$CHECK_OUTPUT" "Remediation steps"
#       exit 1
#   fi
CHECK_OUTPUT=""
run_check() {
    local cmd="$1"
    CHECK_OUTPUT=$(eval "$cmd" 2>&1)
    local exit_code=$?
    return $exit_code
}

# ─────────────────────────────────────────────────────────────────────
# Scanner exclusions (SSOT)
#
# Directories that file-walking checks must never descend into.
# Claude Code's worktree feature creates full nested module checkouts
# under .claude/worktrees/; stale copies there break tree-walking
# scanners — and a scanner that errors on code it cannot build may
# silently SKIP it, degrading coverage behind a green summary.
#
# Mirrored (static configs cannot read this list):
#   - .go-arch-lint.yml excludeFiles has a /\.claude/.* entry
#   - semgrep is covered via .gitignore (.claude/* is ignored); no
#     .semgrepignore on purpose — creating one would REPLACE semgrep's
#     built-in default ignores and lose more than it adds
# The contract for all of them is pinned by
# test/integration/scanner_exclusion_test.go.
CKELETIN_SCAN_EXCLUDED_DIRS=(
    ".claude/worktrees"
)

# ckeletin_go_files: print Go files under a root (default .), pruning
# CKELETIN_SCAN_EXCLUDED_DIRS and any directory that carries its own
# go.mod (a nested module — e.g. another worktree's checkout — is never
# part of this module's scan surface).
ckeletin_go_files() {
    local root="${1:-.}"
    local prunes=()
    local d
    for d in "${CKELETIN_SCAN_EXCLUDED_DIRS[@]}"; do
        prunes+=(-path "${root}/${d}" -prune -o)
    done
    while IFS= read -r d; do
        prunes+=(-path "$d" -prune -o)
    done < <(find "$root" -mindepth 2 -name go.mod -exec dirname {} \; 2>/dev/null)
    find "$root" "${prunes[@]}" -type f -name '*.go' -print 2>/dev/null
}
