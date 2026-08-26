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

**secp256k1 validators break the IBC GNO light client.** The IBC Tendermint light
client verifies a counterparty chain's validator set and their commit signatures.
IBC core does not in principle forbid secp256k1 consensus keys — ed25519 is the
de-facto interoperable consensus key across the Tendermint/IBC ecosystem, but the
concrete restriction is in **our Gno IBC light-client implementation**, whose
signature-verification and validator-set encoding are built around ed25519. A
gno.land chain that admitted a secp256k1 validator would produce headers that this
light client cannot verify, making the chain unbridgeable and putting relayers at
risk of accepting or rejecting headers inconsistently.

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

### Rejecting disallowed validator keys at registration (valopers)

The `EndBlocker` gate rejects a disallowed valset *update*, but it does so **silently**:
`r/gnops/valopers`'s `Register` / `UpdateSigningKey` only validated that the signing key
was well-formed bech32, not that its *type* is accepted for consensus. So an operator
could register or rotate to a secp256k1 signing key, see the tx succeed, and only later
have the `EndBlocker` drop the resulting valset update. Worst case: an operator rotates,
sees success, swaps its `priv_validator_key.json` to the new key, and — because consensus
never accepted the update — signs with a key the set doesn't expect and goes dark.

To fail fast, `valopers` now rejects a signing key whose type is not in the allow-list at
registration/rotation time. Because the allow-list lives in consensus params (which gno
realms cannot read directly), the chain **mirrors** it into the params store:

- `InitChainer` writes `ConsensusParams.Validator.PubKeyTypeURLs` to the chain-managed
  `node:valset:pubkey_types` key (alongside `valset:current`). Consensus params are
  immutable after genesis, so this write-once mirror never drifts.
- `r/sys/params.GetValsetPubKeyTypes()` exposes the mirror to realms (a thin param read;
  it deliberately adds **no** new imports to that widely-imported realm).
- `valopers` reads the allow-list and classifies the bech32 signing key locally (it
  already imports `crypto/bech32`): the amino type URL is carried verbatim as the first
  field of the encoded key, so it is recoverable without an amino decoder. An empty
  allow-list means "accept all", matching the `EndBlocker` fallback.

**Cost.** Decoding the ~103-char `gpub` in the interpreted GnoVM adds ~4–7M gas per
`Register` / `UpdateSigningKey`; affected txtar fixtures had their `-gas-wanted` raised
accordingly. This was preferred over adding a native `chain` primitive (which would be
gas-cheap but requires regenerating stdlib bindings) to keep the change confined to gno.

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

**Expose the allow-list to realms via a native `chain` primitive.**
A `chain.ValidatorPubKeyTypes()` (and a native pubkey-type classifier) would let
`valopers` validate keys gas-cheaply. Rejected here to avoid adding public `chain` stdlib
surface and regenerating native bindings (`go generate`). Instead the allow-list is
mirrored into the params store and the realm classifies the key in pure gno. The tradeoff
is gas: the in-gno bech32 decode costs ~4–7M gas per valoper key op. A native primitive
remains a reasonable future optimization if that cost matters.

## Key files

| File | Role |
|------|------|
| `tm2/pkg/bft/types/params.go` | `validatorPubKeyTypeURLs` set, `DefaultValidatorParams`, `ValidateConsensusParams` |
| `tm2/pkg/bft/types/params_test.go` | Validation test asserting secp256k1 validator params are rejected |
| `tm2/pkg/bft/state/execution.go` | `validateValidatorUpdates` — unchanged; honors the allow-list |
| `tm2/pkg/bft/types/genesis.go` | Applies `DefaultConsensusParams()` when genesis omits params |
| `gno.land/pkg/gnoland/app.go` | `InitChainer` — gates the initial genesis validator set against the allow-list (panics on a disallowed key type); mirrors the allow-list to `node:valset:pubkey_types` |
| `gno.land/pkg/gnoland/node_params.go` | `valsetPubKeyTypesPath` const + chain-only `WillSetParam` gate for the mirror |
| `gno.land/pkg/gnoland/app_test.go` | `TestInitChainer_GenesisValidatorPubKeyType` — accepts ed25519, aborts on secp256k1 |
| `examples/gno.land/r/sys/params/valset.gno` | `GetValsetPubKeyTypes()` — exposes the mirrored allow-list to realms |
| `examples/gno.land/r/gnops/valopers/valopers.gno` | `Register`/`UpdateSigningKey` reject a signing key whose type is not allowed |
| `gno.land/pkg/integration/testdata/*.txtar` | secp256k1 *validator* signing keys swapped to ed25519; `valopers.txtar` asserts secp256k1 registration is rejected and queries the mirror; affected fixtures' `-gas-wanted` raised for the added check |

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
- `r/gnops/valopers.Register` / `UpdateSigningKey` reject a secp256k1 (or otherwise
  disallowed) signing key at tx time with a clear error, instead of succeeding and having
  the `EndBlocker` silently drop the eventual valset update. This adds ~4–7M gas to those
  calls (in-gno bech32 decode of the signing key).
- **Backwards compatibility:** an existing chain whose consensus params already list
  secp256k1, or whose active validator set contains a secp256k1 key, is affected —
  its stored params would no longer re-validate and such a validator could not be
  re-added. In practice gno.land validators are ed25519 (the priv-validator keygen
  path uses ed25519), so no live gno.land validator is impacted. Any chain relying on
  secp256k1 validators must migrate those validators to ed25519.
- secp256k1 remains unaffected for accounts, transaction signing, and keys.
