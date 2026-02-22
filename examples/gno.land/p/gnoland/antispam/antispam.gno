// Package antispam provides a multi-rule spam scoring engine for Gno realms.
//
// It combines content heuristics, rate limiting, reputation scoring,
// Bayesian filtering, duplicate detection, keyword dictionary,
// and address/pattern blocklists into a single weighted score.
//
// The package provides pure scoring functions and state helpers
// (Corpus, FingerprintStore, Blocklist, KeywordDict) that
// are owned and persisted by the calling realm.
//
// Usage:
//
//	var (
//	    corpus = antispam.NewCorpus()
//	    fps    = antispam.NewFingerprintStore()
//	    bl     = antispam.NewBlocklist()
//	    kw     = antispam.NewKeywordDict()
//	)
//
//	result := antispam.Score(antispam.ScoreInput{
//	    Author:  author,
//	    Content: content,
//	    Rate:    rate,
//	    Rep:     rep,
//	    Corpus:  corpus,
//	    Fps:     fps,
//	    Bl:      bl,
//	    Dict:    kw,
//	})
//	if result.Total >= ThresholdHide {
//	    // hide the post
//	}
//
// For gas optimization, set EarlyExitAt to a positive threshold
// (e.g. ThresholdReject) to skip expensive rules when cheap rules
// already exceed it. Leave at EarlyExitDisabled for complete scoring.
//
// Check the filetest.gno files for examples of how to use the package
// and the scenarios it covers.
package antispam

// ---- Scoring thresholds (recommended defaults, callers can use their own)

const (
	ThresholdHide   = 5
	ThresholdReject = 8

	// EarlyExitDisabled means all rules are evaluated (no early exit).
	// Pass this as ScoreInput.EarlyExitAt for a complete score.
	EarlyExitDisabled = 0

	// MaxInputLength caps content length to prevent gas abuse via very long inputs.
	// Content beyond this limit is truncated before scoring.
	MaxInputLength = 4096
)

// ---- Rule weights (centralized for easy tuning)
//
// Each rule returns one of these weights when triggered.
// The final score is the sum of all triggered rules.
//
//	Rule              Weight  Trigger
//	BLOCKED_ADDRESS      99   Address in blocklist
//	BLOCKED_PATTERN       5   Content matches regex pattern
//	NEAR_DUPLICATE        4   MinHash similarity to known spam
//	RATE_BURST            4   Posting frequency exceeds limit
//	ALL_CAPS              2   >50% uppercase letters
//	SHORT_WITH_LINK       3   Body <=30 chars with a URL
//	BAYES_SPAM            3   Bayesian classifier: spam-heavy tokens
//	BAD_REPUTATION        3   High flag ratio or ban history
//	KEYWORD_SPAM          3   Multiple spam keywords (co-occurrence)
//	LINK_HEAVY            2   >3 URLs in content
//	NEW_ACCOUNT           2   Account <1 day old with no posts
//	ZALGO_TEXT            3   Excessive combining diacritical marks (zalgo text)
//	INVISIBLE_CHARS       2   Zero-width or directional override characters
//	HOMOGLYPH_MIX        2   Mixed unicode scripts in single word (e.g. latin + cyrillic)
//	BANNED_BEFORE       1-3   +1 per past ban, capped at 3
//	EXCESSIVE_PUNCT       1   >20% punctuation characters
//	NO_USERNAME           1   No registered username
//	LOW_BALANCE           1   Balance < 1000 GNOT

const (
	WeightBlockedAddress = 99
	WeightBlockedPattern = 5
	WeightNearDuplicate  = 4
	WeightRateBurst      = 4
	WeightAllCaps        = 2
	WeightShortWithLink  = 3
	WeightBayesSpam      = 3
	WeightBadReputation  = 3
	WeightKeyword        = 3
	WeightZalgoText      = 3
	WeightLinkHeavy      = 2
	WeightNewAccount     = 2
	WeightInvisibleChars = 2
	WeightHomoglyphMix   = 2
	WeightExcessivePunct = 1
	WeightNoUsername     = 1
	WeightLowBalance     = 1
)

// ---- Rule names (used in SpamScore.Triggered and for programmatic checks)

const (
	RuleBlockedAddress = "BLOCKED_ADDRESS"
	RuleBlockedPattern = "BLOCKED_PATTERN"
	RuleNearDuplicate  = "NEAR_DUPLICATE"
	RuleRateBurst      = "RATE_BURST"
	RuleAllCaps        = "ALL_CAPS"
	RuleShortWithLink  = "SHORT_WITH_LINK"
	RuleBayesSpam      = "BAYES_SPAM"
	RuleBadReputation  = "BAD_REPUTATION"
	RuleKeywordSpam    = "KEYWORD_SPAM"
	RuleLinkHeavy      = "LINK_HEAVY"
	RuleNewAccount     = "NEW_ACCOUNT"
	RuleZalgoText      = "ZALGO_TEXT"
	RuleInvisibleChars = "INVISIBLE_CHARS"
	RuleHomoglyphMix   = "HOMOGLYPH_MIX"
	RuleExcessivePunct = "EXCESSIVE_PUNCT"
	RuleNoUsername     = "NO_USERNAME"
	RuleLowBalance     = "LOW_BALANCE"
	RuleBannedBefore   = "BANNED_BEFORE"
)

// RuleHit represents a single triggered anti-spam rule with its score.
type RuleHit struct {
	Score int
	Rule  string
}

// ScoreInput holds all parameters for spam scoring.
// State fields (Corpus, Fps, Bl, Dict) can be nil if not available.
//
// EarlyExitAt controls gas optimization: when set to a positive value
// (e.g. ThresholdReject), expensive rules (regex, tokenization, Bayes,
// keywords, fingerprints) are skipped once cheap rules reach that score.
// Use EarlyExitDisabled (the zero-value default) to evaluate all rules
// and get a complete score - required for multi-level decisions
// (hide/reject/ban).
//
// NOTE: Zero-value fields penalize. With Rep at zero-value, NEW_ACCOUNT(2) +
// NO_USERNAME(1) + LOW_BALANCE(1) = 4 points. Callers must fill Rep accurately.
type ScoreInput struct {
	Author      string
	Content     string
	Rate        RateState
	Rep         ReputationData
	Corpus      *Corpus
	Fps         *FingerprintStore
	Bl          *Blocklist
	Dict        *KeywordDict
	EarlyExitAt int // >0: skip expensive rules at this score; EarlyExitDisabled(0): evaluate all
}

// SpamScore is the result of scoring content.
type SpamScore struct {
	Total     int      // sum of all triggered rule weights
	Triggered []string // names of rules that fired
	TopRule   string   // highest-scoring rule
	TopScore  int      // score of the highest-scoring rule
}

// Score evaluates content against all anti-spam rules.
//
// Rules are ordered in two phases for gas efficiency:
//
//	Phase 1 (cheap): address blocklist, content heuristics, rate, reputation.
//	  These are O(1) lookups and single-pass character scans.
//
//	Phase 2 (expensive): regex patterns, tokenization, Bayes, keywords, fingerprints.
//	  These involve regex matching, AVL traversals, and MinHash computation.
//
// When EarlyExitAt is set to a positive threshold, phase 2 is skipped
// if the score from phase 1 already reaches it. When EarlyExitAt is
// EarlyExitDisabled (0, the default), all rules are always evaluated.
func Score(in ScoreInput) SpamScore {
	// Truncate oversized input to cap gas cost.
	if len(in.Content) > MaxInputLength {
		in.Content = in.Content[:MaxInputLength]
	}

	result := SpamScore{}

	// Allowlist check: bypass all scoring
	if in.Bl != nil && in.Bl.IsAllowed(in.Author) {
		return result
	}

	add := func(score int, rule string) {
		if score <= 0 {
			return
		}
		result.Total += score
		result.Triggered = append(result.Triggered, rule)
		if score > result.TopScore {
			result.TopScore = score
			result.TopRule = rule
		}
	}

	addHits := func(hits []RuleHit) {
		for _, h := range hits {
			add(h.Score, h.Rule)
		}
	}

	earlyExit := func() bool {
		return in.EarlyExitAt > 0 && result.Total >= in.EarlyExitAt
	}

	// ---- Phase 1: cheap rules (O(1) lookups, linear scans)

	// Blocked address check (AVL lookup, no regex)
	if in.Bl != nil && in.Bl.IsBlocked(in.Author) {
		add(WeightBlockedAddress, RuleBlockedAddress)
		return result
	}

	// Content heuristics + unicode abuse (single-pass character scan)
	addHits(scoreContentAll(in.Content))

	// Rate limiting (arithmetic comparison)
	if s, r := ScoreRate(in.Rate); s > 0 {
		add(s, r)
	}

	// Reputation (arithmetic comparisons)
	addHits(ScoreReputation(in.Rep))

	if earlyExit() {
		return result
	}

	// ---- Phase 2: expensive rules (regex, tokenization, AVL traversal, MinHash)

	// Regex pattern matching (single combined regex scan)
	if in.Bl != nil && in.Bl.MatchesPattern(in.Content) {
		add(WeightBlockedPattern, RuleBlockedPattern)
	}

	if earlyExit() {
		return result
	}

	// Tokenize only when needed by Bayes or keyword rules.
	var tokens []string
	if in.Corpus != nil || in.Dict != nil {
		tokens = dedup(Tokenize(in.Content))
	}

	// Bayesian filter
	if s, r := scoreBayesTokens(tokens, in.Corpus); s > 0 {
		add(s, r)
	}

	if earlyExit() {
		return result
	}

	// Keyword dictionary
	if s, r := scoreKeywordsTokens(tokens, in.Dict); s > 0 {
		add(s, r)
	}

	if earlyExit() {
		return result
	}

	// Fingerprint near-duplicate detection (MinHash computation)
	if s, r := ScoreFingerprint(in.Content, in.Fps); s > 0 {
		add(s, r)
	}

	return result
}
