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

## Detection Strategies

**Content checks** - Looks for shouting in all caps, excessive punctuation, link spam, and short posts with just a URL.

**Unicode abuse** - Detects three forms of text manipulation: zalgo text (stacking combining diacritical marks to distort rendering), invisible characters (zero-width spaces, directional overrides used for obfuscation), and homoglyph mixing (substituting Latin characters with visually identical Cyrillic or Greek look-alikes in the same word, e.g. Latin "a" replaced by Cyrillic "a" in "payment").

**Rate limiting** - Flags accounts posting too fast. Moderate bursts get half weight, heavy flooding gets full weight.

**Account reputation** - Considers account age, balance, username, how often they've been flagged, and prior bans. New accounts with no username and low balance rack up points. High flag ratios trigger bad reputation scoring.

**Bayesian filter** - Uses Bayesian classification trained on spam and legitimate examples. Checks if enough words in the post are more common in spam than real posts. Needs at least 3 spam-heavy words to trigger. Shares pre-computed deduplicated tokens with the keyword rule to avoid redundant work.

**Duplicate detection** - Catches copy-paste spam waves where the same message is posted across threads. Compares MinHash fingerprints of new posts against recent spam. Threshold set to catch paraphrases but not unrelated content.

**Blocklists/Allowlists** - Regex patterns catch common scam formats like "send X tokens" or "free airdrop". Address blocklist immediately rejects known bad actors (score 99) and short-circuits all further checks. Supports allowlisting trusted addresses to bypass scoring entirely.

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
| BANNED_BEFORE | 1-3 | +1 per past ban, capped at 3 |
| EXCESSIVE_PUNCT | 1 | >20% punctuation characters |
| NO_USERNAME | 1 | No registered username |
| LOW_BALANCE | 1 | Balance < 1000 GNOT |

## State Containers

| Type | Purpose | Bounds |
|------|---------|--------|
| `Corpus` | Bayesian token statistics (spam/ham counts) | Unbounded (caller manages) |
| `FingerprintStore` | MinHash signatures of known spam | Capped at 500 entries (LRU eviction) |
| `Blocklist` | Blocked addresses, allowlist, regex patterns | Patterns capped at 30 |
| `KeywordDict` | Spam keywords with per-word weights (1-3) | Unbounded (caller manages) |

## Gas Optimization

Scoring cost depends on which rules are active. Control it with two levers:

- **Nil state**: pass nil for `Corpus`, `Fps`, `Bl`, or `Dict` to skip those rules entirely. All-nil runs only cheap Phase 1 checks (content heuristics, rate, reputation).
- **Early exit**: set `EarlyExitAt` to a threshold (e.g. `ThresholdReject`) to skip expensive Phase 2 rules when cheap rules already reach the threshold. Use `EarlyExitDisabled` for a complete score.

```gno
// Lightweight: Phase 1 only (nil state, cheapest)
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
- **Two-phase scoring**: cheap rules (O(1) lookups, character scans) run first; expensive rules (regex, Bayes, keywords, fingerprints) run second. Set `EarlyExitAt` to a positive threshold (e.g. `ThresholdReject`) to skip phase 2 when cheap rules already reach it. Leave at `EarlyExitDisabled` for a complete score.
- **Single-pass content analysis**: caps, punctuation, links, and unicode abuse detected in one scan
- **Rule independence**: each rule triggers and scores independently

## Filetests

The `z*_filetest.gno` files demonstrate real scenarios:

- `z1` - Comprehensive: all 18 rules firing across 8 scenarios (start here)
- `z2` - Multi-user: concurrent users with different reputation profiles
- `z3` - Bayes training: corpus training effects on detection
- `z4` - Fingerprint: near-duplicate content detection
- `z5` - Blocklist: address/pattern blocking and allowlist operations
- `z6` - Spam evolution: legitimate user turning into a spammer
- `z7` - Unicode abuse: zalgo text, invisible chars, homoglyph mixing
- `z8` - Early exit: EarlyExitAt gas optimization (Phase 1 vs Phase 2)
- `z9` - Pattern categories: all 20 regex categories + false positive checks
