# r/gnoland/antispam

Shared on-chain anti-spam scoring service for the Gno ecosystem. Provides pre-trained spam detection (Bayesian corpus, ~400 keywords, 21 scam regex patterns) and per-address reputation tracking.

## Quick Start

```gno
import antispamr "gno.land/r/gnoland/antispam"
import engine "gno.land/p/gnoland/antispam"

// Score content - reputation auto-populated from shared state
rep := engine.ReputationData{
    AccountAgeDays: 30,
    Balance:        5000000000,  // in ugnot
    HasUsername:    true,
}
result := antispamr.Score(
    author, content, rate, rep,
    nil, nil, nil, nil,  // use shared state
    engine.EarlyExitDisabled,
)

if result.Total >= engine.ThresholdReject {
    // reject (score >= 8)
} else if result.Total >= engine.ThresholdHide {
    // hide (score >= 5)
}

// Record moderation decision (requires registration)
if result.Total < engine.ThresholdHide {
    antispamr.RecordAccepted(author)
}
```

## API Reference

### Public Functions (anyone can call)

**Scoring:**
```gno
Score(author address, content string, rate RateState, rep ReputationData,
      callerCorpus *Corpus, callerFps *FingerprintStore,
      callerDict *KeywordDict, callerBl *Blocklist,
      earlyExitAt int) SpamScore
```
- Evaluates content using shared state (or caller-provided state if not nil)
- Auto-populates `FlaggedCount`, `BanCount`, `TotalAccepted` from internal reputation
- Caller only provides chain data: `AccountAgeDays`, `Balance`, `HasUsername`
- Pass `nil` for all state params to use shared instances
- **Read-only** - does NOT update reputation counters

**Reputation queries:**
```gno
GetReputation(addr address) ReputationData
ReputationCount() int
TrustedCallerCount() int
```

### Reputation Recording (trusted callers only)

**Registration required:** Call `AdminRegisterCaller(realm)` first to authorize your realm.

```gno
RecordAccepted(realm, addr address)  // Increment accepted content count
RecordFlag(realm, addr address)      // Increment flagged content count
RecordBan(realm, addr address)       // Increment ban count (permanent)
```

**IMPORTANT:** `Score()` does NOT automatically call these functions. You must manually record moderation decisions:

```gno
// After scoring
if result.Total < engine.ThresholdHide {
    antispamr.RecordAccepted(author)  // Content accepted
} else if moderatorFlagged {
    antispamr.RecordFlag(author)      // Moderator confirmed spam
}

if adminBanned {
    antispamr.RecordBan(author)       // Admin banned (permanent)
}
```

**Why manual recording?**
- Gives realms full control over moderation decisions
- Prevents false positives from poisoning shared reputation data
- Allows human review before reputation changes
- Separates detection (scoring) from action (reputation update)

### Admin Functions

**Setup:**
```gno
AdminLoadDefaults(realm)  // Load pre-configured keywords and patterns (call post-deploy)
AdminSetAdmin(realm, newAdmin address)
```

**Reputation management:**
```gno
AdminResetReputation(realm, addr address)
AdminRegisterCaller(realm, callerAddr address)   // Allow realm to record reputation
AdminRemoveCaller(realm, callerAddr address)
```

**Blocklist management:**
```gno
AdminBlockAddress(realm, addr address)
AdminUnblockAddress(realm, addr address)
AdminAllowAddress(realm, addr address)           // Bypass all scoring
AdminRemoveAllow(realm, addr address)
AdminAddPattern(realm, pattern string)           // Add regex pattern
AdminRemovePattern(realm, pattern string)
```

**Keyword management:**
```gno
AdminAddKeyword(realm, word string, weight int)  // weight: 1-3
AdminRemoveKeyword(realm, word string)
AdminBulkAddKeywords(realm, data string)         // newline-separated "word:weight"
```

**Training:**
```gno
AdminTrain(realm, content string, isSpam bool)   // Train Bayesian corpus
AdminAddSpamFingerprint(realm, content string)   // Add known spam fingerprint
```

## Trust Model

| Action | Who can do it |
|--------|---------------|
| **Score content** | Anyone (public, read-only) |
| **Query reputation** | Anyone (public, read-only) |
| **Record reputation** | Registered realms only (via `AdminRegisterCaller`) |
| **Modify shared state** | Admin only (patterns, blocklist, keywords, corpus) |
| **Admin operations** | Admin only (set by `OriginCaller()` at deployment) |

**Security:**
- Reputation recording requires realm registration to prevent data poisoning
- Admin operations use `OriginCaller()` (wallet signature), not `PreviousRealm()`
- Blocklisted addresses short-circuit immediately (score 99, no further checks)
- Allowlisted addresses bypass all scoring

## Shared State

The realm maintains:
- **Bayesian corpus** - Pre-trained on spam/ham examples
- **Keyword dictionary** - ~400 spam keywords with weights (1-3)
- **Blocklist** - Combined regex pattern covering 21 scam categories
- **Fingerprint store** - MinHash signatures of known spam
- **Reputation counters** - Per-address `TotalAccepted`, `FlaggedCount`, `BanCount`

State is created empty on deployment. Admin must call `AdminLoadDefaults()` post-deploy to populate keywords and patterns (separate tx to avoid storage deposit limits).

## Reputation Counters

Lightweight per-address data (no content stored):

| Counter | Meaning | Updated by |
|---------|---------|------------|
| `TotalAccepted` | Content accepted into the system | `RecordAccepted()` |
| `FlaggedCount` | Content flagged by moderators | `RecordFlag()` |
| `BanCount` | Times banned (permanent history) | `RecordBan()` |

Counters are used by the scoring rules:
- `NEW_ACCOUNT` (2 pts) - Account < 1 day old with `TotalAccepted == 0`
- `BAD_REPUTATION` (3 pts) - `FlaggedCount / TotalAccepted > 30%`
- `BANNED_BEFORE` (1-3 pts) - +1 pt per ban, capped at 3

## Gas Optimization

**Early exit** - Stop scoring when threshold is reached:
```gno
// Skip expensive rules if cheap rules already reach reject threshold
result := antispamr.Score(author, content, rate, rep,
    nil, nil, nil, nil,
    engine.ThresholdReject,  // early exit at score >= 8
)
```

**User tiering** - Use reputation history to skip expensive rules for trusted users:
```gno
rep := antispamr.GetReputation(author)

if rep.TotalAccepted > 10 && rep.FlaggedCount == 0 && chainData.AccountAgeDays > 30 {
    // Established user: lightweight scoring
    result = antispamr.Score(author, content, rate, chainData,
        nil, nil, nil, nil, engine.ThresholdReject)
} else {
    // New/suspicious user: full pipeline
    result = antispamr.Score(author, content, rate, chainData,
        nil, nil, nil, nil, engine.EarlyExitDisabled)
}
```

**Custom state** - Override shared state with your own instances:
```gno
// Use shared Bayesian corpus but custom blocklist
myBlocklist := engine.NewBlocklist()
myBlocklist.AddPattern(`my-custom-pattern`)

result := antispamr.Score(author, content, rate, rep,
    nil, nil, nil, myBlocklist,  // custom blocklist, rest shared
    engine.EarlyExitDisabled)
```

## Deployment

The realm is deployed empty. Post-deployment setup:

```gno
// 1. Admin loads default patterns and keywords
antispamr.AdminLoadDefaults()

// 2. Admin registers trusted caller realms
antispamr.AdminRegisterCaller(boardsRealmAddr)
antispamr.AdminRegisterCaller(forumsRealmAddr)

// 3. Optional: train Bayesian corpus with examples
antispamr.AdminTrain("buy cheap viagra now!!!", true)  // spam
antispamr.AdminTrain("governance proposal discussion", false)  // ham
```

## Examples

See filetests for complete examples:

| Filetest | Demonstrates |
|----------|--------------|
| [z1_score_demo](z1_score_demo_filetest.gno) | Basic scoring + manual reputation recording workflow |
| [z2_reputation_lifecycle](z2_reputation_lifecycle_filetest.gno) | Full lifecycle: register caller, record actions, score reflects reputation |

## Technical Details

For architecture, detection strategies, and scoring rules, see the package documentation:
- [p/gnoland/antispam/README.md](../../p/gnoland/antispam/README.md)

## On-Chain Status

Visit the realm on-chain to see current state:
```
https://gno.land/r/gnoland/antispam
```

The `Render()` function displays:
- Admin address
- Corpus size (token count)
- Fingerprint store size
- Keyword dictionary size
- Blocked/allowed address counts
- Active regex patterns
- Tracked addresses count
- Trusted callers count
