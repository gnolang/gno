# PR5949: Remove secp256k1 Support for Validators

## Context

Consensus params carry a `Validator.PubKeyTypeURLs` allow-list that gates which
public-key types a validator may use. Historically the chain accepted two types:

- `ed25519.PubKeyEd25519`
- `secp256k1.PubKeySecp256k1`

Both `DefaultValidatorParams()` and the `validatorPubKeyTypeURLs` validation set in
`tm2/pkg/bft/types/params.go` listed secp256k1, and `ValidateConsensusParams`
accepted it. At the execution layer, `validateValidatorUpdates` (in
`tm2/pkg/bft/state/execution.go`) also honored whatever the params allowed, so a
secp256k1 validator could be added to the set.

### Why remove it

**secp256k1 validators break IBC.** The IBC Tendermint light client verifies a
counterparty chain's validator set and their commit signatures. Its consensus
verification path is built around ed25519 validator keys: the light-client
signature-verification and validator-set encoding used across the IBC ecosystem do
not interoperate with secp256k1 consensus keys. A gno.land chain that admitted a
secp256k1 validator would produce headers that an IBC light client cannot verify,
making the chain unbridgeable and putting relayers at risk of accepting or rejecting
headers inconsistently.

Because the validator set is consensus-critical and its accepted key types are fixed
by the consensus params, this must be enforced at the protocol layer rather than left
to operator convention.

Note this is scoped strictly to **validators**. secp256k1 remains a fully supported
key type for user accounts and transaction signing (`std`, `crypto/keys`, ledger,
`SignGenesisTxs`, session accounts); only its use as a consensus/validator key is
removed.

## Decision

Drop secp256k1 from the validator key-type allow-list and from the defaults, leaving
ed25519 as the only accepted validator key type.

In `tm2/pkg/bft/types/params.go`:

- `validatorPubKeyTypeURLs` — the set consulted by `ValidateConsensusParams` — now
  contains only `ed25519.PubKeyEd25519`. A consensus-params / genesis doc that lists
  secp256k1 for validators is now rejected at validation time.
- `DefaultValidatorParams()` returns ed25519-only. New genesis docs that do not
  specify validator params default to ed25519 (via
  `DefaultConsensusParams().Update(...)` in `genesis.go`).
- The now-unused secp256k1 import is removed.

The execution-layer check (`validateValidatorUpdates` →
`ValidatorParams.IsValidPubKeyTypeURL`) needs no change: it already rejects any
validator pubkey type absent from the params, so removing secp256k1 from the
allow-list automatically causes secp256k1 validator updates to be rejected.

### Gating the initial genesis validator set

The params allow-list and the runtime `EndBlocker` gate together cover consensus-params
validation and runtime valset *updates*, but the **initial** validator set is seeded
into consensus state directly by the gno.land `InitChainer`
(`gno.land/pkg/gnoland/app.go`) — it does not pass through `validateValidatorUpdates`.
Nothing else re-checked those genesis validators against the allow-list, so a genesis
doc with ed25519-only params could still seed a secp256k1 validator into the active set,
leaving the "IBC-compatible by construction" guarantee incomplete.

`InitChainer` now scans `req.Validators` against
`req.ConsensusParams.Validator.PubKeyTypeURLs` before writing `valset:current`, using the
same allow-list logic as the `EndBlocker` gate. A disallowed genesis validator is fatal:
as with the existing valoper-coverage assertion, `ResponseInitChain.Error` is silently
discarded by tm2, so the check **panics** to abort the handshake loudly rather than boot
a non-compliant chain. When the allow-list is empty (no `Validator` params configured),
the scan is skipped — matching the `EndBlocker` "accept all" fallback.

This gate lives in the gno.land app layer, not in tm2's `GenesisDoc.Validate` /
`ValidateAndComplete`, because tm2 genesis validation is deliberately key-type-agnostic
(mock and secp256k1 keys are used pervasively in tm2/gnogenesis test fixtures for
non-consensus purposes). The app layer is where the genesis set is actually consumed into
consensus and where the sibling `EndBlocker` gate already lives.

## Alternatives considered

**Leave secp256k1 in the allow-list, document that operators should not use it.**
Relies on convention for a consensus-critical, bridge-breaking property. A single
misconfigured genesis or validator update would silently make the chain
unbridgeable. Rejected.

**Remove the secp256k1 package entirely.**
Out of scope and incorrect: secp256k1 is a legitimate account/transaction key type
and is still used for user keys, ledger, and genesis-tx signing. Rejected.

**Enforce ed25519-only only at the execution layer.**
Would reject secp256k1 validator updates at runtime but still let a genesis doc pass
`ValidateConsensusParams` with secp256k1 listed, leaving misleading params in state.
Enforcing at the params allow-list is the single authoritative gate. Rejected.

**Gate the initial validator set in tm2 `GenesisDoc.ValidateAndComplete`.**
Would reject secp256k1 (and mock) genesis validators at the tm2 layer. Rejected: tm2
genesis validation is intentionally key-type-agnostic — mock/secp256k1 validator keys
are used across tm2 and gnogenesis test fixtures for non-consensus purposes, and a
key-type policy there is both too broad and the wrong layer. The gate belongs where the
genesis set is consumed into consensus (gno.land `InitChainer`), alongside the existing
runtime `EndBlocker` gate.

## Key files

| File | Role |
|------|------|
| `tm2/pkg/bft/types/params.go` | `validatorPubKeyTypeURLs` set, `DefaultValidatorParams`, `ValidateConsensusParams` |
| `tm2/pkg/bft/types/params_test.go` | Validation test asserting secp256k1 validator params are rejected |
| `tm2/pkg/bft/state/execution.go` | `validateValidatorUpdates` — unchanged; honors the allow-list |
| `tm2/pkg/bft/types/genesis.go` | Applies `DefaultConsensusParams()` when genesis omits params |
| `gno.land/pkg/gnoland/app.go` | `InitChainer` — gates the initial genesis validator set against the allow-list (panics on a disallowed key type) |
| `gno.land/pkg/gnoland/app_test.go` | `TestInitChainer_GenesisValidatorPubKeyType` — accepts ed25519, aborts on secp256k1 |
| `gno.land/pkg/integration/testdata/params_valset_rotation_during_pending_proposal.txtar` | Rotation target switched from a secp256k1 to an ed25519 signing key (secp256k1 no longer publishable to the valset) |

## Consequences

- New chains default to ed25519-only validators and are IBC-compatible by
  construction.
- A genesis doc or consensus-params update that lists `/tm.PubKeySecp256k1` under
  `Validator.PubKeyTypeURLs` now fails `ValidateConsensusParams`.
- A `ValidatorUpdate` carrying a secp256k1 pubkey is rejected by
  `validateValidatorUpdates` (unsupported for consensus).
- A genesis doc whose **initial** validator set contains a validator with a disallowed
  (e.g. secp256k1) pubkey type now aborts `InitChain` (panic), so a chain cannot boot with
  a non-compliant validator already in the active set.
- **Backwards compatibility:** an existing chain whose consensus params already list
  secp256k1, or whose active validator set contains a secp256k1 key, is affected —
  its stored params would no longer re-validate and such a validator could not be
  re-added. In practice gno.land validators are ed25519 (the priv-validator keygen
  path uses ed25519), so no live gno.land validator is impacted. Any chain relying on
  secp256k1 validators must migrate those validators to ed25519.
- secp256k1 remains unaffected for accounts, transaction signing, and keys.
