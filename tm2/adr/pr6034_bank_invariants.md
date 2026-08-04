# PRXXXX: Invariant checks for the two-tier balance layout

## Status

Proposed

## Context

Splitting balances into two homes (`tm2/adr/pr6034_realm_denom_balance_keys.md`) created
correctness properties that span two keyspaces and that no single read can verify. The
point-of-use assertions added with that change fire only on paths that execute: they
catch a violation when someone happens to read the affected account, not when it is
introduced.

`bank/invariants.go` existed on master; that change removed its contents (an unreachable
check with no caller) and this one re-creates the file with the checks below. Two reasons, and
the second is the stronger: it walked accounts checking `acc.GetCoins()` for negative
amounts, which post-split is only the gas denom — but it could never have fired on any
chain, before or after, because a negative amount in an account object **fails amino
decode**, so `GetAllAccounts` panics before returning one. And it had no caller:
`sdk.InvariantRegistry` is an interface with no implementor anywhere in the repo.

## Decision

Add invariant **functions**, typed `sdk.Invariant`, called from tests. Deliberately no
`RegisterInvariants`, no registry implementor, no runner, no check period.

Operationally, follow cosmos's practice rather than its default wiring — checks belong on
non-validating nodes and in CI, not on validators. Two rules are fixed for whenever a
runner does arrive:

1. **The check period is local node config, never a consensus or genesis param.** A
   checking node must produce a byte-identical app hash to one that does not.
2. **A break halts the local node, never the chain.** No `MsgVerifyInvariant`. These
   checks are written by humans against a moving codebase, so false positives are the
   expected failure mode, and a false positive must not be able to stop the network.

A third rule follows from those, and is an acceptance criterion for any future check:
**no invariant may be breakable by an unprivileged actor.** A periodic check that halts
the local node is otherwise a remote DoS. This rule caused a proposed check to be
dropped: "no session address is also a regular account" is inducible by anyone sending
coins to a session address, because `AddCoins` → `ensureAccount` creates the account.

### The checks

`bank.BalanceKeysInvariant` sweeps `/b/`: keys well formed and denoms valid; values
decode to a positive amount; no denom there belongs to the account tier (the allowlist
grew without migrating); a realm-shaped denom has a shape a realm could have issued;
every address holding a balance has an account object, without which it cannot sign and
the funds are unspendable.

`bank.AccountTierInvariant` sweeps account objects: every denom is in the allowlist. One
that is not is stranded — the keeper routes it to the split tier, so `GetCoin` reports
zero while `GetCoins` reports it. A realm-issuable denom there is reported distinctly,
because its consequence is different: the issuer could mint into a shared metered blob.
On the failure path only, it notes whether the denom is *also* split-tier, which is the
state where `GetCoins` would sum both homes.

`auth.AccountKeyspaceInvariant` sweeps `/a/`: key shapes some iterator enumerates;
values that decode to a non-nil account; **the account's own `Address` agrees with the
key it is filed under**; account numbers unique and below the global counter; sessions
naming a master that exists and claiming that same master.

The key/field agreement check was not in the original list and is the most valuable one
here. `SetAccount` files an account under `AddressStoreKey(acc.GetAddress())` while every
reader looks it up by the address it already has, so a disagreement silently redirects
writes — crediting one address lands the coins in another. Demonstrated: with B's account
stored at A's key, one ordinary `AddCoins(A, 1ugnot)` took supply from 1000 to 2001.

### Reading state without inheriting its panics

Every typed accessor panics on exactly the state an invariant must report:
`decodeBalance`, `splitCoins`, `accountTierCoins`, `GetCoins`/`GetCoin`,
`ak.decodeAccount` and therefore `IterateAccounts`/`GetAccount`, and `Coins.AmountOf`.
The invariants call none of them. Instead:

- `tryDecodeBalance` and `parseBalanceKey` are non-panicking twins; `decodeBalance` is
  now a wrapper over the first, so the wire format has one source of truth and the money
  path keeps its identical messages. It panics with `err.Error()`, not `err` — the
  recovered value's type is observable.
- `auth.DecodeAccountSafe` returns an error where `decodeAccount` panics, **and treats a
  nil account as a finding**: a zero-length value is accepted by the store and unmarshals
  to a nil interface with *no error*, so "decoded cleanly" is not enough to call an
  account usable.
- `auth.IterateAccountEntries` walks `/a/` unfiltered and **returns the iterator's
  error**. This matters: the bptree store (which production uses) panics from
  `Valid()`/`Close()` if its error is never read, and the iavl store reports nothing at
  all, which would silently truncate a sweep and read as a healthy keyspace.
- Keys are classified **before** decoding, and by infix as well as length — a key can be
  exactly session-length without the `/s/` infix, and a length-only classifier would hand
  it to a session decoder.
- `sdk.Guard` recovers as a **backstop**, reporting the panic as broken so a bug in a
  check is louder than a violation. It is not the mechanism, and it cannot catch
  everything: an out-of-memory allocation is fatal, which is why the account-number
  uniqueness bitset is bounded rather than sized from the raw counter —
  `NewAccountWithUncheckedNumber` is exported and can drive that counter arbitrarily high.
- `sdk.InvariantReport` counts every finding and formats the first ten. A corrupt chain
  could otherwise build a multi-gigabyte message inside the check meant to diagnose it.
- Iteration passes a **nil gas context**. Under a live meter a whole-keyspace sweep would
  run out of gas partway through and `Guard` would flatten that into one "check panicked"
  line, reporting nothing about the state.

## Alternatives considered

**Register the invariants anyway.** Rejected: `sdk.InvariantRegistry` has no implementor,
so a registered check is unreachable, and a `RegisterInvariants` in the bank module would
read as an active check that isn't — which is what made the deleted file dead code.

**A total-supply invariant**, cosmos's headline check. Not possible *at the time of
this ADR*: there was no supply record, `TotalCoin` was `panic("not yet implemented")`,
and the reverse denom index was deliberately omitted, so summing balances would have
compared them against themselves. **Since superseded** — see
`pr6034_coin_supply.md`, which adds the counter and `SupplyInvariant` with it.

**Checks that were specified and dropped, because they are false on a healthy chain:**

- *Every split-tier denom is realm-qualified.* This re-conflates tier with shape. The
  split tier is everything **not** on the allowlist — `atom`, `zeta`, `ibc/<hash>`, any
  non-gas genesis denom. Reproduced firing on `TestConservation`'s own fixtures. Narrowed
  to a conditional: *if* realm-shaped, then well-shaped. Also deliberately **without**
  `isValidBaseDenom`'s minimum base length, which existing realm denoms in tests
  (`/gno.land/r/a:x`) do not satisfy and which has no consequence for stored state.
- *Vesting is funded across both tiers.* False for every vesting account that has ever
  paid gas: `DeductFees` → `SendCoinsUnrestricted` bypasses the vesting check, and the
  keeper documents the clamp. An existing test asserts exactly this state
  (`TestBankKeeper_VestingUnrestrictedBypass`). Shipped as a periodic invariant it would
  halt nodes on healthy chains. Replaced with schedule *validity*, which is always true.
- *Tier exclusivity as its own sweep.* Implied by the two allowlist checks: a
  double-homed denom is either in the allowlist (so its split key is a violation) or not
  (so its account entry is). Kept as a note on the failure path instead.

## Consequences

- No behaviour change on any chain: nothing calls these outside tests.
- `decodeBalance` and `denomFromBalanceKey` are now thin wrappers. The existing bank suite
  passes unmodified.
- `TestConservation` runs both bank invariants and the auth invariant after **every**
  operation, including rejected ones (0.08 s → 0.10 s). This is two-directional: the
  random walk fuzzes the invariants — 500 keeper-produced states that must all be reported
  healthy, the cheap defence against a check that fires on legitimate state — and the
  invariants act as extra oracles for the keeper, seeing what the map model structurally
  cannot.
- A latent flaw in that test's fixture was fixed: its vesting account was built with
  account number 0 while the global counter was still 0, colliding with the first
  auto-allocated account.
- A future runner needs a registry implementor, app wiring, a trigger, and an explicit
  decision on halt-versus-log. `AllInvariants` per module is the seam.
- Running both `/a/` sweeps decodes every account twice. Accepted: this is an offline
  check, and sharing the pass would need a visitor abstraction across a module boundary.

## Tests

Each check with a violating-state test asserts its specific finding and was
mutation-verified: deleting the check fails that test, and only that test. Asserting the
message rather than merely that the invariant broke is the point — the checks share a
sweep, so a bare "broken" assertion passes whenever any sibling fires.

Three checks have **no** violating test and are known-unpinned rather than
believed-covered: the key-ordering and iterator-error reports in `BalanceKeysInvariant`,
and the iteration-error report in `AccountTierInvariant`. All three fire only on a
store-level fault that no stored state can produce, so pinning them would need a fake
store.

Violating states are built through the public API wherever possible and by raw store
write only where the keeper's guards make a state otherwise unreachable. Two worth
calling out because they are reachable without any raw write: the stranded-denom case
uses a keeper whose allowlist shrank, which is the balance-split ADR's documented
upgrade-without-replay scenario; and the session master-mismatch case exploits the fact
that `NewSessionAccount` sets the master *field* while `SetSessionAccount` decides the
master in the *key*, and neither consults the other.

`TestInvariantReportsWhatIterateAccountsPanicsOn` asserts both halves of the design: the
keeper accessor panics on a poisoned account, and the invariant reports it instead.

## AI assistance

Implemented with AI assistance. Three agents proposed plans independently; the distilled
plan was then reviewed by three more. That process changed the design substantially — the
reviewers refuted two specified invariants as false on healthy chains, found that a
zero-length account value decodes to a nil account with no error, found that the store
iterator itself panics unless its error is read, found that the uniqueness bitset could be
driven to an unrecoverable allocation, and proposed the key/field agreement check that
turned out to be the most valuable one. The human author reviewed and owns the change.
