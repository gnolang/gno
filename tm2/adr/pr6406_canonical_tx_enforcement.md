# Canonical transaction encoding, enforced

## Status

Proposed. Follows `tm2/adr/pr6404_tx_canonical_encoding.md`, which analysed
the problem and added `amino.UnmarshalCanonical` (merged 2026-11-03), and
#6399, which takes the zero session address off the wire with
`amino:"skip_zero"` (merged 2026-11-02).

## Context

#6404 established that any re-encoding of a signed transaction was accepted
under a different hash, because the hash covers the raw amino bytes while the
signatures cover the JSON sign doc, and that the fix is to accept only the
canonical encoding in `runTx`. It also established why that could not ship
with the encoding change: on gnoland-1 about 81% of historical transactions
and about 85% of then-live traffic carried the written-out zero session
address of the old Go encoder.

How gno.land replays history settles the remaining question. The upgrades
ledger (`gno.land/pkg/upgrades`) binds each released version to one block
range, from its own halt height to its successor's, and a node replaying
history switches binaries at those heights. A release is never asked to
execute blocks produced under an earlier rule set.

## Decision

`BaseApp.runTx` decodes every tx, in every mode, with `amino.Unmarshal`, runs
`validateBasicTxMsgs` on it, and only then checks with `amino.CheckCanonical`
that the bytes are the tx's canonical encoding, returning `std.ErrTxDecode`
with the first differing byte in `Log` when they are not. `CheckCanonical` is
the second half of `amino.UnmarshalCanonical`, split out so that validation
comes first: the check re-encodes the tx, and for some values the re-encoding
costs far more than the decoding did (see the cost under Consequences). There
is no height gate and no new parameter.

`validateBasicTxMsgs` rejects a nil message with `std.ErrUnknownRequest`.
An empty `Any` (`0a 00`) decodes to a nil message and is that message's
canonical encoding, so the check passes it; before, `ValidateBasic`
dereferenced it and the recover in `runTx` reported an internal error, at
about sixteen times the CPU of a decode error and charged no gas. The
explicit check closes that path (found in review of #6404).

`vm.MsgAddPackage.ValidateBasic` and `vm.MsgRun.ValidateBasic` refuse a nil
`Package` with `ErrInvalidPkgPath`. Such a message is non-nil and in its
canonical encoding, so it passes both checks above; `ValidateBasic` then
dereferenced it, and the recovered panic cost about thirteen times a clean
rejection (found in review of this change). The other seven messages gno.land
registers already return clean errors. A delivered tx's error type is part of
its result, and so of the block's results hash, which is why the guard rides
this halt rather than a rolling release.

The same two methods refuse, with `ErrInvalidPackage`, a non-nil
`MemPackage.Info` and a `MemPackage.Type` that is not a `gno.MemPackageType`.
Both fields are typed `any`, so they decode to any registered type. The keeper
already refuses `Info` at execution and overwrites `Type`, and gnokey sends
`Type` as a `MemPackageType`, so no client loses anything; what the check buys
is that no other value reaches the re-encoding (found in review of this
change).

The rule reaches a chain through the coordinated halt mechanism from #5368,
which gnoland-1 already runs: a GovDAO halt with `halt_min_version` set to
the enforcing release, so an old binary cannot restart past the halt and a
new one cannot start early, with the entry in the chain's upgrades ledger
(`misc/deployments/mainnet.gno.land/upgrades.json`) written once the chain
runs the release (#6381). From block `halt_height + 2` on, every included tx
is canonical by construction. Block `halt_height + 1` is filled by the old
binaries' mempools and executed by the enforcing release (see "What a halt
actually looks like" in `gno.land/cmd/gnoland/UPGRADES.md`), so a non-canonical
tx they admitted lands there and fails as a decode error. Blocks up to the
halt height are replayed by the binaries that produced them. The release it
rides is the next coordinated release on the
mainnet line after `v1.5.1`, the current rolling release; its release page
and the halt announcement name the height once the proposal has executed.
Merge order:
#6399 and #6404 first, both merged, then this change, then the related `runTx`
tx-encoding change queued behind it, so the interaction is settled before
merge rather than discovered at it.

## Alternatives considered

- **A height gate in `runTx`** (a governance param `node:p:canonical_tx_height`,
  seeded from a genesis field for fresh chains). Implemented and dropped: the
  house mechanism for a state-breaking change is the coordinated halt followed
  by a binary switch, and the ledger already confines each binary to its own
  block range, so a gate would duplicate in state what replay enforces by
  switching binaries, and it adds a genesis field and a param to every fresh
  chain. There is deliberately no gate to look for.
- **`CheckTx` only.** No consensus change, but a byzantine proposer can still
  include a mutated tx, and the mempool refuses old clients' txs just as early.
- **A height in `config.toml`**, like the per-node `halt_height` fallback.
  Every operator would set the same value by hand, and a mismatch is a
  consensus failure.

## Consequences

- **State-breaking.** Ships as a coordinated upgrade, never as a rolling
  release, and only after the written-out zero session address has drained
  from traffic: Go tools rebuilt from the module after #6399, `amino-ts`
  0.2.1 or later (gnolang/amino-ts#2, #4); the ts-proto clients already emit the canonical form.
  Every master-key tx from an older Go client is refused from the first block
  of the enforcing release.
- **Replay needs the earlier binaries.** A node that syncs history from
  genesis on the enforcing release alone fails at the first non-canonical
  block; it has to switch binaries at the ledger's heights, as it does for
  every other state-breaking release.
- **Fresh chains are strict from block 1.** The whole txtar suite runs that
  way; gnokey emits canonical bytes.
- **Clients** must emit canonical amino: coins strings in Go's rendering
  (sorted by denom, no whitespace), lowercase bech32, no explicit defaults,
  no padded varints. The decode error names the first differing byte.
- **The check costs about as much again as decoding, on a tx that passes
  `ValidateBasic`**, and on the inputs an attacker picks signature
  verification is never reached, so that is the relevant comparison (measured
  in review of #6404, `Unmarshal` against `UnmarshalCanonical` on the same
  bytes): 1 KB memo 3.5 to 5.6 µs, 100 KB memo 23.5 to 32.3 µs, 999 KB memo
  121.5 to 190.0 µs, 797 KB with 7000 messages 8.1 to 14.9 ms. It is bounded:
  the mempool rejects txs over `maxTxBytes` before `CheckTx` reaches the app,
  and `Marshal` runs only after `Unmarshal` and `validateBasicTxMsgs` succeeded,
  so garbage and refused messages stay cheap. The order is what bounds it. Run
  before validation, the re-encoding took a 113-byte unsigned `MsgAddPackage`
  whose `Info` holds a `BigdecValue` written `1e999999` from 21 ms of
  `CheckTx` to 102 ms, and a `BigintValue` written as 499,998 hex digits from
  3.4 ms of decoding to 166 ms. The 21 ms that remains is the decode itself,
  and predates this change: `BigdecValue.UnmarshalAmino` accepts an exponent
  form `MarshalAmino` never emits.
- **`Log` on decode failures** carries the decode error's message, clipped,
  where it used to be empty. It deliberately skips `ABCIResultFromError`,
  whose `%#v` adds a stack trace: that made rejecting one undecodable byte cost
  8.8 µs instead of 2.3 µs, and grew each such result in a byzantine
  proposer's block from 24 stored bytes to 496. `Log` is excluded from
  `ResultsHash`; this is a visible RPC response change only.
- **Tests** cover the four encoding classes on a local struct (amino), a
  padded tag and a nil message through `runTx` (baseapp), and on a real app
  with a signed tx: an explicit empty memo, a padded tag, a padded value
  inside the fee, a padded length inside the message `Any`, and the
  written-out zero session address every Go client produced before #6399,
  each decoding to the same signed tx and each rejected. The last case lives
  in #6399's `TestDeliverTxZeroSessionAddr`, flipped as its own comment
  anticipated: the old bytes, built by its `legacyTxBytes`, are refused as a
  decode error where they used to deliver, and the app-hash, tampered
  signature and per-byte gas comparisons that held only while they delivered
  are gone. A nil `Package` in `MsgAddPackage` and `MsgRun` is refused as a
  bad request in both modes, with no recovered panic. A `MsgAddPackage` whose
  `Info` or `Type` holds a `BigdecValue` written non-canonically is refused by
  `ValidateBasic` in both modes, not as a decode error, which fails if the
  canonical check runs first, and `addpkg_mempackage_info.txtar` now expects
  its `Info` payload refused at `CheckTx` rather than by the keeper at
  `DeliverTx`, with the same `InvalidPackageError`. Decode errors carry no
  stack trace in `Log`.
