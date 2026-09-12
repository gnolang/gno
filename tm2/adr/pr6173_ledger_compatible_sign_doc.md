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

Two halves, and the second is what makes the first shippable.

**Render the fee in the Cosmos shape** inside `GetSignaturePayload`, the single
function every signature passes through:

```json
"fee":{"amount":[{"amount":"1000000","denom":"ugnot"}],"gas":"200000"}
```

The conversion happens before amino, not in a `MarshalJSON` on `Fee`: amino has
its own reflection encoder and ignores `json.Marshaler`. A `Fee.MarshalJSON` was
written first and had no effect whatsoever on the output.

`std.Fee` itself is unchanged. Only the signed rendering moves, and only the fee
— every other field already matches the allowlist.

**Accept the previous rendering too.** `GetSignaturePayloadLegacy` reproduces the
pre-change bytes, and verification tries the current rendering first and falls
back to it. Clients build the signature payload themselves — wallets, the genesis
tooling, anything holding a key — so a node that accepted only the new bytes
would reject every client that had not shipped the change on the same day, and
every signature made before it. That second set includes signatures that can
never be reissued: the transactions inside an already-written genesis file.

### Why accepting both is safe

The concern with two valid encodings is malleability: if some legacy rendering of
one transaction could equal the current rendering of a *different* one, a
signature authorising T1 would also authorise T2.

It cannot. The legacy fee object carries `gas_wanted` and `gas_fee`; the current
one carries `amount` and `gas`. Those key sets have no member in common, and two
JSON objects that parse to different key sets are not the same bytes — whatever
else differs between the documents. The fee is also the one field an attacker
choosing the document cannot avoid: memo, chain ID and msgs are all attacker-
chosen, and none of them can remove or rename the fee object.

So a signature still authorises exactly one transaction. What widens is the set
of acceptable *proofs* of that transaction, not the set of transactions behind a
proof. `TestSignaturePayloadEncodingsAreDisjoint` pins this, including against
sign docs whose memo and chain ID spell out the other encoding's own text, and it
is the test to consult before either rendering is touched again.

### Where the fallback is wired

Every path that verifies a signature it did not produce:

| path | why it matters |
|---|---|
| `sdk/auth/ante.go` | consensus. Also the path genesis replay takes, so existing genesis files keep validating |
| `gnogenesis verify` | reads files written in the past, which cannot be re-signed |
| `gnogenesis fork` | brute-forces sequences over txs taken off a live source chain, all signed by whatever their senders were running |
| `gnokey verify` | would otherwise call a signature invalid that the chain accepts |

`auth.GetSignBytes` is exported but has no callers in the tree and is not on any
verification path; its doc comment now says it is a signing helper only, since a
verifier using it would reject legacy signatures the node accepts.

`gnokey verify` prints which rendering a valid signature matched. Only the
amount/gas rendering is one a Ledger will sign, so a wallet developer checking
their payload shape can tell the two apart from the output.

The legacy encoding is computed only when the current one fails, so the ordinary
path pays for one encoding and one curve operation, as before. Gas is charged
once, before either attempt.

That is the honest path. A signature that matches neither rendering costs two
encodings and two curve operations, and none of it is metered: the first encoding
happens before the signature gas is charged, and a transaction the ante handler
rejects pays no fee. Measured on a 1 MB `MsgAddPackage`, one encoding plus
`sortJSON` takes about 7 ms, so garbage-signature spam costs a node roughly twice
what it did before dual verification. This amplifies a cost the attacker already
did not pay rather than opening a new class of attack; the sunset switch below is
the operator's lever against it.

## Consequences

**An upgraded node accepts a superset of what an un-upgraded one does, so this
is still a consensus change and still needs a coordinated upgrade** — but the
failure mode reverses, and that is the point. Before dual verification, the
change broke every existing client at once. Now nothing that works today stops
working; what an un-upgraded node cannot do is accept a *Ledger-signed*
transaction. If validators upgrade piecemeal, an upgraded proposer can include a
Ledger-signed tx in a block that un-upgraded validators reject. Validators must
still upgrade together; users and wallets no longer have to.

That holds in one direction only: nodes before clients. A client built from this
commit signs the amount/gas rendering and nothing else, so against a node that
has not upgraded every one of its transactions is refused with `signature
verification failed`, and a genesis file it signed panics a pre-change node at
`InitChain`. Ship node binaries first and hold client releases until the
validator set has moved. `node:p:halt_height` and `node:p:halt_min_version` are
the coordination tools the chain already has for the validator side.

Verified on `gnoland1` with the fee change applied and no other modification: a
Nano X on Cosmos app 2.39.1 signs both a `send` and a `/vm.m_call`, the node
accepts the signatures and executes the transactions. Each then fails on its own
merits — a chain transfer restriction, and wugnot's minimum-deposit check — which
is what a working signature path looks like.

`addpkg` still fails, with `hidapi: unknown failure` at 8,840 bytes where a
~400-byte call succeeds. That is a transport or device-memory limit on large
payloads and is a **separate, unfixed problem**: package upload appears to be
beyond a Ledger regardless of this change. This ADR does not address it.

A zero fee renders as `"amount":[]`, not as a list holding an empty coin.
Cosmos's `Coins` carries no zero entries, and the device displays every coin it
is given, so `{"denom":"","amount":"0"}` would be both wrong and likely refused —
reintroducing this very failure for zero-fee transactions. That path is live, not
hypothetical: `ante.go` branches on `GasFee.IsZero()`, and every genesis
transaction is signed with `GetSignBytes(chainID, 0, 0)`.

The protobuf wire encoding is untouched. Only the JSON signature payload moves,
so `Tx` on the wire, and every decoder of it, are unaffected.

A multisig whose members run different clients cannot sign during the
transition. `PubKeyMultisigThreshold.VerifyBytes` checks every constituent
signature against one payload, so a signature set split across the two
renderings fails both attempts, and `gnokey multisign` aggregates partial
signatures without verifying them, so the failure surfaces only at broadcast.
Every member must produce the same rendering.

Clients that build the payload themselves — `tm2-js-client` and `gno-js-client`,
Adena, gnonative — keep working through the fallback, and must move to the
amount/gas rendering to sign with a Ledger.

**The legacy path is meant to be temporary, and nothing here retires it.** It has
no sunset switch and no deprecation warning; removing it later is a second
consensus change, needing its own coordination once wallets have moved. A switch
to refuse the legacy rendering would make that transition observable — a node
could report how often the fallback fires — and is the obvious follow-up. It must
be a chain-wide consensus parameter rather than a per-node flag: nodes that
disagree about accepting the legacy rendering fork on the first legacy-signed
transaction. Its zero value must mean *accept*: a parameter absent from the
store reads back as its zero value on an in-place upgrade, so a "reject legacy"
switch defaulting to off is the only shape that does not itself produce the flag
day this change avoids. It was left out here to keep this change to one decision.

### Tests

Sixteen, across four packages. Twelve are in `tm2/pkg/std`, and the first four of
those fail against the pre-change rendering, naming the offending keys:

- the sign doc's keys, and the fee's keys, are within the app's allowlist, and
  every key the app requires is present even when its value is empty
- the fee has the array-of-coins shape with string-typed numbers
- a zero fee is pinned byte-for-byte as an empty list — pinned rather than
  parsed, because `json.Unmarshal` reads `[]` and `null` into the same
  zero-length slice, and a nil slice renders as `null`
- the whole current payload is pinned byte-for-byte, once with nil msgs and once
  with a real `bank.MsgSend`
- the two renderings differ only in the fee; every other field, msgs included,
  is byte-identical
- `signDocPayload` mirrors `SignDoc` field for field, so a field added to one
  and forgotten in the other fails a test instead of silently leaving the signed
  bytes
- `Tx.SignDoc` binds the transaction to the chain ID, account number and sequence
  it is given
- the legacy payload is pinned byte-for-byte, to a literal taken from master
  before the change — anything derived from the current code would follow the
  current code wherever it went
- the two encodings are disjoint, over a table including adversarial memos
- the shared helper takes either rendering, still binds the sign doc it was
  given, still refuses nonsense, and reports a sign doc it cannot encode as an
  error rather than as a bad signature

Plus four behavioural ones: the ante handler admits a legacy-signed transaction,
still rejects a signature made over a different chain ID, sequence, or nothing at
all, `gnogenesis verify` accepts a genesis file signed before the change, and
`gnokey verify` says so when the legacy rendering is the one that matched.

Each was checked by ablation — breaking one thing at a time and confirming which
test fired. Removing the ante fallback fails only the ante acceptance test;
making `GetSignaturePayloadLegacy` return the current bytes fails the legacy pin,
the disjointness test, and the ante test's own guard against the two renderings
coinciding; making the fallback swallow its failure fails all three rejection
cases; making the shared helper return `true` fails the binds-the-document and
garbage arms, which is a different assertion from the one that fires when the
helper's legacy branch is removed. Making `feeAmount` return a nil slice fails
the zero-fee pin; an `omitempty` on the payload's memo fails the required-keys
check; a field added to `SignDoc` alone fails the mirror test; renaming the
payload's msgs tag fails both msgs pins.

The byte-for-byte pins are the ones that would have caught the original problem.
Nothing in the tree pinned the signed bytes before, so the rendering was free to
drift from what any external signer expects, and did.

## Alternatives considered

**Change the rendering without accepting the old one.** The first version of this
change. It is smaller and it needs no malleability argument, but it invalidates
every signature every existing client produces, and every signature already
written into a genesis file. On a chain that has not launched that costs nothing;
on a live one it is a flag day for every wallet simultaneously. Rejected once the
question "could the chain accept both?" was asked, because it can, safely.

**Gate the fallback behind a flag, off by default.** Safer-sounding, but the
default decides everything: off by default is the alternative above wearing a
flag, and on by default is this decision plus an untested code path. A flag
belongs to the *removal* of the legacy rendering, not its introduction — see the
sunset note above.

**Ship a gno Ledger app.** Correct in the long run, and the only route that lets
the device display gno's own message types meaningfully rather than approximating
them as Cosmos messages. Much slower, and it does not help anyone today.

**Do nothing.** Ledger support is already broken for every user who updates, and
silently. Not viable.
