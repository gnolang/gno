# antispam

Multi-rule spam scoring engine for Gno realms. Combines content heuristics, rate limiting, reputation signals, Bayesian filtering, duplicate detection, keyword matching, and blocklists into a single weighted score.

## Quick Start

**Option 1: Use the shared realm (recommended for most cases)**

```gno
import antispamr "gno.land/r/gnoland/antispam"
import engine "gno.land/p/gnoland/antispam"

// Score content - reputation auto-populated from shared state
rep := engine.ReputationData{
    AccountAgeDays: 30,
    Balance:        5000000000,
    HasUsername:    true,
}
result := antispamr.Score(author, content, rate, rep, nil, nil, nil, nil, engine.EarlyExitDisabled)

if result.Total >= engine.ThresholdReject {
    // reject (score >= 8)
} else if result.Total >= engine.ThresholdHide {
    // hide from default view (score >= 5)
}

// Record moderation decision (requires registration via AdminRegisterCaller)
if result.Total < engine.ThresholdHide {
    antispamr.RecordAccepted(author)
}
```

**Option 2: Self-hosted state (for custom rules or isolated scoring)**

```gno
import "gno.land/p/gnoland/antispam"

// Create persistent state in your realm
var (
    corpus = antispam.NewCorpus()
    fps    = antispam.NewFingerprintStore()
    bl     = antispam.NewBlocklist()
    kw     = antispam.NewKeywordDict()
)

// Score content
result := antispam.Score(antispam.ScoreInput{
    Author:  author,
    Content: content,
    Rate:    antispam.RateState{PostCount: 1, WindowSeconds: 3600},
    Rep:     antispam.ReputationData{AccountAgeDays: 30, Balance: 5000000000, HasUsername: true},
    Corpus:  corpus,
    Fps:     fps,
    Bl:      bl,
    Dict:    kw,
})
```

**Important:** A zero-value `ReputationData` triggers NEW_ACCOUNT(2) + NO_USERNAME(1) + LOW_BALANCE(1) = 4 points. Always fill `Rep` accurately to avoid false positives.

## How It Works

**Architecture:**
- **Allowlist bypass** - Trusted addresses skip all scoring
- **Short-circuit** - `BLOCKED_ADDRESS` returns immediately (score 99)
- **Ascending-cost pipeline** - Cheap rules first (O(1) rate/reputation), then content scan (O(n)), then expensive rules (regex, Bayes, fingerprints)
- **Single-pass content analysis** - Caps, punctuation, links, unicode abuse detected in one scan
- **Rule independence** - Each rule triggers and scores independently

**Detection strategies:**

| Strategy | How it works |
|----------|--------------|
| **Content checks** | All caps, excessive punctuation, repeated chars ("BUYYYYYYY"), link spam, short content with just a URL |
| **Unicode abuse** | Zalgo text (stacked diacritics), invisible chars (zero-width spaces), homoglyph mixing (Cyrillic "a" in Latin word) |
| **Rate limiting** | Flags fast posting. Moderate bursts get half weight, heavy flooding gets full weight |
| **Account reputation** | Age, balance, username, flag ratio, ban history. New accounts + no username + low balance = high score |
| **Bayesian filter** | Trained on spam/ham examples. Needs 3+ spam-heavy words to trigger. Shares tokens with keyword rule |
| **Duplicate detection** | MinHash fingerprints catch copy-paste spam waves across realms |
| **Blocklists** | Regex patterns for scam formats ("send X tokens", "free airdrop", email addresses). Address blocklist for known spammers |
| **Keyword detection** | Co-occurrence model: multiple spam keywords together trigger, single words don't. Includes leet-speak normalization |

## Scoring Rules

Scores accumulate across all triggered rules. **Thresholds: Hide >= 5, Reject >= 8**.

| Rule | Weight | Trigger |
|------|--------|---------|
| BLOCKED_ADDRESS | 99 | Address in blocklist |
| BLOCKED_PATTERN | 5 | Content matches a regex pattern |
| NEAR_DUPLICATE | 4 | MinHash similarity to known spam |
| RATE_BURST | 4 | >10 submissions/hour (half weight at 10-20) |
| BAYES_SPAM | 3 | Bayesian classifier flags spam-heavy tokens |
| BAD_REPUTATION | 3 | Flag ratio >30% or flags with no accepted content |
| KEYWORD_SPAM | 3 | >=2 spam keywords with combined weight >=4 |
| SHORT_WITH_LINK | 3 | Body <=30 chars with a URL |
| ZALGO_TEXT | 3 | Excessive combining diacritical marks |
| ALL_CAPS | 2 | >50% uppercase (min 10 letters) |
| LINK_HEAVY | 2 | >3 URLs in content |
| NEW_ACCOUNT | 2 | Account <1 day old with no accepted content |
| INVISIBLE_CHARS | 2 | Zero-width or directional override characters |
| HOMOGLYPH_MIX | 2 | Mixed scripts in a single word (latin + cyrillic) |
| REPEATED_CHARS | 2 | 4+ consecutive identical characters |
| BANNED_BEFORE | 1-3 | +1 per past ban, capped at 3 |
| EXCESSIVE_PUNCT | 1 | >20% punctuation characters |
| NO_USERNAME | 1 | No registered username |
| LOW_BALANCE | 1 | Balance < 1000 GNOT |

## State Containers

| Type | Purpose | Bounds |
|------|---------|--------|
| `Corpus` | Bayesian token statistics (spam/ham counts) | 10,000 tokens max |
| `FingerprintStore` | MinHash signatures of known spam | 500 entries (LRU eviction) |
| `Blocklist` | Blocked addresses, allowlist, regex patterns | 30 patterns max |
| `KeywordDict` | Spam keywords with weights (1-3) | 10,000 keywords max |
| `TrainingGuard` | Circuit breaker for safe auto-training | Configurable (default: 500 max trains) |

All state parameters can be `nil` - scoring degrades gracefully using only available rules.

## Shared Realm (`r/gnoland/antispam`)

The `gno.land/r/gnoland/antispam` realm provides shared on-chain state (pre-trained corpus, ~400 keywords, 21 scam regex patterns) and per-address reputation tracking.

**Trust model:**
- **Scoring is public** - Any realm can call `Score()` (read-only, no side effects)
- **Training is admin-only** - Corpus, patterns, keywords, blocklist modifications require admin
- **Reputation recording requires registration** - Only realms registered via `AdminRegisterCaller()` can call `RecordAccepted()`, `RecordFlag()`, `RecordBan()`

**Reputation tracking:**

The realm stores lightweight per-address counters (no content stored):
- `TotalAccepted` - Content accepted into the system
- `FlaggedCount` - Content flagged by moderators
- `BanCount` - Times the address was banned (permanent history)

**Reputation recording workflow:**

`Score()` is **read-only** and does NOT automatically update reputation. The calling realm must explicitly record moderation actions:

```gno
// 1. Score the content
result := antispamr.Score(author, content, rate, rep, nil, nil, nil, nil, engine.EarlyExitDisabled)

// 2. Decide based on score (Score() does NOT store anything)

// 3. Manually record the moderation decision
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
- Prevents false positives from poisoning reputation data
- Allows human review before reputation changes
- Separates detection (scoring) from action (reputation update)

**Future:** Auto-feedback could be added as an optional mode with safeguards against spiraling false positives.

## Advanced Usage

### Gas Optimization

Control scoring cost with three levers:

**1. Input cap** - Content is truncated to 4096 bytes to prevent gas abuse

**2. Nil state** - Pass `nil` for expensive state containers to skip those rules:

```gno
// Lightweight: cheap rules only
result := antispam.Score(antispam.ScoreInput{Content: content, Rate: rate, Rep: rep})

// Standard: full state (corpus, keywords, etc.)
result := antispam.Score(antispam.ScoreInput{
    Content: content, Rate: rate, Rep: rep,
    Corpus: corpus, Fps: fps, Bl: bl, Dict: kw,
})
```

**3. Early exit** - Set `EarlyExitAt` to stop scoring when threshold is reached:

```gno
// Early exit at reject threshold (skip expensive rules if cheap rules already reach 8)
result := antispam.Score(antispam.ScoreInput{
    Content: content, Rate: rate, Rep: rep,
    Corpus: corpus, Fps: fps, Bl: bl, Dict: kw,
    EarlyExitAt: antispam.ThresholdReject,  // Stop at score >= 8
})

// Full scoring (complete score, no early exit)
result := antispam.Score(antispam.ScoreInput{
    Content: content, Rate: rate, Rep: rep,
    Corpus: corpus, Fps: fps, Bl: bl, Dict: kw,
    EarlyExitAt: antispam.EarlyExitDisabled,  // Get complete score
})
```

**User-based tiering** - Use reputation history to skip expensive rules for trusted users:

```gno
rep := antispamr.GetReputation(author)

if rep.TotalAccepted > 10 && rep.FlaggedCount == 0 && chainData.AccountAgeDays > 30 {
    // Established user: lightweight scoring
    result = antispamr.Score(author, content, rate, chainData,
        nil, nil, nil, nil,  // Skip Bayes/keywords/fingerprints
        engine.ThresholdReject,
    )
} else {
    // New/suspicious user: full pipeline
    result = antispamr.Score(author, content, rate, chainData,
        nil, nil, nil, nil, engine.EarlyExitDisabled,
    )
}
```

### Auto-Training Safety

When auto-training the Corpus from moderation actions, use `TrainingGuard` to prevent feedback loops:

```gno
var (
    corpus = antispam.NewCorpus()
    guard  = antispam.NewTrainingGuard()
)

func ModerateContent(author, content string) {
    score := antispam.Score(antispam.ScoreInput{...})

    if score.Total >= antispam.ThresholdReject {
        if guard.ShouldTrain(score) {
            corpus.Train(content, true)  // Train as spam
            guard.RecordTrain()
        }
    }
}

func AdminResetGuard(_ realm) {
    assertAdmin()
    guard.Reset()  // Admin reviews and resets counter
}
```

**TrainingGuard safeguards:**
1. **Minimum score** (default: 10) - Only train on clear spam, not borderline content
2. **Multi-rule consensus** (default: 3 rules) - Require multiple signals, prevent single-rule bias
3. **Circuit breaker** (default: 500 trains) - Cap total auto-trains until admin review

Without `TrainingGuard`, auto-training creates feedback loops: borderline content -> trained as spam -> corpus bias -> more false positives -> more training -> amplified bias.

Use `NewTrainingGuardCustom(minScore, minRules, maxTrains)` to adjust thresholds.

### Realm Deployment

State containers are created empty by default. For large datasets (keywords, patterns), populate via a separate admin function post-deploy to avoid exceeding storage deposit limits:

```gno
func init() {
    adminAddr = runtime.OriginCaller()
    corpus   = antispam.NewCorpus()
    fps      = antispam.NewFingerprintStore()
    bl       = antispam.NewBlocklist()
    keywords = antispam.NewKeywordDict()
}

// Called by admin after deployment (separate tx, separate storage deposit)
func AdminLoadDefaults(_ realm) {
    assertAdmin()
    bl.AddPattern(defaultPattern)     // 21 scam categories
    keywords.BulkAdd(defaultKeywords)  // ~400 spam keywords
}
```

The `gno.land/r/gnoland/antispam` realm follows this pattern.

## Filetests

The `z*_filetest.gno` files demonstrate real scenarios:

| File | Demonstrates |
|------|--------------|
| `z1_comprehensive` | All 19 rules firing across 8 scenarios (start here) |
| `z2_multi_user` | Concurrent users with different reputation profiles |
| `z3_bayes_training` | Corpus training effects on detection |
| `z4_fingerprint` | Near-duplicate content detection |
| `z5_blocklist` | Address/pattern blocking and allowlist operations |
| `z6_spam_evolution` | Legitimate user turning into a spammer |
| `z7_unicode_abuse` | Zalgo text, invisible chars, homoglyph mixing |
| `z8_early_exit` | EarlyExitAt gas optimization (cheap vs expensive rules) |
| `z9_pattern_categories` | All 21 regex categories + false positive checks |
| `z10_size_comparison` | Scoring across different content sizes (30-4096+ chars) |

**Realm-specific filetests** (in `r/gnoland/antispam/`):

| File | Demonstrates |
|------|--------------|
| `z1_score_demo` | Basic scoring via shared realm + manual reputation recording |
| `z2_reputation_lifecycle` | Full reputation tracking workflow (register caller, record actions, score reflects reputation) |
