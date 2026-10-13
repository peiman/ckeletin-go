#!/usr/bin/env bash
# Fast vulnerability scan for pre-commit hooks
# Uses caching to avoid repeated scans within a time window
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

# Cache configuration
CACHE_DIR="${TMPDIR:-/tmp}/ckeletin-go-vuln-cache"
CACHE_FILE="${CACHE_DIR}/last-scan"
CACHE_RESULT="${CACHE_DIR}/result"
CACHE_TTL=300  # 5 minutes - skip if scanned recently

# Allow skipping for offline work
if [[ "${SKIP_VULN_CHECK:-}" == "1" ]]; then
    echo "  Skipping vulnerability check (SKIP_VULN_CHECK=1)"
    exit 0
fi

check_header "Fast vulnerability scan"

# Create cache directory
mkdir -p "$CACHE_DIR"

# Check if we have a recent scan result
if [[ -f "$CACHE_FILE" ]]; then
    LAST_SCAN=$(cat "$CACHE_FILE")
    NOW=$(date +%s)
    AGE=$((NOW - LAST_SCAN))

    if [[ $AGE -lt $CACHE_TTL ]]; then
        # Check cached result
        if [[ -f "$CACHE_RESULT" ]] && [[ "$(cat "$CACHE_RESULT")" == "0" ]]; then
            check_success "No vulnerabilities (cached ${AGE}s ago)"
            exit 0
        elif [[ -f "$CACHE_RESULT" ]]; then
            check_failure \
                "Vulnerabilities found (cached ${AGE}s ago)" \
                "Previous scan detected issues" \
                "Run: task check:vuln for details"
            exit 1
        fi
    fi
fi

# Run govulncheck
# Use -show verbose to get more info, but parse for just the summary
vuln_code=0
run_check "govulncheck ./... 2>&1" || vuln_code=$?
if [[ "$vuln_code" -eq 0 ]]; then
    # No vulnerabilities found
    date +%s > "$CACHE_FILE"
    echo "0" > "$CACHE_RESULT"
    check_success "No vulnerabilities found"
    exit 0
fi

case "$(classify_govulncheck "$vuln_code" "$CHECK_OUTPUT")" in
    network)
        echo "  Warning: Vulnerability check skipped (network unavailable)"
        echo "  Run 'task check:vuln' when online"
        # Don't cache network errors - retry next time
        exit 0
        ;;
esac

# Real vulnerabilities found
date +%s > "$CACHE_FILE"
echo "1" > "$CACHE_RESULT"
check_failure \
    "Security vulnerabilities detected" \
    "$CHECK_OUTPUT" \
    "Run: task check:vuln for details"$'\n'"Update vulnerable dependencies before committing"
exit 1
