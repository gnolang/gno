#!/usr/bin/env bash
# Regenerate move-list.txt and gnovm-pinned.txt from safe-list.txt + repo state.
#
# Algorithm:
#   1. Compute base "must-stay" set: candidates (in examples/ but not on safe
#      list) that are imported by anything under gnovm/ tests OR by the
#      gno.land/ integration testdata/ harness (loadpkg, etc.). Both code
#      paths resolve via the hardcoded examples/ root and break when their
#      dependency moves.
#   2. Iteratively expand by transitive closure: any package imported by a
#      pinned package must also stay in examples/ to keep the safe set
#      self-consistent.
#   3. Move list = candidates - pinned.
#
# Run from repo root.

set -euo pipefail

cd "$(dirname "$0")/../.."

here="misc/quarantine"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

find examples/gno.land/p examples/gno.land/r examples-quarantine/gno.land/p examples-quarantine/gno.land/r \
    -name gnomod.toml -type f 2>/dev/null \
    | sed 's|/gnomod.toml||; s|^examples/||; s|^examples-quarantine/||' \
    | sort -u > "$tmp/all.txt"

# Sources that reference packages by hardcoded import-path and would break
# if their dependency moves out of examples/.
#   - gnovm/tests/files/    : filetests with `import "gno.land/..."`
#   - gnovm/cmd/, gnovm/pkg : transpiler tests and other tooling
#   - gno.land/pkg/integration/testdata/ : txtar harness `loadpkg` directives
{
    grep -rEoh '"gno.land/[^"]+"' \
        gnovm/tests/files/ gnovm/cmd/ gnovm/pkg/test/ gnovm/pkg/transpiler/ 2>/dev/null \
        | tr -d '"'
    grep -rEoh 'gno\.land/[A-Za-z0-9._/-]+' \
        gno.land/pkg/integration/testdata/ 2>/dev/null
} | sort -u > "$tmp/gnovm-imports.txt"

sort -u "$here/safe-list.txt" > "$tmp/safe.txt"

# Candidates: in repo (either tree) but not on safe list.
comm -23 "$tmp/all.txt" "$tmp/safe.txt" > "$tmp/candidates.txt"

# Base gnovm-pinned: candidates referenced directly by gnovm/.
comm -12 "$tmp/candidates.txt" "$tmp/gnovm-imports.txt" > "$tmp/pinned.txt"

# Helper: list gno.land/... imports for a given package path (look in either tree).
imports_of() {
    local pkg="$1"
    local dir
    for tree in examples examples-quarantine; do
        dir="$tree/$pkg"
        if [[ -d "$dir" ]]; then
            grep -rhE -o '"gno\.land/[^"]+"' "$dir" 2>/dev/null | tr -d '"' | sort -u
            return
        fi
    done
}

# Iterate: each pinned package's imports that are also candidates get pinned.
while true; do
    added=0
    while IFS= read -r pkg; do
        while IFS= read -r imp; do
            [[ -z "$imp" ]] && continue
            if grep -qxF "$imp" "$tmp/candidates.txt" && ! grep -qxF "$imp" "$tmp/pinned.txt"; then
                echo "$imp" >> "$tmp/pinned.txt"
                added=$((added + 1))
            fi
        done <<< "$(imports_of "$pkg")"
    done < "$tmp/pinned.txt"
    [[ "$added" -eq 0 ]] && break
done

sort -u "$tmp/pinned.txt" > "$here/gnovm-pinned.txt"
comm -23 "$tmp/candidates.txt" "$here/gnovm-pinned.txt" > "$here/move-list.txt"

printf 'safe (kept):       %4d packages\n' "$(wc -l < "$tmp/safe.txt")"
printf 'gnovm-pinned:      %4d packages (stay in examples/)\n' "$(wc -l < "$here/gnovm-pinned.txt")"
printf 'move-list:         %4d packages (move to examples-quarantine/)\n' "$(wc -l < "$here/move-list.txt")"
printf 'total candidates:  %4d packages\n' "$(wc -l < "$tmp/candidates.txt")"
