#!/usr/bin/env bash
# Move packages listed in move-list.txt from examples/ to examples-quarantine/
# using `git mv` so history is preserved.
#
# Run from repo root. Idempotent: a package already moved is skipped.

set -euo pipefail

cd "$(dirname "$0")/../.."

list="misc/quarantine/move-list.txt"
src_root="examples"
dst_root="examples-quarantine"

# Filter out entries whose parent is also in the list (parent move drags them
# along). Sort so a deterministic order keeps diff review tractable.
declare -a to_move
while IFS= read -r pkg; do
    parent="$pkg"
    skip=0
    while parent="${parent%/*}"; [[ "$parent" == *"/"* ]]; do
        if grep -qxF "$parent" "$list"; then
            skip=1
            break
        fi
    done
    if (( skip == 0 )); then
        to_move+=("$pkg")
    fi
done < "$list"

printf 'will move %d top-level entries (nested packages move with their parent)\n' "${#to_move[@]}"

for pkg in "${to_move[@]}"; do
    src="$src_root/$pkg"
    dst="$dst_root/$pkg"

    if [[ ! -d "$src" ]]; then
        echo "skip (already moved or missing): $pkg"
        continue
    fi

    mkdir -p "$(dirname "$dst")"
    git mv "$src" "$dst"
done

printf 'done\n'
