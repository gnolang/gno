# ADR: a devnet home for gnodev

## Status

Proposed.

## Context

With a local package, `gnodev local` redirected `/` to the first one; without
one, `/` showed `r/gnoland/home`, the mainnet home, which describes a chain the
developer is not on. Nothing in the browser said which packages gnodev loaded,
what the last transactions did, or how to reach this chain from `gnokey`, an
agent or the testnet. That lived in the terminal, the docs, or nowhere.

## Decision

In local mode, with the default `-web-home ""` and gnoweb on, gnoweb serves
"Your Gno.land devnet" at `/`, built again on every request.

- **Your packages**: the ones gnodev was started on (`-paths`, the directories
  given or the current one), then the other packages the loader reads from the
  workspace, `-extra-root` trees and gnomod-less directories, realms first;
  five shown, then "and N more". Examples resolved on demand and remote
  packages are not listed. A realm links to its page, Actions and State.
- **Transactions**: the last five, newest first, as function or message kind,
  package, failure and gas, held in memory. A reload replays the transactions,
  so the list stays; a reset drops them, so it clears.
- **This chain**: chain ID and RPC from the flags; then static guidance in
  `devnet_home.md.tmpl` (checks, keys, gnomcp, a call and a testnet deploy filled
  with the first realm's path and folder, docs and community links).

How it is built:

- The page is Go, not a realm: what it shows lives in gnodev, outside the
  chain.
- `devnetHome` wraps the event emitter given to the dev node. It records each
  `TxResult` and passes every event on. A tm2 event-switch listener would not
  survive the node rebuilt on every reload.
- The package list is a snapshot, read again on its own goroutine after each
  reset or reload. Rendering never takes the node's lock, which is held for a
  whole rebuild (seconds on a large workspace), and the event path never waits
  for it either: the node emits those events while holding that lock.
- gnoweb gets one addition: `AliasTarget.Render func() string`. On a
  StaticMarkdown alias, it is called on each request in place of `Value`, in
  HTML and in markdown; its output is the operator's, like `Value`. Every other
  rule for static pages (headers, robots, sitemap, page kind) is unchanged.

`-web-home <path>` still redirects, and `/` or `:none:` still show the chain's
home realm. `gnodev staging` defaults to `:none:` and never builds the page,
even with `-web-home ""`; nor does `-no-web`.

## Alternatives considered

- **Keep the redirect and serve the page at `/devnet`.** No change to the
  default, but few would find it. The developer's realm is the first link of
  the new home, and `-web-home` brings the redirect back.
- **A static page written once at startup.** No gnoweb change, but the package
  list goes stale on the first reload and transactions cannot be shown.
- **Rewrite the alias map on each event.** gnoweb reads that map on every
  request without a lock: a data race.
- **A new `AliasKind`, or an interface field.** A new kind needs a case at
  every switch on the kind for what is still a static page. An interface would
  keep `AliasTarget` comparable, at the cost of a type for one method; nothing
  in the repository compares aliases, so the plain func won.
- **Read transactions from the block store.** A reload replays them in genesis,
  where they are not blocks, so the list would empty on each save.

## Consequences

- A developer who relied on `/` landing on their realm gets the home instead;
  `-web-home <path>` restores the redirect. The docs that said to open `/` to
  see a realm now point at the realm.
- `AliasTarget` holds a func, so code elsewhere can no longer compare two of
  them with `==`.
- The list shows what this gnodev process saw. A transaction rejected in
  simulation never reaches the chain and is not listed. Stepping back with `P`
  keeps undone transactions on the list, and a reset to a saved state clears
  it while the saved transactions stay on chain.
- The guidance hard-codes onyx's chain ID, RPC and deploy fee, the docs links
  and gnomcp's local identity; whoever changes those updates the template, as
  for the docs.
- Security: package paths, transaction fields and flag values reach the page
  only through allowlists (path characters, identifiers, flag characters), so
  nothing a transaction or a folder name carries can add markup, a link or a
  `gno-*` tag, even with `-web-html`.
- Performance: a request reads the snapshot and fills the template, with no RPC
  call. A refresh costs one `filepath.Rel` per loaded package and root, about
  0.3 ms for 500 packages, once per reset or reload. Memory is five
  transactions and five packages.
