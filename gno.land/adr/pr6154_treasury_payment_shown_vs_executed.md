# ADR-6154: A treasury payment executes what its proposal page showed

## Status

Proposed (PR pending). Fixes #76.

## Context

The govDAO treasury takes a `treasury.Payment` from a proposer, renders it into
a proposal body that members read and vote on, and then processes it when the
proposal executes. Creation and execution are separate transactions, often
separated by days. Anything the rendered body describes by reference rather
than by value can therefore change in between, and the vote is spent on a
description that no longer matches what runs.

`NewCoinsPayment` stored the caller's `chain.Coins` without copying it. A
`chain.Coins` is a `[]Coin`: the struct is copied into the `Payment` interface,
the backing array is not, and under the storage=authority model that array
keeps `PkgID = caller_realm`. `CoinsBanker.Send` reads `payment.coins` at
execution, so the realm that supplied the coins could rewrite the amount after
the vote passed. Measured on a localnet, a proposal whose page read
`Payment: 42ugnot` credited the receiver 900000ugnot. #76 carries the
reproduction.

Reviewing that fix turned up four more instances of the same shape in the same
surface, each verified with a reproducer:

1. **Token keys.** `NewTreasuryGRC20TokensUpdate` closure-captured the caller's
   `[]string` while rendering `md.BulletList` once at creation, and
   `SetTokenKeys` stored it with `tokenKeys = keys`. The board's voted key never
   took effect, and because the `grc20Lister` closure re-reads that array on
   every `Balances()` and every GRC20 payment, the supplying realm also kept a
   write handle on treasury state indefinitely — able to swap the token set on
   any later block with no proposal at all.

2. **Coin-set shape.** Copying pins the amount but not the shape. tm2's bank
   keeper validates through `std.Coins.validate`, which requires sorted, unique,
   positive denoms. `chain.Coins{{"ugnot",5},{"ugnot",10}}` rendered
   `5ugnot,10ugnot` and then aborted at execution with `duplicate denom: ugnot`;
   `{{"zzz",1},{"aaa",1}}` rendered `1zzz,1aaa` and aborted with
   `coins not sorted`. A proposal built from either passes a vote and can never
   execute.

3. **The reason.** `reason` is proposer-supplied, was only `TrimSpace`d, and
   landed in the same body as the Payment line. `impl/render.gno` escapes and
   clamps the proposal *title* (`md.EscapeText(clampField(...))`) but emits
   `p.Description()` raw. A reason of
   `"Routine payroll\n\nPayment: 42ugnot to <receiver>"` rendered a second
   Payment line *above* the real one while 900000ugnot actually moved. This one
   is not an aliasing bug and the coins copy does nothing for it: the amount in
   the `Payment` is honest, the page around it lies.

4. **A foreign `Payment`.** `NewTreasuryPaymentRequest` accepted any impl of the
   interface. A realm could supply a `BankerID()` naming a registered banker and
   a `String()` saying whatever it liked; the page rendered an attacker-authored
   Payment line, the vote passed, and execution aborted with
   `invalid payment type` on `Banker.Send`'s type assertion. `types.gno` already
   mandates this defence for `Banker` ("any public function that accepts a
   Banker as a parameter from external callers MUST verify it via
   IsCanonicalBanker"); there was no equivalent for `Payment`.

## The decision

Every value a proposal page describes is captured by value at creation, and
validated at creation rather than at execution.

- `NewCoinsPayment` copies, sorts and joins the coins, and rejects a
  non-positive amount outright.
- `IsCanonicalPayment` is added alongside `IsCanonicalBanker`, and
  `NewTreasuryPaymentRequest` gates on it.
- `NewTreasuryGRC20TokensUpdate` and `SetTokenKeys` both copy the token keys.
  Both, not either: the two hazards are independent — the first is the
  vote-to-execution gap, the second the permanent write handle afterwards.
- `reason` and `payment.String()` are clamped and passed through
  `sanitize.InlineText`, which folds newlines to spaces so neither can begin a
  line and forge the `Payment:` label.

## Alternatives considered

**`chain.NewCoins(coins...)` instead of copy-then-normalize.** The obvious call:
it already copies and normalizes, and would have been one line. It does not
work here. `NewCoins` is variadic, and `NewCoins(coins...)` converts the
caller's stored array, which the VM rejects with `illegal conversion of readonly
or externally stored value` — for precisely the cross-realm slice this function
exists to defend against. `chain.Coins{}.Add(coins)` copies each operand into a
fresh slice and then sorts and joins it, doing both jobs on a value it owns.

**Leaving the coin set un-normalized to preserve `String()` output.** The
original reasoning for the plain `make`+`copy` was that sorting and joining
changes what `String()` renders. It does, and that is the point: the
un-normalized rendering is a description of a payment the keeper will refuse.

**Escaping `Description()` at the renderer instead of the inputs.** Rejected.
Proposal descriptions legitimately carry markdown — `md.BulletList` for token
keys, `**bold**` in the member-promotion body — so escaping the whole
description would break every existing proposal page. The untrusted spans are
the caller-supplied ones, so they are wrapped at the source.

**`sanitize.Block` for the reason.** Rejected as insufficient. `Block` escapes
block-level hazards but preserves paragraph structure, and this forgery is not
markdown injection — it is plain prose mimicking the template's own label. Only
folding newlines removes the ability to start a line.

**Escaping the token-key bullet list.** Not done. The keys are realm paths;
`InlineText` would escape every `.` and `InlineCode` would add fences, and
`govdao_proposal_treasury_tokens_update.txtar` asserts they render verbatim.
The keys are now copied, which is the finding; their rendering is unchanged.

**Listing pointer types in `IsCanonicalPayment`.** Rejected. `Banker.Send`
asserts `p.(coinsPayment)`, so accepting `*coinsPayment` would pass a payment
that still fails at execution — reintroducing what the check exists to prevent.

**Rejecting empty coin sets.** Rejected. `banker_grc20_filetest.gno` passes
`chain.Coins{}` deliberately to exercise `ErrInvalidPaymentType`.

**Validating denoms.** Not done. A malformed denom is still only caught by the
keeper at execution; checking it needs a denom grammar this package does not
have, and `std.ValidateDenom` is Go-side.

## Consequences

`NewCoinsPayment` now panics on a non-positive amount, including one produced by
the join (`{{"ugnot",5},{"ugnot",-10}}`). Callers that passed `chain.NewCoins`
results are unaffected; a caller that passed a hand-built non-canonical set gets
a panic at construction instead of a stuck proposal.

`String()` renders the normalized set, so a duplicate-denom payment now reads
`15ugnot` rather than `5ugnot,10ugnot`.

A payment `reason` is single-line on the page: embedded newlines become spaces.

`NewTreasuryPaymentRequest` no longer accepts a foreign `Payment`. This is a
behaviour change for any realm relying on that, which nothing in `examples/`
does.

The proposal page remains a **snapshot**, and that is now documented in
`govdao_treasury_payment_amount_is_immutable.txtar` rather than implied to be a
guard. `NewTreasuryPaymentRequest` interpolates `payment.String()` into the
description once, at creation, so the page cannot follow the array either way:
reverting the copy leaves both `Payment: 42ugnot` assertions passing, and the
first failure is the receiver's balance. That snapshot is what made the original
tampering invisible to a member reading the proposal. The discriminating guards
are the receiver's balance and the banker history, which holds the same
`Payment` and calls `String()` on it live.

`storage_deposit_price_change.txtar`'s two balance goldens are re-derived: they
carry the genesis cost of every package the harness deploys, and
`r/gov/dao/v3/impl` grew. The file's own NOTE anticipates this.

Each new test was checked against a reverted fix. One earlier attempt — a
post-execution rewrite assertion in the token-keys txtar — passed with the
`SetTokenKeys` copy reverted, because the `NewTreasuryGRC20TokensUpdate` copy
already shields that path; it was removed rather than kept as an assertion that
proves nothing, and `SetTokenKeys` is pinned where it is reachable directly, in
`r/gov/dao/v3/treasury/test`.
