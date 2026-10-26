# Canonical transaction encoding: wire format first, enforcement staged

## Status

Accepted for the encoding change in this PR. Enforcement follows in a
separate, height-gated PR (see "Rollout").

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
tx. Nine such mutations were reproduced against a real `gnoland` app.

Fixing this surfaced a second, pre-existing wire mismatch.
`std.Signature.SessionAddr` is a `crypto.Address` whose amino repr is its
bech32 string, so a zero address is written as the 40-character
`g1qqqq…luuxe` (42 bytes with tag and length) on every master-key signature.
The session design (`gno.land/adr/adr-002-session-subaccounts.md`, the struct
comment, the proto comment "empty for master-key signatures") specifies that
the field is omitted when zero, and the ts-proto based clients
(`tm2-js-client`, `gno-js-client`, `gno-types`) omit it. Go clients and
`amino-ts` wrote it. Both forms decode to the same tx, so the chain accepts
both; a canonical-encoding rule has to pick one.

On gnoland-1 the written-out form is not an edge case: a reviewer's scan of
213,029 txs up to height 1,066,551 found about 81% carrying it, and about 85%
of recent traffic. Enforcing the canonical form immediately would therefore
break block replay and refuse most live traffic until clients upgrade.

## Decision

### This PR: the wire format, backward compatible

1. **Zero session address has no wire presence.** `Signature.SessionAddr`
   becomes `std.SessionAddress`, a `crypto.Address` whose `MarshalAmino`
   returns `""` for the zero value, so both encoder paths omit the field in
   binary exactly as they already did in JSON. This realises the documented
   session design and makes the Go canonical form the one the ts-proto clients
   already emit. Call sites convert with `std.SessionAddress(addr)` and read
   with `.Address()`; the generated code needs no change because it only calls
   `MarshalAmino`/`UnmarshalAmino`. Decoding is unchanged: an absent field and
   `""` both decode to the zero address, so nodes keep accepting the old form.

2. **`amino.UnmarshalCanonical(bz, ptr)`**: `Unmarshal`, then `Marshal` the
   decoded value and require the result to equal `bz` byte for byte. It
   rejects any encoding this encoder would not produce for the decoded value,
   whatever the mechanism. The converse does not hold: `Marshal` can emit
   bytes the check then rejects where a repr type's producer is laxer than its
   consumer (an unsorted `std.Coins` string round-trips sorted), and bytes
   the check accepts can still be invalid (`0a 00`, a nil message, is the
   canonical encoding of `Tx{Msgs: {nil}}`). It ships here as a library
   function with its tests; nothing calls it yet.

### Follow-up PR: enforcement, unconditional in the enforcing binary

`BaseApp.runTx` decodes with `UnmarshalCanonical` in every mode, with no
height gate. History that predates the rule is not replayed by that binary:
the upgrades ledger (`gno.land/pkg/upgrades`) binds each released version to
the block range from its own halt height to its successor's, and a node
replaying history switches binaries at those heights. The enforcing release
therefore only ever executes blocks produced after its halt height, where
every included tx is canonical by construction.

## Rollout

1. Merge and release this PR. Nodes accept both encodings, as they do today.
2. Release the clients: Go tools rebuilt from this module, `amino-ts`
   (gnolang/amino-ts#2); the ts-proto clients already emit the canonical form.
3. Wait for the written-out zero session address to drain from chain traffic.
4. Ship the follow-up (#6406) as a coordinated upgrade, the mechanism from
   #5368 that gnoland-1 already runs: a GovDAO halt with `halt_min_version`
   set to the enforcing release, so an old binary cannot restart past the
   halt and a new one cannot start early, and an entry in
   `misc/deployments/mainnet.gno.land/upgrades.json` so replay switches
   binaries at that height. The release it rides is the next coordinated
   release on the mainnet line after `v1.5.1`, the current rolling release;
   the ledger entry names it when the halt is scheduled.

## Alternatives considered

- **Enforce in this PR.** Rejected on the reviewer's data above: replay of
  gnoland-1 fails from the first non-canonical block, and `CheckTx`
  enforcement alone refuses most live traffic until clients upgrade.
- **A height gate in `runTx`** (a governance param `node:p:canonical_tx_height`,
  seeded from genesis for fresh chains). Implemented and dropped: the house
  mechanism for a state-breaking change is the coordinated halt followed by a
  binary switch, and the upgrades ledger already confines each binary to its
  own block range, so a gate would duplicate in state what replay enforces by
  switching binaries, at the cost of a genesis field and a param on every
  fresh chain. There is deliberately no gate to look for.
- **A strict amino decoder** that rejects explicit defaults and non-minimal
  varints everywhere. It would apply to persisted state, p2p and consensus
  messages as well as txs, has to be implemented twice (reflection and
  genproto2) and kept in sync, and every non-canonical form has to be
  enumerated by hand. The re-encode check is a few lines and complete.
- **Hash the canonical re-encoding instead of the raw bytes.** The mempool,
  the block, and the indexer all key on raw bytes, `Tx.Hash` would have to
  decode, and blocks would still carry several byte forms of one tx.
- **Canonicalise in the mempool.** The hash a client computes from its own
  bytes would then differ from the chain's, and a proposer can bypass the
  mempool.
- **Keep the written zero session address as canonical.** Every ts-proto
  client would have to start writing `g1qqqq…luuxe`, against the session ADR.
- **Make `crypto.Address` marshal `""` when zero.** It would also change the
  JSON sign doc of any message carrying a zero address, breaking signatures,
  and the binary encoding of every struct with a possibly-zero address
  (`auth.Params.FeeCollector`, `bft.Header.ProposerAddress`, …).
- **Treat any `AminoMarshaler` whose Go value is zero as a default field,
  generically.** The rule would reach every repr type in the tree, including
  `std.Coins`, `params.Param` and gnovm's `ObjectID`/`ValueHash` in persisted
  realm state, so it changes app hashes on existing chains, and it would have
  to be mirrored in the genproto2 generator and in JSON.
- **`*crypto.Address` pointer field.** Two encodings for a master-key
  signature (nil and `&zero`) and regenerated code; the dedicated type has one.

## Consequences

- **No consensus change in this PR.** Old and new nodes decode both forms to
  the same tx; tx-size gas is charged on the bytes as received, identically on
  every node.
- **Tx bytes and hashes change for Go clients.** Master-key txs shrink by
  43 bytes (the field plus one byte of the signature length prefix), so at
  `TxSizeCostPerByte` 10 every single-signature tx costs 430 gas less; the
  12 pinned `GAS USED` values in 8 txtar files under
  `gno.land/pkg/integration/testdata`, and the two gas-estimate lines in
  `gnokey_gasfee.txtar`, move by exactly that.
- **Go API.** `Signature.SessionAddr` changes type; the few call sites convert.
- **Sibling repositories**, checked by encoding nine txs (`send`, memo,
  multi-denom, `m_call` with and without args/send, `m_addpkg`, `m_run`,
  session-signed, two msgs and two sigs) and diffing bytes against this
  encoder: `tm2-js-client` 3.3.0, `gno-js-client` 3.1.0 and `gno-types` 1.1.1
  are identical on all nine; `amino-ts` 0.2.0 differed only by the zero
  session address and is fixed in gnolang/amino-ts#2; `gnonative` and
  `tx-indexer` (pinned after #5307) write the old form until they bump this
  module; `supernova` and `faucet` (pinned before #5307) already omit it;
  `ledger-ts`/`ledger-app-gno` sign the JSON sign doc only.
- **Once enforcement lands**, clients must emit canonical amino: coins strings
  in Go's rendering (sorted by denom, no whitespace), lowercase bech32, no
  explicit defaults, no padded varints.
