# antispam

Multi-rule spam scoring engine for Gno realms. Combines content heuristics, rate limiting, reputation signals, Bayesian filtering, duplicate detection, keyword matching, and blocklists into a single weighted score.

Realms own and persist all state. The package provides pure scoring functions: no side effects.

## Usage

```gno
import "gno.land/p/gnoland/antispam"

// Realm creates persistent state
var (
    corpus = antispam.NewCorpus()
    fps    = antispam.NewFingerprintStore()
    bl     = antispam.NewBlocklist()
    kw     = antispam.NewKeywordDict()
)

// Score each post
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

if result.Total >= antispam.ThresholdReject {
    // reject
} else if result.Total >= antispam.ThresholdHide {
    // hide from default view
}
```

All state parameters (`Corpus`, `Fps`, `Bl`, `Dict`) can be nil: scoring degrades gracefully using only the rules that have state available.

**Important:** A zero-value `ReputationData` triggers NEW_ACCOUNT(2) + NO_USERNAME(1) + LOW_BALANCE(1) = 4 points. Callers must fill `Rep` accurately to avoid false positives.

## Realm Deployment

State containers (`Blocklist`, `KeywordDict`, etc.) are created empty by default.

**Recommended pattern:** create empty state in `init()`, then populate via a separate admin function post-deploy:

```gno
func init() {
    adminAddr = runtime.OriginCaller()
    corpus   = antispam.NewCorpus()
    fps      = antispam.NewFingerprintStore()
    bl       = antispam.NewBlocklist()
    keywords = antispam.NewKeywordDict()
}

// AdminLoadDefaults loads regex patterns and keywords.
// Called by admin after deployment (separate tx, separate storage deposit).
func AdminLoadDefaults(_ realm) {
    assertAdmin()
    bl.AddPattern(defaultPattern)
    keywords.BulkAdd(defaultKeywords)
}
```

The `gno.land/r/gnoland/antispam` realm follows this pattern with ~400 pre-configured spam keywords and a combined regex covering 21 scam categories.

## Detection Strategies

**Content checks** - Looks for shouting in all caps, excessive punctuation, repeated characters (e.g. "BUYYYYYYY"), link spam, and short posts with just a URL.

**Unicode abuse** - Detects three forms of text manipulation: zalgo text (stacking combining diacritical marks to distort rendering), invisible characters (zero-width spaces, directional overrides used for obfuscation), and homoglyph mixing (substituting Latin characters with visually identical Cyrillic or Greek look-alikes in the same word, e.g. Latin "a" replaced by Cyrillic "a" in "payment").

**Rate limiting** - Flags accounts posting too fast. Moderate bursts get half weight, heavy flooding gets full weight.

**Account reputation** - Considers account age, balance, username, how often they've been flagged, and prior bans. New accounts with no username and low balance rack up points. High flag ratios trigger bad reputation scoring.

**Bayesian filter** - Uses Bayesian classification trained on spam and legitimate examples. Checks if enough words in the post are more common in spam than real posts. Needs at least 3 spam-heavy words to trigger. Shares pre-computed deduplicated tokens with the keyword rule to avoid redundant work.

**Duplicate detection** - Catches copy-paste spam waves where the same message is posted across threads. Compares MinHash fingerprints of new posts against recent spam. Threshold set to catch paraphrases but not unrelated content.

**Blocklists/Allowlists** - Regex patterns catch common scam formats like "send X tokens", "free airdrop", or personal email addresses (gmail, yahoo, etc.) used for off-platform contact redirection. Address blocklist immediately rejects known bad actors (score 99) and short-circuits all further checks. Supports allowlisting trusted addresses to bypass scoring entirely.

**Keyword detection** - Uses a co-occurrence model: one suspicious word alone won't trigger, but multiple spam keywords together will. Keywords are weighted by how spam-specific they are. Includes leet-speak normalization (e.g. "fr33 a1rdr0p" is recognized as "free airdrop") to catch common bypass attempts.

## Scoring Rules

Scores accumulate across all triggered rules. Thresholds: **Hide >= 5**, **Reject >= 8**.

| Rule | Weight | Trigger |
|------|--------|---------|
| BLOCKED_ADDRESS | 99 | Address in blocklist |
| BLOCKED_PATTERN | 5 | Content matches a regex pattern |
| NEAR_DUPLICATE | 4 | MinHash similarity to known spam |
| RATE_BURST | 4 | >10 posts/hour (half weight at 10-20) |
| BAYES_SPAM | 3 | Bayesian classifier flags spam-heavy tokens |
| BAD_REPUTATION | 3 | Flag ratio >30% or flags with no visible posts |
| KEYWORD_SPAM | 3 | >=2 spam keywords with combined weight >=4 |
| SHORT_WITH_LINK | 3 | Body <=30 chars with a URL |
| ZALGO_TEXT | 3 | Excessive combining diacritical marks |
| ALL_CAPS | 2 | >50% uppercase (min 10 letters) |
| LINK_HEAVY | 2 | >3 URLs in content |
| NEW_ACCOUNT | 2 | Account <1 day old with no posts |
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
| `Corpus` | Bayesian token statistics (spam/ham counts) | Capped at 10,000 unique tokens |
| `FingerprintStore` | MinHash signatures of known spam | Capped at 500 entries (LRU eviction) |
| `Blocklist` | Blocked addresses, allowlist, regex patterns | Patterns capped at 30 |
| `KeywordDict` | Spam keywords with per-word weights (1-3) | Capped at 10,000 keywords |
| `TrainingGuard` | Circuit breaker for safe auto-training | Configurable (default: 500 max trains) |

## Auto-Training Safety

When auto-training the Corpus from moderation actions (e.g. training rejected posts as spam), use `TrainingGuard` to prevent feedback loops:

```gno
var (
    corpus = antispam.NewCorpus()
    guard  = antispam.NewTrainingGuard()
)

func ModeratePost(author, content string) {
    score := antispam.Score(antispam.ScoreInput{...})

    if score.Total >= antispam.ThresholdReject {
        if guard.ShouldTrain(score) {
            corpus.Train(content, true)
            guard.RecordTrain()
        }
        // reject post regardless
    }
}

func AdminResetGuard(_ realm) {
    assertAdmin()
    guard.Reset()
}
```

`TrainingGuard` enforces three safeguards:

1. **Minimum score** (default: 10) - Only train on clear spam, not borderline content (score 8-9)
2. **Multi-rule consensus** (default: 3 rules) - Require multiple detection signals, not single-rule bias
3. **Circuit breaker** (default: 500 trains) - Cap total auto-trains until manual admin review

Without `TrainingGuard`, auto-training can create a feedback loop: borderline content gets auto-trained as spam, biasing the corpus, which causes similar legitimate posts to score higher, which get auto-trained, amplifying the bias. The token decay mechanism in `Corpus` slows this drift but does not prevent it alone.

Use `NewTrainingGuardCustom(minScore, minRules, maxTrains)` to adjust thresholds for your realm.

## Gas Optimization

Scoring cost depends on which rules are active. Control it with two levers:

- **Input cap**: content is truncated to `MaxInputLength` (4096 bytes) to prevent gas abuse via oversized inputs.
- **Nil state**: pass nil for `Corpus`, `Fps`, `Bl`, or `Dict` to skip those rules entirely. All-nil runs only cheap rules (rate, reputation, content heuristics).
- **Early exit**: set `EarlyExitAt` to a threshold (e.g. `ThresholdReject`) to skip expensive rules when cheap rules already reach the threshold. Use `EarlyExitDisabled` for a complete score.

```gno
// Lightweight: cheap rules only (nil state, cheapest)
result := antispam.Score(antispam.ScoreInput{Content: content, Rate: rate, Rep: rep})

// Standard: full state with early exit
result := antispam.Score(antispam.ScoreInput{
    Content: content, Rate: rate, Rep: rep,
    Corpus: corpus, Fps: fps, Bl: bl, Dict: kw,
    EarlyExitAt: antispam.ThresholdReject,
})

// Full: all rules, complete score
result := antispam.Score(antispam.ScoreInput{
    Content: content, Rate: rate, Rep: rep,
    Corpus: corpus, Fps: fps, Bl: bl, Dict: kw,
    EarlyExitAt: antispam.EarlyExitDisabled,
})
```

## Architecture

- **Allowlist bypass**: allowed addresses skip all scoring
- **Short-circuit**: `BLOCKED_ADDRESS` returns immediately, no further rules evaluated
- **Ascending-cost pipeline**: rules ordered by cost (O(1) rate/reputation, O(n) content scan, regex, tokenization + Bayes, keywords, fingerprint hashing) with `earlyExit` checks between each group. Set `EarlyExitAt` to a positive threshold (e.g. `ThresholdReject`) to skip costlier rules when cheap rules already reach it. Leave at `EarlyExitDisabled` for a complete score.
- **Single-pass content analysis**: caps, punctuation, links, and unicode abuse detected in one scan
- **Rule independence**: each rule triggers and scores independently

## Boards2 Integration

boards2 may call `Score()` on the realm, not the package directly. The realm holds shared persistent state: one Bayesian corpus, one blocklist, one fingerprint store, and ~400 pre-loaded keywords that all calling realms benefit from. When an admin blocks an address or trains the corpus, every realm using the shared state picks it up immediately.

### Scoring

1. boards2 might call `Score()` during post creation (CreateThread/CreateReply)
2. boards2 might call `Score()` again in `Render()` to filter posts that originate from a newly banished address.

Today the boards2 realm owner can already hide a post with a single flag (no need to reach the threshold). That's fast but manual. To go further:

- **Option A: Scoring at Render()** - Instead of scoring only at creation time, boards2 re-scores during every `Render()`. This eliminates delays and maximizes coverage, but consumes CPU on every render. Can be mitigated by caching the score on the post and only re-scoring posts from recently blocked addresses.

- **Option B: Retroactive cascade on ban/block** - When an admin calls `AdminBlockAddress()` in `r/gnoland/antispam`, the boards2 realm iterates over all posts from that address and marks them as Hidden = true. Less elegant but more efficient.

### Gas Cost for Legitimate Content

Most messages are not spam. For legitimate users (score 0), `EarlyExitAt` provides no savings because the score never reaches the threshold - all rules execute. This is the "tax" on legitimate content, but scoring is 100% read-only (no state writes), so it remains cheaper than storing the message itself.

To reduce cost for the common case (established users posting normal content):

```gno
// Established users: lightweight scoring (skip expensive rules)
if user.AccountAgeDays > 30 && user.TotalPosts > 10 {
    result = antispam.Score(antispam.ScoreInput{
        Content: content, Rate: rate, Rep: rep,
        Bl: bl, // blocklist only - catches blocked addresses + regex patterns
        EarlyExitAt: antispam.ThresholdReject,
    })
} else {
    // New/unknown users: full pipeline
    result = antispam.Score(antispam.ScoreInput{
        Content: content, Rate: rate, Rep: rep,
        Corpus: corpus, Fps: fps, Bl: bl, Dict: dict,
    })
}
```

Passing nil for `Corpus`, `Fps`, and `Dict` skips tokenization, Bayes, keywords, and fingerprinting. For established users this is safe: blocklist + content heuristics + rate limiting catch the obvious cases. Full pipeline runs only for new/unknown accounts where the risk is higher.

### Reputation

The antispam package takes `ReputationData` as input - it does not track reputation itself. The calling realm (or `r/gnoland/antispam`) is responsible for gathering reputation data:

- **Chain data**: account age, balance (from `std` or `banker`)
- **Moderation data**: flag counts, ban counts (tracked by the realm)
- **Activity data**: post counts, username (from boards2 or user registry)

The `r/gnoland/antispam` realm is the natural place to centralize moderation state (flags, bans, post counts). boards2 would call `antispamRealm.Score(author, content)` and the realm gathers reputation internally, so boards2 doesn't need to track flags/bans separately.

### Admin Notification

When a post's spam score exceeds a threshold, the realm can emit an event that off-chain indexers read. An off-chain bot listens for these events and sends alerts to Telegram, Discord, etc. This is beyond the current PR's scope but represents the correct approach to alerting.

### Auto-Learning

The Bayesian corpus can learn automatically from moderation actions: no dedicated admin transactions needed. When a moderator flags or hides a post, boards2 calls `corpus.Train(body, true)`. Posts that survive a certain amount of time without flags get `corpus.Train(body, false)`. The corpus improves organically as moderation happens.

Optionally, posts scoring above ThresholdReject can be auto-trained as spam without moderator action. Use `TrainingGuard` to prevent feedback loops (see Auto-Training Safety above).

## Filetests

The `z*_filetest.gno` files demonstrate real scenarios:

- `z1` - Comprehensive: all 19 rules firing across 8 scenarios (start here)
- `z2` - Multi-user: concurrent users with different reputation profiles
- `z3` - Bayes training: corpus training effects on detection
- `z4` - Fingerprint: near-duplicate content detection
- `z5` - Blocklist: address/pattern blocking and allowlist operations
- `z6` - Spam evolution: legitimate user turning into a spammer
- `z7` - Unicode abuse: zalgo text, invisible chars, homoglyph mixing
- `z8` - Early exit: EarlyExitAt gas optimization (cheap vs expensive rules)
- `z9` - Pattern categories: all 21 regex categories + false positive checks
- `z10` - Gas benchmark: full pipeline with realistic state volumes for profiling
