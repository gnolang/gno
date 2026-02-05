# Anti-Squatting System Design Document

**Package:** `gno.land/p/gnoland/antisquat`
**Issue:** [#2727](https://github.com/gnolang/gno/issues/2727)
**Author:** @politas180

---

## 1. Problem Statement

gno.land's current namespace system (`r/gnoland/users/v1`) uses a flat 1 GNOT fee and basic regex validation (`^[a-z]{3}[_a-z0-9]{0,14}[0-9]{3}`). This is insufficient to prevent:

- **Name squatting:** Registering valuable names with intent to resell
- **Bulk registration:** Automated scripts claiming namespaces at minimal cost
- **Trademark conflicts:** Impersonation of projects or individuals

This document proposes `p/gnoland/antisquat`, a pluggable anti-squatting library that any realm can import.

## 2. Architecture Overview

### 2.1 Separation of Concerns

```
p/gnoland/antisquat (stateless library, instantiated per-realm)
    |
    +-- Rate limiting (per-address, windowed)
    +-- Protected names registry (dynamic add/remove)
    +-- Length-based pricing calculator
    +-- Name lifecycle management (register, expire, renew)

r/gnoland/users/v2 (stateful realm, imports antisquat)
    |
    +-- GovDAO authorization gating
    +-- Coin collection and fee forwarding
    +-- Cross-realm calls to r/sys/users
    +-- Event emission for indexers
```

The package handles **state and logic**. The realm handles **authorization and coins**.

### 2.2 Controller Pattern Integration

`r/sys/users` is the storage layer with a controller whitelist managed via GovDAO:

```
r/sys/users (storage)
    +-- controllers whitelist (addrset.Set)
    |     +-- r/gnoland/users/v1 (current: pattern + flat fee)
    |     +-- r/gnoland/users/v2 (proposed: antisquat integration)
    +-- nameStore (avl.Tree: name -> *UserData)
    +-- addressStore (avl.Tree: address -> *UserData)
```

The v2 realm calls `susers.RegisterUser(cross, name, caller)` exactly as v1 does.

### 2.3 Storage Design

The `AntiSquat` struct uses four AVL trees for O(log n) operations:

| Tree | Key | Value | Purpose |
|------|-----|-------|---------|
| `names` | name (string) | `*NameRecord` | Name lookup |
| `addresses` | addr.String() | `*avl.Tree` (name -> bool) | Per-address name listing |
| `rateLimit` | addr.String() | `*rateLimitEntry` | Rate limit tracking |
| `reserved` | name (string) | `ReservedCategory` | Protected name registry |

**No global iteration is ever required for core operations.**

## 3. Review Feedback Addressed

This design proactively addresses all six concerns raised by @notJoon on PR #5074:

### 3.1 Pricing (1 GNOT = 1,000,000 ugnot)

All prices use correct ugnot conversion:

| Name Length | Annual Fee (GNOT) | Annual Fee (ugnot) |
|-------------|-------------------|---------------------|
| 1-3 chars | 1,000 | 1,000,000,000 |
| 4 chars | 100 | 100,000,000 |
| 5 chars | 50 | 50,000,000 |
| 6-10 chars | 10 | 10,000,000 |
| 11+ chars | 5 | 5,000,000 |

### 3.2 Performance (O(log n), No Global Iteration)

- `CanRegister`: AVL tree lookup O(log n) for name, reserved, rate limit checks
- `isRateLimited`: O(log n) AVL lookup + O(k) backward scan where k <= MaxRegistrationsPerWindow (default 3)
- `GetNamesForAddress`: O(log n) outer tree lookup + O(m) inner tree iteration where m = names owned by address
- No operation iterates over ALL registered names or ALL addresses

### 3.3 Pruning (Per-Address Only)

`recordRegistration` only prunes old block heights for the specific address being recorded:

```gno
func (as *AntiSquat) recordRegistration(addr std.Address, height int64) {
    // O(log n) AVL lookup
    entry := getRateLimitEntry(addr)
    
    // Prune only THIS address's old entries
    windowStart := height - as.config.RateLimitWindowBlocks
    pruneIdx := 0
    for pruneIdx < len(entry.Heights) {
        if entry.Heights[pruneIdx] >= windowStart {
            break
        }
        pruneIdx++
    }
    if pruneIdx > 0 {
        entry.Heights = entry.Heights[pruneIdx:]
    }
    
    entry.Heights = append(entry.Heights, height)
}
```

Inactive accounts with old records cause zero overhead.

### 3.4 Dynamic Protected Names

Protected names are managed via `AddReservedName` and `RemoveReservedName`:

```gno
// In realm's GovDAO proposal handler:
func ProposeProtectName(name string, cat antisquat.ReservedCategory) dao.ProposalRequest {
    cb := func(cur realm) error {
        as.AddReservedName(name, cat)
        return nil
    }
    return dao.NewProposalRequest(
        "Protect namespace: "+name,
        "Add to protected names registry",
        dao.NewSimpleExecutor(cb, ""),
    )
}
```

Default names are seeded in `init()` and can also be removed via GovDAO.

### 3.5 Auction Penalties for Non-Revealers

For Phase 2 Vickrey auctions, non-revealers forfeit a configurable percentage (default 50%) of their deposit:

- **0% burn:** No deterrence. Attackers submit fake commitments for free.
- **100% burn:** Too harsh. Honest users who lose their salt lose everything.
- **50% burn:** Meaningful penalty that deters griefing while allowing partial recovery.

The exact percentage is configurable via GovDAO.

### 3.6 Fee Handling

Fee flow is explicit:

1. `antisquat.CalculateFee(name)` returns the required fee in ugnot
2. The realm checks `std.GetOrigSend()` against the required fee
3. The realm forwards fees to a configurable `feeCollector` address (defaulting to GovDAO treasury)
4. Overpayment is refunded to the caller

The package never touches coins directly.

## 4. Three-Tier Namespace System

### Tier 1: Provisional (Phase 1)
- **Format:** `{name}-{hash6}` (e.g., `alice-a1b2c3`)
- **Cost:** Gas only
- **Limit:** 3 per address per 30-day window
- **Purpose:** Permissionless entry for new users

### Tier 2: Verified (Phase 1 - Current Implementation)
- **Format:** `{name}` (e.g., `alice`)
- **Cost:** Length-based fees (see pricing table above)
- **Verification:** Economic cost serves as verification
- **Renewal:** Annual, with 30-day grace period

### Tier 3: Premium (Phase 2)
- **Format:** Short names (1-3 chars)
- **Cost:** Vickrey sealed-bid auction
- **Purpose:** Fair allocation of scarce, high-value names

## 5. Phase 2: Auction Mechanisms

### 5.1 Vickrey Sealed-Bid Auction

For premium names and contested verified names:

```
COMMIT phase (~7 days): Submit hash(name|bid|salt|address) with deposit
REVEAL phase (~3 days): Reveal actual bids, contract verifies hash
RESOLUTION: Highest bidder wins, pays SECOND-highest price
            Losers get full deposit refund
            Non-revealers forfeit 50% of deposit
```

### 5.2 Dutch Auction for Expired Names

When a name passes its grace period:

```
Starting price: 10x last renewal fee
Decay: Linear decrease over 30 days
Floor price: Standard tier pricing
First to claim at current price wins
```

## 6. Phase 2: External Identity Integration

Per issue #4937, external identity verification provides natural anti-squatting:

- **Format:** `{username}@{provider}` (e.g., `kouteki@github`)
- **Inherently anti-squat:** Tied to verified external account
- **Pricing:** Rate limiting applies, but length-based pricing may be waived
- **Verification:** Realm handles OAuth proof (requires oracle integration)

## 7. Security Analysis

| Attack | Mitigation |
|--------|-----------|
| Front-running | Provisional: hash suffix. Verified: economic cost. Premium: commit-reveal. |
| Sybil (bulk wallets) | Economic cost per name + annual renewal. 100 six-char names = 1000 GNOT/year. |
| Rate limit bypass | Rate limiting is defense-in-depth; economic cost is primary deterrent. |
| Protected name enumeration | Protected list is public by design. Economic cost guards unprotected names. |
| Auction shill bidding | Each bid requires deposit. Shill bids cost real money. |

## 8. Implementation Status

### Phase 1 (This PR)

| Component | Status |
|-----------|--------|
| `p/gnoland/antisquat` types and errors | Done |
| Name validation and normalization | Done |
| Rate limiting with per-address pruning | Done |
| Protected names registry (dynamic) | Done |
| Length-based pricing calculator | Done |
| Name lifecycle (register, expire, renew) | Done |
| Comprehensive unit tests | Done |

### Phase 2 (After Phase 1 Feedback)

| Component | Status |
|-----------|--------|
| Vickrey sealed-bid auction | Designed, ready to implement |
| Dutch auction for expired names | Designed |
| External identity integration (#4937) | Designed |
| `r/gnoland/users/v2` realm | Ready to implement after Phase 1 review |

## 9. Deployment Plan

1. Deploy `p/gnoland/antisquat` via `gnokey maketx addpkg`
2. Deploy `r/gnoland/users/v2` that imports the package
3. GovDAO proposal to whitelist v2 in `r/sys/users`
4. Test period: both v1 and v2 active simultaneously
5. GovDAO proposal to de-whitelist v1 (optional, based on v2 stability)

## 10. References

- [Issue #2727: Anti-Squatting System](https://github.com/gnolang/gno/issues/2727)
- [Issue #2827: r/sys/users v2](https://github.com/gnolang/gno/issues/2827) (closed)
- [Issue #3020: Moderation DAOs](https://github.com/gnolang/gno/issues/3020) (closed as stale)
- [Issue #4937: External Identity](https://github.com/gnolang/gno/issues/4937)
- [PR #5074: Competing design document](https://github.com/gnolang/gno/pull/5074) and @notJoon's review
- [ENS Name Registration](https://docs.ens.domains/registry/eth)
- [Vickrey Auction](https://en.wikipedia.org/wiki/Vickrey_auction)
