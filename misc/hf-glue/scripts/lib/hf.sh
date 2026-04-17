#!/usr/bin/env bash
# misc/hf-glue/scripts/lib/hf.sh — helpers used by migrate.sh (and friends).
#
# The intent is that migrate.sh reads like a config: each line describes a
# piece of the migration (where to get the genesis, where to get the txs,
# what to patch, etc). Plumbing lives here.
#
# Env set by the caller's Makefile:
#   OUT, REPO  absolute paths
#   ORIGINAL_CHAIN_ID, CHAIN_ID
#   HALT_HEIGHT (may be empty — auto-detected from RPC by hf_fetch_txs_via_rpc)

set -euo pipefail

# ---- state ----------------------------------------------------------------
# Filled by the hf_* functions. Read at the end by hf_assemble.
_HF_STAGE=""           # staging dir (gnoland-data layout)
_HF_STAGE_GEN=""       # path to base genesis.json
_HF_STAGE_TXS=""       # path to historical txs.jsonl (empty until step 2)
_HF_PATCHES=()         # list of "pkgpath=srcdir" entries for --patch-realm
_HF_OVERLAYS=()        # overlay tx files (pre-history, not yet supported)
_HF_MIGRATIONS=()      # migration tx files (post-history, not yet supported)

# ---- presentation ---------------------------------------------------------
hf_banner() {
  printf '\n\033[1;36m━━━ %s ━━━\033[0m\n' "$*"
}

hf_kv() {
  printf "  %-22s \033[36m%s\033[0m\n" "$1" "$2"
}

hf_die() {
  printf '\033[1;31mERROR:\033[0m %s\n' "$*" >&2
  exit 1
}

# ---- setup ----------------------------------------------------------------
# hf_init — must be the first call. Prints a header, creates the staging dir.
hf_init() {
  : "${OUT:?OUT is required}"
  : "${REPO:?REPO is required}"
  : "${ORIGINAL_CHAIN_ID:?ORIGINAL_CHAIN_ID is required}"
  : "${CHAIN_ID:?CHAIN_ID is required}"

  _HF_STAGE="$OUT/source"
  _HF_STAGE_GEN="$_HF_STAGE/config/genesis.json"
  _HF_STAGE_TXS="$_HF_STAGE/txs.jsonl"
  mkdir -p "$_HF_STAGE/config"

  hf_banner "hardfork migration"
  hf_kv "original chain id" "$ORIGINAL_CHAIN_ID"
  hf_kv "new chain id"      "$CHAIN_ID"
  hf_kv "halt height"       "${HALT_HEIGHT:-<auto>}"
  hf_kv "output genesis"    "$OUT/genesis.json"
  hf_kv "staging dir"       "$_HF_STAGE"
  echo ""
}

# ---- step 1: base genesis -------------------------------------------------
# hf_fetch_genesis_from_url URL
#   Direct .json asset (e.g. GitHub release).
hf_fetch_genesis_from_url() {
  local url="$1"
  hf_banner "step 1 — base genesis (URL)"
  if [[ -f "$_HF_STAGE_GEN" ]]; then
    hf_kv "cached" "$(_hf_size "$_HF_STAGE_GEN") bytes"
    return 0
  fi
  hf_kv "url" "$url"
  curl -fSL --retry 3 --retry-delay 5 --max-time 600 --progress-bar \
    -o "$_HF_STAGE_GEN" "$url"
  hf_kv "size" "$(_hf_size "$_HF_STAGE_GEN") bytes"
}

# hf_fetch_genesis_from_rpc RPC_URL
#   Fetches ${RPC_URL}/genesis and unwraps the JSON-RPC envelope.
hf_fetch_genesis_from_rpc() {
  local rpc="$1"
  hf_banner "step 1 — base genesis (RPC)"
  if [[ -f "$_HF_STAGE_GEN" ]]; then
    hf_kv "cached" "$(_hf_size "$_HF_STAGE_GEN") bytes"
    return 0
  fi
  local env="${rpc%/}/genesis"
  hf_kv "url" "$env"
  curl -fSL --retry 3 --retry-delay 5 --max-time 600 --progress-bar \
    -o "$_HF_STAGE/envelope.json" "$env"
  jq -c '.result.genesis' < "$_HF_STAGE/envelope.json" > "$_HF_STAGE_GEN"
  rm -f "$_HF_STAGE/envelope.json"
  hf_kv "size" "$(_hf_size "$_HF_STAGE_GEN") bytes"
}

# hf_fetch_genesis_from_file PATH
#   Local file copy.
hf_fetch_genesis_from_file() {
  local src="$1"
  hf_banner "step 1 — base genesis (file)"
  [[ -f "$src" ]] || hf_die "genesis file not found: $src"
  if [[ -f "$_HF_STAGE_GEN" ]]; then
    hf_kv "cached" "$(_hf_size "$_HF_STAGE_GEN") bytes"
    return 0
  fi
  hf_kv "from" "$src"
  cp "$src" "$_HF_STAGE_GEN"
  hf_kv "size" "$(_hf_size "$_HF_STAGE_GEN") bytes"
}

# ---- step 2: historical txs -----------------------------------------------
# hf_fetch_txs_via_rpc RPC_URL
#   Uses contribs/tx-archive with batching. Auto-detects HALT_HEIGHT from
#   the RPC's /status if HALT_HEIGHT is empty.
hf_fetch_txs_via_rpc() {
  local rpc="$1"
  hf_banner "step 2 — historical txs (RPC)"
  if [[ -z "${HALT_HEIGHT:-}" ]]; then
    HALT_HEIGHT=$(curl -fsS --max-time 30 "${rpc%/}/status" \
      | jq -r '.result.sync_info.latest_block_height')
    hf_kv "halt (auto)" "$HALT_HEIGHT"
  else
    hf_kv "halt" "$HALT_HEIGHT"
  fi
  if [[ -f "$_HF_STAGE_TXS" ]]; then
    hf_kv "cached" "$(wc -l < "$_HF_STAGE_TXS" | tr -d ' ') txs"
    return 0
  fi
  hf_kv "rpc" "$rpc"
  hf_kv "range" "1..$HALT_HEIGHT"
  ( cd "$REPO/contribs/tx-archive" && go run ./cmd backup \
      -remote "$rpc" \
      -from-block 1 \
      -to-block "$HALT_HEIGHT" \
      -batch 1000 \
      -output-path "$_HF_STAGE_TXS" \
      -overwrite )
  hf_kv "total" "$(wc -l < "$_HF_STAGE_TXS" | tr -d ' ') txs"
}

# hf_fetch_txs_from_jsonl PATH
#   Copy a pre-exported txs.jsonl. Still requires HALT_HEIGHT.
hf_fetch_txs_from_jsonl() {
  local src="$1"
  hf_banner "step 2 — historical txs (jsonl)"
  [[ -f "$src" ]] || hf_die "txs.jsonl not found: $src"
  : "${HALT_HEIGHT:?HALT_HEIGHT is required when pulling txs from a file}"
  hf_kv "from" "$src"
  cp "$src" "$_HF_STAGE_TXS"
  hf_kv "total" "$(wc -l < "$_HF_STAGE_TXS" | tr -d ' ') txs"
}

# hf_skip_txs
#   No historical txs at all (genesis-only hardfork).
hf_skip_txs() {
  hf_banner "step 2 — historical txs (none)"
  : "${HALT_HEIGHT:?HALT_HEIGHT is required when skipping tx pull}"
  hf_kv "halt" "$HALT_HEIGHT"
  : > "$_HF_STAGE_TXS"
}

# ---- patches + overlays ---------------------------------------------------
# hf_patch_addpkg PKGPATH SRCDIR
#   Rewrites the genesis-mode addpkg tx for PKGPATH in-place with the
#   *.gno + gnomod.toml files from SRCDIR. Source genesis on disk stays
#   untouched — the patch is applied in memory during hf_assemble.
hf_patch_addpkg() {
  local pkg="$1" src="$2"
  [[ -d "$src" ]] || hf_die "patch srcdir not found: $src"
  _HF_PATCHES+=("$pkg=$src")
}

# hf_overlay_txs PATH
#   Future: inject extra genesis-mode txs BEFORE historical tx replay
#   (post-genesis-mode, pre-history). Not yet plumbed in misc/hardfork —
#   hf_assemble will refuse if any overlay was requested.
hf_overlay_txs() {
  local src="$1"
  _HF_OVERLAYS+=("$src")
}

# hf_migration_tx PATH
#   Future: inject a migration tx that runs AFTER historical replay
#   (e.g. to update r/sys/validators/v2 to the new valset, to reset
#   chain params, etc). Conceptually the last step of the hardfork —
#   "reproduce history, then mutate".
#
#   Currently a no-op placeholder so callers can express intent. Will
#   fail loudly from hf_assemble once any migration tx is registered.
#
#   Valset-swap note: gnoland1 seeds its valset via govdao_prop1.gno
#   at genesis. A hardfork inherits that state, so r/sys/validators/v2
#   still lists the original 7 validators even though tm2 consensus is
#   driven by whatever is in GenesisDoc.Validators. For a coherent fork
#   we want a migration tx that:
#     (a) registers the new valset via a govDAO proposal
#         (r/sys/validators/v2.NewPropRequest + MustCreateProposal +
#          MustVoteOnProposal + ExecuteProposal — signed by a T1 member)
#     (b) optionally drops unneeded members / realms
#   The testbed works today because consensus only reads
#   GenesisDoc.Validators; queries to r/sys/validators/v2 will still
#   return the stale set until this lands.
hf_migration_tx() {
  local src="$1"
  _HF_MIGRATIONS+=("$src")
}

# ---- step 3: assemble -----------------------------------------------------
# hf_assemble
#   Runs `misc/hardfork genesis` against the staged source dir, applying
#   any accumulated --patch-realm entries.
hf_assemble() {
  hf_banner "step 3 — assemble hardfork genesis"
  : "${HALT_HEIGHT:?HALT_HEIGHT must be set (auto-detected earlier, or pass explicitly)}"

  if [[ ${#_HF_OVERLAYS[@]} -gt 0 ]]; then
    hf_die "hf_overlay_txs is not supported by misc/hardfork yet (${#_HF_OVERLAYS[@]} requested)"
  fi
  if [[ ${#_HF_MIGRATIONS[@]} -gt 0 ]]; then
    hf_die "hf_migration_tx is not supported by misc/hardfork yet (${#_HF_MIGRATIONS[@]} requested)"
  fi

  local args=(
    genesis
    --source            "$_HF_STAGE"
    --chain-id          "$CHAIN_ID"
    --original-chain-id "$ORIGINAL_CHAIN_ID"
    --halt-height       "$HALT_HEIGHT"
    --output            "$OUT/genesis.json"
  )
  local p
  for p in "${_HF_PATCHES[@]:-}"; do
    [[ -z "$p" ]] && continue
    hf_kv "patch" "$p"
    args+=(--patch-realm "$p")
  done

  ( cd "$REPO/misc/hardfork" && go run . "${args[@]}" )

  echo ""
  if command -v sha256sum >/dev/null 2>&1; then
    hf_kv "sha256" "$(sha256sum "$OUT/genesis.json" | cut -d' ' -f1)"
  elif command -v shasum >/dev/null 2>&1; then
    hf_kv "sha256" "$(shasum -a 256 "$OUT/genesis.json" | cut -d' ' -f1)"
  fi
  hf_kv "output" "$OUT/genesis.json"
}

# ---- internal -------------------------------------------------------------
_hf_size() { wc -c < "$1" | tr -d ' '; }
