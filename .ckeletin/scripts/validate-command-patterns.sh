#!/bin/bash
# validate-command-patterns.sh
#
# Validates that command files follow ckeletin-go ultra-thin command patterns
# (ADR-001). Runs in `task check` and CI to enforce consistency.
#
# Enforced contract (ADR-001 "Enforcement"):
#   - Each run* function targets <=30 lines, measured from its `func` line
#     through its closing brace. 31-35 lines: warning. More than 35: error.
#   - Whole command files above 80 lines get an advisory warning.
#   - Whitelist mechanism: // ckeletin:allow-custom-command skips the pattern
#     checks, but the marker MUST carry a short justification on the marker
#     line or on a comment line directly above/below it. A bare marker errors.

set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

RUN_FUNC_TARGET=30 # run* functions should stay at or below this
RUN_FUNC_LIMIT=35  # 31-35 lines tolerated with a warning; beyond is an error
FILE_WARN_LIMIT=80 # advisory ceiling for whole command files

ERRORS=0
WARNINGS=0
ERROR_DETAILS=""

# An exemption marker must carry a human-readable reason: trailing text on the
# marker line, or a comment with prose on the line directly above or below it.
# File-path header comments (e.g. "// cmd/foo.go") do not count as a reason.
marker_has_justification() {
    local file="$1"
    local lineno="$2"
    local line adj

    line=$(awk -v n="$lineno" 'NR == n' "$file")
    if printf '%s' "${line#*ckeletin:allow-custom-command}" | grep -q '[A-Za-z]'; then
        return 0
    fi

    for adj in $((lineno - 1)) $((lineno + 1)); do
        if [ "$adj" -lt 1 ]; then
            continue
        fi
        line=$(awk -v n="$adj" 'NR == n' "$file")
        case "$line" in
        //go:build*) continue ;;
        esac
        if printf '%s' "$line" | grep -qE '^//[[:space:]]*cmd/[^[:space:]]+\.go[[:space:]]*$'; then
            continue
        fi
        if printf '%s' "$line" | grep -qE '^[[:space:]]*//.*[A-Za-z]'; then
            return 0
        fi
    done

    return 1
}

check_header "Validating ADR-001: Ultra-thin command pattern"

# Get all command files (exclude framework files, tests, and helper files)
COMMAND_FILES=$(find cmd -name "*.go" -not -name "*_test.go" -not -name "root.go" -not -name "flags.go" -not -name "helpers.go" -not -name "*_helpers.go" -not -name "template*.go")

for cmd_file in $COMMAND_FILES; do
    cmd_name=$(basename "$cmd_file" .go)

    # Whitelisted files skip the pattern checks, but the marker itself must
    # say why the exemption exists (ADR-001, ADR-014)
    marker_line=$(awk '/\/\/ ckeletin:allow-custom-command/ { print NR; exit }' "$cmd_file")
    if [ -n "$marker_line" ]; then
        if ! marker_has_justification "$cmd_file" "$marker_line"; then
            ERROR_DETAILS+="$cmd_name: ckeletin:allow-custom-command marker has no justification (add a short reason on or next to the marker line)"$'\n'
            ((++ERRORS))
        fi
        continue
    fi

    # Check 1: run* function length (the 30-line contract)
    # Measured from the `func runX(...)` line through its closing brace.
    # Known heuristic blind spots (accepted): only column-0 `func run...`
    # declarations match, so receiver methods (`func (x T) runY(...)`) are
    # not measured; the end is the next `}` at column 0, which gofmt
    # guarantees for top-level functions but hand-written or generated
    # code without gofmt could skew the count.
    run_funcs=$(awk '
        /^func run/ { name = $2; sub(/\(.*/, "", name); start = NR }
        /^}/        { if (start) { print name, NR - start + 1; start = 0 } }
    ' "$cmd_file")
    while read -r func_name func_lines; do
        if [ -z "$func_name" ]; then
            continue
        fi
        if [ "$func_lines" -gt "$RUN_FUNC_LIMIT" ]; then
            ERROR_DETAILS+="$cmd_name: ${func_name}() is ${func_lines} lines (target ${RUN_FUNC_TARGET}, hard limit ${RUN_FUNC_LIMIT}) - move logic to internal/"$'\n'
            ((++ERRORS))
        elif [ "$func_lines" -gt "$RUN_FUNC_TARGET" ]; then
            ERROR_DETAILS+="$cmd_name: ${func_name}() is ${func_lines} lines (target ${RUN_FUNC_TARGET}, tolerated up to ${RUN_FUNC_LIMIT})"$'\n'
            ((++WARNINGS))
        fi
    done <<<"$run_funcs"

    # Check 2: whole-file size (advisory; run* functions are the hard contract)
    line_count=$(($(wc -l <"$cmd_file")))
    if [ "$line_count" -gt "$FILE_WARN_LIMIT" ]; then
        ERROR_DETAILS+="$cmd_name: Command file is $line_count lines (advisory ceiling ${FILE_WARN_LIMIT})"$'\n'
        ((++WARNINGS))
    fi

    # Skip remaining structure checks for parent-only commands (files that
    # group subcommands with AddCommand but have no RunE/Run)
    if grep -q "AddCommand(" "$cmd_file" && ! grep -qE "(RunE:|Run:)" "$cmd_file"; then
        continue
    fi

    # Check 3: Command metadata exists (check both project and framework locations)
    # For subcommands like note_get.go, also check parent config (note_config.go)
    metadata_found=false
    if find internal/config/commands -name "${cmd_name}_config.go" 2>/dev/null | grep -q .; then
        metadata_found=true
    elif find .ckeletin/pkg/config/commands -name "${cmd_name}_config.go" 2>/dev/null | grep -q .; then
        metadata_found=true
    else
        # Check parent config: note_get -> note
        parent_name="${cmd_name%%_*}"
        if [ "$parent_name" != "$cmd_name" ]; then
            if find internal/config/commands -name "${parent_name}_config.go" 2>/dev/null | grep -q .; then
                metadata_found=true
            fi
        fi
    fi

    if ! $metadata_found; then
        ERROR_DETAILS+="$cmd_name: Missing metadata file internal/config/commands/${cmd_name}_config.go"$'\n'
        ((++ERRORS))
    fi

    # Check 4: Uses NewCommand helper
    if ! grep -q "NewCommand(" "$cmd_file"; then
        ERROR_DETAILS+="$cmd_name: Does not use NewCommand() helper"$'\n'
        ((++WARNINGS))
    fi

    # Check 5: Uses MustAddToRoot helper
    if ! grep -q "MustAddToRoot(" "$cmd_file"; then
        if grep -q "RootCmd.AddCommand" "$cmd_file" && grep -q "setupCommandConfig" "$cmd_file"; then
            ERROR_DETAILS+="$cmd_name: Manual RootCmd setup (consider MustAddToRoot)"$'\n'
            ((++WARNINGS))
        fi
    fi

    # Check 6: Business logic detection (simple heuristic)
    # Look for complex control flow outside of run* functions
    # Exclude common command-layer patterns: JSON encoding, formatted output, envelope wrapping
    if grep -v "^func run" "$cmd_file" | grep -E "(for\s+.*\{|if\s+.*\{\s*$|switch\s+.*\{)" | grep -v "^//" | grep -vE "(json\.(NewEncoder|Marshal)|fmt\.Fprint|envelope\.)" | grep -q .; then
        ERROR_DETAILS+="$cmd_name: Possible business logic in command file (should be in internal/$cmd_name/)"$'\n'
        ((++WARNINGS))
    fi
done

# Summary
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    check_success "All commands follow the pattern"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    check_success "All commands pass (${WARNINGS} warning(s))"
    check_note "Warnings are suggestions and won't fail the build."
    exit 0
else
    check_failure \
        "${ERRORS} error(s) found, ${WARNINGS} warning(s)" \
        "$ERROR_DETAILS" \
        "Keep run* functions <=${RUN_FUNC_TARGET} lines; move logic to internal/.
To whitelist a command, add: // ckeletin:allow-custom-command - <short reason>"
    exit 1
fi
