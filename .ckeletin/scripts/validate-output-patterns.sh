#!/usr/bin/env bash
# Enforces the output rules: business code under internal/ does not print
# directly. *_test.go files are exempt. A framework file that must print is
# listed by its exact path in framework_paths below. A directory named ui,
# progress, logger, or testutil anywhere else is scanned.
#
# Every other path that must write directly is listed, with a reason, in
# .ckeletin-output-allowlist at the repository root. A line may pin the
# number of matching lines, or the exact line numbers:
#
#   internal/ui/check.go count=1 check printer writes to stdout
#   internal/progress/terminal.go lines=41 detects a terminal on stdout
#
# Every entry needs count= or lines=. count= is greater than zero. A path
# is listed once. A pin that does not match fails, including a count for a
# file that no longer prints and a path that is gone. A directory is not an
# entry. Pin each file. Comments are not matches. A write to os.Stderr is.
# The allowlist is always .ckeletin-output-allowlist at the repository root.
# An environment variable does not select another file.
# A missing or malformed allowlist is its own error. It is not a scan failure
# and it does not dump a traceback.

set -u

allowlist=".ckeletin-output-allowlist"

# Exact framework paths. Do not add a directory name here. An application's
# internal/ui is not a framework path and is scanned unless the application
# allowlists the file.
framework_paths=(
)

if [[ ! -f "$allowlist" ]]; then
	echo "Output allowlist is missing"
	echo "$allowlist"
	echo "Create this file at the repository root. Each line is a relative path, then a reason."
	echo "An empty file allows no extra paths. The rendering layer is not exempt by directory name."
	echo "Pin a count or line numbers so a new print in an allowlisted file fails."
	echo "Example: internal/tui/screen.go count=1 full-screen TUI writes the frame"
	exit 1
fi

paths=()
pins=()
line_no=0
while IFS= read -r line || [[ -n "$line" ]]; do
	line_no=$((line_no + 1))
	trimmed="${line#"${line%%[![:space:]]*}"}"
	if [[ -z "$trimmed" || "$trimmed" == \#* ]]; then
		continue
	fi
	path="${trimmed%%[[:space:]]*}"
	rest="${trimmed#"$path"}"
	rest="${rest#"${rest%%[![:space:]]*}"}"
	pin="-"
	if [[ "$rest" == count=* || "$rest" == line=* || "$rest" == lines=* ]]; then
		pin="${rest%%[[:space:]]*}"
		rest="${rest#"$pin"}"
		rest="${rest#"${rest%%[![:space:]]*}"}"
		if [[ "$pin" == count=* ]]; then
			num="${pin#count=}"
			if [[ ! "$num" =~ ^[1-9][0-9]*$ ]]; then
				echo "Output allowlist is malformed"
				echo "${allowlist}:${line_no}"
				echo "count= must be greater than zero."
				exit 1
			fi
		else
			nums="${pin#*=}"
			if [[ ! "$nums" =~ ^[0-9]+(,[0-9]+)*$ ]]; then
				echo "Output allowlist is malformed"
				echo "${allowlist}:${line_no}"
				echo "lines= is a comma-separated list of line numbers."
				exit 1
			fi
		fi
	fi
	if [[ "$pin" == "-" ]]; then
		echo "Output allowlist is malformed"
		echo "${allowlist}:${line_no}"
		echo "Every entry needs count= or lines=. An unpinned path would allow new prints."
		exit 1
	fi
	if [[ -d "$path" ]]; then
		echo "Output allowlist is malformed"
		echo "${allowlist}:${line_no}"
		echo "not a file; pin each file"
		exit 1
	fi
	if [[ -z "$rest" || "$path" == /* || "$path" == *..* ]]; then
		echo "Output allowlist is malformed"
		echo "${allowlist}:${line_no}"
		echo "Each line is a relative path, an optional count= or lines= pin, then a reason."
		exit 1
	fi
	if ((${#paths[@]} > 0)); then
		for existing in "${paths[@]}"; do
			if [[ "$existing" == "$path" ]]; then
				echo "Output allowlist is malformed"
				echo "${allowlist}:${line_no}"
				echo "This path is already listed."
				exit 1
			fi
		done
	fi
	paths+=("$path")
	pins+=("$pin")
done < "$allowlist"

echo "Validating output patterns..."

if [[ ! -d internal ]]; then
	echo "Output patterns: no internal/ tree to scan"
	exit 0
fi

# Comments are blanked first. A mention of fmt.Println in a comment is not
# a write. Strings are left intact, so a string that names the call still
# matches. os.Stderr is a write to the status stream.
if ! hits=$(python3 - <<'PY'
import os
import sys

def strip_go_comments(src):
    out = []
    i = 0
    n = len(src)
    while i < n:
        c = src[i]
        nxt = src[i + 1] if i + 1 < n else ""
        if c == "/" and nxt == "/":
            while i < n and src[i] != "\n":
                out.append(" ")
                i += 1
            continue
        if c == "/" and nxt == "*":
            out.append(" ")
            out.append(" ")
            i += 2
            while i < n and not (src[i] == "*" and i + 1 < n and src[i + 1] == "/"):
                out.append("\n" if src[i] == "\n" else " ")
                i += 1
            if i < n:
                out.append(" ")
                out.append(" ")
                i += 2
            continue
        tick = chr(96)
        if c == tick:
            out.append(c)
            i += 1
            while i < n and src[i] != tick:
                out.append(src[i])
                i += 1
            if i < n:
                out.append(tick)
                i += 1
            continue
        if c == '"' or c == "'":
            quote = c
            out.append(c)
            i += 1
            while i < n and src[i] != quote:
                if src[i] == "\\" and i + 1 < n:
                    out.append(src[i])
                    out.append(src[i + 1])
                    i += 2
                    continue
                out.append(src[i])
                i += 1
            if i < n and src[i] == quote:
                out.append(quote)
                i += 1
            continue
        out.append(c)
        i += 1
    return "".join(out)

needles = ("fmt.Print", "os.Stdout", "os.Stderr")
for dirpath, _, names in os.walk("internal"):
    for name in sorted(names):
        if not name.endswith(".go") or name.endswith("_test.go"):
            continue
        path = os.path.join(dirpath, name).replace(os.sep, "/")
        try:
            text = open(path, encoding="utf-8").read()
        except OSError as exc:
            print(f"cannot read {path}: {exc}", file=sys.stderr)
            sys.exit(1)
        stripped = strip_go_comments(text).splitlines()
        for number, line in enumerate(stripped, start=1):
            if any(needle in line for needle in needles):
                print(f"{path}:{number}:{line}")
PY
); then
	echo "Output patterns: could not scan Go files"
	exit 1
fi

hit_files=()
hit_lines=()
hit_full=()
while IFS= read -r hit; do
	[[ -z "$hit" ]] && continue
	file="${hit%%:*}"
	rest="${hit#*:}"
	line="${rest%%:*}"
	skip=0
	if ((${#framework_paths[@]} > 0)); then
		for entry in "${framework_paths[@]}"; do
			if [[ "$file" == "$entry" ]]; then
				skip=1
				break
			fi
		done
	fi
	if [[ "$skip" -eq 1 ]]; then
		continue
	fi
	hit_files+=("$file")
	hit_lines+=("$line")
	hit_full+=("$hit")
done <<< "$hits"

violations=""
seen_files="|"
if ((${#hit_files[@]} > 0)); then
	for i in "${!hit_files[@]}"; do
		file="${hit_files[$i]}"
		case "$seen_files" in
		*"|$file|"*) continue ;;
		esac
		seen_files="${seen_files}${file}|"

		count=0
		lines=""
		for j in "${!hit_files[@]}"; do
			if [[ "${hit_files[$j]}" == "$file" ]]; then
				count=$((count + 1))
				lines="${lines}${hit_lines[$j]},"
			fi
		done
		lines="${lines%,}"

		pin=""
		allowed=0
		if ((${#paths[@]} > 0)); then
			for p in "${!paths[@]}"; do
				entry="${paths[$p]}"
				if [[ "$file" == "$entry" ]]; then
					pin="${pins[$p]}"
					allowed=1
					break
				fi
			done
		fi

		if [[ "$allowed" -eq 0 ]]; then
			for j in "${!hit_files[@]}"; do
				if [[ "${hit_files[$j]}" == "$file" ]]; then
					violations="${violations}${hit_full[$j]}"$'\n'
				fi
			done
			continue
		fi

		if [[ "$pin" == count=* ]]; then
			want="${pin#count=}"
			if [[ "$count" -ne "$want" ]]; then
				violations="${violations}${file}: expected ${want} direct output lines, found ${count}"$'\n'
			fi
			continue
		fi
		want="${pin#*=}"
		if [[ "$lines" != "$want" ]]; then
			violations="${violations}${file}: expected output lines ${want}, found ${lines}"$'\n'
		fi
	done
fi

for p in "${!paths[@]}"; do
	entry="${paths[$p]}"
	pin="${pins[$p]}"
	if [[ ! -f "$entry" ]]; then
		violations="${violations}${entry}: allowlist path does not exist"$'\n'
		continue
	fi
	count=0
	lines=""
	if ((${#hit_files[@]} > 0)); then
		for j in "${!hit_files[@]}"; do
			if [[ "${hit_files[$j]}" == "$entry" ]]; then
				count=$((count + 1))
				lines="${lines}${hit_lines[$j]},"
			fi
		done
	fi
	lines="${lines%,}"
	if [[ "$pin" == count=* ]]; then
		want="${pin#count=}"
		if [[ "$count" -eq 0 && "$want" -ne 0 ]]; then
			violations="${violations}${entry}: expected ${want} direct output lines, found 0"$'\n'
		fi
		continue
	fi
	if [[ "$count" -eq 0 ]]; then
		violations="${violations}${entry}: expected output lines ${pin#*=}, found none"$'\n'
	fi
done

if [[ -n "$violations" ]]; then
	echo "Direct output in business logic"
	echo "Write through progress.Reporter, internal/ui, or .ckeletin/pkg/output."
	echo "Do not call fmt.Print* or write os.Stdout or os.Stderr from internal/."
	echo "To keep an existing path, add it to ${allowlist} with a reason. A count= or lines= pin fails when the file gains or loses a match."
	printf '%s' "$violations"
	exit 1
fi

echo "Output patterns are compliant"
exit 0
