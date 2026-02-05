# Anti-Squatting System Design Document

**Issue:** https://github.com/gnolang/gno/issues/2727
**Author:** @bimakw
**Status:** Draft
**Created:** 2026-01-19

---

## 1. Executive Summary

This document proposes `p/gnoland/antisquat`, a pluggable anti-squatting library for gno.land namespaces. The system prevents username squatting through three mechanisms:

1. **Rate Limiting** - Max 3 registrations per address per 30 days
2. **Protected Names** - DAO-managed list of reserved names
3. **Length-Based Pricing** - Shorter names cost more (ENS-style)

Phase 2 adds **Vickrey Sealed-Bid Auctions** for premium/contested names.

---

## 2. Problem Statement

Current anti-squatting in `r/gnoland/users/v1` is weak:

| Mechanism | Current | Effectiveness |
|-----------|---------|---------------|
| Registration Fee | 1 GNOT | LOW - too cheap |
| Username Pattern | Must end with 3 digits | MEDIUM |
| Pre-registered Names | 10 hardcoded | LOW |
| Per-address Limit | 1 name only | Bypassable with multiple wallets |

**Gaps:**
- No rate limiting window
- No DAO-managed protected names
- No premium pricing for short names
- No auction for contested names

---

## 3. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         ARCHITECTURE                                     │
└─────────────────────────────────────────────────────────────────────────┘

                          ┌──────────────────────┐
                          │   User/Frontend      │
                          └──────────┬───────────┘
                                     │
                                     ▼
                    ┌────────────────────────────────┐
                    │  r/gnoland/users/v2            │
                    │  (Controller Realm)            │
                    │                                │
                    │  - Register()                  │
                    │  - Validates payment           │
                    │  - Calls antisquat.Validate()  │
                    └───────────────┬────────────────┘
                                    │
                 ┌──────────────────┼──────────────────┐
                 │                  │                  │
                 ▼                  ▼                  ▼
    ┌────────────────┐  ┌────────────────┐  ┌────────────────┐
    │ p/gnoland/     │  │ p/gnoland/     │  │ p/gnoland/     │
    │ antisquat/     │  │ antisquat/     │  │ antisquat/     │
    │ ratelimit      │  │ protected      │  │ pricing        │
    │                │  │                │  │                │
    │ - CanRegister  │  │ - IsProtected  │  │ - CalcFee      │
    │ - Record       │  │ - Add/Remove   │  │ - GetTier      │
    └────────────────┘  └────────────────┘  └────────────────┘
                                    │
                                    ▼
                          ┌──────────────────┐
                          │   r/sys/users    │
                          │   (Storage)      │
                          └──────────────────┘
```

---

## 4. Design Rationale

This section explains WHY this architecture was chosen over alternatives.

### 4.1 Why a Multi-Layered Approach?

A single anti-squatting mechanism is insufficient. Each layer addresses a different attack vector:

| Layer | Addresses | Limitation if Used Alone |
|-------|-----------|-------------------------|
| Rate Limiting | Bulk registration from single actor | Bypassable via multiple wallets |
| Length Pricing | Economic barrier for valuable names | Doesn't prevent wealthy attackers |
| Protected Names | Safeguards critical system names | Doesn't scale to all valuable names |
| Vickrey Auction | Fair price discovery for contested names | Overkill for regular registrations |

**Combined Defense:** Rate limiting + pricing creates economic friction. Protected names guard critical infrastructure. Auctions handle high-value contested cases.

### 4.2 Why Length-Based Pricing (ENS Model)?

**Alternatives considered:**

| Approach | Pros | Cons |
|----------|------|------|
| Flat fee (current) | Simple | Underprices valuable short names |
| Popularity-based | Reflects demand | Complex oracle needed |
| **Length-based** | Simple, predictable, proven | May not match true market value |
| Auction-only | True price discovery | Too slow for regular registration |

**Decision:** Length-based pricing was chosen because:
1. **Proven model** - ENS has validated this approach at scale
2. **Predictable UX** - Users know fee before attempting registration
3. **No oracle dependency** - Fully on-chain deterministic
4. **Low implementation complexity** - Simple lookup table
5. **Reasonable correlation** - Short names are generally more valuable

### 4.3 Why Vickrey Sealed-Bid Auction?

**Alternatives considered:**

| Auction Type | Pros | Cons |
|--------------|------|------|
| English (ascending) | Familiar | Front-running vulnerable |
| Dutch (descending) | Fast resolution | Overpayment common |
| First-price sealed | No front-running | Winner's curse |
| **Vickrey (2nd-price sealed)** | Truthful bidding incentive | Two-phase complexity |

**Decision:** Vickrey auction was chosen because:
1. **Truthful bidding** - Bidders incentivized to bid true value
2. **Front-running resistant** - Commit-reveal prevents MEV exploitation
3. **Fair pricing** - Winner pays second-highest, not their max
4. **Blockchain proven** - Used by ENS for premium auctions

### 4.4 Why Pluggable Library Design?

**Alternatives considered:**

| Approach | Pros | Cons |
|----------|------|------|
| Hardcoded in r/gnoland/users | Simple | Not reusable |
| **Importable package** | Reusable, testable | Slight indirection |
| Separate realm | Isolated state | Cross-realm complexity |

**Decision:** Pluggable package (`p/gnoland/antisquat`) because:
1. **Reusability** - Other realms can import (packages, organizations)
2. **Testability** - Isolated unit tests without realm state
3. **Maintainer guidance** - @moul explicitly requested "generic, importable library"
4. **Separation of concerns** - Anti-squatting logic separate from user management

### 4.5 Why GovDAO Integration for Protected Names?

Direct admin control creates trust issues. GovDAO integration ensures:
1. **Transparency** - All additions/removals are public proposals
2. **Community consensus** - Changes require voting approval
3. **Accountability** - Proposal history creates audit trail
4. **Decentralization** - No single point of control

### 4.6 Comparison with Other Proposals

| Proposal | Our Adoption | Reasoning |
|----------|--------------|-----------|
| @thinhnx-var: Sealed auction | ✅ Phase 2 | Best for contested names |
| @miguelito766: DNS verification | ❌ Not included | External dependency, complexity |
| @AnhVAR: Dual-tier (verified/unverified) | ⚠️ Partial (pricing tiers) | Simplified to pricing tiers |
| @thinhnx-var: DID/KYC integration | ❌ Not in v1 | Future enhancement |

**We adopted** the core sealed auction mechanism but simplified the tier system to length-based pricing for Phase 1.

---

## 5. Package Structure

```
examples/gno.land/
├── p/gnoland/antisquat/           # Core Library (pluggable)
│   ├── gno.mod
│   ├── antisquat.gno              # Main facade
│   ├── types.gno                  # Interfaces & types
│   ├── errors.gno                 # Error definitions
│   ├── ratelimit.gno              # Rate limiting implementation
│   ├── protected.gno              # Protected names registry
│   ├── pricing.gno                # Length-based pricing
│   └── antisquat_test.gno         # Tests
│
├── p/gnoland/auction/vickrey/     # Auction Library (Phase 2)
│   ├── gno.mod
│   ├── auction.gno                # Vickrey auction logic
│   ├── types.gno                  # Auction types
│   └── auction_test.gno           # Tests
│
└── r/gnoland/users/v2/            # Updated Controller (optional)
    ├── gno.mod
    ├── users.gno                  # Registration with antisquat
    └── admin.gno                  # GovDAO integration
```

---

## 6. Core Types & Interfaces

### 6.1 Main Interface

```gno
package antisquat

// AntiSquat is the main facade
type AntiSquat interface {
    // Validate checks all anti-squatting rules
    Validate(req RegistrationRequest) ValidationResult

    // Record logs a successful registration
    RecordRegistration(req RegistrationRequest)

    // Component access
    RateLimiter() RateLimiter
    Protected() ProtectedRegistry
    Pricing() PricingCalculator
}

type RegistrationRequest struct {
    Name      string
    Caller    address
    Payment   int64      // ugnot sent
    Timestamp time.Time
}

type ValidationResult struct {
    Allowed     bool
    Reason      string
    RequiredFee int64
}
```

### 6.2 Rate Limiter

```gno
type RateLimiter interface {
    CanRegister(addr address, now time.Time) bool
    RecordRegistration(addr address, name string, now time.Time)
    GetRemaining(addr address, now time.Time) int
    SetConfig(config RateLimitConfig)
}

type RateLimitConfig struct {
    MaxRegistrations int           // default: 3
    WindowDuration   time.Duration // default: 30 days
    Enabled          bool
}
```

### 6.3 Protected Names

```gno
type ProtectedRegistry interface {
    IsProtected(name string) bool
    GetProtection(name string) (ProtectedName, bool)
    AddProtected(name string, category Category, reason string) error
    RemoveProtected(name string) error
    ListByCategory(category Category) []ProtectedName
}

type Category int
const (
    CategorySystem Category = iota + 1  // gno, admin, root
    CategoryBrand                        // google, apple
    CategoryGovernance                   // dao, voting
    CategoryInfrastructure              // validator, rpc
)

type ProtectedName struct {
    Name       string
    Category   Category
    Reason     string
    AddedBy    address
    AddedAt    time.Time
    ProposalID string  // GovDAO proposal ID
}
```

### 6.4 Pricing Calculator

```gno
type PricingCalculator interface {
    CalculateFee(name string) int64
    GetTier(length int) PriceTier
    SetTiers(tiers []PriceTier)
}

type PriceTier struct {
    MinLength  int
    MaxLength  int   // -1 for unlimited
    PriceUgnot int64
}

// Default tiers
var DefaultPriceTiers = []PriceTier{
    {1, 2, 100_000_000_000},   // 100 GNOT
    {3, 3, 50_000_000_000},    // 50 GNOT
    {4, 4, 10_000_000_000},    // 10 GNOT
    {5, 5, 5_000_000_000},     // 5 GNOT
    {6, 7, 2_000_000_000},     // 2 GNOT
    {8, -1, 1_000_000_000},    // 1 GNOT (default)
}
```

---

## 7. Implementation Details

### 7.1 Rate Limiting

**Storage:** AVL tree keyed by `address.String()`

```gno
var histories = avl.NewTree()  // addr -> *registrationHistory

type registrationHistory struct {
    records []RegistrationRecord
}

type RegistrationRecord struct {
    Name      string
    Timestamp time.Time
}

func (rl *rateLimiter) CanRegister(addr address, now time.Time) bool {
    history := rl.getHistory(addr)
    if history == nil {
        return true
    }

    windowStart := now.Add(-rl.config.WindowDuration)
    count := 0
    for _, rec := range history.records {
        if rec.Timestamp.After(windowStart) {
            count++
        }
    }
    return count < rl.config.MaxRegistrations
}

func (rl *rateLimiter) RecordRegistration(addr address, name string, now time.Time) {
    history := rl.getOrCreateHistory(addr)
    history.records = append(history.records, RegistrationRecord{
        Name:      name,
        Timestamp: now,
    })
    // Prune expired records
    rl.pruneOldRecords(history, now)
}
```

### 7.2 Protected Names

**Storage:** Two AVL trees - main + category index

```gno
var (
    names      = avl.NewTree()  // name -> *ProtectedName
    categories = avl.NewTree()  // "cat:system" -> *avl.Tree (name -> bool)
)

// Default protected names initialized in init()
func init() {
    systemNames := []string{
        "gno", "gnoland", "gnolang", "admin", "root", "system",
        "official", "support", "api", "www",
    }
    for _, name := range systemNames {
        addProtectedInternal(name, CategorySystem, "system reserved")
    }

    govNames := []string{"dao", "govdao", "governance", "voting", "treasury"}
    for _, name := range govNames {
        addProtectedInternal(name, CategoryGovernance, "governance reserved")
    }
}

func (r *registry) IsProtected(name string) bool {
    normalized := strings.ToLower(name)
    _, exists := names.Get(normalized)
    return exists
}
```

### 7.3 Pricing

```gno
func (c *calculator) CalculateFee(name string) int64 {
    length := len(name)
    for _, tier := range c.tiers {
        if length >= tier.MinLength {
            if tier.MaxLength == -1 || length <= tier.MaxLength {
                return tier.PriceUgnot
            }
        }
    }
    return 1_000_000_000 // fallback: 1 GNOT
}
```

**Pricing Table:**

| Length | Price (GNOT) | Example |
|--------|--------------|---------|
| 1-2    | 100          | "x", "ab" |
| 3      | 50           | "bob" |
| 4      | 10           | "john" |
| 5      | 5            | "alice" |
| 6-7    | 2            | "charlie" |
| 8+     | 1            | "username123" |

---

## 8. Vickrey Auction (Phase 2)

For premium names or contested registrations.

### 8.1 Flow

```
Phase 1: COMMIT (7 days)
─────────────────────────
Bidder computes: hash = sha256(amount|salt|address)
Bidder calls: CommitBid(username, hash)
  - Sends deposit >= bid amount
  - Hash stored, amount hidden

Phase 2: REVEAL (3 days)
─────────────────────────
Bidder calls: RevealBid(username, amount, salt)
  - Contract verifies hash
  - Bid recorded if valid

Phase 3: RESOLUTION
─────────────────────────
  - Highest bidder wins
  - Winner pays SECOND-highest price
  - Losers get full refund
```

### 8.2 Types

```gno
type AuctionPhase uint8
const (
    PhaseCommit AuctionPhase = iota
    PhaseReveal
    PhaseFinished
    PhaseCanceled
)

type Commitment struct {
    Bidder   address
    Hash     string   // sha256(amount|salt|address)
    Deposit  int64
    Revealed bool
}

type AuctionConfig struct {
    CommitDuration time.Duration  // 7 days
    RevealDuration time.Duration  // 3 days
    MinimumBid     int64          // floor price
}
```

### 8.3 Security

- **Front-running prevention:** Bids hidden until reveal
- **Commitment binding:** Hash includes bidder address
- **Deposit requirement:** Must lock funds >= bid
- **Griefing mitigation:** Non-revealers just lose opportunity

---

## 9. GovDAO Integration

### 9.1 Protected Names Proposals

```gno
import "gno.land/r/gov/dao"

func ProposeAddProtectedName(name string, category Category, reason string) dao.ProposalRequest {
    cb := func(cur realm) error {
        return antisquat.Protected().AddProtected(name, category, reason)
    }

    return dao.NewProposalRequest(
        ufmt.Sprintf("Protect name: %s", name),
        ufmt.Sprintf("Add '%s' to protected names (%s): %s", name, category, reason),
        dao.NewSimpleExecutor(cb, ""),
    )
}

func ProposeRemoveProtectedName(name string, reason string) dao.ProposalRequest {
    cb := func(cur realm) error {
        return antisquat.Protected().RemoveProtected(name)
    }

    return dao.NewProposalRequest(
        ufmt.Sprintf("Unprotect name: %s", name),
        reason,
        dao.NewSimpleExecutor(cb, ""),
    )
}
```

### 9.2 Config Updates

```gno
func ProposeUpdateRateLimits(maxRegs int, windowDays int) dao.ProposalRequest {
    cb := func(cur realm) error {
        config := RateLimitConfig{
            MaxRegistrations: maxRegs,
            WindowDuration:   time.Duration(windowDays) * 24 * time.Hour,
            Enabled:          true,
        }
        antisquat.RateLimiter().SetConfig(config)
        return nil
    }

    return dao.NewProposalRequest(
        "Update rate limits",
        ufmt.Sprintf("Set max %d registrations per %d days", maxRegs, windowDays),
        dao.NewSimpleExecutor(cb, ""),
    )
}
```

---

## 10. Integration with Existing System

### 10.1 Controller Update

```gno
// r/gnoland/users/v2/users.gno
package users

import (
    "time"
    "chain/banker"
    "chain/runtime"

    "gno.land/p/gnoland/antisquat"
    susers "gno.land/r/sys/users"
)

var as antisquat.AntiSquat

func init() {
    as = antisquat.New()
}

func Register(_ realm, username string) {
    // 1. Check caller is EOA
    if !runtime.PreviousRealm().IsUser() {
        panic("only EOA can register")
    }

    // 2. Build request
    req := antisquat.RegistrationRequest{
        Name:      username,
        Caller:    runtime.PreviousRealm().Address(),
        Payment:   banker.OriginSend().AmountOf("ugnot"),
        Timestamp: time.Now(),
    }

    // 3. Validate against antisquat rules
    result := as.Validate(req)
    if !result.Allowed {
        panic(result.Reason)
    }

    // 4. Register in sys/users
    if err := susers.RegisterUser(cross, username, req.Caller); err != nil {
        panic(err)
    }

    // 5. Record for rate limiting
    as.RecordRegistration(req)

    // 6. Emit event
    chain.Emit("Registration",
        "address", req.Caller.String(),
        "name", username,
        "fee", result.RequiredFee,
    )
}
```

---

## 11. Validation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         VALIDATION FLOW                                  │
└─────────────────────────────────────────────────────────────────────────┘

                    ┌─────────────────┐
                    │ Validate(req)   │
                    └────────┬────────┘
                             │
                             ▼
               ┌─────────────────────────┐
               │ 1. Is name protected?   │
               └────────────┬────────────┘
                            │
              ┌─────────────┴─────────────┐
              │ YES                       │ NO
              ▼                           ▼
    ┌───────────────────┐    ┌─────────────────────────┐
    │ REJECT            │    │ 2. Rate limit check     │
    │ "name protected:  │    └────────────┬────────────┘
    │  {reason}"        │                 │
    └───────────────────┘   ┌─────────────┴─────────────┐
                            │ EXCEEDED                  │ OK
                            ▼                           ▼
              ┌───────────────────┐    ┌─────────────────────────┐
              │ REJECT            │    │ 3. Calculate fee        │
              │ "rate limit       │    │    (based on length)    │
              │  exceeded"        │    └────────────┬────────────┘
              └───────────────────┘                 │
                                                   ▼
                                    ┌─────────────────────────┐
                                    │ 4. Payment >= required? │
                                    └────────────┬────────────┘
                                                 │
                               ┌─────────────────┴─────────────────┐
                               │ NO                                │ YES
                               ▼                                   ▼
                 ┌───────────────────┐             ┌───────────────────┐
                 │ REJECT            │             │ ALLOW             │
                 │ "insufficient     │             │ Registration      │
                 │  fee: need X"     │             │ proceeds          │
                 └───────────────────┘             └───────────────────┘
```

---

## 12. Testing Strategy

### 12.1 Unit Tests

```gno
func TestRateLimiting(t *testing.T) {
    rl := ratelimit.New()
    addr := testutils.TestAddress("user1")
    now := time.Now()

    // First 3 should succeed
    for i := 0; i < 3; i++ {
        uassert.True(t, rl.CanRegister(addr, now))
        rl.RecordRegistration(addr, ufmt.Sprintf("name%d", i), now)
    }

    // 4th should fail
    uassert.False(t, rl.CanRegister(addr, now))

    // After window passes, should work again
    future := now.Add(31 * 24 * time.Hour)
    uassert.True(t, rl.CanRegister(addr, future))
}

func TestProtectedNames(t *testing.T) {
    pr := protected.New()

    uassert.True(t, pr.IsProtected("gno"))
    uassert.True(t, pr.IsProtected("GNO"))  // case insensitive
    uassert.False(t, pr.IsProtected("myname"))
}

func TestPricing(t *testing.T) {
    pc := pricing.New()

    uassert.Equal(t, int64(100_000_000_000), pc.CalculateFee("ab"))    // 2 chars
    uassert.Equal(t, int64(50_000_000_000), pc.CalculateFee("bob"))    // 3 chars
    uassert.Equal(t, int64(1_000_000_000), pc.CalculateFee("longname")) // 8 chars
}
```

### 12.2 Integration Tests (filetest)

```gno
// PKGPATH: gno.land/r/test/antisquat
// SEND: 5000000000ugnot

package main

import "gno.land/r/gnoland/users/v2"

func main() {
    // Should succeed with 5 GNOT for 5-char name
    users.Register(cross, "alice")
    println("registered: alice")
}

// Output:
// registered: alice
```

---

## 13. Migration Plan

1. **Deploy `p/gnoland/antisquat`** - No state migration needed
2. **Deploy `r/gnoland/users/v2`** (or update v1)
3. **GovDAO proposal** to whitelist v2 controller
4. **Existing users** remain in `r/sys/users` (no change)
5. **New registrations** go through antisquat validation

---

## 14. Security Considerations

| Risk | Mitigation |
|------|------------|
| Sybil attack (many wallets) | Rate limiting + pricing makes it expensive |
| Front-running auctions | Commit-reveal scheme |
| Griefing protected list | GovDAO approval required |
| Admin key compromise | Multi-sig / GovDAO control |
| Storage spam | Prune expired rate limit records |

---

## 15. Operational Considerations

This section addresses deployment, monitoring, failure scenarios, and ongoing operations.

### 15.1 Deployment Strategy

**Phased Rollout:**

| Phase | Action | Risk Level |
|-------|--------|------------|
| 1. Testnet | Deploy to test4.gno.land | LOW |
| 2. Audit | Code review + security audit | - |
| 3. Soft Launch | Deploy to mainnet, disabled by default | LOW |
| 4. Limited Activation | Enable for new registrations only | MEDIUM |
| 5. Full Activation | Enable all features including auctions | HIGH |

**Deployment Steps:**
1. **Package deployment** - `gnokey maketx addpkg` for `p/gnoland/antisquat`
2. **Realm deployment** - Deploy `r/gnoland/users/v2` (if new controller)
3. **Whitelist update** - GovDAO proposal to whitelist new controller in `r/sys/users`
4. **Configuration** - Set initial rate limits, protected names, pricing tiers
5. **Monitoring setup** - Configure event indexing and alerting

**Rollback Plan:**
- If critical issues found: GovDAO proposal to remove v2 from whitelist
- Existing users unaffected (stored in r/sys/users)
- v1 controller remains functional as fallback

### 15.2 Monitoring & Observability

**On-chain Events Emitted:**

```gno
chain.Emit("Registration",
    "address", addr,
    "name", name,
    "fee", fee,
    "tier", tier,
)

chain.Emit("RateLimitHit",
    "address", addr,
    "remaining", remaining,
    "window_resets", windowResets,
)

chain.Emit("ProtectedNameAttempt",
    "address", addr,
    "name", name,
    "category", category,
)

chain.Emit("AuctionCreated",
    "name", name,
    "commit_end", commitEnd,
    "reveal_end", revealEnd,
)
```

**Metrics to Track:**

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| registrations_per_hour | New registrations | > 100/hr (possible attack) |
| rate_limit_hits | Blocked attempts | > 50/hr |
| protected_name_attempts | Attempts on reserved names | > 20/hr |
| avg_registration_fee | Average fee paid | < 1 GNOT (possible bug) |
| unique_addresses_per_day | New registrants | Trend analysis |
| auction_participation | Bids per auction | < 2 (low interest indicator) |

**Indexer Integration:**
- Events indexed by tx-indexer for dashboard queries
- Grafana/equivalent dashboards for ops team
- PagerDuty/equivalent alerts for anomalies

### 15.3 Failure Scenarios & Recovery

**Scenario 1: Rate Limiter State Corruption**
- *Detection:* Users unable to register despite being within limits
- *Recovery:* GovDAO proposal to reset rate limiter state for affected addresses
- *Prevention:* Comprehensive test coverage, state validation on reads

**Scenario 2: Protected Names List Manipulation**
- *Detection:* Unauthorized names added/removed without proposal
- *Recovery:* GovDAO proposal to restore correct list
- *Prevention:* All modifications require GovDAO approval

**Scenario 3: Auction Stuck in Commit Phase**
- *Detection:* Auction past commit deadline with no transitions
- *Recovery:* Admin function to force phase transition
- *Prevention:* Block-based deadlines (not wall-clock time)

**Scenario 4: Price Oracle Manipulation (Future)**
- *Detection:* Unusual pricing patterns
- *Recovery:* Emergency price freeze via GovDAO
- *Prevention:* Use block-based pricing, no external oracles in v1

**Scenario 5: Storage Bloat from Rate Limit Records**
- *Detection:* Increasing gas costs for registration
- *Recovery:* Batch prune expired records
- *Prevention:* Automatic pruning on each registration, periodic cleanup jobs

### 15.4 Upgrade Path

**Package Upgrades (`p/gnoland/antisquat`):**
- Packages are immutable once deployed
- New versions deployed as `p/gnoland/antisquat/v2`, etc.
- Realms must be updated to import new version

**Realm Upgrades (`r/gnoland/users/v2`):**
- State persists in `r/sys/users` (separate realm)
- New controller can be deployed alongside old
- GovDAO proposal to whitelist new, de-whitelist old

**Configuration Updates:**
- Rate limits: GovDAO proposal via `ProposeUpdateRateLimits()`
- Protected names: GovDAO proposal via `ProposeAddProtectedName()`
- Pricing tiers: GovDAO proposal via `ProposeUpdatePricing()`

**Backward Compatibility:**
- All existing usernames remain valid
- No migration of existing users required
- Old name lookups continue to work

### 15.5 Performance Considerations

**Storage Costs:**

| Component | Storage per Entry | Estimated Growth |
|-----------|------------------|------------------|
| Rate limit record | ~100 bytes | O(active users × 3) |
| Protected name | ~150 bytes | O(protected names) |
| Auction commitment | ~200 bytes | O(active auctions × bidders) |

**Gas Costs (Estimated):**

| Operation | Gas Units |
|-----------|-----------|
| Register (standard) | ~50,000 |
| Register (with rate limit check) | ~60,000 |
| Commit bid | ~40,000 |
| Reveal bid | ~45,000 |
| Resolve auction | ~80,000 |

**Optimization Strategies:**
1. **Lazy pruning** - Remove expired rate limit records only when address registers again
2. **Batch operations** - GovDAO can add multiple protected names in single proposal
3. **Efficient lookups** - AVL trees provide O(log n) access

**Scalability Limits:**
- Rate limit storage: ~1M addresses before noticeable impact
- Protected names: ~10K names is practical upper bound
- Concurrent auctions: ~100 recommended maximum

### 15.6 Incident Response

**Severity Levels:**

| Level | Description | Response Time | Example |
|-------|-------------|---------------|---------|
| P0 | Complete system failure | Immediate | Registration completely broken |
| P1 | Security vulnerability | < 4 hours | Bypass allowing unlimited registrations |
| P2 | Degraded functionality | < 24 hours | Rate limiting not working |
| P3 | Minor issue | < 1 week | UI displaying wrong fee |

**Response Procedures:**
1. **Detect** - Monitoring alerts or user reports
2. **Assess** - Determine severity and impact scope
3. **Contain** - GovDAO emergency pause if needed
4. **Fix** - Deploy patched version
5. **Recover** - Restore normal operations
6. **Review** - Post-incident analysis and documentation

**Emergency Contacts:**
- GovDAO multi-sig holders for emergency proposals
- Core team Discord/Telegram for coordination

---

## 16. Future Enhancements

1. **Proof of Humanity** - Integration with external verification
2. **Reputation-based pricing** - Active users get discounts
3. **Name expiration** - Annual renewal requirement
4. **Transfer mechanism** - Allow name transfers with fee
5. **Dispute resolution** - Trademark claims via GovDAO

---

## 17. Implementation Timeline

| Phase | Scope | Priority |
|-------|-------|----------|
| 1 | Core library: types, rate limit, protected, pricing | HIGH |
| 2 | Tests + documentation | HIGH |
| 3 | Controller integration | MEDIUM |
| 4 | GovDAO admin functions | MEDIUM |
| 5 | Vickrey auction | LOW |

---

## 18. Open Questions

1. **Protected names list** - Should it live in antisquat package or separate `p/gnoland/protectednames`?
2. **Pricing tiers** - Are default values appropriate for mainnet?
3. **Rate limit window** - 30 days reasonable? Should it be block-based or time-based?
4. **Auction triggers** - When should auction kick in vs direct registration?

---

## References

- Issue #2727: https://github.com/gnolang/gno/issues/2727
- Issue #2827: https://github.com/gnolang/gno/issues/2827 (r/sys/users v2)
- Issue #3020: https://github.com/gnolang/gno/issues/3020 (Moderation DAOs)
- ENS Pricing: https://docs.ens.domains/registry/eth
- Vickrey Auction: https://en.wikipedia.org/wiki/Vickrey_auction
