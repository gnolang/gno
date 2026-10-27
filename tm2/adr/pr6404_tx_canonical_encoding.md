# Canonical transaction encoding: the check, and the plan to enforce it

## Status

Accepted for `amino.UnmarshalCanonical`, which this PR adds as a library
function. Enforcement follows in #6406; the wire-format fix the enforcement
depends on is #6399.

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

Fixing this surfaced a second, pre-existing wire mismatch. A zero
`std.Signature.SessionAddr`, which every master-key signature carries, was
written as the 40-character bech32 string `g1qqqq…luuxe` (42 bytes with tag
and length), although the session design (`adr-002-session-subaccounts.md`),
the struct comment and the proto comment all specify omission, and the
ts-proto clients (`tm2-js-client`, `gno-js-client`, `gno-types`) omit it. Go
clients and `amino-ts` wrote it. Both forms decode to the same tx, so the
chain accepted both; a canonical-encoding rule has to pick one, and the
documented one is omission. On gnoland-1 the written-out form is not an edge
case: a reviewer's scan of 213,029 txs up to height 1,066,551 found about 81%
carrying it, and about 85% of recent traffic.

This PR first fixed that field with a dedicated `std.SessionAddress` type whose
zero value marshals to `""`. #6399, opened three days earlier, fixes the same
field with a generic amino tag, `amino:"skip_zero"`, keeps the field's Go type,
and carries codec tests and generator support for the tag. Review of both
preferred #6399 (no Go API change for `std.Signature`; the tag is reusable), so
the session-address fix was removed from this PR in its favour.

## Decision

**`amino.UnmarshalCanonical(bz, ptr)`**: `Unmarshal`, then `Marshal` the
decoded value and require the result to equal `bz` byte for byte. It rejects
any encoding this encoder would not produce for the decoded value, whatever
the mechanism. The converse does not hold: `Marshal` can emit bytes the check
then rejects where a repr type's producer is laxer than its consumer (an
unsorted `std.Coins` string round-trips sorted), and bytes the check accepts
can still be invalid (`0a 00`, a nil message, is the canonical encoding of
`Tx{Msgs: {nil}}`). It ships here as a library function with its tests;
nothing calls it yet.

**Enforcement is a separate PR, #6406**: `BaseApp.runTx` decodes every tx
with `UnmarshalCanonical`, unconditionally. It cannot ship before the wire
fix, because under the current encoder the written-out zero session address
is canonical and the omitted form is not, the reverse of what the clients
emit and of what #6399 makes canonical.

## Rollout

1. #6399 merges and ships as a rolling release: the zero session address
   leaves the wire, nodes accept both encodings as they do today.
2. This PR merges; it changes no behaviour.
3. Clients release: Go tools rebuilt from the module after #6399, `amino-ts`
   (gnolang/amino-ts#2, which mirrors `skip_zero` as a field option); the
   ts-proto clients already emit the canonical form.
4. The written-out zero session address drains from chain traffic.
5. #6406 ships as a coordinated upgrade, the mechanism from #5368 that
   gnoland-1 already runs: a GovDAO halt with `halt_min_version` set to the
   enforcing release and an entry in
   `misc/deployments/mainnet.gno.land/upgrades.json`. Blocks before the halt
   are replayed by the binaries that produced them, so no height gate is
   needed. The release it rides is the next coordinated release on the mainnet
   line after `v1.5.1`; the ledger entry names it when the halt is scheduled.

## Alternatives considered

- **Enforce in this PR.** Rejected on the reviewer's data above: replay of
  gnoland-1 fails from the first non-canonical block, and `CheckTx`
  enforcement alone refuses most live traffic until clients upgrade.
- **A dedicated `std.SessionAddress` type** with an empty repr when zero (this
  PR's first version). Produces the same bytes as #6399 but changes the Go
  type of `Signature.SessionAddr` for gnokey, gnoclient, the ante handler and
  users outside the repo, and a defined type inherits none of
  `crypto.Address`'s methods. Superseded by #6399.
- **A height gate in `runTx`** (a governance param seeded from genesis).
  Implemented and dropped: the house mechanism for a state-breaking change is
  the coordinated halt followed by a binary switch, and the upgrades ledger
  already confines each binary to its own block range, so a gate would
  duplicate in state what replay enforces by switching binaries. There is
  deliberately no gate to look for.
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

## Consequences

- **No behaviour change in this PR.** One exported function and its tests.
- **Once enforcement lands**, clients must emit canonical amino: coins strings
  in Go's rendering (sorted by denom, no whitespace), lowercase bech32, no
  explicit defaults, no padded varints. `ParseCoins` still accepts unsorted
  or padded input on decode, but such a tx is then rejected as non-canonical.
- **Sibling repositories**, checked by encoding nine txs (`send`, memo,
  multi-denom, `m_call` with and without args/send, `m_addpkg`, `m_run`,
  session-signed, two msgs and two sigs) and diffing bytes against an encoder
  that omits the zero session address: `tm2-js-client` 3.3.0,
  `gno-js-client` 3.1.0 and `gno-types` 1.1.1 are identical on all nine;
  `amino-ts` 0.2.0 differed only by the zero session address and is fixed in
  gnolang/amino-ts#2; `gnonative` and `tx-indexer` (pinned after #5307) write
  the old form until they bump this module; `supernova` and `faucet` (pinned
  before #5307) already omit it; `ledger-ts`/`ledger-app-gno` sign the JSON
  sign doc only.
