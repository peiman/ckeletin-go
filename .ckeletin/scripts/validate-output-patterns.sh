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
# A pin that does not match the file fails. An unpinned path allows every
# current and future match in that file, or in files under that directory.
# A missing or malformed allowlist is its own error. It is not a scan failure
# and it does not dump a traceback.

set -u

allowlist="${OUTPUT_ALLOWLIST:-.ckeletin-output-allowlist}"

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
			if [[ ! "$num" =~ ^[0-9]+$ ]]; then
				echo "Output allowlist is malformed"
				echo "${allowlist}:${line_no}"
				echo "count= must be a non-negative integer."
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
	if [[ -z "$rest" || "$path" == /* || "$path" == *..* ]]; then
		echo "Output allowlist is malformed"
		echo "${allowlist}:${line_no}"
		echo "Each line is a relative path, an optional count= or lines= pin, then a reason."
		exit 1
	fi
	paths+=("$path")
	pins+=("$pin")
done < "$allowlist"

echo "Validating output patterns..."

if [[ ! -d internal ]]; then
	echo "Output patterns: no internal/ tree to scan"
	exit 0
fi

hits=$(grep -R -n -E 'fmt\.Print|os\.Stdout' internal \
	--include='*.go' \
	--exclude='*_test.go' \
	|| true)

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
			if [[ "$allowed" -eq 0 ]]; then
				for p in "${!paths[@]}"; do
					entry="${paths[$p]}"
					entry_pin="${pins[$p]}"
					if [[ "$entry_pin" == "-" && "$file" == "$entry"/* ]]; then
						pin="-"
						allowed=1
						break
					fi
				done
			fi
		fi

		if [[ "$allowed" -eq 0 ]]; then
			for j in "${!hit_files[@]}"; do
				if [[ "${hit_files[$j]}" == "$file" ]]; then
					violations="${violations}${hit_full[$j]}"$'\n'
				fi
			done
			continue
		fi

		if [[ "$pin" == "-" ]]; then
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

if [[ -n "$violations" ]]; then
	echo "Direct output in business logic"
	echo "Write through progress.Reporter, internal/ui, or .ckeletin/pkg/output."
	echo "Do not call fmt.Print* or os.Stdout from internal/."
	echo "To keep an existing path, add it to ${allowlist} with a reason. A count= or lines= pin fails when the file gains or loses a match."
	printf '%s' "$violations"
	exit 1
fi

echo "Output patterns are compliant"
exit 0
