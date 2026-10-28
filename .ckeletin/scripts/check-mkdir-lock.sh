#!/usr/bin/env bash
# Flag mkdir used as a mutual-exclusion lock.
#
# ubuntu-26.04 ships uutils mkdir 0.10, where two concurrent mkdir calls on
# one path can both exit 0. A lock has to be noclobber (O_EXCL), which does
# not depend on coreutils:
#   ( set -C; printf '%s\n' "$$" > "$LOCK" ) 2>/dev/null
#
# `mkdir -p` of an ordinary directory is not a lock.
# Usage: check-mkdir-lock.sh [file...]
# With no files, scan tracked shell scripts and shell blocks in Taskfiles
# and GitHub workflows.
set -euo pipefail

errors=0

code_without_comment() {
    local line=$1
    local out="" quote="" prev="" char
    local i len=${#line}
    for ((i = 0; i < len; i++)); do
        char=${line:i:1}
        if [ -n "$quote" ]; then
            out+=$char
            if [ "$char" = "$quote" ] && [ "$prev" != "\\" ]; then
                quote=""
            fi
            prev=$char
            continue
        fi
        if [ "$char" = "#" ]; then
            break
        fi
        case $char in
            "'" | '"') quote=$char ;;
        esac
        out+=$char
        prev=$char
    done
    printf '%s' "$out"
}

is_mkdir_command() {
    [[ $1 =~ (^|[^[:alnum:]_-])mkdir([[:space:]]|$) ]]
}

names_a_lock() {
    [[ $1 =~ [Ll][Oo][Cc][Kk] ]]
}

is_lock_condition() {
    [[ $1 =~ (^|[^[:alnum:]_])(until|if)[[:space:]]+mkdir([^[:alnum:]_]|$) ]]
}

is_mkdir_or_exit() {
    is_mkdir_command "$1" && [[ $1 =~ \|\|[[:space:]]*(return|exit)([^[:alnum:]_]|$) ]]
}

report() {
    local file=$1 number=$2 line=$3 why=$4
    printf 'mkdir lock: %s:%s: %s\n' "$file" "$number" "$why"
    printf '  %s\n' "$line"
    printf '  Replace the mkdir lock with noclobber:\n'
    printf '    ( set -C; printf '\''%%s\n'\'' "$$" > "$LOCK" ) 2>/dev/null\n'
    errors=$((errors + 1))
}

# The shell command on a line, after a YAML list dash or a cmd:/run: key.
# A task name such as validate:mkdir-lock is not a command.
shell_text() {
    local text=$1
    text="${text#"${text%%[![:space:]]*}"}"
    if [[ $text == "- "* ]]; then
        text=${text#- }
        text="${text#"${text%%[![:space:]]*}"}"
    fi
    case $text in
        cmd:* | run:*)
            text=${text#*:}
            text="${text#"${text%%[![:space:]]*}"}"
            ;;
    esac
    printf '%s' "$text"
}

first_word() {
    printf '%s' "${1%%[[:space:]]*}"
}

check_command() {
    local file=$1 number=$2 raw=$3 command=$4
    local word
    word=$(first_word "$command")
    case $word in
        until | if)
            if is_lock_condition "$command"; then
                report "$file" "$number" "$raw" "mkdir is used as a lock condition"
            fi
            ;;
        mkdir)
            if is_mkdir_or_exit "$command"; then
                report "$file" "$number" "$raw" "mkdir failure is used as a lock"
            elif names_a_lock "$command"; then
                report "$file" "$number" "$raw" "mkdir creates a path named like a lock"
            fi
            ;;
    esac
}

check_file() {
    local file=$1
    local number=0 line code shell part rest
    while IFS= read -r line || [ -n "$line" ]; do
        number=$((number + 1))
        code=$(code_without_comment "$line")
        if [ -z "${code//[[:space:]]/}" ]; then
            continue
        fi
        rest=$(shell_text "$code")
        while [ -n "$rest" ]; do
            part=${rest%%;*}
            part="${part#"${part%%[![:space:]]*}"}"
            part="${part%"${part##*[![:space:]]}"}"
            if [ -n "$part" ]; then
                check_command "$file" "$number" "$line" "$part"
            fi
            case $rest in
                *\;*) rest=${rest#*;} ;;
                *) break ;;
            esac
            rest="${rest#"${rest%%[![:space:]]*}"}"
        done
    done < "$file"
}

scan_repo() {
    local file
    while IFS= read -r -d '' file; do
        case $file in
            .ckeletin/scripts/check-mkdir-lock.sh) continue ;;
        esac
        check_file "$file"
    done < <(git ls-files -z -- '*.sh' '*.bash' 'Taskfile.yml' '.ckeletin/Taskfile.yml' '.github/workflows/*.yml' '.github/workflows/*.yaml')
}

if [ "$#" -gt 0 ]; then
    for file in "$@"; do
        check_file "$file"
    done
else
    scan_repo
fi

if [ "$errors" -gt 0 ]; then
    printf 'mkdir lock: %s shell lock(s). See the noclobber form above.\n' "$errors"
    exit 1
fi

printf 'mkdir lock: no mkdir lock in shell scripts\n'
