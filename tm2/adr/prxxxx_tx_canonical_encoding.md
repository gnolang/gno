# Canonical transaction encoding at tx ingress

## Status

Proposed

## Context

A transaction is identified by `sha256(txBytes)` (`bft/types.Tx.Hash`, the
mempool's `txKey`, `broadcast_tx_*` results, the tx indexer). Its signatures,
however, cover the amino **JSON** sign doc (`std.GetSignaturePayload`), which
is computed from the decoded value and says nothing about the binary bytes
that carried it.

Amino binary, like protobuf, accepts more than one encoding of a value:

- a default-valued field written out instead of omitted (an empty memo as
  `0x22 0x00`, a zero varint, an empty string anywhere in a nested message);
- a non-minimal varint for any tag, length prefix or integer
  (`0x0a` written as `0x8a 0x00`, a length `0x40` as `0xc0 0x00`);
- combinations of the above at every nesting level: the tx, the fee, the
  message `Any`, the message body, each signature.

Both decoder paths (`decodeReflectBinaryStruct` and the genproto2 generated
`UnmarshalBinary2`) accept all of them; the existing strictness only rejects
unknown, duplicate and out-of-order fields.

So anyone who sees a signed tx (mempool gossip, RPC) can re-encode it, and the
variant passes `CheckTx` and `DeliverTx` with the original signatures under a
different hash. Once the variant is included, the sender's own bytes fail on
the consumed sequence, and anything that waits on the hash the sender computed
(`broadcast_tx_commit` callers, indexers, bridges and relayers) never sees the
tx. Nine such mutations were reproduced against a real `gnoland` app
(`gno.land/pkg/gnoland/tx_canonical_test.go` keeps two of them).

While fixing this a second, pre-existing wire mismatch surfaced.
`std.Signature.SessionAddr` is a `crypto.Address` whose amino repr is its
bech32 string, so a zero address is written as the 40-character
`g1qqqq…luuxe` (42 bytes with tag and length) on every master-key signature.
The session design (`gno.land/adr/adr-002-session-subaccounts.md`, the struct
comment, the proto comment "empty for master-key signatures") specifies that
the field is omitted when zero, and the ts-proto based clients
(`tm2-js-client`, `gno-js-client`, `gno-types`) omit it. Go clients and
`amino-ts` wrote it. Both forms decode to the same tx, so the chain accepted
both; a canonical-encoding rule has to pick one.

## Decision

1. **`amino.UnmarshalCanonical(bz, ptr)`**: `Unmarshal`, then `Marshal` the
   decoded value and require the result to equal `bz` byte for byte. It is
   complete by construction: any encoding the encoder would not have produced
   is rejected, whatever the mechanism. `BaseApp.runTx` decodes the tx with
   it in every mode (check, simulate, deliver) and returns
   `std.ErrTxDecode` with the offset of the first differing byte in `Log`.

2. **Zero session address has no wire presence.** `Signature.SessionAddr`
   becomes `std.SessionAddress`, a `crypto.Address` whose `MarshalAmino`
   returns `""` for the zero value, so both encoder paths omit the field in
   binary exactly as they already did in JSON. This realises the documented
   session design and makes the Go canonical form the one the ts-proto
   clients already emit. Call sites convert with `std.SessionAddress(addr)`
   and read with `.Address()`; the generated code needs no change because it
   only calls `MarshalAmino`/`UnmarshalAmino`.

## Alternatives considered

- **A strict amino decoder** that rejects explicit defaults and non-minimal
  varints everywhere. It would apply to persisted state, p2p and consensus
  messages as well as txs, has to be implemented twice (reflection and
  genproto2) and kept in sync, and every non-canonical form has to be
  enumerated by hand. The re-encode check is a few lines and complete. Minimal
  varint enforcement remains a reasonable defense-in-depth follow-up.
- **Hash the canonical re-encoding instead of the raw bytes.** The mempool,
  the block, and the indexer all key on raw bytes, `Tx.Hash` would have to
  decode, and blocks would still carry several byte forms of one tx.
- **Canonicalise in the mempool** (re-encode before admission). The hash a
  client computes from its own bytes would then differ from the chain's, and a
  proposer can bypass the mempool.
- **Enforce in `CheckTx` only.** No consensus change, closes the attack for
  everyone without proposal power, but a byzantine proposer can still include
  a mutated tx. The code enforces in every mode; a staged rollout on a live
  chain would gate this same `runTx` branch on `mode != RunTxModeDeliver` (or
  on a height) until the coordinated upgrade.
- **Keep the written zero session address as canonical.** Every ts-proto
  client would have to start writing `g1qqqq…luuxe`, against the session ADR.
- **Make `crypto.Address` marshal `""` when zero.** It would also change the
  JSON sign doc of any message carrying a zero address, breaking signatures.
- **Treat any `AminoMarshaler` whose Go value is zero as a default field,
  generically.** The rule would reach every repr type in the tree, including
  `std.Coins`, `params.Param` and gnovm's `ObjectID`/`ValueHash` in persisted
  realm state, so it changes app hashes on existing chains, and it would have
  to be mirrored in the genproto2 generator and in JSON.
- **`*crypto.Address` pointer field.** Two encodings for a master-key
  signature (nil and `&zero`) and regenerated code; the dedicated type has one.

## Consequences

- **Consensus-relevant.** `DeliverTx` now rejects bytes older nodes accept, so
  all validators must switch together. Blocks that already contain
  non-canonical txs cannot be replayed under the new rule, and because of the
  session address every master-key tx sent by a Go client since #5307 (April
  2026) is non-canonical, so this ships at a chain restart (genesis txs are
  re-marshalled and therefore canonical) or behind a height gate.
  `CheckTx`/simulate enforcement is safe to deploy at any time.
- **Tx hashes and sizes change for Go clients.** Master-key txs shrink by
  43 bytes (the field plus one byte of the signature length prefix), so at
  `TxSizeCostPerByte` 10 every single-signature tx costs 430 gas less; the
  eight txtar gas pins in `gno.land/pkg/integration/testdata` move by
  exactly that.
- **Clients must emit canonical amino.** Coins strings in Go's rendering
  (sorted by denom, no whitespace), lowercase bech32, no explicit defaults, no
  padded varints. `ParseCoins` still accepts unsorted or padded input on
  decode, but such a tx is now rejected as non-canonical.
- **Sibling repositories**, checked by encoding the same nine txs (`send`,
  memo, multi-denom, `m_call` with and without args/send, `m_addpkg`, `m_run`,
  session-signed, two msgs and two sigs) and diffing bytes against this
  encoder:
  - `tm2-js-client` 3.3.0, `gno-js-client` 3.1.0, `gno-types` 1.1.1
    (ts-proto): all vectors identical.
  - `amino-ts` 0.2.0: identical except the zero session address, which it
    writes like the old Go encoder (`sessionAddr: reprs.Address` in
    `src/gno/types.gen.ts`). It needs a repr that encodes the zero address as
    `""`, mirroring `std.SessionAddress`, before the canonical check goes live.
  - Go repositories import this module: `gnonative` and `tx-indexer` are
    pinned after #5307 and write the zero session address until they bump;
    `supernova` and `faucet` are pinned before #5307, omit the field, and
    already produce the canonical form (they cannot decode session-signed
    txs, which is a separate matter).
  - `ledger-ts` / `ledger-app-gno` sign the JSON sign doc only; unaffected.
- One extra `Marshal` per tx in `runTx`, negligible next to signature
  verification.
