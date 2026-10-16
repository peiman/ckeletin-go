#!/usr/bin/env bash
# Stop a framework update when .ckeletin/ is not the upstream tree at the
# version the project records, after the rewrites the update itself applies.
#
# Usage:
#   framework-local-diff.sh --mode update|dry-run|compatibility \
#     [--discard 0|1] [--result FILE]
#
# update and compatibility exit 1 on a stop and print one
# CKELETIN_UPDATE_RESULT line. dry-run prints the difference, says the update
# would stop, and exits 0. --discard 1 may continue only for a committed
# difference, and only in update mode. Uncommitted changes always stop.
#
# --result receives discarded=true when a committed difference is discarded,
# otherwise discarded=false when the command may continue.
#
# Bash 3.2. No namerefs. No sort -V.

set -euo pipefail

mode=
discard=0
result=
snap=

# shellcheck disable=SC2329 # invoked from the EXIT trap
cleanup() {
	if [ -n "$snap" ] && [ -d "$snap" ]; then
		rm -rf "$snap"
	fi
}
trap cleanup EXIT

usage() {
	echo "framework-local-diff: --mode update|dry-run|compatibility [--discard 0|1] [--result FILE]" >&2
	exit 2
}

while [ $# -gt 0 ]; do
	case "$1" in
	--mode | --discard | --result)
		if [ $# -lt 2 ]; then
			echo "framework-local-diff: missing value for $1" >&2
			exit 2
		fi
		case "$1" in
		--mode) mode=$2 ;;
		--discard) discard=$2 ;;
		--result) result=$2 ;;
		esac
		shift 2
		;;
	*)
		echo "framework-local-diff: unknown argument: $1" >&2
		usage
		;;
	esac
done

case "$mode" in
update | dry-run | compatibility) ;;
*) usage ;;
esac

write_result() {
	if [ -n "$result" ]; then
		printf 'discarded=%s\n' "$1" >"$result"
	fi
}

read_version() {
	if [ ! -f .ckeletin/VERSION ]; then
		printf 'unknown'
		return
	fi
	tr -d '[:space:]' < .ckeletin/VERSION
}

# reason is uncommitted, local_difference, or unknown_version.
# Differences have already been printed.
stop_for() {
	reason=$1
	if [ "$mode" = "dry-run" ]; then
		echo "The update would stop (${reason}). Nothing was changed."
		write_result false
		exit 0
	fi
	printf 'CKELETIN_UPDATE_RESULT={"status":"stopped","reason":"%s","from":"%s","to":null,"committed":false,"discarded":false}\n' \
		"$reason" "$from_version"
	exit 1
}

print_uncommitted() {
	echo "Uncommitted changes in .ckeletin/:"
	git --no-pager status --porcelain -- .ckeletin/
	echo "Diff:"
	git --no-pager diff HEAD -- .ckeletin/ || true
	while IFS= read -r path; do
		[ -z "$path" ] && continue
		if [ -f "$path" ]; then
			diff -u /dev/null "$path" || true
		else
			echo "Untracked: $path"
		fi
	done <<EOF
$(git ls-files --others --exclude-standard -- .ckeletin/)
EOF
}

# show_tree_diff prints a content diff with stable labels.
show_tree_diff() {
	left=$1
	right=$2
	set +e
	diff -ru "$left" "$right"
	set -e
	return 0
}

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
	echo "framework-local-diff: not a git repository" >&2
	exit 1
fi

export GIT_TERMINAL_PROMPT=0
export COPYFILE_DISABLE=1

from_version=$(read_version)

porcelain=$(git --no-pager status --porcelain -- .ckeletin/ || true)
if [ -n "$porcelain" ]; then
	print_uncommitted
	stop_for uncommitted
fi

if ! git remote get-url ckeletin-upstream >/dev/null 2>&1; then
	echo "framework-local-diff: remote ckeletin-upstream is not configured" >&2
	exit 1
fi

if ! git fetch ckeletin-upstream main; then
	echo "framework-local-diff: failed to fetch ckeletin-upstream main" >&2
	exit 1
fi

match=
while IFS= read -r rev; do
	[ -z "$rev" ] && continue
	ver=$(git show "${rev}:.ckeletin/VERSION" 2>/dev/null | tr -d '[:space:]' || true)
	if [ "$ver" = "$from_version" ]; then
		match=$rev
		break
	fi
done <<EOF
$(git rev-list ckeletin-upstream/main)
EOF

if [ -z "$match" ]; then
	echo "No commit on ckeletin-upstream/main has .ckeletin/VERSION ${from_version}."
	stop_for unknown_version
fi

snap=$(mktemp -d)
upstream_tree="$snap/upstream"
local_tree="$snap/local"
mkdir -p "$upstream_tree" "$local_tree"

if ! git archive "$match" .ckeletin | tar -x -C "$upstream_tree"; then
	echo "framework-local-diff: failed to read .ckeletin at ${match}" >&2
	exit 1
fi
if ! git archive HEAD .ckeletin | tar -x -C "$local_tree"; then
	echo "framework-local-diff: failed to read the committed .ckeletin tree" >&2
	exit 1
fi

# Same rewrites task ckeletin:update applies after checkout. Shell scripts are
# left alone. The string split keeps scaffold-init from replacing the upstream
# identity when this file is copied into a derived project.
upstream_module="github.com/peiman/ckeleti""n-go"
current_module=$(awk 'NR==1 { print $2; exit }' go.mod)
if ! env -u GOWORK go run ./.ckeletin/scripts/rewrite-imports/ \
	-old "$upstream_module" -new "$current_module" -dir "$upstream_tree/.ckeletin" >/dev/null; then
	echo "framework-local-diff: failed to rewrite imports in the upstream snapshot" >&2
	exit 1
fi

upstream_name="ckeletin""-go"
# awk exits after the first match. grep | head under pipefail can lose the name.
current_name=$(awk '/BINARY_NAME:/ { print $2; exit }' Taskfile.yml || true)
if [ -n "$current_name" ] && [ "$current_name" != "$upstream_name" ]; then
	bash .ckeletin/scripts/rewrite-derived-names.sh \
		"$upstream_module" "$current_module" "$upstream_name" "$current_name" \
		"$upstream_tree/.ckeletin"
fi

set +e
diff -rq "$upstream_tree/.ckeletin" "$local_tree/.ckeletin" >/dev/null
diff_status=$?
set -e
if [ "$diff_status" -eq 0 ]; then
	write_result false
	exit 0
fi
if [ "$diff_status" -ne 1 ]; then
	echo "framework-local-diff: diff failed" >&2
	exit 1
fi

echo "Local framework difference:"
show_tree_diff "$upstream_tree/.ckeletin" "$local_tree/.ckeletin" \
	| while IFS= read -r line; do
		line=${line//"$upstream_tree/.ckeletin"/upstream}
		line=${line//"$local_tree/.ckeletin"/local}
		printf '%s\n' "$line"
	done

if [ "$mode" = "update" ] && [ "$discard" = "1" ]; then
	echo "Discarding local framework changes in .ckeletin/."
	write_result true
	exit 0
fi

stop_for local_difference
