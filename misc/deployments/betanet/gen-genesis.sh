#!/usr/bin/env bash
# Generate betanet genesis.json.
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"

# Resolve listed realms + all transitive deps in topological order.
pkg_dirs=$(cd "$EXAMPLES_DIR" && gno tool deplist \
    ./gno.land/r/sys/... \
    ./gno.land/r/gov/... \
    ./gno.land/r/gnoland/blog/... \
    ./gno.land/r/gnoland/wugnot/... \
    ./gno.land/r/gnoland/coins/... \
    ./gno.land/r/gnoland/boards2/... \
)

# Copy resolved packages into a staging directory.
STAGING=$(mktemp -d)
trap 'rm -rf "$STAGING"' EXIT
while IFS= read -r dir; do
    [[ -z "$dir" ]] && continue
    rel="${dir#$EXAMPLES_DIR/}"
    mkdir -p "$(dirname "$STAGING/$rel")"
    cp -r "$dir" "$STAGING/$rel"
done <<< "$pkg_dirs"

# Build genesis.
gnogenesis generate -chain-id betanet -output-path "$SCRIPT_DIR/genesis.json"
#gnogenesis validator add \
#    -name placeholder \
#    -power 1 \
#    -address g1foobar \
#    -pub-key g1pubbaz \
#    --genesis-path "$SCRIPT_DIR/genesis.json"
gnogenesis txs add packages "$STAGING" \
    --genesis-path "$SCRIPT_DIR/genesis.json"

echo "Done: genesis.json"
