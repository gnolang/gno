# ADR: gno.land Network Registry API

## Context

Network configuration metadata for gno.land (Betanet, Staging, Test11, future
testnets) is currently duplicated in several unrelated places:

- `docs/resources/gnoland-networks.md` — a static Markdown table.
- CLI tools such as `gnokey` and `gnodev` that accept a `--remote` flag
  without any built-in awareness of which chain a given RPC endpoint belongs
  to.
- Downstream consumers (wallets, explorers, docs sites) that hardcode their
  own copies of the list.

Every time a testnet is rotated (e.g. Test11 → Test12) each copy goes stale
independently, and there is no programmatic way for a tool to ask "which
networks exist right now?". Recent work in this area (#5543, #5557, the
upcoming interactive `gnokey maketx` wizard) has reinforced the need for a
single, machine-readable source of truth that ships with the monorepo.

See issue [#5584](https://github.com/gnolang/gno/issues/5584) for the
original proposal.

## Decision

Introduce a canonical network registry inside the monorepo and expose it
over HTTP from gnoweb.

### Canonical data file

- Location: `gno.land/pkg/networks/networks.json`.
- The file is the single source of truth for the list of known gno.land
  networks.
- Schema per entry: `name`, `chain_id`, `rpc_endpoint`, `gnoweb_url`
  (optional), `faucet_url` (optional), `status` (`active` | `deprecated` |
  `offline`).
- Updating a testnet is a one-line change to this JSON file; all consumers
  pick it up automatically.

### Go package

`gno.land/pkg/networks` exposes two minimal entry points so consumers do not
need to reimplement parsing:

- `networks.Raw()` — returns the embedded JSON bytes verbatim (used by
  gnoweb to avoid a round-trip through `encoding/json`).
- `networks.Load()` — returns a typed `Registry` for programmatic
  consumption (future CLI clients, tests, etc.).

### HTTP endpoint

gnoweb serves the registry at `GET /api/networks` via a new handler in
`gno.land/pkg/gnoweb/networks.go`, wired into `NewRouter` alongside the
existing `/status.json`, `/liveness`, and `/ready` handlers. The response
body is the embedded JSON verbatim; `Content-Type: application/json` and a
5-minute `Cache-Control` header are set. Tools that want live data can hit
`https://gno.land/api/networks`, cache it, and fall back to their own
hardcoded defaults when offline.

### Staleness detection

`networks_live_test.go` in `gno.land/pkg/networks` hits each `active`
network's `/status` endpoint and asserts the reported `chain_id` matches the
registry. The test skips under `-short` so default `go test ./...` runs stay
hermetic; CI can run it on a schedule (or contributors locally) to catch a
renamed/retired testnet or a wrong RPC URL before users do.

### Documentation

`docs/resources/gnoland-networks.md` now points at the canonical JSON file
and the HTTP endpoint and instructs contributors to update the JSON — not
the Markdown table — when networks change. The table remains for human
readers.

## Alternatives considered

1. **Keep the Markdown table as the source of truth.** Rejected: not
   machine-readable, easy to forget to update, can't be served over HTTP
   without scraping.
2. **Put `networks.json` at the repo root (`gno.land/networks.json`).**
   Matches the exact path in the issue, but Go's `//go:embed` cannot reach
   a sibling directory. We'd need a build step to copy/generate it into
   gnoweb, or a `go generate` pass. Keeping the canonical file inside a Go
   package avoids that complexity while preserving the spirit of the
   proposal (one file, one source of truth).
3. **Query the chain for network metadata.** Would require on-chain
   configuration that does not exist today and could not describe *other*
   networks (by definition a running node only knows itself). A monorepo
   registry is additive and strictly simpler.

## Consequences

### Positive

- Contributors have one place to update when rotating a testnet.
- CLI tools, wallets, and explorers can fetch a fresh list at runtime with a
  plain HTTP GET, and fall back to a hardcoded copy when offline.
- Interactive flows (the upcoming `gnokey maketx` wizard, `gno init`
  prompts) can show a live, curated network picker without hardcoding one
  per tool.
- The registry schema is intentionally small; extending it with new fields
  (websocket URL, explorer URL, deprecation notice, etc.) is a backwards
  compatible JSON addition.

### Negative / follow-ups

- CLI fetch/cache wiring for `gnokey` and `gnodev` is deliberately out of
  scope for this PR — each tool will integrate the registry in a dedicated
  follow-up so reviewers can focus on one consumer at a time.
- Downstream integrations (wallets, explorers) need to be notified that the
  registry exists; this ADR and the updated docs page are the initial
  signal.
- The `/api/networks` payload is served from a compiled-in copy; any change
  requires a gnoweb redeploy. This matches how the existing Markdown table
  worked and is acceptable given how infrequently networks change.
