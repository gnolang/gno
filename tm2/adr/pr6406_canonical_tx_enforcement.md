# Canonical transaction encoding, enforced

## Status

Proposed. Follows `tm2/adr/pr6404_tx_canonical_encoding.md`, which shipped
the backward-compatible wire format and `amino.UnmarshalCanonical`.

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

`BaseApp.runTx` decodes every tx, in every mode, with
`amino.UnmarshalCanonical`, and returns `std.ErrTxDecode` with the first
differing byte in `Log` when the bytes are not the canonical encoding. There
is no height gate and no new parameter.

`validateBasicTxMsgs` rejects a nil message with `std.ErrUnknownRequest`.
An empty `Any` (`0a 00`) decodes to a nil message and is that message's
canonical encoding, so the check passes it; before, `ValidateBasic`
dereferenced it and the recover in `runTx` reported an internal error, at
about sixteen times the CPU of a decode error and charged no gas. The
explicit check closes that path (found in review of #6404).

The rule reaches a chain through the coordinated halt mechanism from #5368,
which gnoland-1 already runs: a GovDAO halt with `halt_min_version` set to
the enforcing release, so an old binary cannot restart past the halt and a
new one cannot start early, and an entry in the chain's upgrades ledger
(`misc/deployments/mainnet.gno.land/upgrades.json`). From the halt height on,
every included tx is canonical by construction; blocks before it are replayed
by the binaries that produced them. The release it rides is the next
coordinated release on the mainnet line after `v1.5.1`, the current rolling
release; the ledger entry names it when the halt is scheduled. Merge order:
#6404, then this change, then the related `runTx` tx-encoding change queued
behind it, so the interaction is settled before merge rather than discovered
at it.

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
  from traffic: Go tools rebuilt from #6404, `amino-ts` with
  gnolang/amino-ts#2; the ts-proto clients already emit the canonical form.
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
- **The check costs about as much again as decoding**, and on the inputs an
  attacker picks signature verification is never reached, so that is the
  relevant comparison (measured in review of #6404, `Unmarshal` against
  `UnmarshalCanonical` on the same bytes): 1 KB memo 3.5 to 5.6 µs, 100 KB
  memo 23.5 to 32.3 µs, 999 KB memo 121.5 to 190.0 µs, 797 KB with 7000
  messages 8.1 to 14.9 ms. It is bounded: the mempool rejects txs over
  `maxTxBytes` before `CheckTx` reaches the app, `Marshal` runs only after
  `Unmarshal` succeeded so garbage stays cheap, and the pre-existing unpriced
  decode is the larger term.
- **`Log` on decode failures.** Decode errors now go through
  `ABCIResultFromError`, so `Log` carries the reason where it used to be
  empty. `Log` is excluded from `ResultsHash`; this is a visible RPC response
  change only.
- **Tests** cover the four encoding classes on a local struct (amino), a
  padded tag and a nil message through `runTx` (baseapp), and on a real app
  with a signed tx: an explicit empty memo, a padded tag, a padded value
  inside the fee, a padded length inside the message `Any`, and the
  written-out zero session address every Go client produced before #6404,
  each decoding to the same signed tx and each rejected.
