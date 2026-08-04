# PR6034: Store non-gas balances in their own keys

## Status

Proposed

## Context

An account's coin balances live inside the account object, in its `Coins`
field, which amino-encodes to a single comma-joined string
(`Coins.MarshalAmino` → `Coins.String()`). Store gas is charged per byte of the
value read and written, and the account object is read and rewritten on **every
transaction its owner sends**, because the ante handler bumps the sequence.

Separately, `chain/banker`'s `IssueCoin` lets a realm mint an arbitrary denom to
an **arbitrary recipient with no consent** (`SDKBanker.IssueCoin` →
`bank.AddCoins`; no recipient check anywhere on that path). A realm may mint
unlimited *distinct* denoms — `assertCoinDenom` constrains the shape
(`/pkgPath:base`) but not the number.

Put together, an attacker can make a victim's account object arbitrarily large,
and the victim then pays for it on every transaction they send, forever:

| junk denoms | account object | extra gas per victim tx |
|---|---|---|
| 32 | ~8.9 KB | ~+274,000 |
| 128 | ~35 KB | ~+1,095,000 |
| 1,000 | ~276 KB | **~+8,556,000** |

Measured with **maximal-length denoms** — ~283 bytes each including the amount and
separator, the worst case an attacker can choose freely — which is why the object
grows so fast. The gas follows from gno.land's pinned constants: the victim's own
transaction reads and rewrites the object at `17 + 14` gas per byte, so the extra
cost is ~31× the bytes added. Scale it for a shorter denom: a typical
`/gno.land/r/demo/foo:gold` costs about 31 bytes, so 1,000 of those is ~+960,000
per victim transaction rather than ~+8.5M. Still a permanent unilateral tax on
someone else's account.

The victim cannot cheaply undo it. There is no burn message anywhere in the tree
and `RemoveCoin` is issuer-only, so the only disposal is to transfer the junk to
another address — one transaction per cleanup, against an attacker who can re-mint
for less and chooses the timing. Minting is also cheap for the attacker: `cacheStore` refunds the
prior charge for repeated writes to the same key, so minting N denoms in one
transaction costs one write at the final size — roughly 4.0M gas for 1,000
denoms, i.e. payback on the victim's second transaction. There is no
self-limiting effect.

Two amplifications make this worse than a per-victim grief:

- The **fee collector** is credited by `DeductFees` on every fee-paying
  transaction, and its address is `AddressFromPreimage("fee_collector")` —
  publicly derivable and **keyless**. Polluting it taxes every transaction on
  the chain, and nobody can ever sign for it to clean up.
- There is no maximum value size in the store, so a large enough blob makes the
  victim's own read+write exceed the block gas limit, at which point the account
  is unrecoverable rather than merely expensive.

## Decision

Store realm-issued balances outside the account object, one key per denom:

```
/b/ || addr (20 raw bytes) || denom   ->   amount, 8-byte big-endian
```

**The tier boundary is an explicit allowlist of the chain's gas denoms**, injected
at keeper construction (gno.land passes `[]string{ugnot.Denom}`). Everything else
is split out.

An allowlist rather than a pattern, and closed rather than open, because the
justification for the account tier is narrow: a balance there is free *only*
because the account object is written on every transaction anyway, for the
sequence bump. That argument holds for the denoms used to pay gas and for nothing
else. So a denom reaching the account tier should require an explicit decision,
and anything new should be split by default.

An earlier draft keyed the tier on a leading `/` (`std.IsRealmDenom`), on the
grounds that `/`-prefixed denoms are exactly the permissionlessly-mintable set
today. That was wrong in a way worth recording: **it does not survive IBC.** An
`ibc/<hash>` voucher starts with a letter, so a shape-based rule would file it in
the account object — and IBC denoms are permissionlessly creatable, unbounded in
count, and arrive without the recipient's consent, which is precisely the attack
this ADR exists to close. The same applies to any non-gas genesis denom. The
allowlist is immune because it does not ask what a denom looks like.

`std.IsRealmDenom` still exists, but only to answer a different question — which
denoms a realm may *issue* — and is enforced at the banker. The two must not be
conflated.

Consequence for `GetCoins`: the tiers are no longer separable by first byte (a
split-tier `atom` sorts before an account-tier `ugnot`, a split-tier `zeta` after
it), so the two must be **merged**, not concatenated. The merge is
`Coins.AddUnsafe`, not `Add`. `Add` revalidates every denom in the result and
panics if it dislikes one, and there is nothing for it to find: both inputs
already passed `ValidateDenom` on the way in and the tiers are disjoint, so the
one property iteration could break is ordering, which `splitCoins` asserts
itself. `bank/balances` reaches this unauthenticated, so corrupt state belongs in
the answer where the invariants can report it, not in a panic.

An earlier draft justified this by cost — the revalidation measured 74% of a
`GetCoins` call at 128 denoms. That figure no longer applies: `ValidateDenom` was a
regexp when it was taken, and this same change replaced it with a byte scan
(4,446ns → 174ns on a maximal denom). The choice stands on the redundancy and the
panic, which do not depend on how fast validation is.

`chain/banker` gains `GetCoin(addr, denom) int64`, the Gno-visible form of the
O(1) read. Without it a realm wanting one balance had to call `GetCoins` and pay
for every denom the address holds — the same O(n) cost this change removes from
the money path, left in place for contracts. Note this widens the Gno `Banker`
**interface**, so it breaks any implementor or mock, not just callers. A sweep
found the practical impact small — two implementors in this repo (the stdlib and
one test mock), none at all in gnoswap — because the concrete type is unexported
and the APIs accepting a `Banker` reject non-canonical ones, so a custom implementor
cannot be plugged into either of the interfaces that accept one today.

Balances are per-denom throughout. `AddCoins`/`SubtractCoins` touch one key per
denom moved instead of rewriting the whole set; `HasCoins`, the new `GetCoin`,
and the vesting check read only the denoms involved. `AddCoins`/`SubtractCoins`
now return only `error` — no production caller used the returned `Coins`, and
returning the full set would have reinstated the O(n) read the change exists to
remove.

Layout choices, and why:

- **No length prefix on the address.** cosmos-sdk length-prefixes it because
  `sdk.AccAddress` is 20 or 32 bytes; `crypto.Address` is a fixed `[20]byte`, so
  the split is at a constant offset. If address size ever becomes variable, this
  format needs one.
- **Fixed-width value, not amino.** `amino.Marshal(int64(0))` returns zero bytes
  and the store panics on a nil value, so an accidentally-persisted zero would
  surface as a panic. Fixed width also makes write gas independent of the
  balance's magnitude.
- **Zero balances are deleted**, never stored. Required, not stylistic: a zero
  would make the reconstructed `Coins` invalid.
- **No reverse denom→address index and no supply tracking.** cosmos needs the index
  only for `DenomOwners`, which tm2 does not have, and `TotalCoin` was
  `panic("not yet implemented")` when this was decided. Omitting the index halves the
  write cost of a balance update.

  **Superseded for supply:** `pr6034_coin_supply.md` adds a per-denom counter and
  implements `TotalCoin` on top of it — without a reverse index, which is still
  deliberately absent, because a per-denom total needs one number and not a list of
  holders.
- **Same store as accounts** (`mainKey`). A separately mounted store would add a
  commit root for no benefit, since gno.land pins depth gas so tree size does not
  affect cost.
- **The allowlist is injected, not a param.** The tier decides where bytes
  physically live, so changing it moves state. As a governance param it could be
  flipped without a migration and strand balances in the tier they were written
  to; as a construction argument it is fixed per binary and changing it is a
  coordinated upgrade, like any other consensus change.

The `/` prefix is now also re-asserted **in Go** at `SDKBanker.IssueCoin`/
`RemoveCoin`. It was previously enforced only in interpreted `.gno` stdlib
source; that was acceptable while it was a naming convention, but a bare denom
reaching issuance would now let a realm mint into the genesis tier, including
the native token.

## Changing the allowlist

The allowlist decides where a balance physically lives, so editing it moves state.
It is compiled into the binary on purpose — not a flag, not a config value, not a
governance param — because two nodes with different lists route the same denom to
different keys and produce different app hashes. That is a silent fork, not a
failed startup. Editing it is therefore a coordinated upgrade.

**Adding a non-realm denom — an IBC voucher, or any second gas denom.** Say the
chain decides to accept `ibc/<hash>` for fees, so it should ride along in the
account object like `ugnot`.

1. Add it to `accountTierDenoms` in `gno.land/pkg/gnoland/app.go` and build.
2. Do **not** restart validators on the existing database. Existing holders have
   balances in `/b/` keys, and the new binary looks for them in the account
   object: transfers fail with `InsufficientCoinsError` — not
   `InsufficientFundsError`, which gnokey reports differently — a later credit
   creates a second home for the same denom, and the first `GetCoins` panics on the
   exclusivity assertion. Nothing is lost, but the balance is frozen. All three are
   pinned by `TestAMisMigratedBalanceIsFrozenNotSpendable` and
   `TestSplitKeyForAnAccountTierDenomFailsLoudly`.
3. Regenerate state instead. `gnogenesis fork generate` assembles a new genesis
   from the source chain's state *and* its transaction history — note the
   subcommand; bare `gnogenesis fork` only prints help. Smoke-test it first with
   `gnogenesis fork test`, which runs the replay in memory, so a bad migration
   surfaces before any validator is asked to start on it.
4. Start the new chain from that genesis, with every validator on the new binary.
   The replay happens here, not in step 3: since a balance can only be produced by
   a transaction or a genesis entry, replaying both rewrites every balance under
   the new routing, and both tiers come out consistent with no migration code. The
   app hash changes; that is expected and unavoidable.
5. Tell integrators: `auth/accounts` will now include the new denom in `coins` for
   holders. Anything treating that field as "the balance" sees a different set.
   `bank/balances` is unaffected.

The denom must satisfy `ValidateDenom`, which is lowercase-only — so a cosmos-style
`ibc/` + uppercase-hex hash has to be lowercased on ingress (or the grammar
widened, which is its own consensus decision) before any of this applies.

**Adding a realm-issued denom: don't.** `NewViewKeeper` refuses one outright, and
this is not the same decision as the IBC case even though both are "just another
denom". The account tier is a blob every account-tier denom shares, and the
question that matters is who can put a denom into a stranger's blob:

- `ugnot` requires someone to send it, which costs them the coins.
- an IBC voucher requires a real transfer from another chain.
- a **realm denom** can be minted from nothing, to any address, without consent —
  `IssueCoin` has no recipient check.

The mechanism is the same one described in Context. The account object is read and
rewritten on **every transaction its owner sends**, because the ante handler bumps
the sequence, and store gas is charged per byte of that value. So any denom sitting
in the account tier is paid for on every transaction that address ever makes. A
realm denom is ~40 bytes (`/pkgPath:base`), which at 17 gas/byte read plus 14/byte
write is roughly **1,200 gas per transaction, forever** — imposed on any address
the issuing realm picks, at no cost to the issuer.

The victim is not without recourse, and it is worth being precise about this: they
can clear it by transferring the balance away, since a zero balance drops out of
the `Coins` set. But that costs them a transaction, and the issuer can re-mint for
less than the cleanup cost. So it is a griefing loop rather than a permanent brick
— the same asymmetry noted in Context for the multi-denom case, where with no burn
path a holder can only relocate junk, at cost parity favouring whoever chooses the
timing.

Exact matching holds the damage to a single denom rather than unlimited ones, which
is what makes it a tax rather than a brick. It is still a non-consensual recurring
cost imposed by a third party, which is the thing this ADR exists to remove.

If a realm-issued token genuinely becomes a fee token, leave it in the split tier
and accept one extra key per holder (~306,600 gas on a transfer that touches it).
That cost falls on the party using the token, which is where it belongs.

## Alternatives considered

**Full cosmos ADR-004 — split every denom.** Rejected on measured cost. gno.land
pins store depth gas (`Fixed*Depth100` = 100/200/540), so per-key cost is
independent of tree size: a charged read is `59,000 + 17/byte` and a charged
write `247,600 + 14/byte`. `cacheStore` makes a repeat read free and
refund-dedups repeat writes, so today a transaction pays for each *account key*
exactly once. Splitting the gas denom out buys a second key per address touched.

Measured by simulating the same transaction against a binary whose allowlist is
empty — which is exactly how a chain would move to full ADR-004:

| transaction | this design | full ADR-004 | delta |
|---|---|---|---|
| `MsgSend`, existing recipient | 932,933 | 1,355,208 | **+45%** |
| `MsgSend`, fresh recipient | 1,238,405 | 1,909,148 | +54% |
| fee-only (`maketx call`) | 620,817 | 985,098 | +59% |

The per-key attribution is **not** uniform, and an earlier draft of this ADR got
that wrong by quoting a flat "+306,600 per address touched". Only the *signer*
pays a full extra read+write (+306,197), because only the signer's account object
is written anyway for the sequence bump. For the fee collector and the recipient
the account object is written *only* to carry a balance, so splitting turns their
write into a read and the true cost is +58,000, not +306,600. The aggregate
conclusion holds — the numbers above are measured end to end — but the mechanism
is "the signer's write is already paid for", not "every account write is free".

Whether this gap would erode over time was the one open question, since it is a
function of gas constants rather than of structure. Resolved in favour of keeping
the account tier: store-write gas is not expected to fall, because gas prices
track worst-case validator cost and permanent state only becomes more expensive to
hold as the tree grows. The measured gap is a floor.

**Caveat worth knowing: this is weaker for session-signed transactions.** The
ante handler bumps the sequence on the *session* account (`/a/<master>/s/<session>`),
never the master's, so a session-signed transaction writes the master's account
object only for the balance — the one thing the account tier assumes is free.
Measured, the advantage roughly halves (+115,001 fee-only, +173,133 `MsgSend`,
versus +364,281 / +422,275 unsessioned). As sessions become the default signing
path, the tier earns less. It does not go negative: the fee collector's and
recipient's avoided existence-read remain.

The design adopted here is **never worse than full ADR-004 and identical to
today's cost on every native-only path**, because the genesis balance rides along
in an object that is being written regardless. It also avoids ADR-004's incidental costs. Emptying the allowlist needs no amino
change — `BaseAccount.Coins` would simply always be empty — but it does make
`auth/accounts` report `"coins": ""` for every account, which breaks 20
assertions across 11 txtar files, `gnoclient.QueryAccount`, and every wallet or
explorer reading that endpoint. Vesting constructors also keep working here, though
note they validate `OriginalVesting` against `baseAcc.Coins`, which is now only the
account tier: that is correct at genesis by ordering (`applyBalance` writes the full
pre-split amount before `SetCoins` narrows it) but would silently pass for any future
caller building a vesting account from an already-split account object.

**A per-account denom or byte cap.** Attractive — smallest diff, near-zero gas
cost, and one reviewer recommended it. Rejected because the cap must fire in
`AddCoins`, so a *legitimate* inbound transfer would fail based on the
recipient's state; because with no burn path a victim can only relocate junk, at
cost parity with an attacker who chooses the timing; because without a native-denom
exemption an attacker can fill a never-funded address's slots and permanently
brick it (cannot receive the native token ⇒ cannot pay gas ⇒ cannot clean up);
and because a cap does not fix the underlying layout defect — someone
*legitimately* holding 200 realm coins would still pay ~+1.7M gas per
transaction. Note that adding the native exemption a cap needs is this design's
split in disguise. A cap remains a reasonable addition **on top** of this change
if bounding key count is also wanted.

**Recipient consent for `IssueCoin`.** Does not close the attack: an attacker
mints to their own address (self-consented) and then pushes via `MsgSend`, which
has no recipient consent either and caps neither coin count nor size. Closing it
would require consent on every inbound transfer of every denom, a much larger
semantic change. It also leaves the layout defect for consenting holders.

**Keying the split on "the fee denom" rather than the `/` prefix.** Rejected as
unsound — note this alternative was weighed against the *earlier* `/`-prefix draft,
which the Decision above also rejects, for a different reason (it does not survive
IBC). Neither shape survived; the allowlist replaced both. tm2 has no native-denom concept: `ugnot` is defined in
`gno.land/pkg/gnoland/ugnot`, and the only mentions inside `tm2/` are in tests and
comments, never in logic. The fee denom is whatever `tx.Fee.GasFee.Denom` names, and the one
place it is checked (`EnsureSufficientMempoolFees`) is CheckTx-only and therefore
non-consensus. The `/` prefix, by contrast, is a property of the denom string
alone.

## Consequences

- **Consensus-visible.** The storage layout changes for split-tier balances, so
  the app hash moves for any account holding one. The storage change alone would
  leave the existing app-hash pin untouched, since the pinned fixture holds no
  such balance — but the pin *does* move on this change, because adding
  `banker.GetCoin` adds bytes to `banker.gno`, and stdlib source is itself chain
  state. That is the only reason it moves.
- **Split-tier operations cost one more key** (~+306,600 gas per address) than
  before, since they no longer ride along in the account object. Native-denom
  operations, including the entire fee path, are unchanged. This is the trade:
  the cost lands on the party using a realm denom instead of on every account
  that has ever been sent one.
  Measured end to end, by running the same transaction against both binaries
  (`realm_banker_issued_coin_denom.txtar`, identical tx lines, gas-wanted raised
  only because the true cost now exceeds the old ceiling):

  | transaction | master | this branch | delta |
  |---|---|---|---|
  | send `1330/gno.land/r/test/realm_banker:ugnot` | 1,240,231 | 1,851,884 | **+611,653 (+49%)** |
  | mint a realm denom from a maximum-length package path | 2,648,993 | 3,221,477 | +572,484 (+22%) |

  The transfer touches two addresses, so it pays for two new keys: 2 × 306,600 =
  613,200, within 0.3% of the measured delta. That is the per-key figure above,
  confirmed through the whole stack rather than derived from the gas constants. The
  mint is lower than two keys' worth because the account object it writes also got
  smaller by the length of the denom it no longer carries.
- **Bulk denom spam becomes ~76× dearer per denom** (a fresh key at ~306,600 vs
  ~4,030 amortised inside one blob rewrite) and imposes **zero** ongoing cost on
  the recipient.
- **`GetCoins` got a higher floor**: `splitCoins` opens an iterator
  unconditionally, and an open charges a flat `ReadCostFlat` even over an empty
  range, so `banker.GetCoins` on an account holding only the gas denom went from
  ~60,300 to ~119,300 gas (**+59,000**). Realms wanting one denom should call the
  new `GetCoin`: ~59,100 for a split-tier denom, ~60,300 for a gas denom, which
  reads the account object. All figures here are under gno.land's pinned depth
  config; tm2's default config scales reads by tree depth and gives different
  absolute numbers for the same operations.
- `GetCoins(addr)` remains O(number of realm denoms held). That cost now falls
  only on callers that genuinely want every balance — the `bank/balances` query
  and Gno's `banker.GetCoins` — and never on a transfer. A realm
  calling `banker.GetCoins` on a hostile account can still be pushed to
  out-of-gas.
- Denom **count** remains unbounded, and denom bytes move from a metered value
  into an **unmetered key**.
- **The denom grammar moves in both directions, and both are consensus-visible.**
  It was `[a-z/][a-z0-9_.:/]{2,}` with no length limit; it is now the same class
  plus `-`, bounded at `MaxDenomLength` (274). Rejecting: a denom longer than 274
  bytes no longer validates, which matters because store keys are unmetered, so an
  unbounded denom is unbounded free key. Accepting: `-` was excluded, so a realm at
  a path as ordinary as `gno.land/r/my-org/token` could deploy and then fail to
  issue its own coin. Two nodes disagreeing on either half fork, so both must land
  in the same upgrade — and a chain that has already admitted an over-long denom
  cannot adopt the bound without replay.
- `AddCoins`/`SubtractCoins` signatures changed (dropped the returned `Coins`), which is
  a breaking change to `bank.BankKeeperI`. They were later **removed from
  `vm.BankKeeperI`** entirely in favour of `MintCoins`/`BurnCoins`/`TotalSupply`, so a
  realm cannot reach a supply-blind credit — see `pr6034_coin_supply.md`.
- Receiving coins still creates the recipient's account. This is deliberate and
  load-bearing: an address with no account cannot sign, so the funds would
  otherwise be visible and permanently unspendable. Account creation allocates
  from a global counter, so *when* it happens is consensus state.
- Moving to full ADR-004 later means emptying the allowlist. No in-place
  migration code is needed *if the move is done by replay*: `gnogenesis fork
  generate` assembles a genesis from the source chain's state and transaction
  history, and starting the new binary on it regenerates every balance in whatever
  layout that binary implements. It is not
  free, though, and calling it "a one-line change" would be wrong: a plain binary
  upgrade on an existing database is **not** a replay and would strand every gas
  balance in the account object — the `accountTierCoins` assertion exists to halt
  such a node rather than let it under-report. The `auth/accounts` wire-format
  break above applies either way.

## Tests

- `tm2/pkg/sdk/bank/balance_test.go` — pins the key layout as a format, the
  address/denom split, fixed-width encoding and its corruption panics,
  zero-balance deletion, iteration ordering, the merge property `GetCoins`
  depends on, tier routing (including that a realm balance is *not* in
  the account object), replace-not-merge semantics for `SetCoins`, account
  creation on receive for both tiers, and that the account and balance keyspaces
  cannot reach each other in either direction.
- Each was mutation-verified: swapping the address/denom order in the key,
  dropping the zero-balance delete, reversing the merge order, removing
  account creation on receive, and breaking tier routing all fail.
- `tm2/pkg/std/coin_test.go` — `TestIsRealmDenom` against the leading-class
  property it derives from, including interior-slash denoms that a
  substring test would mis-tier, plus a sweep over every byte confirming the
  predicate keys on the first byte alone.
- Tier exclusivity is asserted in **both** directions, and each assertion has its
  own test that fails if the assertion is deleted: a split key for an allowlisted
  denom (`splitCoins`, covering the allowlist growing) and an account-object
  balance for a non-allowlisted one (`accountTierCoins`, covering it shrinking,
  which is the documented path to a fully split layout).
- `TestConservation` runs 500 deterministic operations against a map oracle,
  comparing after every one including failures, over `SendCoins`, `MintCoins`,
  `BurnCoins`, `SetCoins` and `InputOutputCoins` (the credit and debit arms became
  mint and burn when the supply counter landed), plus a directed vesting probe. Arms
  skip when their randomly drawn amount is unusable, so somewhat fewer than 500
  operations execute.
- `TestGetCoinCostIsFlat` pins the property the split exists to buy: reading
  one denom costs the same whether the address holds 1 or 200 others. It is
  written comparatively, holding key count equal across both arms, because tm2's
  *default* gas config scales reads by tree depth even though gno.land pins it
  flat.
- `gno.land/pkg/integration/testdata/realm_banker_issued_coin_denom.txtar` —
  end-to-end issuance over the split storage path, and a `vm/qeval` read through
  the new `banker.GetCoin`.

## AI assistance

Implemented with AI assistance. The plan and the storage-design choice were each
taken through three independent review rounds — the design decision was escalated
to the repo owner with measured gas numbers rather than decided by the agent, and the
dissenting recommendation is recorded above. The human author reviewed and owns the
change.
