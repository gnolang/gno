# PRXXXX: Track per-denom coin supply

## Status

Proposed

## Context

`SDKBanker.TotalCoin(denom)` was `panic("not yet implemented")` and is reachable from
Gno through the `Banker` interface, so a realm calling it aborted its transaction.

More importantly, the invariants added in `pr6034_bank_invariants.md` are all
*structural*: they check key shapes, tier membership, and that an account's address
agrees with the key it is filed under. None is a **redundancy** check — two numbers
maintained by different code that must agree. That is what makes cosmos's
`total-supply` invariant the one that catches minting bugs, and it was impossible
here because there was nothing to compare a sum of balances against. That ADR
recorded "no supply invariant is possible" as a consequence; this change removes the
reason.

## Decision

Store a per-denom counter:

```
/supply/ || denom   ->   amount, 8-byte big-endian
```

**Spelled out, not abbreviated.** Every module shares one store in tm2, so the
prefix is what namespaces the keyspace. Cosmos can afford an opaque single byte
(`x/bank` uses prefix `0` for supply, `2` for balances) because each of its modules
gets its own mounted store; that precedent does not transfer. Store keys are not
gas-metered, so legibility is free. Deliberately not `/s/`: that is
`auth.SessionStoreKeyInfix` verbatim, and although a session key begins `/a/` so the
ranges cannot overlap, sharing the spelling would make a search for one return the
other.

The value reuses the balance codec unchanged — same domain (a positive `int64`),
same reason for fixed width, and the key is deleted at zero exactly as balances are,
so a fully burned denom leaves no trace and `encodeBalance`'s positivity guard keeps
serving both keyspaces.

### Only mint and burn move the counter

`bank.MintCoins` and `bank.BurnCoins` own it. `AddCoins` and `subtract` deliberately
do **not** touch it, and this is the central design decision rather than an
oversight: they also carry transfers, fees and storage-deposit refunds, so a counter
that followed them would make an unpaired credit look legitimate instead of breaking
the invariant. Keeping a second number is only useful if it is maintained by
different code than the number it checks.

`AddCoins`/`SubtractCoins` were also **removed from `vm.BankKeeperI`**. Issuance from
gno.land now goes through mint/burn, so an unaccounted mint is not expressible from a
realm — a structural guarantee rather than a convention. It cost nothing: each had
exactly one caller.

Every path that creates or destroys value was traced: `SDKBanker.IssueCoin`,
`SDKBanker.RemoveCoin`, genesis balances, and the height-0 signer auto-funding (which
is a genuine mint — it is guarded by the account not existing, so there is no prior
balance). `auth.RemoveAccount` would destroy account-tier coins, and its own comment
already says it violates the supply invariant; it has no production caller and should
be deleted or given a burn hook.

### Supply is capped at MaxInt64 per denom

This is a new consensus rule, and a real one. Before the counter, two addresses could
each hold `MaxInt64` of a single denom: `AddCoins` bounds each *balance* with
`overflow.Add`, and nothing bounded the sum. Verified independently by three
reviewers. So the counter closes an existing hole rather than only adding
bookkeeping.

A mint that would exceed the cap returns an **error**, not a panic — it is reachable
by an ordinary realm minting too much, which is a caller's mistake.
`SDKBanker.IssueCoin` already panics on a bank error, so from Gno the behaviour is an
ordinary transaction abort. Everything fallible happens before anything is written,
so a rejected mint leaves neither the balance nor the counter changed.

### Genesis seeds by sweeping, not by delta

`RecomputeSupply` totals every balance across both tiers and rewrites the records. It
runs at `InitChainer`, after the balance loop, in both the in-memory and streaming
paths.

A delta inside `SetCoins` looks sufficient and is provably wrong: `applyBalance`
writes the account object with the **full pre-split amount before** calling
`SetCoins`, so `SetCoins` reads `old == new` and the delta would be **zero for every
vesting account**. Three reviewers reproduced this. The pre-write cannot be removed
either, because `NewContinuousVestingAccount`/`NewDelayedVestingAccount` validate
`OriginalVesting` against it. So `SetCoins` stays supply-blind, its doc comment says
so, and the sweep is authoritative.

`RecomputeSupply` is for genesis and offline tooling only, never a live chain: it
writes state outside any block, so a node that ran it and one that did not would
diverge.

## Alternatives considered

**Hook `SDKBanker.IssueCoin`/`RemoveCoin` only.** Rejected: it puts a tm2 invariant's
correctness in a gno.land caller's hands, cannot cover the height-0 funding, and
leaves the supply-blind `AddCoins` where the next contributor reaches for it.

**Hook `AddCoins`/`subtract` symmetrically.** The counter would be numerically
correct — transfers are paired, and `InputOutputCoins` validates `sum(in) == sum(out)`
— but it defeats the purpose. A future unpaired `AddCoins` would silently become a
legitimate mint because the counter follows it. It also roughly doubles the cost of
every transfer rather than of issuance.

**No counter; sum on demand for `TotalCoin`.** Makes the invariant a tautology
(comparing a sum against itself) and makes a Gno-reachable call O(all balances).

## Consequences

- **Consensus-visible.** The `/supply/` keyspace appears at genesis for every denom
  in the balances file. Migration is fork-and-replay, the same answer
  `pr6034_realm_denom_balance_keys.md` committed to and for the same reason: replay
  regenerates supply through the code that maintains it, so there is no migration
  code to get wrong.
- **A mint or burn costs ~+306,700 gas** — one extra read and write on the supply
  key. Within one transaction, N mints of the *same* denom pay it once
  (`cacheStore` refunds the repeat), so the cost falls on distinct-denom issuance,
  which also makes denom spam dearer. One txtar ceiling needed raising, from
  3,200,000 to 3,400,000; the other three Mint/Burn calls absorbed it within
  existing headroom.
- The **app-hash pin does not move**. Predicted and verified: the pinned fixture
  funds via `SetCoins` and never calls `RecomputeSupply`, and `SetCoins` is
  supply-blind. It *would* have moved under the rejected delta design.
- `bank.BankKeeperI` and `bank.ViewKeeperI` gained methods; `vm.BankKeeperI` traded
  `AddCoins`/`SubtractCoins` for `MintCoins`/`BurnCoins`/`TotalSupply`.
- `bankerTotalCoin`'s native gas base went from 89 to 349, matching `bankerGetCoin`'s
  conservative placeholder. The 89 was calibrated against a mock returning `0`, so it
  cannot see the real path's `ValidateDenom` regexp or store read. Wants
  re-derivation on the bench box.
- **`TotalCoin` validates its denom**, for the reason `GetCoin` does: it is a
  realm-supplied string reaching a store key and nothing on the `.gno` side bounds
  it. `TestBanker.TotalCoin` is implemented too, so `gno test` and the chain agree.

## Tests

`SupplyInvariant` reports in both directions — held-but-unrecorded (what an
unaccounted credit produces) and recorded-but-not-held (what a lost balance write
produces) — plus a disagreement in amount and a corrupt record. It shares
`computeSupply` with `RecomputeSupply`, so the recorded number and the checked number
come from one piece of code.

`TestConservation` now runs it after every operation. That required converting its
credit and debit arms from `AddCoins`/`SubtractCoins` to `MintCoins`/`BurnCoins` and
reseeding after its replace-all arm — the invariant correctly flagged the old
arms, which used raw credit *as* mint.

Six mutations verified: stop updating the counter, drop the overflow check, store a
zero instead of deleting, skip the account tier when totalling, drop the
held-but-unrecorded report, and drop the amount comparison. Each fails its own test.

`TestRecomputeSupplyCoversBothTiersAndGenesisShape` builds the exact genesis shape
that defeats a delta hook — account object prewritten, then `SetCoins` — and asserts
the sweep gets it right.

## AI assistance

Implemented with AI assistance. Three agents planned independently and converged,
each verifying by execution rather than reading: all three found that supply could
already exceed `int64`, and all three found the genesis delta defect. One additionally
noted that `applyBalance` on a repeat address draws a fresh account number and
overwrites the account. That was investigated and **deliberately left alone**: the
account is recreated either way (a plain entry after a vesting one must clear the
schedule), the balance and the supply seed are both replace-all/swept, and the only
trace is a gap in account numbering, which is harmless. Closing the gap was tried and
reverted — it shifts every later account number, changing the genesis state of any
chain whose balance file repeats an address and breaking nine integration goldens, for
no correctness gain. The behaviour is now pinned by
`TestApplyBalanceWithARepeatedAddress`.
The human author reviewed and owns the change.
