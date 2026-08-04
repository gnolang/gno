# Invariant checking for gno.land — a plan

Status: **proposal**. Nothing here is implemented. Written alongside the balance-split
branch (`tm2/adr/prxxxx_realm_denom_balance_keys.md`), which added several point-of-use
assertions and made the absence of global ones conspicuous.

## The shape, and why

Follow cosmos's operational model rather than its default wiring:

> Run validators with checking **off**. Run invariant checks on separate
> non-validating nodes, and in CI.

A full sweep is O(accounts + balance keys). Enumerating balances measures ~1.5 µs per
denom, so a chain with a million balance keys costs ~1.5 s per sweep — unacceptable in
`EndBlock` on a validator, and entirely acceptable on an archive node every thousand
blocks or in CI on every commit.

Two rules follow, and they are the load-bearing part of this plan:

1. **The check period is local node config, never a consensus or genesis param.** A
   node that checks must produce a byte-identical app hash to one that does not.
2. **A broken invariant halts the local node, never the chain.** No
   `MsgVerifyInvariant`, no consensus-visible halt. A false positive must not be able
   to stop the network — and given these checks will be written by humans against a
   moving codebase, false positives are the expected failure mode.

This is deliberately weaker than cosmos's `x/crisis`, which panics and (because every
node runs the same code) effectively halts the chain. We get most of the value —
detection during development, upgrades, and dry-runs, which is where cosmos's
invariants earned their keep — without arming a liveness weapon.

## What exists today

- `sdk.Invariant` — `func(Context) (string, bool)`. Fine as-is.
- `sdk.InvariantRegistry` — an interface with **no implementor anywhere in the repo**.
- `sdk.FormatInvariant` — a message formatter.
- `bank/invariants.go` — **deleted** on the balance-split branch. It walked accounts and
  checked `acc.GetCoins()`, which post-split is only the gas denom, so it would have
  inspected a fraction of balances while reporting that it checked all of them. It also
  had no caller. See RISKS.md §3f.

So: the type is right, the registry is a stub, and there is no runner, no trigger, and
no way to invoke a check.

## Phasing

### Phase 0 — invariant functions, called from tests (no registry)

Write the checks as `sdk.Invariant`-returning constructors in each module. Call them
directly from tests. This needs no registry and is where most of the value is.

Every invariant lands with **two** tests: one that passes on healthy state, and one that
constructs a violating state and asserts detection. An invariant without a violation
test is indistinguishable from `return "", false`.

### Phase 1 — a way to run them out of band

- An `InvariantRegistry` implementor (a route table) plus per-module
  `RegisterInvariants`, wired in `gno.land/pkg/gnoland/app.go`.
- `gnoland invariants check --data-dir <dir>` — offline, against a halted node's
  multistore. `contribs/gnogenesis/internal/fork/source_txs_data_dir.go` already shows
  how to open a store outside a running node.
- An ABCI query route (`invariants/check`, optionally `invariants/check/<module>`) so an
  operator can ask a running non-validating node. Read-only, infinite gas meter, no
  state writes — same class as `bank/balances`.

### Phase 2 — periodic local checking

- `--inv-check-period=N` (default **0 = off**), a local flag, checked in `EndBlock` on
  that node only. Logs on break.
- `--inv-halt-on-break` (default off) to stop the local node on detection.
- Document that validators should leave both off, and that archive/sentry nodes are the
  place to enable them.

### Phase 3 — supply tracking, and the invariant that needs it

The most valuable cosmos invariant is `total-supply`: two independently maintained
numbers that must agree. gno.land has no supply record — `SDKBanker.TotalCoin` is
`panic("not yet implemented")` — so that check has nothing to compare against.

Adding one is cheap because **transfers are supply-neutral**: only `IssueCoin`,
`RemoveCoin`, and genesis change supply. A `/s/<denom> -> int64` counter costs one extra
key per mint and **zero on every transfer**. The hook must sit at
`SDKBanker.IssueCoin`/`RemoveCoin`, not at `bank.AddCoins`/`subtract` — those carry
transfers too, and wiring it there would make the counter drift and the invariant lie.

This is its own PR (new keyspace, moves the app hash, needs genesis seeding). It also
implements `TotalCoin`, which is currently a panic reachable from Gno.

## The invariants

Cost classes: **O(1)** cheap; **O(A)** one pass over accounts; **O(B)** one pass over
`/b/` keys; **O(A+B)** both.

### Balances — the keyspace this branch introduced

| # | Invariant | Cost | Catches |
|---|---|---|---|
| B1 | **Tier exclusivity.** No `(addr, denom)` has both a `/b/` key and an entry in that address's account-object `Coins`. | O(A+B) | The failure the two point-of-use panics catch per-read, checked globally — including for addresses nothing has touched. |
| B2 | **Positive amounts.** Every `/b/` value decodes to a strictly positive `int64`; no account-object `Coins` entry is non-positive. | O(A+B) | A zero left behind instead of deleted; a negative from arithmetic inversion. |
| B3 | **Well-formed keys.** Every key under `/b/` is exactly `"/b/" ‖ addr(20B) ‖ denom`, with `ValidateDenom(denom)` passing and `len(denom) ≤ MaxDenomLength` (274). | O(B) | Key-construction bugs, truncated addresses, and denoms that entered the keyspace without validation. |
| B4 | **No orphan balances.** Every address holding any balance has an account object at `/a/<addr>`. | O(A+B) | Unspendable funds: an address with no account cannot sign, so a balance without an account is permanently stuck. `AddCoins` calls `ensureAccount` precisely to prevent this. |
| B5 | **Account tier ⊆ allowlist.** Every denom in every account object is in the keeper's `accountDenoms`. | O(A) | An allowlist that shrank without migrating — the documented path to a fully split layout. |
| B6 | **Split tier ∩ allowlist = ∅.** No denom in `/b/` is in `accountDenoms`. | O(B) | An allowlist that grew without migrating. B5 and B6 together are B1 stated per-tier, and are cheaper to check. |
| B7 | **Total supply.** For each denom, `supply[denom] == Σ balances` across both tiers. | O(A+B) | Mints and burns that bypass the accounting — the classic detector. **Blocked on Phase 3.** |

### Issuance

| # | Invariant | Cost | Catches |
|---|---|---|---|
| I1 | **Every split-tier denom is realm-qualified**: `IsRealmDenom(denom)` holds, and it parses as `/<pkgPath>:<base>` with `len(base) ≤ 16`. | O(B) | A denom that reached the store without passing `assertCoinDenom`/`assertIssuable`. Both live on the issuance path; this verifies the *result* in state rather than trusting the gate. |
| I2 | **No realm-issuable denom sits in the account tier.** | O(A) | The construction-time check in `NewViewKeeper`, verified against live state — a realm denom in a metered account blob reinstates the griefing this branch closes. |
| I3 | **Issuer packages exist.** For every realm denom in state, its `pkgPath` resolves to a deployed package. | O(B) + store lookups | Denoms attributed to a realm that was never deployed. Cross-module (bank + vm); the most expensive check here, and the first candidate to drop if cost matters. |
| I4 | **Supply attribution.** Each realm denom's supply was created only by its own realm. | — | **Not implementable retroactively** — there is no issuance log. Noted so nobody assumes B7 gives this. |

### Accounts

| # | Invariant | Cost | Catches |
|---|---|---|---|
| A1 | **Account numbers are unique and below `globalAccountNumber`.** | O(A) | Duplicate or reused account numbers, which are consensus state. |
| A2 | **Session accounts are well-placed.** Everything at `/a/<master>/s/<session>` has a live master at `/a/<master>`, and nothing at a session path is enumerated as a regular account. | O(A) | The session-nesting hazard that makes `IterateAccounts` filter on key length. |
| A3 | **Vesting is funded.** For every vesting account, `OriginalVesting ≤` the address's actual total across **both** tiers. | O(A+B) | The known gap in `NewContinuousVestingAccount`/`NewDelayedVestingAccount`, which validate `OriginalVesting` against `baseAcc.Coins` — now only the gas denom. Turns a documented weakness into a detectable state. See RISKS.md. |

### Storage and params

| # | Invariant | Cost | Catches |
|---|---|---|---|
| S1 | **Bank keys carry no storage deposit.** The realm storage ledger accounts for no key under `/b/`. | O(B) | This is true today by construction (`RealmStorageDiffs()` is a VM-internal map keyed by PkgID, not a store diff). Asserting it means a future change that starts counting bank keys is caught rather than silently charging users. |
| P1 | **`RestrictedDenoms` entries are valid denoms.** | O(1) | A governance param that can never match anything, silently disabling a restriction. |

## What this plan deliberately excludes

- **No `MsgVerifyInvariant`** and no consensus-visible halt (see the two rules above).
- **No genesis or governance param** controlling the check period.
- **No invariant that mutates state**, including lazily repairing what it finds. A check
  that fixes things cannot be run safely on a validator, and hides the bug it found.
- **No blanket `RegisterInvariants` before Phase 1.** Registering into an interface with
  no implementor produces something that looks like an active check and is unreachable —
  which is exactly why the pre-existing `bank/invariants.go` was deleted rather than
  ported.

## Open questions

1. Should Phase 1's query route be gated behind a node flag? An unauthenticated
   `invariants/check` is an O(A+B) sweep on request, i.e. a cheap RPC DoS on any node
   exposing it. Probably: off by default, same flag as periodic checking.
2. Should a break on an archive node emit a structured alert (metric, log field) rather
   than only a log line, so it can page someone?
3. Is B3's `ValidateDenom` per key affordable at scale? It is a regexp per denom. If
   not, a length-and-charset check is most of the value.
4. Where do these live — `tm2/pkg/sdk/<module>/invariants.go` per module, matching
   cosmos, or one `gno.land`-level package that can reach across modules (I3 and S1
   need both bank and vm)? Cross-module checks argue for a small `invariants` package at
   the app level that imports what it needs.
