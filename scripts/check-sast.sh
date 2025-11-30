#!/usr/bin/env bash
# Run Semgrep static analysis for security and code quality
# Complements gosec with more advanced pattern matching
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

# Mode: "local" uses .semgrep.yml, "registry" uses semgrep registry rules
MODE="${1:-local}"

# Allow skipping
if [[ "${SKIP_SAST:-}" == "1" ]]; then
    echo "  Skipping SAST scan (SKIP_SAST=1)"
    exit 0
fi

# Check for semgrep installation
if ! command -v semgrep &> /dev/null; then
    echo "Error: semgrep is not installed"
    echo ""
    echo "Install with:"
    echo "  brew install semgrep                     # macOS"
    echo "  pip install semgrep                      # Python"
    echo "  pipx install semgrep                     # Isolated Python"
    echo ""
    echo "Or run: task setup  # Installs all dev tools"
    exit 1
fi

check_header "Running static analysis (SAST)"

case "$MODE" in
    local)
        # Use local rules only (fast, no network)
        echo "  Mode: local rules (.semgrep.yml)"
        if ! run_check "semgrep scan --config .semgrep.yml --error --quiet 2>&1"; then
            check_failure \
                "SAST issues detected" \
                "$CHECK_OUTPUT" \
                "Fix the issues or add exceptions to .semgrep.yml"
            exit 1
        fi
        ;;

    registry)
        # Use semgrep registry rules (comprehensive, requires network)
        echo "  Mode: registry rules (p/golang, p/security-audit)"
        if ! run_check "semgrep scan --config p/golang --config p/security-audit --error --quiet 2>&1"; then
            check_failure \
                "SAST issues detected" \
                "$CHECK_OUTPUT" \
                "Review and fix detected issues"
            exit 1
        fi
        ;;

    full)
        # Run both local and registry rules
        echo "  Mode: full (local + registry)"
        if ! run_check "semgrep scan --config .semgrep.yml --config p/golang --config p/security-audit --error --quiet 2>&1"; then
            check_failure \
                "SAST issues detected" \
                "$CHECK_OUTPUT" \
                "Review and fix detected issues"
            exit 1
        fi
        ;;

    report)
        # Generate detailed report
        echo "  Mode: generating report"
        mkdir -p reports
        REPORT_FILE="reports/semgrep-report.json"
        semgrep scan --config .semgrep.yml --json --output "$REPORT_FILE" 2>/dev/null || true
        echo "  Report saved to: $REPORT_FILE"

        # Also generate human-readable summary
        FINDINGS=$(jq '.results | length' "$REPORT_FILE" 2>/dev/null || echo "0")
        echo "  Findings: $FINDINGS"
        exit 0
        ;;

    *)
        echo "Usage: $0 [local|registry|full|report]"
        echo ""
        echo "Modes:"
        echo "  local     Use local .semgrep.yml rules only (default, fast)"
        echo "  registry  Use semgrep registry rules (comprehensive)"
        echo "  full      Run both local and registry rules"
        echo "  report    Generate JSON report to reports/semgrep-report.json"
        exit 1
        ;;
esac

check_success "No SAST issues detected"
