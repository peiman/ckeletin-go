#!/usr/bin/env bash
# Scan for secrets using gitleaks
# Prevents accidental commit of API keys, passwords, tokens, etc.
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

# Mode: "staged" for pre-commit, "all" for full scan
MODE="${1:-staged}"

# Allow skipping for specific cases
if [[ "${SKIP_SECRET_SCAN:-}" == "1" ]]; then
    echo "  Skipping secret scan (SKIP_SECRET_SCAN=1)"
    exit 0
fi

# Check for gitleaks installation
if ! command -v gitleaks &> /dev/null; then
    echo "Error: gitleaks is not installed"
    echo ""
    echo "Install with:"
    echo "  brew install gitleaks                    # macOS"
    echo "  go install github.com/gitleaks/gitleaks/v8@latest"
    echo ""
    echo "Or run: task setup  # Installs all dev tools"
    exit 1
fi

check_header "Scanning for secrets"

case "$MODE" in
    staged)
        # Scan only staged changes (fast, for pre-commit)
        echo "  Mode: staged changes only"
        if ! run_check "gitleaks protect --staged --config .gitleaks.toml --verbose 2>&1"; then
            # Check if it's a "no commits" error (empty repo)
            if echo "$CHECK_OUTPUT" | grep -qi "no commits"; then
                check_success "No commits to scan"
                exit 0
            fi
            check_failure \
                "Secrets detected in staged changes" \
                "$CHECK_OUTPUT" \
                "Remove secrets before committing:"$'\n'"  1. Remove the secret from the file"$'\n'"  2. Use environment variables instead"$'\n'"  3. Add to .gitleaks.toml allowlist if false positive"
            exit 1
        fi
        ;;

    all)
        # Scan entire repository history
        echo "  Mode: full repository scan"
        if ! run_check "gitleaks detect --config .gitleaks.toml --verbose 2>&1"; then
            check_failure \
                "Secrets detected in repository" \
                "$CHECK_OUTPUT" \
                "Review and remediate detected secrets"$'\n'"Consider using git-filter-repo to remove from history"
            exit 1
        fi
        ;;

    baseline)
        # Create baseline report (for legacy repos with existing secrets)
        echo "  Mode: creating baseline report"
        BASELINE_FILE="reports/gitleaks-baseline.json"
        mkdir -p reports
        if gitleaks detect --config .gitleaks.toml --report-path "$BASELINE_FILE" --report-format json 2>/dev/null; then
            echo "  No secrets found - baseline not needed"
        else
            echo "  Baseline created: $BASELINE_FILE"
            echo "  Add --baseline-path $BASELINE_FILE to ignore these in future scans"
        fi
        exit 0
        ;;

    *)
        echo "Usage: $0 [staged|all|baseline]"
        echo ""
        echo "Modes:"
        echo "  staged    Scan staged changes only (default, for pre-commit)"
        echo "  all       Scan entire repository history"
        echo "  baseline  Create baseline report for legacy repos"
        exit 1
        ;;
esac

check_success "No secrets detected"
