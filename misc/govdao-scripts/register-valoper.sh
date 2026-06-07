#!/usr/bin/env bash
# Self-register a valoper profile in r/gnops/valopers.
#
# This is the prerequisite for governance to promote an operator into
# the active valset via add-validator.sh — v3's
# NewValidatorProposalRequest rejects any operator that isn't already
# in valoperCache, which is populated only by valopers.Register and
# valopers.UpdateKeepRunning.
#
# Auth: at runtime (ChainHeight > 0), r/gnops/valopers enforces
# ErrOperatorSquatGuard (caller == operator address). $GNOKEY_NAME's
# address MUST equal the <operator_address> argument; this is why this
# script is signed by the operator themselves, NOT by the GovDAO T1.
#
# Usage:
#   ./register-valoper.sh <moniker> <description> <server_type> \
#                         <operator_address> <signing_pubkey>
#
# server_type: one of "cloud", "on-prem", "data-center" (validated by
# the realm).
#
# Environment: see README.md. Override $GNOKEY_NAME to the operator's
# local key name.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:?GNOKEY_NAME is required}"
CHAIN_ID="${CHAIN_ID:?CHAIN_ID is required}"
REMOTE="${REMOTE:?REMOTE is required}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

if [ $# -lt 5 ]; then
  echo "Usage: $0 <moniker> <description> <server_type> <operator_address> <signing_pubkey>"
  echo ""
  echo "Example:"
  echo "  $0 'aeddi-1' 'Aeddi node #1' cloud \\"
  echo "     g1s2ht24e85qq3t66gc9sgdvk5kzc38yy68aaqvr \\"
  echo "     gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqfr74tgql2cvzadga2uts62v3f8a5dx66dauaq6sphg3ynuhgl286cce2mn"
  exit 1
fi

MONIKER="$1"
DESCRIPTION="$2"
SERVER_TYPE="$3"
OPERATOR_ADDR="$4"
SIGNING_PUBKEY="$5"

echo "Registering valoper: ${MONIKER} (${SERVER_TYPE})"
echo "  Operator addr:  ${OPERATOR_ADDR}"
echo "  Signing pubkey: ${SIGNING_PUBKEY}"
echo "  Key:    ${GNOKEY_NAME}   (must control ${OPERATOR_ADDR} — squat guard)"
echo "  Chain:  ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey maketx call \
  --pkgpath gno.land/r/gnops/valopers \
  --func Register \
  --args "$MONIKER" \
  --args "$DESCRIPTION" \
  --args "$SERVER_TYPE" \
  --args "$OPERATOR_ADDR" \
  --args "$SIGNING_PUBKEY" \
  --gas-wanted "$GAS_WANTED" \
  --gas-fee "$GAS_FEE" \
  --chainid "$CHAIN_ID" \
  --remote "$REMOTE" \
  --broadcast \
  "$GNOKEY_NAME"
