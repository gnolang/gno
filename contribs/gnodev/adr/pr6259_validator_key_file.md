# ADR: let gnodev validate with a provided key

## Status

Proposed ([#6259](https://github.com/gnolang/gno/pull/6259)).

## Context

gnodev runs a single-validator in-memory node. `newNodeConfig` in
`pkg/dev/node.go` created the validator identity with `bft.NewMockPV()`, and it
is called on every node rebuild — so the validator key, and with it the
validator address, changed on every reload, reset and restart.

That is invisible to most realm development, but it gets in the way as soon as
something refers to the validator by address or public key: a realm or genesis
state that lists validators, tooling that compares the dev node against a known
validator set, or a test that needs the same identity across runs.

## Decision

Add a `-validator-key-file <file>` flag (both `local` and `staging`). The file
is a `priv_validator_key.json` in the format `gnoland secrets init` writes,
loaded with `signer.LoadFileKey`, which also checks that the public key and
address in the file match the private key.

- `pkg/dev.NodeConfig` gains `ValidatorKey crypto.PrivKey`. When set,
  `newNodeConfig` builds the validator with `bft.NewMockPVWithPrivKey`; when
  nil, behavior is unchanged (a random key per rebuild).
- `setupDevNodeConfig` loads the file and logs the validator address. A missing
  or invalid file fails startup instead of falling back to a generated key.

The config holds the private key rather than a `bft.PrivValidator`: the tm2
node closes its `PrivValidator` on stop, and gnodev stops and rebuilds the node
on every reload, so a fresh one is built from the key each time.

The signer stays the mock one, without a sign-state file. gnodev rebuilds the
chain from genesis on every reload, so it signs heights 1..N again with
different block contents. The file-backed `privval.PrivValidator` exists to
refuse exactly that, and would stop the node from signing after the first
reload.

## Alternatives considered

- **Pass the key on the command line** (hex or base64). Puts a private key in
  shell history and the process list, and invents a second format next to the
  one `gnoland secrets` already produces.
- **Generate the key once and keep it for the process lifetime.** Fixes the
  change across reloads but not across restarts, and does not let the user
  choose the identity.
- **Load or create the file** (`LoadOrMakeFileKey`). A mistyped path would
  silently produce a new key, which defeats the point of pinning one.
- **Take the validator from `-genesis`.** A genesis file carries validator
  public keys, not private ones, so gnodev could not sign as them anyway.

## Consequences

- The validator address is stable across reloads, resets and restarts when the
  flag is set. Without it nothing changes.
- The key gets no double-sign protection. A key that validates on a live
  network must never be passed here: gnodev re-signs the same heights with
  conflicting blocks on every reload. The flag help and the docs say so.
- The node key (p2p identity) is still generated per rebuild in
  `gnoland.NewInMemoryNode`. gnodev never peers, so it is left alone.
