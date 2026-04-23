#!/usr/bin/env bash
# Fetch historical txs from an RPC in chunks with retries, concatenate into
# out/source/txs.jsonl. Use when the RPC is flaky and tx-archive's single
# run can't complete — tx-archive has no resume, so chunking is the only way
# to make progress on unreliable connections.
#
# Inputs (env):
#   RPC_URL       (default https://rpc.gno.land)
#   HALT_HEIGHT   (required)
#   OUT           (default misc/hf-glue/out)
#   REPO          (required — repo root, to locate contribs/tx-archive)
#   CHUNK         (default 20000 blocks per fetch)
#   MAX_RETRIES   (default 8)
set -euo pipefail

: "${REPO:?REPO is required}"
: "${HALT_HEIGHT:?HALT_HEIGHT is required}"
RPC_URL="${RPC_URL:-https://rpc.gno.land}"
OUT="${OUT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/out}"
CHUNK="${CHUNK:-20000}"
MAX_RETRIES="${MAX_RETRIES:-8}"

STAGE_DIR="$OUT/source"
FINAL="$STAGE_DIR/txs.jsonl"
CHUNK_DIR="$STAGE_DIR/txs-chunks"
mkdir -p "$CHUNK_DIR"

echo "── chunked RPC fetch ──────────────────────────────────────────"
echo "  rpc          $RPC_URL"
echo "  range        1..$HALT_HEIGHT"
echo "  chunk size   $CHUNK blocks"
echo "  chunk dir    $CHUNK_DIR"
echo ""

fetch_chunk() {
  local from="$1" to="$2" out_path="$3"
  local attempt=1
  while ((attempt <= MAX_RETRIES)); do
    echo "  [$from..$to] attempt $attempt/$MAX_RETRIES"
    if (cd "$REPO/contribs/tx-archive" && go run ./cmd backup \
      -remote "$RPC_URL" \
      -from-block "$from" \
      -to-block "$to" \
      -batch 100 \
      -output-path "$out_path" \
      -overwrite); then
      return 0
    fi
    echo "  [$from..$to] failed, sleeping $((attempt * 5))s"
    sleep $((attempt * 5))
    ((attempt++))
  done
  return 1
}

from=1
while ((from <= HALT_HEIGHT)); do
  to=$((from + CHUNK - 1))
  ((to > HALT_HEIGHT)) && to=$HALT_HEIGHT
  chunk_file="$CHUNK_DIR/${from}-${to}.jsonl"
  if [[ -s "$chunk_file" ]] || [[ -f "$chunk_file.done" ]]; then
    echo "  [$from..$to] cached ($(wc -l <"$chunk_file" 2>/dev/null | tr -d ' ') txs)"
  else
    if ! fetch_chunk "$from" "$to" "$chunk_file"; then
      echo "ERROR: chunk $from..$to failed after $MAX_RETRIES attempts" >&2
      exit 1
    fi
    touch "$chunk_file.done"
  fi
  from=$((to + 1))
done

echo ""
echo "── assembling final txs.jsonl ─────────────────────────────────"
: >"$FINAL"
for f in $(ls "$CHUNK_DIR"/*.jsonl | sort -t- -k1 -n); do
  cat "$f" >>"$FINAL"
done
echo "  wrote $FINAL"
echo "  total txs: $(wc -l <"$FINAL" | tr -d ' ')"
