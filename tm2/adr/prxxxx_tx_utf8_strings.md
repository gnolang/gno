# ADR: Refusing Transactions Whose Strings Are Not Valid UTF-8

## Context

A transaction is signed over `sortJSON(amino.MarshalJSON(signDoc))`. JSON
encoding turns every byte of a string that is not valid UTF-8 into U+FFFD, and
`sortJSON` re-encodes the result, so `"a\uFFFDb"` and `"a\xffb"` have the same
sign bytes. The binary decoder keeps the raw bytes (`amino.DecodeString` is
`string(bz)`), and nothing on the transaction path checked them afterwards.

So a signature does not cover such bytes. Anyone relaying a pending
transaction can re-encode it with a signed U+FFFD swapped for an invalid byte,
or an invalid byte for another, and broadcast the copy. The copy passes CheckTx
under a new hash, the original is then refused for its used sequence, and the
handlers receive bytes the signer did not sign:

- a realm's `string` parameter gets a different value (`len`, `==`, map keys,
  hashes differ);
- a realm that rejects invalid UTF-8 panics after the ante handler took the
  fee and the sequence;
- a `.gno` file stops parsing, and other package files, such as `README.md`,
  are stored with altered bytes.

The relayer needs no account and pays nothing. Amounts, coins, addresses,
function and package names are typed or strictly parsed and are not affected.

## Decision

`runTx` refuses, with `ErrTxDecode`, a transaction that holds a string that is
not valid UTF-8, after `validateBasicTxMsgs` and before the ante handler.
`std.ValidateUTF8` walks the decoded `Tx`, its signatures left out, as amino
encodes it.
It goes through pointers, interfaces, structs, slices and arrays, over the
fields amino does not skip. For a type with a `MarshalAmino` method it walks
the representation instead of the Go fields. It checks every string: the memo
and every string of every message, package files included. It skips byte
slices and byte arrays, which hold binary data.

Signers refuse too: building a sign payload (`std.GetSignaturePayload` and the
legacy rendering, so `Tx.GetSignBytes`, gnokey and every Go client) returns an
error for such a string, rather than a payload that would equally verify the
transaction with those bytes swapped for U+FFFD. The ante handler builds its
payload the same way, so it walks the signed fields again for each payload it
builds; `runTx` has already refused such a transaction, so that refusal never
fires there.

The response's `Log` names the field by its JSON path, such as
`msg[0].args[0]`. `Log` is not part of the results hash; `Error` is, and it is
the same `TxDecodeError` as for any undecodable transaction.

CheckTx, DeliverTx, Simulate and the `Check`/`Deliver` helpers all reach
`runTx` with bytes, so the one check covers them. Genesis transactions read
from a genesis file or the streamed JSONL go through JSON decoding, which
already turns invalid bytes and lone surrogates into U+FFFD, so they always
pass. A genesis built in memory (gnodev, the integration harness, the in-memory
node) passes package files as they are, so a file that is not valid UTF-8 now
fails there, as it would on chain.

## Alternatives considered

- *Check in `amino.DecodeString`*: the VM loads every realm object through
  amino on each transaction (`amino.MustUnmarshal` in the store's
  `loadObjectSafe`), and a realm can legally store `string([]byte{0xff})`.
  Every load of such an object would panic, bricking the realm.
- *Walking the Go value's fields*: misses strings that amino reaches only
  through `MarshalAmino`, and revisits values the decoder aliases, so nested
  aliased values cost exponential time. The representation has neither
  problem.
- *Per-message `ValidateBasic` checks*: the memo and fee need their own checks,
  and every message type, present and future, has to remember them.
- *A lossless JSON rendering of invalid bytes*: changes the sign bytes of
  every transaction holding them, and the Ledger app could not display them.
- *Signing the binary encoding*: breaks Ledger signing, every wallet, and the
  signatures already in genesis files.
- *An unchecked payload for the ante handler*: it would save the repeated
  walks, but an exported helper that silently skips validation is easy to
  misuse. Use the same checked payload builder for signing and verification.
  The extra work is one walk of the full payload per transaction, whatever its
  number of signers, since #267 derives each later signer's payload from the
  first. A failed current-format verification adds two walks of the fee alone,
  to build the legacy payload.
- *Requiring canonical bytes (re-encoding must reproduce the input)*: the
  swapped copy is already canonical, so this does not close the issue. It
  belongs to the separate hash-malleability work.

## Consequences

- **Consensus-breaking, shipped with the public-key and multisig checks as a
  halt-height binary switch.** A transaction that was valid becomes
  undecodable. The chain halts at an agreed height and restarts on the new
  binary. Blocks below that height keep their old binary, so the check never
  re-executes them and needs no activation height.
- **No live client sends such strings.** All 233,680 gnoland-1 transactions up
  to height 1,182,449 were checked, failed ones included.
  - The raw bytes came from indexer.gno.land; each hashes to its listed hash,
    and the count matches the header's `total_txs`.
  - Every one decodes and passes; one holds a U+FFFD in a string, which is
    valid.
  - A rescan just before release is a sanity check that this still holds. It
    gates nothing, and other networks can run the same scan.
- **Re-executing history on the new binary would need a gate.** A node that
  replays blocks below the halt height on the new binary, rather than
  switching binaries, refuses any transaction the check now rejects.
  - The smallest gate is a per-chain height compared with `ctx.BlockHeight()`
    in `runTx`, set by gno.land through a BaseApp option.
  - The ante handler would need a gate too, since its payload build refuses
    such a string. It would need an unchecked build below that height,
    selected through `AnteOptions`, so that a pre-activation transaction
    verifies as it did.
  - On gnoland-1 no such gate would let one binary re-execute its blocks: the
    multisig check shipped with this one refuses one historical transaction
    and is deliberately not gated (see *A height gate* in
    `prxxxx_multisig_exact_threshold.md`). So a node re-executing gnoland-1's
    blocks needs the earlier binary up to the halt height.
- **Refusal costs the signer nothing.** A refused transaction costs the signer
  no fee and no sequence, so a relayed copy no longer burns the original.
- **The reverse swap still runs.** A transaction signed over invalid bytes is
  refused too, but its U+FFFD rewrite carries the same sign bytes and runs.
  Go signers now refuse to sign such a string; other clients, such as gno-js,
  need the same check.
- **`gnogenesis fork` loses a step on such a transaction.** It brute-forces a
  historical transaction's signer sequence through `VerifySignaturePayload`.
  For a transaction holding invalid UTF-8 that now errors, so it falls back to
  its counter with a warning. None exists on gnoland-1; other source chains
  could have some.
- **The node pays, unmetered.** Over the sign payload's fields the walk follows
  the representation that encoding the payload follows, so it costs at most
  about one such encoding. It runs once in `runTx`, and the ante handler repeats
  it before each encoding it already pays for. `runTx`'s walk leaves out the
  signatures: they hold no text, the payload leaves them out, and the ante
  handler bounds their count only later. Walking them cost 756 ms for 499,000
  empty signatures in a 998 KB transaction, against 92 ms without. It is
  unmetered, like decoding, and its recursion is no deeper than the value amino
  built.
- **The walk runs after `ValidateBasic`.** A representation can cost far more
  to render than its bytes cost to decode, and can panic: a 900 KB hex
  `BigintValue` renders to decimal in 197 ms after decoding in 11 ms, and a zero
  `params.Param` panics. Both can sit in `MemPackage.Info`, typed `any`, which
  gno.land's `ValidateBasic` refuses since gnolang/gno#6406. So the walk runs on
  what `ValidateBasic` accepted, as #6406's canonical check does. An application
  whose messages accept `any` must refuse such values in `ValidateBasic` too.
- **Binary data in `string` arguments must be encoded.** A client passing
  binary as a `string` argument must encode it (base64, hex, or a `[]byte`
  parameter). There was none on gnoland-1.
- **Package files must be valid UTF-8.** Every file of a package, `.md` and
  `.toml` included, must be valid UTF-8.
- **Unchanged:** realms can still build and store invalid UTF-8 strings, and
  queries do not go through `runTx`.
- **Not addressed here:** other transaction hash malleability, which this does
  not change. gnolang/gno#6406's canonical check refuses non-minimal varints and
  explicit zero fields in the transaction (gno-fixes #23), and
  `prxxxx_multisig_exact_threshold.md` in the multisignature nested in
  `Signature.Signature`. Two rewrites that keep the bytes canonical are closed
  by the sibling ADRs, within the limits each states: stripping a signature's `PubKey`
  (`prxxxx_tx_signature_pubkey.md`) and dropping spare multisig signatures
  (`prxxxx_multisig_exact_threshold.md`).
