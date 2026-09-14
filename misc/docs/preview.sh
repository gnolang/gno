#!/usr/bin/env bash
# Boot the docs.gno.land Docusaurus frontend against the local working-tree
# docs/. Invoked by the `preview` target in docs/Makefile.
#
# The rendered site lives in a separate repo (gnolang/docs.gno.land); this repo
# only holds the Markdown. We clone that frontend once, copy the local docs/ into
# it, regenerate the sidebar from the local README, and run its dev server.
#
# The docs are copied (not symlinked): Docusaurus 3.x cannot generate doc
# metadata when its docs path resolves outside the site directory, so a symlink
# makes every page crash with "metadata is undefined". A real in-project copy is
# what the frontend's own deploy script uses too.
#
# Env vars:
#   NO_WATCH=1  disable the 1s re-sync poll (preview a static snapshot instead
#               of hot-reloading edits — on by default since a frozen preview
#               looks identical to a live one)
#   UPDATE=1    git pull the frontend clone before starting
set -euo pipefail

FRONTEND_REPO="https://github.com/gnolang/docs.gno.land"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GNO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DOCS_DIR="$GNO_ROOT/docs"
CLONE_DIR="$DOCS_DIR/.preview/docs.gno.land"
DEST="$CLONE_DIR/docs"

# Safety: sync_docs runs `rsync --delete` into $DEST (derived from this script's
# location). Confirm the path derivation landed in the gnolang/gno repo before
# touching anything, so a moved or symlinked script can't rsync an unexpected
# tree. The module line is structural and stable; if it matches, $GNO_ROOT/docs
# is by definition the gno docs. (--delete only prunes the clone under $DEST;
# it never deletes from the source $DOCS_DIR.)
grep -qxF 'module github.com/gnolang/gno' "$GNO_ROOT/go.mod" 2>/dev/null || {
	echo "error: $GNO_ROOT is not the gnolang/gno repo root; refusing to run" >&2
	exit 1
}
# And the rsync source must exist and be a directory named docs/. `rsync --delete`
# makes $DEST mirror the source, so fail fast here if $DOCS_DIR is missing or
# resolved to the wrong tree, rather than letting rsync wipe/corrupt the clone.
# ($DOCS_DIR is set to $GNO_ROOT/docs above, so this also catches a future edit
# that points it somewhere else.)
[ -d "$DOCS_DIR" ] && [ "${DOCS_DIR##*/}" = docs ] || {
	echo "error: $DOCS_DIR is not a docs/ directory; refusing to run" >&2
	exit 1
}

# 1. Clone the frontend once; reuse it afterwards.
if [ ! -d "$CLONE_DIR/.git" ]; then
	echo "==> Cloning $FRONTEND_REPO"
	mkdir -p "$(dirname "$CLONE_DIR")"
	git clone --depth 1 "$FRONTEND_REPO" "$CLONE_DIR"
fi

if [ -n "${UPDATE:-}" ]; then
	echo "==> Updating frontend clone"
	git -C "$CLONE_DIR" pull --ff-only
fi

# 2. Copy the local docs into the clone (everything but the nested .preview clone
#    and the generated sidebar). --delete keeps it in sync across re-runs.
sync_docs() {
	rsync -a --delete --exclude='.preview' --exclude='sidebar.json' "$DOCS_DIR/" "$DEST/"
}
echo "==> Copying docs into the frontend"
mkdir -p "$DEST"
sync_docs

# 3. Regenerate the sidebar from the local README into the copy, where
#    sidebars.js reads it (docs/sidebar.json). Only the indexparser half of
#    `make generate` — not embedmd, which would rewrite the .md files in place.
echo "==> Generating sidebar"
go run -C "$GNO_ROOT/misc/docs/tools/indexparser" . -path "$DEST/README.md" > "$DEST/sidebar.json"

# 4. Re-sync the copy on a 1s poll so working-tree edits hot-reload (no inotify).
#    On by default; NO_WATCH=1 skips it for a static snapshot. `|| true` keeps the
#    loop alive on a mid-copy rsync failure; it stops when the server exits.
if [ -z "${NO_WATCH:-}" ]; then
	echo "==> Watching docs/ for changes (re-sync every 1s)"
	( while sleep 1; do sync_docs || true; done ) &
	watch_pid=$!
	trap 'kill "$watch_pid" 2>/dev/null || true' EXIT
fi

# 5. Run the dev server on the Node version the frontend pins (.nvmrc). Prefer
#    fnm to select it; otherwise require the current Node's major to match.
NODE_VERSION="$(tr -dc '0-9.' < "$CLONE_DIR/.nvmrc")"
run_node() {
	if command -v fnm >/dev/null 2>&1; then
		fnm install "$NODE_VERSION" >/dev/null 2>&1 || true
		fnm exec --using="$NODE_VERSION" -- "$@"
	else
		"$@"
	fi
}

if ! command -v fnm >/dev/null 2>&1; then
	case "$(node -v 2>/dev/null)" in
		v"${NODE_VERSION%%.*}".*) ;;
		*) echo "error: frontend needs Node v$NODE_VERSION; found $(node -v 2>/dev/null || echo none). Install fnm or switch Node." >&2; exit 1 ;;
	esac
fi

cd "$CLONE_DIR/docusaurus"
echo "==> Installing dependencies"
run_node corepack yarn install
echo "==> Starting dev server at http://localhost:3000/"
run_node corepack yarn start
