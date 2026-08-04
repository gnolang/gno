# RISKS.md — denom/account hardening

Working notes for Jae, written during the account/denom fix. Everything here is
either a decision I did not feel entitled to make, or a thing worth re-reviewing
after the work looks done.

Two reviewers argued this file should not ship to master and that its live
content belongs in the ADR. You asked for it explicitly, so it is here — but if
you agree with them, §1 and §2 are historical and only the items below are
load-bearing.

## What actually needs your decision

0. ~~Keep the account tier, or remove it now?~~ — **resolved: keep it.** The open
   question was whether the tier's value would erode, since it rests on gas
   constants the chain sets for itself: full ADR-004 measures **+45% on a warm
   `MsgSend`** (932,933 → 1,355,208), and removing the tier now would cost no
   migration where removing it later costs a state-wide one. Settled by the repo
   owner on the ground that **store-write gas will not fall**: prices track
   worst-case validator cost, and permanent state only gets more expensive to hold
   as the tree grows. The +45% is therefore a floor, not a decaying constant, and
   the timing asymmetry does not argue for acting now. Recorded because the tier
   carries **no security value** — measured, both designs close the griefing attack
   identically — so anyone who later revisits it should know they are trading gas
   only, and see §3e for the measurements.
1. ~~Genesis can allocate into the realm tier~~ — **resolved** by switching the
   tier to an allowlist of gas denoms. Tier no longer depends on a denom's shape
   or its provenance, so a genesis-allocated `/`-prefixed denom simply lands in
   the split tier, which is where it belongs. Nothing to enforce.
2. **The vesting constructors still read `baseAcc.Coins`** as if it were the whole
   balance (`vesting_account.go:158,250`). Fine while genesis only ever vests the
   gas denom, which is the account tier; a genesis file vesting anything else would
   have its `OriginalVesting` validated against a tier that no longer holds it.
   Enforcement at spend time reads the right tier either way. Still the only
   in-tree reader treating an account's `Coins` as complete.
3. ~~One unreproduced flake~~ (§3d) — **diagnosed, and it was a real defect.** Its
   first symptom turned out to be reachable: `subtract`'s split-tier loop let a
   **non-positive** amount past its guard, so an unvalidated debit credited (100 →
   600 on a −500 debit) and a `MinInt64` one overflowed into the exact panic
   reported. A regression from this branch — master's `subtract` revalidated via
   `Coins.SubUnsafe`, which the split path replaced with raw arithmetic. Fixed by
   making `subtract` defend its own precondition; pinned by
   `TestNoPathAdmitsANonPositiveDebit`. Two of the three symptoms remain
   unexplained and are most likely review agents building in this shared worktree
   while it was being edited — **process fix: agents get their own worktree.** One
   CI confirmation run is still worth it.
4. ~~Gno has no per-denom balance read~~ — **resolved.** `banker.GetCoin(addr,
   denom)` now exposes the O(1) read, plumbed through `vm.BankKeeperI.GetCoin`
   and `SDKBanker`. Its cost is pinned as independent of how many other denoms the
   address holds (`TestGetCoinCostIsFlat`), and it is exercised end to end from
   Gno in the txtar. Two consequences worth knowing: this is new Gno stdlib
   surface, and editing stdlib `.gno` source moved the app-hash pin — which is the
   *only* reason this branch moves it, since the balance split alone does not.
   The native gas entry for `bankerGetCoin` is a conservative placeholder (the
   same base as `bankerGetCoins`, which does strictly more work) and wants
   calibration on the bench box.

   **The cost you should know about:** adding a method to the Gno `Banker`
   *interface* breaks every implementor of it, not just callers. `NewBanker` and
   `NewReadonlyBanker` return the interface, so the method has to be on it to be
   reachable at all — a method on the concrete type alone would need a type
   assertion at every call site. In-tree that cost one mock
   (`p/jaekwon/allowancesender`'s test `mockBanker`). Out of tree, any realm with
   a Banker mock, fake, or decorator stops compiling until it adds the method.
   That is unavoidable given the interface shape, but it is a real
   backwards-compatibility break and belongs in the release notes.
5. **Should this be one PR or two?** Two reviewers said two, stacked: the
   denom-validation commits have nothing open, while the storage commit has
   item 4 open and could stall on it.
6. **The `.gitignore` hunk is unrelated** to denoms — a reviewer asked for it to
   be split out. I kept it because the seven contribs/misc binaries it ignores got
   swept into a commit once already (90 MB, purged), and reappeared within one
   command of my removing the entries. Split it into a `chore:` PR if you prefer.

Carried-forward limitations, all disclosed in the PR body and not needing a
decision: denom *count* is still unbounded; denom bytes moved from a metered
value into an unmetered store key; `GetCoins` is still O(denoms held);
`bankerGetCoins` native gas wants recalibration; `gno test`'s `TestBanker` does
not model the split.

---

## 1. Background — why not full ADR-004 (settled; kept for the record)

> **Superseded in one respect.** This section argues for splitting on a leading
> `/`. What shipped splits on an allowlist of gas denoms instead — see §2 — after
> the `/` rule was found not to survive IBC. The gas reasoning below is unchanged
> and is why full ADR-004 was rejected; ignore the parts that describe the
> boundary itself, including the claim that `GetCoins` can concatenate.

**This is the big one. Please read before merging anything in this area.**

You asked me to port cosmos-sdk ADR-004 ("Split Denomination Keys"): move each
denom's balance out of the account's amino blob into its own store key. I wrote
the plan and had three independent subagents review it. All three came back with
the same finding, by three different derivations:

| Reviewer | Regression | Break-even point |
|---|---|---|
| A | ≈ +18% per address touched | ~7 denoms (at 276 B/denom) |
| B | ≈ +32% per single-denom transfer | — |
| C | ≈ +366k gas/tx, ≈ +35% | ~43 denoms |

### Why it is structural, not an implementation detail

gno.land **pins** store depth gas to fixed values, so cost is independent of
tree size (`gno.land/pkg/sdk/vm/params.go`: Get 100 / SetRead 200 / Write 540).
With `ReadCostFlat 59_000` and `WriteCostFlat 24_000`:

- a charged read  = `1.0 × 59,000` = **59,000** + 17/byte
- a charged write = `2.0 × 59,000 + 5.4 × 24,000` = **247,600** + 14/byte

And `cacheStore` makes a repeat read **free** and refund-dedups repeat writes to
the same key. So today a transaction pays for each *account key* exactly once,
no matter how many times the bank touches it.

The sender's account object is written anyway — the ante handler bumps the
sequence. So after the split you pay for **two** keys on the sender side
(`/a/S` for the sequence, `/b/S/ugnot` for the balance) where today you pay for
one. That is +1 read +1 write ≈ **+306,600 gas per address**, and an ordinary
transfer touches 2–3 addresses. The fee path alone accounts for ~+366k, so even
a transaction that moves no coins gets more expensive.

For scale: `gnokey_gasfee.txtar` pins a trivial `maketx call` at 1,024,011 gas.

### The trade, stated plainly

ADR-004 buys: a balance operation costs O(1) instead of O(total denom bytes), so
the griefing vector — mint junk denoms into a victim's account and make all of
their transactions expensive — is closed. At 1,000 junk denoms a victim
currently pays ~8.5M gas per transaction.

ADR-004 costs: ~+35% on every transaction for everyone, forever, and it only
starts paying for itself above roughly 40 denoms per account.

Denom *count* stays unbounded either way, and denom bytes move from a metered
value into an **unmetered key**.

> **Correction (two errors in my first draft, both caught independently by two
> reviewers).** I wrote that spam is "self-limiting" today because N denoms cost
> O(N²). That is wrong: `cacheStore.Set` refunds the prior charge for the same
> key, so minting N denoms in **one transaction** costs one read plus one write
> at the *final* blob size — O(N), not O(N²). Measured attacker cost for 1,000
> denoms: ~4.0M gas one-time, to impose ~8.5M gas on the victim's every
> transaction forever. Payback on the victim's second transaction. There is no
> self-limiting.
>
> And I had the sign backwards on spam cost: splitting makes bulk denom spam
> **~76× dearer per denom** (306,600 for a fresh key vs ~4,030 amortised inside
> one blob rewrite), not cheaper — while making it free for the victim. That
> flips this from an argument against the split into one for it.

### Why not full ADR-004

Splitting *every* denom, including the gas denom, was measured at +18%/+32%/+35%
gas on every transaction (three independent derivations), break-even around 40
denoms held. The reason is structural: the sender's account object is written
anyway for the sequence bump, so a balance kept there is free, while splitting it
buys a second key at ~306,600 gas. That is why an account tier exists at all.

The alternatives considered and rejected — a per-account denom cap, recipient
consent for issuance, and keying the tier on the denom's shape — are recorded in
the ADR's Alternatives section, along with the dissenting recommendation. The
earlier draft of this file argued for the shape-based rule; it does not survive
IBC, and the ADR explains why.

---

## 2. What I implemented

(c′), as described above. The tier decision lives in exactly one predicate so that
moving to full (d) later is a small code change plus a genesis-boundary migration
  (the ADR is the authority here and is blunter: calling it "a one-line change" would
  be wrong — it also breaks the `auth/accounts` wire format and 20 txtar assertions) —
`gnogenesis fork` already rebuilds chains from exported balances, so there is no
in-place migration code to write either way. The keeper API is per-denom
throughout, which is the expensive, review-heavy part of ADR-004 and is shared by
(c′) and (d): *the layout is a parameter of the design; the API is the design.*

---

## 3. Correctness traps the reviews found (all confirmed against source)

These were found while planning, and applied to *any* option that moves balances
out of the account. Marked with the outcome under the design actually shipped.

**Made moot by keeping genesis denoms in the account object:** items 5 (amino
field renumbering across seven account types, `std.proto`/`pb3_gen.go`
regeneration) and 6 (`test_common.go` build break) never arose, because
`BaseAccount.Coins` was not removed. Likewise the §4 note about gnokey breaking
against an old chain: the `auth/accounts` wire format is unchanged.

**Handled:** 1 and 2 (account creation on receive, and its account-number
timing) — `ensureAccount` fires for both tiers, pinned by test. 3 (`SpendableCoins`
reintroducing O(n)) — the vesting check loops per denom against `LockedCoins`
instead, and `std.SpendableCoins` now has no production caller. 7
(`auth.BankKeeperI` lacking a balance read) — widened with `GetCoin`; the error
stays `InsufficientFundsError`. 8 and 9 (zero-value encoding, overflow) — fixed
8-byte big-endian with delete-on-zero, and `overflow.Add`, both pinned.

**Still open:** 4 (the vesting constructors validate against `baseAcc.Coins`,
which for a realm-denom genesis allocation is now the wrong tier — see §3c) and
10 (`upgradeVestingAccount` drops `GnoAccount.Attributes`, pre-existing).
Item 11 is closed: `bank/invariants.go` was deleted.

1. **Receiving coins currently creates the account.** `bank.SetCoins` calls
   `NewAccountWithAddress` when the account is missing, and
   `auth.GetSignerAcc` **rejects a transaction from an address with no
   account**. Move balances out without replicating that and a fresh recipient's
   funds become permanently unspendable. `maketx_call_send.txtar` pins the
   current behaviour.
2. **Account-creation timing is consensus state.** `GetNextAccountNumber`
   increments a global counter, so changing *when* accounts are created shifts
   every later account number and the app hash.
3. **`std.SpendableCoins` must not be given the balance.** It iterates the whole
   balance set, so threading the balance in would reintroduce O(N) on every send
   from a vesting account — worse than today. Cosmos never calls it on the send
   path; it loops per denom in `amt` against `LockedCoins` (account-local). Port
   cosmos's `subUnlockedCoins`, including its "locked amount exceeds account
   balance" guard.
4. **The vesting constructors validate against `baseAcc.Coins`**
   (`vesting_account.go:151,243`), and genesis sets `Coins` specifically to
   satisfy them. Remove the field and that validation silently vanishes —
   `Balance.Verify()` is *not* called on the InitChain path.
5. **Removing `BaseAccount.Coins` renumbers every following amino field**
   (`PubKey` 3→2, etc.) across seven account types. Needs
   `_ [0]struct{} \`amino:"reserved"\`` — precedent in
   `tm2/adr/pr5301_amino_reserved_field_numbers.md` — plus regeneration of
   `std.proto` and every `pb3_gen.go`.
6. **`tm2/pkg/sdk/auth/test_common.go` is not a `_test.go` file** and mutates
   account coins directly. It breaks the build first, and `auth` cannot import
   `bank` to fix it the easy way.
7. **`auth.BankKeeperI` has no balance read**, so the ante pre-check cannot be
   made per-denom without widening it. Not widening it changes the ABCI error
   from `InsufficientFundsError` to `InsufficientCoinsError`, which tests and
   gnokey surface.
8. **`amino.Marshal(int64(0))` returns zero bytes** and `AssertValidValue`
   panics on a nil value — so "delete zero balances" is load-bearing, not
   stylistic. Fixed 8-byte big-endian avoids the whole class.
9. **Per-denom `int64` arithmetic must use `overflow.Add`.** `Coins.Add` panics
   on overflow today, which means `AddCoins`'s `!newCoins.IsValid()` check is
   dead code. A plain `+` would silently wrap into a negative balance — minting.
10. **`upgradeVestingAccount` downgrades `*GnoAccount` → `*std.BaseAccount`**,
    silently dropping `Attributes` — including the token-lock whitelist bit read
    by `canSendCoins`. Pre-existing bug, but in code this work touches.
11. **`bank/invariants.go` has zero callers** — dead code. Don't port it.

---

## 3b. Separate issues the reviews surfaced — worth their own fixes

These are not the storage layout. Each is independently actionable.

1. **`X_bankerIssueCoin` validates nothing in Go.** `gnovm/stdlibs/chain/banker/
   banker.go` forwards straight to `Banker.IssueCoin` with no denom-prefix check
   and no banker-type check — both live *only* in `banker.gno`'s
   `assertCoinDenom`. `SDKBanker.IssueCoin` validates nothing either. Under (c′)
   "starts with `/`" is a **security** invariant, so I re-assert it in Go. Worth
   doing regardless of which option you pick — defence in depth for an invariant
   currently enforced only in interpreted stdlib source.
2. **The fee collector is a chain-wide amplifier.** `DeductFees` calls `AddCoins`
   into `auth.FeeCollectorAddress` on **every fee-paying transaction**, and that
   address is `AddressFromPreimage([]byte("fee_collector"))` — publicly derivable
   and **keyless**, so nobody can ever sign for it to clean it up, and only the
   issuing realm can `RemoveCoin`. Polluting *it* adds cost to every transaction
   on the chain, permanently, recoverable only by a governance param rotation the
   attacker can immediately re-pollute. (c′) closes this (junk lands in `/b/`
   keys, never in the collector's blob). Same applies to `storage_fee_collector`.
   **This is the strongest argument for fixing it now rather than later.**
3. **Unmetered O(N²) CPU inside `AddCoins`.** `Coins.AddUnsafe` allocates a fresh
   ~N-element slice per call, and `bankerIssueCoin` is metered at a flat 141 gas,
   so a realm looping `IssueCoin` N times does O(N²) copying for O(N) gas. One
   reviewer measured 1,000 mints in one transaction at 138 MB marshaled / **1.755 s
   wall** for 4.17M gas charged — roughly 420× underpriced against the stated
   ~1 gas ≈ 1 ns calibration. A validator-CPU vector independent of layout.
   (c′)/(d) fix it as a side effect (per-denom read-modify-write is O(1)).
   Not measured by me; hand it to the bench box.
4. **Possible permanent account loss, not just expense.** There is no maximum
   value size in the store (`AssertValidValue` rejects only nil). At ~97 MB of blob
   the victim's own read+write (31 gas/byte) alone exceeds the 3B block gas limit,
   so the account becomes unusable *and* unrecoverable — even sending the junk away
   cannot fit in a block. Reachability depends on item 3's per-iteration gas, which
   is unmeasured. **If reachable, this is loss of funds, not griefing.** Please
   have someone measure it.
5. **The write-refund dedup is itself mispriced.** Repeated writes to any large
   value are unpriced beyond the final one, which is what makes item 3 possible.
   Fixing that class (e.g. not refunding the per-byte component) is separate from
   this work.

---

## 3c. Found during implementation review — decisions for you

1. **"Genesis cannot define a realm denom" is false, and I had asserted it.** All
   three reviewers independently checked: `ValidateDenom` accepts a leading `/`,
   `gnoland.Balance.Verify()` does not reject one, and `Verify()` is not even
   called on the InitChain path (only by `gnogenesis verify`). So a genesis file
   *can* allocate into the realm tier — at a path a realm may later be deployed
   at, which would then hold `RemoveCoin` authority over a genesis allocation.
   The runtime behaves correctly either way (I removed the false premise from the
   comments and made the vesting check explicitly tier-agnostic), but **the
   invariant is unenforced**. Options: reject `IsRealmDenom` in `Balance.Verify()`
   *and* in `applyBalance` (Verify alone is not on the chain path), or accept it
   and document that genesis is trusted. I did not want to add a genesis
   validation rule without your call.
2. **Vesting genesis with a realm denom leaves the account object inconsistent.**
   `applyBalance` sets `baseAcc.Coins = bal.Amount` so the vesting constructor's
   `IsAllGTE(OriginalVesting)` check passes, then `SetCoins` strips the realm
   denom back out — so `OriginalVesting` can reference a denom the account object
   no longer shows. Enforcement still reads the right tier. Falls out of item 1.
3. **`DeductFees` now does one extra account decode per transaction** for a
   genesis-denom fee: it calls `bk.GetCoin`, which re-fetches and amino-decodes
   the account the caller already holds. Gas is unchanged (`cacheStore` refunds the
   repeat read) but it is real CPU on the hottest path. A one-line fix is to branch
   on `IsRealmDenom` and use the account in hand otherwise; I left it alone because
   it trades clarity for CPU on a path I have not measured.
4. **`bank.SetCoins` now does a prefix-iterator seek per address at genesis**
   (to find stale realm keys). Correct, but at InitChain the keyspace is empty, so
   a large genesis pays one wasted seek per balance.
5. **`gnoclient` has no balance API.** `QueryAccount` returns `*std.BaseAccount`,
   whose `Coins` is now the genesis tier only. Its real callers want
   accNum/sequence, so nothing is broken, but a `QueryBalance` helper is missing
   and the omission invites the same mistake gnodev made.
6. **`TestBanker` (used by `gno test`) does not model the split** and does not
   enforce the new Go-side prefix check, so a realm can pass `gno test` and panic
   on chain. Test-fixture only, but worth aligning.

---

## 3d. The reported flake — diagnosed, and it found a real defect

A round-2 reviewer reported intermittent failures with three symptoms:
`encodeBalance` panicking on a negative balance while the guard three lines above
compared the same two values; an iavl prefix iterator returning keys in insertion
order; and one `TestAppHashCrossrealm38` Merkle mismatch. I could not reproduce
any of it (15/15 clean, `-race -count=3` clean, 6/6 on `sdk/vm`), and previously
recorded it here as probable host-level nondeterminism.

**That was the wrong conclusion. Symptom 1 was a real defect in this branch, and
it is now fixed.**

`subtract`'s split-tier loop guards with `old < coin.Amount` and then computes
`old - coin.Amount` as raw `int64`. A **non-positive** amount passes that guard,
because `old` is never negative. Reproduced by removing the caller-side
`amt.IsValid()` that every public path happens to perform:

    debit of -500  ->  err=nil, balance 100 -> 600      (a debit that mints)
    debit of MinInt64 -> panic: refusing to encode non-positive balance
                         -9223372036854775208

The second line is the reported symptom verbatim, including the apparent
impossibility: the guard really does compare the same two values three lines
above, and really does pass.

This was **a regression introduced by this branch**, not pre-existing. Master's
`subtract` went through `Coins.SubUnsafe` followed by `IsValid()`, which rejects a
negative result. Splitting balances replaced that, for the split tier only, with
raw arithmetic — the account tier still goes through `SubUnsafe` and was never
exposed. So the reviewer was seeing something real; what I could not reproduce was
the *trigger*, because on the committed tree every caller validates.

Fixed by making `subtract` defend its own precondition rather than trusting five
callers, with the reasoning stated at the guard. With `0 < amount <= old` the
subtraction cannot overflow. Pinned by
`TestNoPathAdmitsANonPositiveDebit`, which drives every public entry point with
negative and `MinInt64` amounts *and* calls `subtract` directly to bypass caller
validation; removing the new guard fails it.

**Still unexplained: symptoms 2 and 3.** An iterator returning insertion order and
an app-hash mismatch are not producible by a non-positive debit, and I have no
mechanism for them. The most probable cause is the one I can evidence rather than
prove: review agents were running builds and tests in this **shared worktree while
other agents were editing it** — two round-4 agents independently reported files
changing underneath them mid-run. A tree that is inconsistent between the moment a
guard was compiled and the moment a fixture was written produces exactly that pair
of symptoms, and explains every piece of negative evidence: a pre-built binary
never failed in 100+ runs, `-race` never fired, a clean temp worktree was 0/8, and
master in adjacent windows was 0/8 because nobody was editing master. **Process
fix: review agents must work in their own worktree, not the branch's.** Worth one
CI confirmation run regardless.

## 3e. Round-4 adversarial design review (three agents, briefed to argue against)

All three were asked to build the case for removing the account tier. All three
concluded **keep it**, and the strongest argument for removal is worth recording
because it is about the future, not the present.

**The gas case got stronger, not weaker.** Measured by simulating the same
transaction against a binary with an empty allowlist — literally the path to full
ADR-004 — a warm `MsgSend` goes 932,933 → 1,355,208 (**+45%**), and a fee-only
call 620,817 → 985,098 (+59%). The ADR previously quoted +18/32/35% from earlier
reviews; nobody had measured a plain `MsgSend` on a warm chain. Corrected in the
ADR, along with the per-key attribution, which was wrong: only the *signer* pays a
full extra read+write, because only the signer's account object is written anyway.
For the fee collector and recipient the true cost is ~+58,000, not +306,600.

**The argument against, which I did not adopt and you may want to.** The tier's
entire value is a function of `Fixed*Depth100` and `WriteCostFlat` — numbers this
chain sets for itself, which currently make one store write ~24% of a trivial
transaction. Work on the bptree fast index is actively making writes cheaper. The
tier trades a structural property (one value, two homes, forever) for a transient
constant, and **removing it later costs a state migration that removing it now
costs nothing**, since nothing has shipped. If you expect store-write gas to fall
by more than ~3x over this chain's life, remove the tier now and accept the 45%.
I left it in because a 45% regression on the hottest path is not something to pay
for a hypothetical.

**The tier is weaker for session-signed transactions than the ADR claimed.** The
ante handler bumps the sequence on the session account, not the master's, so a
session-signed transaction writes the master's account object *only* to carry the
balance — the one thing the tier assumes is free. The advantage roughly halves
(+115,001 fee-only, +173,133 `MsgSend`). Now stated in the ADR. As sessions become
the default signing path, the tier earns less.

**Both designs close the attack identically** — a `MsgSend` with 200 junk denoms
on the sender measured byte-identical to the 0-junk case in both arms. The tier is
a pure gas optimisation with no security content. That is worth knowing: if it is
ever removed, no security property is lost.

Layout review found **no cheaper encoding**, and the reason is worth keeping: at
14 gas/byte, the entire 8-byte value is 112 gas out of a 247,712-gas write —
0.045%. Key bytes are unmetered entirely. Encoding is not where the cost is; key
*count* is. Rejected with numbers: hashing the denom (an 8-byte truncation is a
targeted second-preimage at ~2^64, which would let an attacker mint into a
valuable denom's slot — and it destroys `GetCoins`); denom-first keys (silently
wrong without a separator, and nothing wants denom→holders); amino values (saves
≤98 gas of 247,712); storing zeros instead of deleting (a delete is 112 gas
*cheaper*, and each dead key then costs 1,136 gas in every future iteration);
a packed per-address blob (loses at k=1, the dominant case, and its win region is
exactly its attack region).

### New findings from this round, still open

- **Bank balance keys are the cheapest permanent state on the chain.** One fresh
  key costs ~306,712 gas ≈ 307 ugnot for 66 bytes ≈ **4.65 ugnot/byte** (approx; derived from the flat write cost), falling
  to ~1.01 at a maximal denom. Realm state costs 100 ugnot/byte — and that is a
  *refundable deposit*, whereas `processStorageDeposit` works from
  `RealmStorageDiffs()` and never sees bank keys. So this state carries no deposit
  and, with no burn path and issuer-only `RemoveCoin`, the victim can never reclaim
  it. Denom count remains unbounded. Bounding key count is the lever the ADR
  leaves open; charging for key bytes is not (it would raise an attacker's cost
  1.5% worst case).
- **`SetCoins` is now O(denoms held) full writes** — 32.2M gas at 128 denoms. Every
  caller today runs at genesis with a known count, but it is exported on
  `BankKeeperI`, so a future handler reaching for it would be an instant
  block-gas-limit DoS. Now carries a doc warning; consider unexporting it.
- **`DeductFees` re-reads the account it already holds** (`bk.GetCoin` on an
  address whose `acc` is in hand): ~3.6 µs of amino decode per fee-paying
  transaction, gas-free because `cacheStore` refunds the repeat read. It exists
  only to keep the ABCI error as `InsufficientFundsError`. Fixable by exposing
  `InAccountTier` on `ViewKeeperI`, at the cost of interface surface.
- **Amino account decode is ~an order of magnitude underpriced per denom** (a
  35-byte denom is charged ~595 gas and costs ~8.4 µs, because `ParseCoin` runs a
  regexp per coin). So the griefing case was worse in validator CPU than the gas
  table showed. The split removes it from the money path; `bank/balances` still
  pays it.

### Done, not deferred — realm callers migrated to `GetCoin`

An earlier draft deferred this. A measurement reversed the decision, and it is
worth recording because it is the strongest single argument for `GetCoin`:

Measured under tm2's **default** gas config (which scales reads by tree depth);
gno.land pins depth flat, so its absolute numbers differ. The ratio — and the fact
that repeats are free on one and charged on the other — is what matters here.

| | `GetCoins` | `GetCoin` |
|---|---|---|
| first call | 179,360 | 118,136 |
| **each repeat** | **60,136** | **0** |

`cacheStore` refunds a repeated *read*, but `splitCoins` opens an **iterator**, and
iterator opens are charged every time. So a realm that reads balances in a loop
now pays per iteration where before this change it paid for one account read and
got the rest free. That regression is caused by this PR, which makes fixing the
in-tree callers remedial rather than scope creep. Migrated:

- `r/demo/disperse/disperse.gno` — two sites, both **inside** loops over the coins
  being dispersed, so the cost was quadratic in denom count.
- `r/gnoland/boards2/v1/permissions_validators_open.gno` — read every denom to
  check one, on every non-member post, for a caller-controlled address.
- `r/gnoland/coins/coins.gno` — a single-coin balance view that read them all.

Pinned by `TestRepeatedGetCoinsIsChargedButRepeatedGetCoinIsFree`, which will
fail if iterator opens ever become refundable and the argument above stops holding.
The `examples/` change does **not** move the app-hash pin — verified — so the pin
comment still correctly attributes its shift solely to `banker.GetCoin`.

Not migrated, deliberately — a swept grep found three more single-denom
`GetCoins().AmountOf()` readers, all outside the main workspace:
`quarantined/.../r/docs/soliditypatterns/banker/banker.gno:110`,
`quarantined/.../r/sacha/coinflip/coinflip.gno:43`, and
`r/gnoland/wugnot/filetests/z0_filetest.gno:45`. The first two are quarantined
(resolvable by `ResolveExamplePath` but outside `examples/gnowork.toml`, so not in
the default build, lint or test sweep) and the third is a filetest, so none is on
a production gas path. Worth fixing whenever the quarantine is lifted. Note that
`quarantined/.../commondao/v0/proposal_treasury.gno:53` returns *every* balance by
design — a legitimate `GetCoins` use, not a candidate.

### One reported finding that is a false positive, recorded so it is not re-raised

`TestBanker.IssueCoin` does not mirror the new Go-side `assertIssuable`, which
looks like it would let a realm pass `gno test` and panic on chain. It cannot:
`assertCoinDenom` runs in `banker.gno`, which is the *same* interpreted source in
both environments, so realm code is gated identically either way. The Go-side
assertion is defence in depth for a bypassed interpreter, which realm code cannot
arrange.

## 3f. Round-5 adversarial review of the final tree (four agents, own worktrees)

Four agents attacked the then-final tree with distinct lenses. Each worked in its own git
worktree — the process fix from §3d. Findings acted on:

- **Critical, introduced by this PR and now fixed: `banker.GetCoin`'s denom was
  unvalidated.** Two agents found it independently. Every other denom entry point
  validates (`assertCoinDenom` for issuance, `Coins.IsValid` for transfers and
  fees); this one took a realm-supplied string straight into a store key. The
  charge is flat while the work is proportional to length — the tier lookup hashes
  the whole string, the key copies it, and the cache store retains that copy for
  the transaction. Measured 0.2 µs at a legal denom versus 103.8 µs at 1 MiB for
  identical gas, and one agent drove it through `vm/qeval` — **no transaction, no
  fee, no signature** — for ~15 GiB retained and ~12 s of CPU per block-gas-limit
  of unauthenticated query. Fixed in two places: `SDKBanker.GetCoin` validates and
  panics like every other entry point, and `ViewKeeper.GetCoin` rejects an
  over-long denom before hashing or building a key, which makes the bound
  `balance.go` documents structurally true rather than caller-dependent.
- **`subtract` trusted its callers on both tiers, not just the split one.** The
  previous round hardened the split loop and justified leaving the account tier
  alone on the grounds that `Coins.SubUnsafe` revalidates. That reasoning was
  wrong: `SubUnsafe` validates the *result*, and a negative debit yields a larger
  positive result, which is valid — subtracting −500 from 100 gives 600. The guard
  now covers the whole amount before the tier split.
- **`subtract` wrote the infallible split keys before the fallible account write**,
  so its "past this point nothing can fail" comment was false and a failing account
  write would have kept the split debits. Reordered to match `AddCoins`.
- **Test gaps closed**: `splitCoins`'s exclusivity panic had no test (and the ADR
  claimed both directions were tested — corrected); the `bank/balances` missing-return
  fix had none; `AddCoins`'s `IsValid` gate had none; `SDKBanker.GetCoin` had no
  Go-level test. Three were mutation-verified at the time; the `bank/balances` one was
  **not**, and a later audit found it hollow — the pre-fix code did set `res.Error`, so
  asserting the error alone could not fail. It now asserts `res.Data` is empty, which
  is what the fix actually changes. Recorded because the round-5 note claiming
  otherwise was wrong. One test I wrote this round was itself
  hollow — it used `BaseSessionAccount`, which reports no coins, so `subtract`
  failed its solvency check before reaching the code under test. Replaced with a
  local account type that holds coins but refuses to store them.
- **Six documentation errors corrected**, including a fee-path percentage computed
  against the wrong base (+36% should be +59%), a fresh-key gas figure that charged
  per-byte on a read miss (306,848 → 306,712), two tables quoting the same
  operations under different gas configs without saying which, and a "reached only
  by" comment that omitted `SetCoins`.

Reported and deliberately **not** changed:

- **`sendCoins`'s debit and credit are not atomic as a pair.** With the recipient
  holding `MaxInt64` of a denom, the credit panics after the debit is written. The
  transaction rolls back, so it is only visible to a caller that recovers and
  continues, and it is unchanged from master. Left alone because fixing it means
  restructuring the transfer pair, which is beyond a storage change.
- **Bank balance keys carry no storage deposit** (`processStorageDeposit` works off
  `RealmStorageDiffs()` and never sees them), so permanent state costs ~1/99th what
  realm state does. Not a regression — junk in the account object carried no
  deposit either, and per unit of attacker cost the split made state growth ~69×
  harder — but the brake is a gas constant rather than a deposit, so a future gas
  recalibration could reopen it. Bounding denom *count* is the lever, recorded as a
  possible follow-up.
- **A realm can still be made expensive to call** by loading an address it reads
  with `banker.GetCoins`: ~0.62 GNOT to exceed a 10M gas-wanted, versus ~4,200
  ugnot before the split — 147× dearer, not closed. `GetCoin` is the fix for realms
  that need one balance.
- Two pre-existing bugs outside this PR's scope, both reproduced on master: a
  failed `SubtractCoins` still rewrites the account object via
  `upgradeVestingAccount`, and on gno.land that replacement is a `*std.BaseAccount`
  rather than a `*GnoAccount`, so a genesis file listing a vesting address in
  `unrestricted_addrs` aborts InitChain with an interface-conversion panic. Also
  `Balance.MarshalAmino`/`Parse` cannot round-trip a multi-denom vesting schedule,
  so `gnogenesis export` of such state emits an unbootable file.

## 3g. Round-6 sweep: classes no earlier round had examined

Checked and clean, recorded so they are not re-checked:

- **Determinism.** `accountDenoms` is a Go map, whose iteration order is random. It
  is only ever written at construction (ranging the ordered input slice) and read
  by key lookup — never iterated — so it cannot influence consensus state.
- **State export.** Neither `bank.ExportGenesis` nor `auth.ExportGenesis` exports
  balances; both return `Params` only. So there is no module export path that could
  silently drop split-tier balances.
- **`gnogenesis fork`.** It queries `auth/accounts`, which now reports a partial
  balance — but only for `(accNum, sequence)` to resolve transaction ordering. Zero
  balance reads in that package. Balances come back via replay, as the ADR assumes.
- **`gnogenesis balances add/export/remove`** operate on the genesis file's
  `state.Balances` only, never a live chain, so they route through `InitGenesis` →
  `SetCoins`.
- **`gnokey`.** `maketx` and `verify` read `auth/accounts` for account number,
  sequence and pubkey — never coins.
- **Balance events.** Every `EmitEvent` in the bank module is commented out
  (pre-existing), so no event stream can miss a split-tier movement.
- **Storage deposits.** `RealmStorageDiffs()` is a VM-internal map keyed by PkgID
  built from realm object tracking, not a store-level diff, so a `/b/` write cannot
  be miscounted as realm storage and charged a deposit.
- **`canSendCoins` / `RestrictedDenoms`** test the denoms in the *request* plus the
  account's whitelist flag. Tier-agnostic; unaffected.
- **Every module that builds against this branch** (13 via a local `replace`)
  compiles: gnodev, gnogenesis, tx-archive, gnohealth, github-bot, gnobr, gnokms,
  gnobro, gnomigrate, gnokeykc, misc/loop, misc/autocounterd, stress-test. Of these
  only gnodev read a balance, and it was changed. `gnofaucet` and
  `misc/docs/tools/linter` pin a released version and are unaffected until bumped.
- **The shipped `genesis_balances.txt`** contains `ugnot` only — zero slash-prefixed
  denoms — so nothing in it changes tier.

Gap found and fixed: **the guidance realm authors read had not been updated.** Only
the API reference (`gno-stdlibs.md`) mentioned `GetCoin`. Reading one balance with
`GetCoins(addr).AmountOf(denom)` is now both slower per call and a griefing surface
whenever `addr` is caller-supplied, and three in-tree realms had exactly that bug,
so it is a pattern people write. Added to `docs/resources/effective-gno.md`, as case
11 of `docs/resources/gno-ai-contract-review.md` (with a checklist entry), and as a
rule in `AGENTS.md` — whose "(10 cases)" count was stale as a result.

## 3h. Naming, and a pinned extension point

- **`ViewKeeper.GetBalance` renamed to `GetCoin`.** Every other balance method on the
  keeper uses the "Coins" vocabulary — `GetCoins`, `HasCoins`, `SetCoins`, `AddCoins`,
  `SubtractCoins`, `SendCoins`, `InputOutputCoins` — so `GetBalance` was the sole
  outlier, importing half of cosmos's naming (`GetBalance`/`GetAllBalances`) into a
  keeper that does not use the other half. `TotalCoin(denom) int64` was already
  precedent for a singular `...Coin` returning a scalar. 60 sites across 9 files; the
  Go and Gno layers now share one name. The `bank/balances` query route is unchanged —
  it is an external wire path, and plural is right for it.
- **`TestSecondGasDenomInAccountTier`** pins the allowlist's growth path, which nothing
  otherwise exercised: a second gas denom (an IBC voucher, lowercase-normalised) held
  in the account tier alongside `ugnot`. It matters because such a denom sorts *before*
  `ugnot`, so the account tier is neither a prefix nor a suffix of the sorted result
  and `GetCoins` has to interleave three ways. The test deliberately includes split-tier
  denoms on **both** sides of the account-tier ones — with all split denoms sorting
  first, concatenating the tiers coincidentally yields sorted output and the merge is
  not actually under test. Verified: it now fails both when the merge is replaced by a
  concatenation and when the tier rule is hardcoded to `ugnot`.

## 3i. Audit of the finished branch (four agents, own worktrees)

Findings acted on. Every fix below is mutation-verified — deleting it fails its own test.

- **`computeSupply` swallowed the account-tier overflow error.** `IterateAccountEntries`
  returns the *iterator's* error, so returning true from the callback only stopped the
  walk: `computeSupply` returned **truncated totals with `err == nil`**. That meant
  `RecomputeSupply` seeded a wrong genesis counter instead of refusing to start, and
  `SupplyInvariant` re-derived the same wrong total and reported healthy — the one
  redundancy check in the set, self-consistently wrong. The auditor reproduced it, and
  it was **untested after I fixed it**, which is why the audit mattered twice over.
  Now carried out explicitly and pinned by
  `TestComputeSupplyReportsAnUntotallableSum`.
- **The `applyBalance` fix I had just made was itself a regression.** Reusing the
  account meant a plain genesis entry after a vesting entry for the same address no
  longer cleared the schedule, so the funds stayed locked until `EndTime` — on master
  the plain entry replaced the account and they were spendable. The account's *kind*
  is now replace-all too, with the number carried over. This also restores the
  `*GnoAccount` concrete type, which `applyUnrestrictedAddrs` asserts unchecked.
- **`ValidateDenom`'s regexp was ~40× under-metered on a path `GetCoin` put it on.**
  Measured 4,446ns for a maximal 274-byte denom against a 349-gas native charge.
  Replaced with an equivalent byte scan: **174ns**, a 25× improvement that turns the
  under-charge into an over-charge, and speeds the invariants (which validate every
  denom in state) by roughly the same factor. Equivalence is gated by
  `TestValidDenomMatchesRegexp`, which compares against the regexp over every byte
  value in each of the first three positions.
- **`ParseCoin` ran its regexp before any length check.** `Coins` amino-encode as a
  string, so every transaction decode reaches it — and decode happens *before* the
  ante handler installs a gas meter, so a ~1 MB coin expression bought ~270ms of
  validator CPU for zero gas, spammable through CheckTx. The 274-byte denom cap is
  documented as bounding key size; it did not bound this. One length guard, moved
  above the pattern.
- **`RecomputeSupply` could brick a fork.** It panicked above `maxSupplyDenoms`
  (262,144 distinct denoms, reachable for ~160B gas), so a chain past that bound could
  never be re-genesised from an export. The bound now applies only to the invariant,
  which must not risk an unrecoverable allocation on a node; genesis is trusted,
  one-time, and must not refuse to start. The comment claiming parity with
  `auth.maxUniquenessBits`' memory was also wrong — 80 MB, not 8 MB — and is corrected.
- **`SupplyInvariant`'s report was nondeterministic.** It ranged a Go map and the
  report is truncated at ten, so two operators inspecting identical state saw
  different findings (measured: 20 distinct messages over 20 runs). Sorted.
  `RecomputeSupply` also ranges a map, but for *writes*, and that is safe only because
  the cache store sorts keys before flushing — verified as one app hash over 25
  identical genesis runs, and now stated at the site, since nothing else says it.
- **`subtract` and `nextSupply` trusted their callers on two more preconditions**: that
  the account passed in belongs to the address, and that denoms are strictly ascending.
  A mismatched pair debited two different addresses; a duplicated denom debited or
  minted once while the caller believed twice. Both now checked, both internal-only.
- **`std.ErrInvalidCoins` discards its message** (`%v` renders only "invalid coins
  error"), so these internal guards used plain errors instead — a caller bug needs to
  say which precondition failed. The wart itself is pre-existing and worth its own fix.
- **`TestBanker.TotalCoin` summed without an overflow check**, so `gno test` would
  report a wrapped negative where the chain refuses the mint. Overflow-checked.

Reported, pre-existing, and **not** fixed here:

- `DeliverTx` has no fee-denom gate, so a self-minted realm denom can pay gas
  (`EnsureSufficientMempoolFees` is CheckTx-only by design, since `minGasPrices` is
  node-local). Unchanged in kind from master, and the split *improves* the
  consequence — the collector's junk now lands in its own keys rather than bloating
  the account blob every fee-paying transaction rewrites. Worth a decision against the
  threat model.
- `bank/balances` runs under an infinite gas meter with no fee: at 20,000 denoms of 274
  bytes, ~160ms and 5.6MB of response from a ~60-byte request. Master was worse
  (~2.5×), and the victim no longer pays — node operators do.
- Exposing the invariants as a query or in `BeginBlock` would be an unmetered DoS
  (~29µs/key before the `ValidateDenom` fix), and a nil-gas-context read warms the
  cache store so a later metered read on the same key is free. No production caller
  exists today; the mitigation is a nested `CacheContext` around the sweep.

## 3j. What this file does and does not cover

Written during the storage split and extended as the branch grew, so it is organised by
review round rather than by subject. Three later changes are documented primarily in
their own ADRs rather than here:

- **The invariant functions** — `tm2/adr/prxxxx_bank_invariants.md`. Called from tests
  only; no registry, no runner, deliberately.
- **Per-denom supply tracking** — `tm2/adr/prxxxx_coin_supply.md`. Note the one
  consequence a reviewer should not miss: **per-denom supply is now capped at
  `MaxInt64`**, which is a new consensus rule. It closes a hole (two addresses could
  each hold `MaxInt64` before) but it is a tightening, and a mint past the cap now
  fails.
- **`INVARIANTS.md`** — the roadmap for what was deliberately not built. Its invariant
  table is the original specification and four of its rows were refuted; read the ADR
  first.

### Known-unpinned checks, stated rather than implied

Six invariant checks have no violating test and are known-unpinned: the key-ordering and
iterator-error reports in `BalanceKeysInvariant`, the iteration-error report in
`AccountTierInvariant`, and the three session checks in `AccountKeyspaceInvariant`. The
first three fire only on a store-level fault that no state can produce, so they are
untestable without a fake store. **The three session ones are a real gap** — a session
whose master has no account object, a session claiming a different master, and a
delegated account filed at a regular path all survive deletion with the suites green.
Worth closing.

Also unpinned: `sdk.Guard`'s recover and `InvariantReport`'s ten-finding cap (nothing
references either in a test), `std.ParseRealmDenom`'s charset and pkgPathLimit branches,
`SubtractCoins`'s vesting clamp, `auth.DecodeAccountSafe`'s recover, and
`peekGlobalAccountNumber`'s must-not-bump property.

## 4. Things to review even after this looks done

- **The `/b/` key prefix inventory.** Three reviewers independently enumerated
  every prefix in the shared `mainKey` store and agreed `/b/` collides with
  nothing, and that no prefix iterator can reach it. Worth re-checking if anyone
  adds a store prefix. Note the store already contains **unprefixed** keys
  (GnoVM escaped-object hashes, `[0-9a-f]`-leading), so "check against `/…`
  prefixes" is not sufficient.
- **No length prefix on the address in the balance key.** Safe only because
  `crypto.Address` is fixed `[20]byte`. If address size ever becomes variable,
  this key format needs cosmos's `MustLengthPrefix`.
- **`bankerGetCoins` native gas needs recalibration**, not just golden
  regeneration: `gnovm/stdlibs/native_gas.go` fitted `PostSlope: 36206` against
  a single-blob read.
- **`gnogenesis fork` reads account state from old chains** — the data-dir path
  *panics* on unknown amino fields, the RPC path degrades silently to empty
  balances. Decide whether cross-version account decode must keep working.
- **`GetCoins(addr)` stays O(N)** under any split design — the `bank/balances`
  query, `banker.GetCoins()` in Gno, and any realm iterating balances still pay.
- **Denom count is unbounded under every option except (b).**
- **Most banker methods in `gnovm/stdlibs/chain/banker/banker.gno` still have no
  doc comments**, and neither does `chain.CoinDenom`, so `gno doc` readers get
  nothing while Go readers get the full explanation. `GetCoins` and `GetCoin` now
  have them, since those two are the ones this change makes a choice between — and
  they were free, because adding `GetCoin` already moved the app-hash pin. The
  remaining methods are unrelated to this change; documenting them is worth doing
  and belongs with a change that touches them.
