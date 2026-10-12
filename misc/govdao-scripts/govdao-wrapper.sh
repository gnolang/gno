#!/usr/bin/env bash
# Common wrapper logic for govdao scripts.
# Sourced by per-deployment wrappers that set GOVDAO_LABEL and env defaults.
# Expects: GOVDAO_LABEL, GNOKEY_NAME, CHAIN_ID, REMOTE to be exported.
set -eo pipefail

SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ $# -eq 0 ]; then
  echo "govdao — ${GOVDAO_LABEL}"
  echo ""
  echo "  chain-id: $CHAIN_ID"
  echo "  remote:   $REMOTE"
  echo "  key:      $GNOKEY_NAME"
  echo ""
  echo "commands:"
  for script in "$SCRIPTS_DIR"/*.sh; do
    [ "$(basename "$script")" = "govdao-wrapper.sh" ] && continue
    name=$(basename "$script" .sh)
    desc=$(grep -m1 '^# [A-Z]' "$script" | sed 's/^# //')
    printf "  %-35s %s\n" "$name" "$desc"
  done
  exit 0
fi

CMD="$1"
shift

SCRIPT="$SCRIPTS_DIR/${CMD}.sh"
if [ ! -x "$SCRIPT" ]; then
  echo "error: unknown command '$CMD'" >&2
  echo "run '$0' without arguments to list available commands" >&2
  exit 1
fi

exec "$SCRIPT" "$@"
