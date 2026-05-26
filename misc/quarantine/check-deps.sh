#!/usr/bin/env bash
# Check which packages in examples/ have imports that now live in
# examples-quarantine/ (i.e. broken cross-tree deps that must be fixed by
# either pulling the import back into examples/ or moving the importer
# into examples-quarantine/).

set -euo pipefail

cd "$(dirname "$0")/../.."

# Map: import-path -> tree
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

find examples/gno.land -name gnomod.toml | sed 's|/gnomod.toml||; s|^examples/||' | sort -u > "$tmp/in-examples.txt"
find examples-quarantine/gno.land -name gnomod.toml | sed 's|/gnomod.toml||; s|^examples-quarantine/||' | sort -u > "$tmp/in-quar.txt"

# For each package still in examples/, collect its gno.land/... imports
# and flag any that point to packages in examples-quarantine/.
while IFS= read -r pkg; do
    dir="examples/$pkg"
    imports=$(grep -rhE -o '"gno\.land/[^"]+"' "$dir" 2>/dev/null \
        | tr -d '"' | sort -u)
    while IFS= read -r imp; do
        [[ -z "$imp" ]] && continue
        if grep -qxF "$imp" "$tmp/in-quar.txt"; then
            printf '%s -> %s\n' "$pkg" "$imp"
        fi
    done <<< "$imports"
done < "$tmp/in-examples.txt"
