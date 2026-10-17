#!/usr/bin/env bash
# Re-apply the names `task init` wrote into a framework tree.
#
# `task ckeletin:update` checks out upstream `.ckeletin/`, which brings the
# upstream module path and binary name back. `task init` had already rewritten
# markdown, YAML, and Go files (including tests) to the derived project's
# names. Leaving the upstream names in place is a `.ckeletin/` diff with no
# changelog edit, so `task check` fails after a same-version update.
#
# Shell scripts are left alone. `task init` does not rewrite them.
#
# A Go file that contains `ckeletin:keep-upstream-names` is left alone too.
# The rename tools' own fixtures must keep the upstream module and binary
# name as literals. Rewriting them makes the tests compare a name to itself.
#
# Usage: rewrite-derived-names.sh OLD_MODULE NEW_MODULE OLD_NAME NEW_NAME DIR

set -euo pipefail

if [ "$#" -ne 5 ]; then
	echo "Usage: $0 OLD_MODULE NEW_MODULE OLD_NAME NEW_NAME DIR" >&2
	exit 1
fi

old_module=$1
new_module=$2
old_name=$3
new_name=$4
dir=$5

if [ -z "$old_module" ] || [ -z "$new_module" ] || [ -z "$old_name" ] || [ -z "$new_name" ] || [ -z "$dir" ]; then
	echo "rewrite-derived-names: all arguments must be non-empty" >&2
	exit 1
fi

if [ ! -d "$dir" ]; then
	echo "rewrite-derived-names: directory not found: $dir" >&2
	exit 1
fi

to_env() {
	# Uppercase, then any character that is not a letter or digit becomes "_".
	# "ckeletin-go" -> "CKELETIN_GO". Matches scaffold toEnvPrefix.
	printf '%s' "$1" | tr '[:lower:]' '[:upper:]' | tr -c 'A-Z0-9' '_'
}

github_path() {
	# task init strips the literal prefix "github.com/" and nothing else.
	local module=$1
	local prefix="github.com/"
	if [ "${module#"$prefix"}" != "$module" ]; then
		printf '%s' "${module#"$prefix"}"
	else
		printf '%s' "$module"
	fi
}

owner_of() {
	# host/owner/repo -> owner. Fewer than three segments -> empty.
	local module=$1
	local rest=${module#*/}
	if [ "$rest" = "$module" ]; then
		printf ''
		return
	fi
	case "$rest" in
	*/*) printf '%s' "${rest%%/*}" ;;
	*) printf '' ;;
	esac
}

old_env=$(to_env "$old_name")
new_env=$(to_env "$new_name")
old_github=$(github_path "$old_module")
new_github=$(github_path "$new_module")
old_owner=$(owner_of "$old_module")
new_owner=$(owner_of "$new_module")

# Literal replace. awk gsub would treat dots and slashes in a module path as
# a regular expression. A pair whose old and new text are equal is skipped so
# the index loop cannot spin.
lreplace_fn='
function lreplace(s, old, new,    i, out) {
	if (old == "" || old == new) return s
	out = ""
	while ((i = index(s, old)) > 0) {
		out = out substr(s, 1, i - 1) new
		s = substr(s, i + length(old))
	}
	return out s
}
'

# True when the file carries ckeletin:keep-upstream-names. grep's miss is
# not an error; set -e would otherwise exit the script.
keeps_upstream_names() {
	grep -q -F 'ckeletin:keep-upstream-names' "$1"
}

file_mode() {
	# macOS stat uses -f and prints octal like 644. GNU stat -f is a different
	# flag, so accept only an octal mode and fall back to stat -c.
	local mode
	mode=$(stat -f '%OLp' "$1" 2>/dev/null || true)
	case "$mode" in
	[0-7][0-7][0-7] | [0-7][0-7][0-7][0-7]) printf '%s' "$mode" ;;
	*) stat -c '%a' "$1" ;;
	esac
}

rewrite_text() {
	local file=$1
	local tmp mode
	mode=$(file_mode "$file")
	tmp=$(mktemp "${file}.XXXXXX")
	# -v must come before the program. BSD awk treats a later -v as a filename.
	awk \
		-v old_module="$old_module" -v new_module="$new_module" \
		-v old_github="$old_github" -v new_github="$new_github" \
		-v old_owner="$old_owner" -v new_owner="$new_owner" \
		-v old_name="$old_name" -v new_name="$new_name" \
		"${lreplace_fn}"'
		{
			s = $0
			s = lreplace(s, old_module, new_module)
			s = lreplace(s, old_github, new_github)
			if (old_owner != "" && new_owner != "") {
				s = lreplace(s, "@" old_owner, "@" new_owner)
				s = lreplace(s, "\"" old_owner "\"", "\"" new_owner "\"")
			}
			s = lreplace(s, old_name, new_name)
			print s
		}
	' "$file" >"$tmp"
	if cmp -s "$file" "$tmp"; then
		rm -f "$tmp"
	else
		chmod "$mode" "$tmp"
		mv "$tmp" "$file"
	fi
}

rewrite_go() {
	local file=$1
	local tmp mode
	mode=$(file_mode "$file")
	tmp=$(mktemp "${file}.XXXXXX")
	awk \
		-v old_name="$old_name" -v new_name="$new_name" \
		-v old_env="$old_env" -v new_env="$new_env" \
		"${lreplace_fn}"'
		{
			trimmed = $0
			sub(/^[ \t]+/, "", trimmed)
			if (trimmed ~ /^import \(/) {
				in_import = 1
				print
				next
			}
			if (in_import) {
				if (trimmed == ")") in_import = 0
				print
				next
			}
			if (trimmed ~ /^import /) {
				print
				next
			}
			s = $0
			s = lreplace(s, old_name, new_name)
			s = lreplace(s, old_env, new_env)
			print s
		}
	' "$file" >"$tmp"
	if cmp -s "$file" "$tmp"; then
		rm -f "$tmp"
	else
		chmod "$mode" "$tmp"
		mv "$tmp" "$file"
	fi
}

while IFS= read -r -d '' file; do
	if keeps_upstream_names "$file"; then
		continue
	fi
	rewrite_go "$file"
done < <(find "$dir" \( -name .git -o -name vendor -o -name dist -o -name .task \) -prune -o -type f -name '*.go' -print0)

while IFS= read -r -d '' file; do
	rewrite_text "$file"
done < <(find "$dir" \( -name .git -o -name vendor -o -name dist -o -name .task \) -prune -o -type f \( -name '*.md' -o -name '*.yml' -o -name '*.yaml' \) -print0)
