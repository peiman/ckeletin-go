#!/usr/bin/env bash
# Enforces the output rules: business code under internal/ does not print
# directly. The rendering layer and *.go tests are exempt. Every other path
# that must write directly is listed, with a reason, in
# .ckeletin-output-allowlist at the repository root.
#
# A missing or malformed allowlist is its own error. It is not a scan failure
# and it does not dump a traceback.

set -u

allowlist="${OUTPUT_ALLOWLIST:-.ckeletin-output-allowlist}"

if [[ ! -f "$allowlist" ]]; then
	echo "Output allowlist is missing"
	echo "$allowlist"
	echo "Create this file at the repository root. Each line is a relative path, then a reason."
	echo "An empty file allows no extra paths. The rendering layer is already exempt."
	echo "Example: internal/tui/screen.go full-screen TUI writes the frame"
	exit 1
fi

allowed=()
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
	if [[ -z "$rest" || "$path" == /* || "$path" == *..* ]]; then
		echo "Output allowlist is malformed"
		echo "${allowlist}:${line_no}"
		echo "Each line is a relative path, then a reason. Comments start with #."
		exit 1
	fi
	allowed+=("$path")
done < "$allowlist"

echo "Validating output patterns..."

if [[ ! -d internal ]]; then
	echo "Output patterns: no internal/ tree to scan"
	exit 0
fi

hits=$(grep -R -n -E 'fmt\.Print|os\.Stdout' internal \
	--include='*.go' \
	--exclude='*_test.go' \
	--exclude-dir=ui \
	--exclude-dir=progress \
	--exclude-dir=logger \
	--exclude-dir=testutil \
	|| true)

violations=""
while IFS= read -r hit; do
	[[ -z "$hit" ]] && continue
	file="${hit%%:*}"
	keep=1
	if ((${#allowed[@]} > 0)); then
		for entry in "${allowed[@]}"; do
			if [[ "$file" == "$entry" || "$file" == "$entry"/* ]]; then
				keep=0
				break
			fi
		done
	fi
	if [[ "$keep" -eq 1 ]]; then
		violations="${violations}${hit}"$'\n'
	fi
done <<< "$hits"

if [[ -n "$violations" ]]; then
	echo "Direct output in business logic"
	echo "Write through progress.Reporter, internal/ui, or .ckeletin/pkg/output."
	echo "Do not call fmt.Print* or os.Stdout from internal/."
	echo "To keep an existing path, add it to ${allowlist} with a reason. New paths are not allowed until they are listed."
	printf '%s' "$violations"
	exit 1
fi

echo "Output patterns are compliant"
exit 0
