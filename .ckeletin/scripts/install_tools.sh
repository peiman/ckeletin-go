#!/bin/bash
# Install development tools for ckeletin-go project
# This script is called automatically via SessionStart hook

set -e

echo "Setting up development environment for ckeletin-go..."

# Add go/bin to PATH if not already there
export PATH="/root/go/bin:$PATH"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install a tool if not present
install_tool() {
    local tool_name=$1
    local tool_package=$2

    if command_exists "$tool_name"; then
        echo "✅ $tool_name already installed"
    else
        echo "📦 Installing $tool_name..."
        go install "$tool_package" 2>&1 | grep -v "^go: downloading" || true
        echo "✅ $tool_name installed"
    fi
}

# Install task runner first (required for all task commands)
install_tool "task" "github.com/go-task/task/v3/cmd/task@latest"

# Install essential development tools
install_tool "goimports" "golang.org/x/tools/cmd/goimports@latest"
install_tool "govulncheck" "golang.org/x/vuln/cmd/govulncheck@latest"
install_tool "gotestsum" "gotest.tools/gotestsum@latest"
install_tool "golangci-lint" "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
install_tool "go-mod-outdated" "github.com/psampaz/go-mod-outdated@latest"

# Optional: lefthook (may fail due to network/version issues, not critical)
if ! command_exists "lefthook"; then
    echo "📦 Installing lefthook (optional)..."
    go install github.com/evilmartians/lefthook@latest 2>&1 | grep -v "^go: downloading" || echo "⚠️  lefthook skipped (not critical)"
fi

echo ""
echo "✅ Development environment ready!"
echo "   All essential tools installed in /root/go/bin"
echo ""
echo "   You can now use task commands:"
echo "   - task format     # Format code"
echo "   - task lint       # Run linters"
echo "   - task test       # Run tests"
echo "   - task check      # Run all checks"
echo "   - task bench      # Run benchmarks"
echo ""
