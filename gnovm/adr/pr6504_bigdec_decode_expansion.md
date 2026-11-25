# ADR: Refuse exponents when decoding a `BigdecValue`

## Status

Proposed

## Context

A transaction's `any` fields (`MemPackage.Type`, `MemPackage.Info`, a nested
`TypedValue.V`) decode into any registered concrete type. `runTx` does this
first, before the signature, fee or account is looked at.

`BigdecValue` is registered, and its `UnmarshalAmino` handed the wire string to
`big.Rat.SetString`, which accepts exponents and materialises them. The 8 bytes
`"1e999999"` become a 1,000,000-digit numerator at about 20 ms each. About
22,000 of them fit in one `ArrayValue` under the 1 MB `MaxTxBytes`, for about
7 minutes of `CheckTx` paid by nobody. The in-process ABCI client shares one
mutex between the mempool and consensus, so block production stops for as long
as the decode runs.

## Decision

The rat branch of `UnmarshalAmino` refuses any string with a character outside
`0123456789-/`, before calling `SetString`.

`MarshalAmino` writes the rat form only through `big.Rat.RatString`, which emits
an optional minus sign, decimal digits and at most one fraction bar. So anything
the encoder wrote still decodes, at any magnitude. The float form (`"f:"`) is
untouched: `big.ParseFloat` works at a fixed 512-bit precision, and 22,000
`"f:1e999999"` values decode in about 80 ms.

## Alternatives considered

- **Refuse only `e` and `E`, as the issue suggested.** `SetString` also takes a
  binary exponent with a decimal mantissa. `"1p999999"` parses cheaply into a
  301,030-digit number, about 30 ms to print back when the value is re-encoded,
  and `"1p9999999"` parses too. Listing what is allowed closes every spelling
  at once.
- **Cap the magnitude.** Honest values can be large, and refusing one would make
  stored state undecodable.
- **Decode exponents into the float form.** The decode would stay cheap, but the
  value re-encodes differently, which changes its sign bytes. It would keep a
  form no encoder writes, for no user.
- **Refuse GnoVM values in a transaction's `any` fields during decode.** This is
  the wider fix. It needs a type check inside amino's decode of `any`, so it is a
  separate change.

## Consequences

- A value that used to decode is now a decode error, so this changes consensus
  behaviour and needs a coordinated upgrade.
  - Against a binary with `validatePackageAnyFields` (#6406), such a transaction
    was already refused before the ante handler, as `InvalidPackageError`. It is
    now `TxDecodeError`. No fee or state differs, but `ABCIResult` hashes the
    error, so `LastResultsHash` does.
  - Against `v1.7.0`, which lacks that check, a signed transaction carrying
    `"1e5"` passes decode and is charged, and on a new binary fails decode
    uncharged, so state differs as well. This comes from reading the source; no
    mixed network was run.
- Replay looks unaffected on `gnoland-1`. None of its 264,430 exported block
  transactions, which are decoded JSON with `@type` fields, mentions `Bigdec`.
  Its genesis binary (`v1.2.0`) already has the `big.Rat` encoder, and
  `MarshalAmino` is unchanged from there to `v1.7.0`, so its state holds no older
  spelling. The genesis document was scanned only in part: the first 278 MB have
  no match before the server cut the stream, and its full size is unknown.
- Stored state written by an `apd` binary (`v1.0.0`, `v1.1.0`, `chain/gnoland1`)
  can hold `"1.5"` or `"1E+3"`, which `SetString` accepts and this refuses. #5867
  treated that change as needing no migration, and the public RPC for `gnoland1`
  shows its last block on 2026-09-14, but nothing here proves no such chain
  upgrades in place.
- The cost of decoding long digit strings is outside this change.
