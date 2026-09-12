# Render the signature payload's fee in the shape the Ledger Cosmos app accepts

## Context

gno does not ship a Ledger app. It borrows the Cosmos one: `gnokey` speaks
`CLA 0x55` and asks for `SIGN_MODE_LEGACY_AMINO_JSON`.

That app does not blind-sign. It parses the amino sign doc so it can display what
it is about to sign, and — since commit `a2006518` in `cosmos/ledger-cosmos`,
*"Restrict amino SignDoc to a known field allowlist"*, 2026-06-25 — it refuses
any document containing a key outside a fixed allowlist
(`app/src/tx_validate.c`):

| level | permitted keys |
|---|---|
| sign doc | `account_number`, `chain_id`, `fee`, `memo`, `msgs`, `sequence`, `tip`, `timeout_height` |
| `fee` | `amount`, `gas`, `granter`, `payer` |

`std.Fee` renders as `{"gas_wanted":"200000","gas_fee":"1000000ugnot"}`. Both of
its keys are outside that list, so the device answers `parser_unexpected_field`
— surfaced to the user as `Unexpected field` — and stops.

`Fee` is present in **every** sign doc (`Tx.GetSignBytes`), so this is not
specific to a message type or a transaction size. Measured against `gnoland1`
with a Ledger Nano X on Cosmos app 2.39.1: a 200-byte `send`, a `/vm.m_call`
and an `addpkg` all fail identically, before the device displays anything.

The effect is a delayed-action break. Nothing in gno changed; the app did. A user
who has been signing gno transactions for years loses the ability the moment they
accept a Ledger Live update, and the error they are shown is
`please open Cosmos app on the Ledger device`, which points away from the cause.

Three differences separate `std.Fee` from what the app expects, and all three
must be addressed together:

1. **names** — `gas_wanted`/`gas_fee` against `gas`/`amount`
2. **cardinality** — a single `Coin` against an array of coins
3. **numeric type** — amino renders `Coin` as the string `"1000000ugnot"`; the
   app expects `{"denom":...,"amount":"..."}` with the amount as a string

## Decision

Render the fee in the Cosmos shape inside `GetSignaturePayload`, the single
function every signature passes through:

```json
"fee":{"amount":[{"amount":"1000000","denom":"ugnot"}],"gas":"200000"}
```

The conversion happens before amino, not in a `MarshalJSON` on `Fee`: amino has
its own reflection encoder and ignores `json.Marshaler`. A `Fee.MarshalJSON` was
written first and had no effect whatsoever on the output.

`std.Fee` itself is unchanged. Only the signed rendering moves, and only the fee
— every other field already matches the allowlist.

## Consequences

**This changes the signed bytes, and is therefore a consensus change.** Every
signature produced by an older client is invalid against a node carrying this
change, and vice versa. It cannot ship quietly in a patch release. On a chain
that has not launched it costs nothing; on a live chain it needs a coordinated
upgrade.

Verified on `gnoland1` with the change applied and no other modification: a Nano X
on Cosmos app 2.39.1 signs both a `send` and a `/vm.m_call`, the node accepts the
signatures and executes the transactions. Each then fails on its own merits — a
chain transfer restriction, and wugnot's minimum-deposit check — which is what a
working signature path looks like.

`addpkg` still fails, with `hidapi: unknown failure` at 8,840 bytes where a
~400-byte call succeeds. That is a transport or device-memory limit on large
payloads and is a **separate, unfixed problem**: package upload appears to be
beyond a Ledger regardless of this change. This ADR does not address it.

Three tests are added. They fail against the current rendering, naming the
offending keys, and pass with the change:

- the sign doc's keys, and the fee's keys, are within the app's allowlist
- the fee has the array-of-coins shape with string-typed numbers
- the whole payload is pinned byte-for-byte

The last is the one that would have caught this. Nothing in the tree pinned the
signed bytes before, so the rendering was free to drift from what any external
signer expects, and did.

## Alternatives considered

**Accept either encoding at verification.** `ante.go` verifies against one
rendering; it could try the current bytes and fall back to the Cosmos-shaped
ones. That keeps existing clients working and avoids a coordinated upgrade. It
was not taken here because two valid encodings per transaction is a malleability
surface that needs its own analysis — both renderings must provably bind the
same fee. It remains the better option for a live chain, and is a larger change
than this one.

**Ship a gno Ledger app.** Correct in the long run, and the only route that lets
the device display gno's own message types meaningfully rather than approximating
them as Cosmos messages. Much slower, and it does not help anyone today.

**Do nothing.** Ledger support is already broken for every user who updates, and
silently. Not viable.
