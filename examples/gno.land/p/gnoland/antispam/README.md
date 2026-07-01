# antispam

Multi-rule spam scoring for Gno realms. Combines content heuristics, rate limits, reputation, Bayesian filter, duplicates, keywords, and blocklists.

## Usage

Most realms should use the shared realm at `r/gnoland/antispam`:

```gno
import antispamr "gno.land/r/gnoland/antispam"
import engine "gno.land/p/gnoland/antispam"

rep := engine.ReputationData{
    AccountAgeDays: 30,
    Balance:        5000000000,
    HasUsername:    true,
}
result := antispamr.Score(author, content, rate, rep, nil, nil, nil, nil, engine.EarlyExitDisabled)

if result.Total >= engine.ThresholdReject {
    // reject (score >= 8)
} else if result.Total >= engine.ThresholdHide {
    // hide (score >= 5)
}

if result.Total < engine.ThresholdHide {
    antispamr.RecordAccepted(cross, author)
}
```

For custom rules or isolated scoring, host your own state:

```gno
import "gno.land/p/gnoland/antispam"

var (
    corpus = antispam.NewCorpus()
    fps    = antispam.NewFingerprintStore()
    bl     = antispam.NewBlocklist()
    kw     = antispam.NewKeywordDict()
)

result := antispam.Score(antispam.ScoreInput{
    Author:       author,
    Content:      content,
    Rate:         antispam.RateState{PostCount: 1, WindowSeconds: 3600},
    Reputation:   antispam.ReputationData{AccountAgeDays: 30, Balance: 5000000000, HasUsername: true},
    Corpus:       corpus,
    Fingerprints: fps,
    Blocklist:    bl,
    Keywords:     kw,
})
```

Note: Zero-value ReputationData gives you 4 points (NEW_ACCOUNT + NO_USERNAME + LOW_BALANCE). Fill it properly to avoid false positives.

## How it works

Trusted addresses skip scoring entirely. Blocked addresses get 99 points immediately. Everyone else goes through a pipeline ordered by cost: cheap reputation/rate checks first, then content scan, then expensive rules (regex, Bayes, fingerprints).

Detection strategies:

- Content checks: all caps, excessive punctuation, repeated chars ("BUYYYYYYY"), link spam, short messages that are just a URL
- Unicode abuse: zalgo text, invisible characters, homoglyphs (mixing Cyrillic "a" in Latin words)
- Rate limiting: flags fast posting, scales weight based on severity
- Account reputation: combines age, balance, username, flag ratio, ban history
- Bayesian filter: trained on spam/ham, needs multiple spam-heavy words to trigger
- Duplicates: MinHash fingerprints catch copy-paste spam across realms
- Blocklists: regex patterns for scam formats, address blocklist for known spammers
- Keywords: co-occurrence model where multiple spam words trigger together, minimum matches scale with content length

## Scoring rules

Scores add up across triggered rules. Thresholds: hide at 5, reject at 8.

Rules and weights:
- BLOCKED_ADDRESS (99) - address in blocklist
- BLOCKED_PATTERN (5) - matches regex pattern
- NEAR_DUPLICATE (4) - similar to known spam (MinHash)
- RATE_BURST (4) - posting too fast
- BAYES_SPAM (3) - Bayesian classifier flags it
- BAD_REPUTATION (3) - high flag ratio or prior bans
- KEYWORD_SPAM (3) - multiple spam keywords (min matches scale with length)
- SHORT_WITH_LINK (3) - under 30 chars with a URL
- ZALGO_TEXT (3) - excessive diacritical marks
- ALL_CAPS (2) - over 50% uppercase
- LINK_HEAVY (2) - more than 3 URLs
- NEW_ACCOUNT (2) - under 1 day old, no accepted posts
- INVISIBLE_CHARS (2) - zero-width or directional override chars
- HOMOGLYPH_MIX (2) - mixed scripts in one word
- REPEATED_CHARS (2) - 4+ consecutive identical chars
- BANNED_BEFORE (1-3) - scales with ban count
- EXCESSIVE_PUNCT (1) - over 20% punctuation
- NO_USERNAME (1) - no registered username
- LOW_BALANCE (1) - under 1000 GNOT

## State containers

You can pass nil for any state container and scoring works with what's available:

- Corpus: Bayesian token stats (10k tokens max)
- FingerprintStore: MinHash signatures (500 entries, LRU)
- Blocklist: blocked addresses, allowlist, regex (30 patterns max)
- KeywordDict: spam keywords with weights 1-3 (10k max)
- TrainingGuard: circuit breaker for auto-training (default 500 trains max)

## Shared realm

The `r/gnoland/antispam` realm hosts shared state (pre-trained corpus, 400 keywords, 21 scam patterns) and tracks per-address reputation.

Trust model:
- Anyone can call Score() (read-only, no side effects)
- Only admin can train corpus, modify patterns/keywords/blocklist
- Only registered realms can record reputation (RecordAccepted/Flag/Ban)

Reputation tracking:

The realm stores lightweight counters per address (no content):
- TotalAccepted: content accepted
- FlaggedCount: content flagged by mods
- BanCount: times banned (permanent)

Recording workflow:

Score() doesn't update reputation automatically. You have to record decisions explicitly:

```gno
result := antispamr.Score(author, content, rate, rep, nil, nil, nil, nil, engine.EarlyExitDisabled)

if result.Total < engine.ThresholdHide {
    antispamr.RecordAccepted(cross, author)
} else if moderatorFlagged {
    antispamr.RecordFlag(cross, author)
}

if adminBanned {
    antispamr.RecordBan(cross, author)
}
```

Why manual? Gives you full control, prevents false positives from poisoning shared data, allows human review before changing reputation.

## Gas optimization

Three ways to control cost:

1. Input cap: content truncated at 4096 bytes automatically

2. Nil state: pass nil for containers you don't want to use:

```gno
// Cheap rules only
result := antispam.Score(antispam.ScoreInput{Content: content, Rate: rate, Reputation: rep})

// Full state
result := antispam.Score(antispam.ScoreInput{
    Content:      content,
    Rate:         rate,
    Reputation:   rep,
    Corpus:       corpus,
    Fingerprints: fps,
    Blocklist:    bl,
    Keywords:     kw,
})
```

3. Early exit: stop when threshold reached:

```gno
// Stop at 8
result := antispam.Score(antispam.ScoreInput{
    Content:      content,
    Rate:         rate,
    Reputation:   rep,
    Corpus:       corpus,
    Fingerprints: fps,
    Blocklist:    bl,
    Keywords:     kw,
    EarlyExitAt:  antispam.ThresholdReject,
})

// Full score
result := antispam.Score(antispam.ScoreInput{
    Content:      content,
    Rate:         rate,
    Reputation:   rep,
    Corpus:       corpus,
    Fingerprints: fps,
    Blocklist:    bl,
    Keywords:     kw,
    EarlyExitAt:  antispam.EarlyExitDisabled,
})
```

User-based tiering:

```gno
rep := antispamr.GetReputation(author)

if rep.TotalAccepted > 10 && rep.FlaggedCount == 0 && chainData.AccountAgeDays > 30 {
    // Established user: cheap rules only
    result = antispamr.Score(author, content, rate, chainData,
        nil, nil, nil, nil, engine.ThresholdReject)
} else {
    // New/suspicious: full pipeline
    result = antispamr.Score(author, content, rate, chainData,
        nil, nil, nil, nil, engine.EarlyExitDisabled)
}
```

## Auto-training safety

Use TrainingGuard to prevent feedback loops when auto-training:

```gno
var (
    corpus = antispam.NewCorpus()
    guard  = antispam.NewTrainingGuard()
)

func ModerateContent(author, content string) {
    score := antispam.Score(antispam.ScoreInput{...})

    if score.Total >= antispam.ThresholdReject {
        if guard.ShouldTrain(score) {
            corpus.Train(content, true)
            guard.RecordTrain()
        }
    }
}

func AdminResetGuard(_ realm) {
    assertAdmin()
    guard.Reset()
}
```

TrainingGuard prevents:
- Training on borderline content (default min score: 10)
- Single-rule bias (default min rules: 3)
- Runaway training (default max trains: 500 until admin review)

Without it, you get feedback loops where false positives train the corpus, causing more false positives, causing more training.

Use NewTrainingGuardCustom(minScore, minRules, maxTrains) to adjust.

## Deployment

State containers start empty. For large datasets, load them post-deploy to avoid storage deposit limits:

```gno
func init() {
    adminAddr = runtime.OriginCaller()
    corpus   = antispam.NewCorpus()
    fps      = antispam.NewFingerprintStore()
    bl       = antispam.NewBlocklist()
    keywords = antispam.NewKeywordDict()
}

func AdminLoadDefaults(_ realm) {
    assertAdmin()
    bl.AddPattern(defaultPattern)
    keywords.BulkAdd(defaultKeywords)
}
```

## Examples

Check the filetest files for real scenarios:

- z1_comprehensive: all 19 rules across 8 scenarios
- z2_multi_user: different reputation profiles
- z3_bayes_training: corpus training effects
- z4_fingerprint: duplicate detection
- z5_blocklist: blocking and allowlist
- z6_spam_evolution: legitimate user going rogue
- z7_unicode_abuse: zalgo, invisible chars, homoglyphs
- z8_early_exit: gas optimization
- z9_pattern_categories: all 21 regex patterns
- z10_size_comparison: scoring at different content lengths

In r/gnoland/antispam:
- z1_score_demo: basic scoring + reputation recording
- z2_reputation_lifecycle: how reputation affects scores
