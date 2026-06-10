#!/bin/bash
# Install development tools for ckeletin-go project
# This script is called automatically via SessionStart hook

set -e

echo "Setting up development environment for ckeletin-go..."

# Add go/bin to PATH if not already there
export PATH="${HOME}/go/bin:$PATH"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Tool versions are pinned in the framework Taskfile (SSOT); read them from
# there so this hook installs the same versions as `task setup`. A missing
# pin falls back to latest, with a warning on stderr so the gap is visible.
FRAMEWORK_TASKFILE="${SCRIPT_DIR}/../Taskfile.yml"

tool_version() {
    local pin
    pin=$(grep -E "^[[:space:]]*TOOL_${1}_VERSION:" "$FRAMEWORK_TASKFILE" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    if [ -z "$pin" ]; then
        echo "⚠️  No TOOL_${1}_VERSION pin in ${FRAMEWORK_TASKFILE}; falling back to latest" >&2
        pin="latest"
    fi
    echo "$pin"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Tools that failed to install; reported (with exit 1) at the end so one
# failure doesn't hide the others
FAILED_TOOLS=""

# Function to install a tool if not present
install_tool() {
    local tool_name=$1
    local tool_package=$2
    local output

    if command_exists "$tool_name"; then
        echo "✅ $tool_name already installed"
    else
        echo "📦 Installing $tool_name..."
        # Capture output so go install's real exit status isn't masked by grep
        if output=$(go install "$tool_package" 2>&1); then
            echo "✅ $tool_name installed"
        else
            echo "$output" | grep -v "^go: downloading" >&2 || true
            echo "❌ Failed to install $tool_name ($tool_package)" >&2
            FAILED_TOOLS="${FAILED_TOOLS} ${tool_name}"
        fi
    fi
}

# Install task runner first (required for all task commands)
install_tool "task" "github.com/go-task/task/v3/cmd/task@latest"

# Install essential development tools
install_tool "goimports" "golang.org/x/tools/cmd/goimports@$(tool_version GOIMPORTS)"
install_tool "govulncheck" "golang.org/x/vuln/cmd/govulncheck@$(tool_version GOVULNCHECK)"
install_tool "gotestsum" "gotest.tools/gotestsum@$(tool_version GOTESTSUM)"
install_tool "golangci-lint" "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(tool_version GOLANGCI_LINT)"
install_tool "go-mod-outdated" "github.com/psampaz/go-mod-outdated@$(tool_version GO_MOD_OUTDATED)"
install_tool "yq" "github.com/mikefarah/yq/v4@$(tool_version YQ)"  # mikefarah/yq (Go): conform.sh parses the conformance mapping with it

# Optional: lefthook (may fail due to network/version issues, not critical)
if ! command_exists "lefthook"; then
    echo "📦 Installing lefthook (optional)..."
    # Capture output so go install's real exit status isn't masked by grep
    if output=$(go install "github.com/evilmartians/lefthook/v2@$(tool_version LEFTHOOK)" 2>&1); then
        echo "✅ lefthook installed"
    else
        echo "$output" | grep -v "^go: downloading" >&2 || true
        echo "⚠️  lefthook skipped (not critical)"
    fi
fi

if [ -n "$FAILED_TOOLS" ]; then
    echo "" >&2
    echo "❌ Some tools failed to install:${FAILED_TOOLS}" >&2
    echo "   Re-run: bash .ckeletin/scripts/install_tools.sh (or: task setup)" >&2
    exit 1
fi

echo ""
echo "✅ Development environment ready!"
echo "   All essential tools installed in ${HOME}/go/bin"
echo ""
echo "   You can now use task commands:"
echo "   - task format     # Format code"
echo "   - task lint       # Run linters"
echo "   - task test       # Run tests"
echo "   - task check      # Run all checks"
echo "   - task bench      # Run benchmarks"
echo ""
