#!/usr/bin/env bash
# Generate genesis.json with and without airdrop, then update the Makefile
# checksums (GENESIS_SHA256 and GENESIS_NO_AIRDROP_SHA256).
#
# Usage:
#   ./utils/update-checksums.sh              # full build (run from deployment dir)
#   ./utils/update-checksums.sh --no-install # reuse previously built binaries
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
GEN_SCRIPT="$DEPLOY_DIR/gen-genesis.sh"
MAKEFILE="$DEPLOY_DIR/Makefile"
GENESIS_FILE="$DEPLOY_DIR/genesis.json"
GENESIS_AIRDROP_TMP="$(mktemp /tmp/genesis-airdrop.XXXXXX.json)"
trap 'rm -f "$GENESIS_AIRDROP_TMP"' EXIT

NO_INSTALL_FLAG=""
for arg in "$@"; do
  case "$arg" in
  --no-install) NO_INSTALL_FLAG="--no-install" ;;
  *)
    echo "Unknown argument: $arg"
    exit 1
    ;;
  esac
done

printf "\n=== Step 1/4: Generating genesis with airdrop ===\n"
"$GEN_SCRIPT" $NO_INSTALL_FLAG

GENESIS_SHA256=$(shasum -a 256 "$GENESIS_FILE" | awk '{print $1}')
printf "  SHA256: %s\n" "$GENESIS_SHA256"

printf "\n=== Step 2/4: Saving airdrop genesis ===\n"
cp "$GENESIS_FILE" "$GENESIS_AIRDROP_TMP"
printf "  Saved to %s\n" "$GENESIS_AIRDROP_TMP"

printf "\n=== Step 3/4: Generating genesis without airdrop ===\n"
"$GEN_SCRIPT" --no-airdrop --no-install

GENESIS_NO_AIRDROP_SHA256=$(shasum -a 256 "$GENESIS_FILE" | awk '{print $1}')
printf "  SHA256: %s\n" "$GENESIS_NO_AIRDROP_SHA256"

printf "\n=== Step 4/4: Restoring airdrop genesis and updating Makefile ===\n"
cp "$GENESIS_AIRDROP_TMP" "$GENESIS_FILE"
printf "  Restored airdrop genesis.json\n"

sed -i '' \
  "s|^GENESIS_SHA256 :=.*|GENESIS_SHA256 := $GENESIS_SHA256|" \
  "$MAKEFILE"
sed -i '' \
  "s|^GENESIS_NO_AIRDROP_SHA256 :=.*|GENESIS_NO_AIRDROP_SHA256 := $GENESIS_NO_AIRDROP_SHA256|" \
  "$MAKEFILE"

printf "  Updated Makefile\n"

printf "\n=== Done ===\n"
printf "  GENESIS_SHA256            = %s\n" "$GENESIS_SHA256"
printf "  GENESIS_NO_AIRDROP_SHA256 = %s\n" "$GENESIS_NO_AIRDROP_SHA256"
