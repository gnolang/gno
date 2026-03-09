#!/usr/bin/env bash

# Generate betanet genesis.json with only the realms listed in packages.txt
# and their transitive dependencies (auto-resolved via `gno tool deplist`).

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"

CHAIN_ID="${CHAIN_ID:-betanet}"
GENESIS_FILE="${GENESIS_FILE:-$SCRIPT_DIR/genesis.json}"
MANIFEST="${MANIFEST:-$SCRIPT_DIR/packages.txt}"

# Read package list: one entry per line, skip blanks and comments.
patterns=()
while IFS= read -r line; do
    line="${line%%#*}"          # strip comments
    line="${line## }"           # trim
    line="${line%% }"
    [[ -n "$line" ]] && patterns+=("./gno.land/$line/...")
done < "$MANIFEST"

if [[ ${#patterns[@]} -eq 0 ]]; then
    echo "ERROR: no packages found in $MANIFEST" >&2
    exit 1
fi

# Resolve all packages (listed realms + transitive deps) in topological order.
echo "Resolving dependencies..."
pkg_dirs=$(cd "$EXAMPLES_DIR" && gno tool deplist "${patterns[@]}")

# Build staging directory with resolved packages.
STAGING=$(mktemp -d)
trap 'rm -rf "$STAGING"' EXIT

while IFS= read -r dir; do
    [[ -z "$dir" ]] && continue
    # dir is absolute; compute path relative to examples dir.
    rel="${dir#$EXAMPLES_DIR/}"
    echo "  + $rel"
    mkdir -p "$(dirname "$STAGING/$rel")"
    cp -r "$dir" "$STAGING/$rel"
done <<< "$pkg_dirs"

# Strip test files — not needed in genesis.
find "$STAGING" -name '*_test.gno' -delete
find "$STAGING" -name '*_filetest.gno' -delete

# Generate genesis.
echo ""
echo "Generating genesis for chain=$CHAIN_ID..."
gnogenesis generate \
    -chain-id "$CHAIN_ID" \
    -output-path "$GENESIS_FILE"

echo "Adding placeholder validator..."
gnogenesis validator add \
    -name placeholder-val \
    -power 1 \
    -address g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 \
    -pub-key gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj \
    --genesis-path "$GENESIS_FILE"

echo "Adding packages from staging..."
gnogenesis txs add packages "$STAGING" \
    --genesis-path "$GENESIS_FILE"

# Summary.
realm_count=$(find "$STAGING" -path "*/r/*" -name "gnomod.toml" 2>/dev/null | wc -l)
pkg_count=$(find "$STAGING" -path "*/p/*" -name "gnomod.toml" 2>/dev/null | wc -l)
echo ""
echo "Done: $GENESIS_FILE"
echo "Included: $realm_count realms + $pkg_count pure packages"
