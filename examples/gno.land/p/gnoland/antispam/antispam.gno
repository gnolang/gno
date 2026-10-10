// Package antispam provides spam scoring for Gno realms.
// Combines content checks, rate limits, reputation, Bayes filter, duplicates, keywords, and blocklists.
// See README for usage and examples.
package antispam

// Scoring thresholds (recommended defaults, callers can use their own)

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

// Rule weights (centralized for easy tuning)
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
//	REPEATED_CHARS        2   4+ consecutive identical characters
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
	WeightRepeatedChars  = 2
	WeightExcessivePunct = 1
	WeightNoUsername     = 1
	WeightLowBalance     = 1
)

// Rule names (used in SpamScore.Triggered and for programmatic checks)

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
	RuleRepeatedChars  = "REPEATED_CHARS"
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
// State fields (Corpus, Fingerprints, Blocklist, Keywords) can be nil if not available.
//
// EarlyExitAt controls gas optimization: when set to a positive value
// (e.g. ThresholdReject), expensive rules (regex, tokenization, Bayes,
// keywords, fingerprints) are skipped once cheap rules reach that score.
// Use EarlyExitDisabled (the zero-value default) to evaluate all rules
// and get a complete score - required for multi-level decisions
// (hide/reject/ban).
//
// NOTE: Zero-value fields penalize. With Reputation at zero-value,
// NEW_ACCOUNT(2) + NO_USERNAME(1) + LOW_BALANCE(1) = 4 points.
// Callers must fill Reputation accurately.
type ScoreInput struct {
	Author       string
	Content      string
	Rate         RateState
	Reputation   ReputationData
	Corpus       *Corpus
	Fingerprints *FingerprintStore
	Blocklist    *Blocklist
	Keywords     *KeywordDict
	EarlyExitAt  int // >0: skip expensive rules at this score; EarlyExitDisabled(0): evaluate all
}

// SpamScore is the result of scoring content.
type SpamScore struct {
	// Total is the sum of all triggered rule weights.
	Total int

	// Triggered contains the names of rules that fired.
	Triggered []string

	// TopRule is the name of the highest-scoring rule.
	TopRule string

	// TopScore is the score of the highest-scoring rule.
	TopScore int
}

// Score evaluates content against all anti-spam rules.
//
// Rules are ordered by ascending cost with earlyExit checks between
// each group, so expensive work is skipped as soon as possible:
//
//  1. Address blocklist      O(1) AVL lookup   - immediate return
//  2. Rate limiting          O(1) arithmetic
//  3. Reputation             O(1) arithmetic
//  4. Content heuristics     O(n) single-pass character scan
//  5. Regex pattern matching O(n) compiled regex
//  6. Tokenization + Bayes   O(n) + AVL traversals
//  7. Keyword dictionary     AVL traversal (reuses tokens from 6)
//  8. Fingerprint/MinHash    O(n) hashing
//
// When EarlyExitAt is set to a positive threshold, each earlyExit
// check can short-circuit before reaching costlier rules.
// Use EarlyExitDisabled (0, the default) to evaluate all rules.
func Score(in ScoreInput) SpamScore {
	// Truncate oversized input to cap gas cost.
	if len(in.Content) > MaxInputLength {
		in.Content = in.Content[:MaxInputLength]
	}

	result := SpamScore{}

	// Allowlist check: bypass all scoring
	if in.Blocklist != nil && in.Blocklist.IsAllowed(in.Author) {
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

	// O(1) rules: lookups and arithmetic

	// Blocked address (AVL lookup, instant kill)
	if in.Blocklist != nil && in.Blocklist.IsBlocked(in.Author) {
		add(WeightBlockedAddress, RuleBlockedAddress)
		return result
	}

	// Rate limiting (arithmetic comparison)
	if s, r := ScoreRate(in.Rate); s > 0 {
		add(s, r)
	}

	// Reputation (arithmetic comparisons)
	addHits(ScoreReputation(in.Reputation))

	if earlyExit() {
		return result
	}

	// O(n) single-pass: content heuristics + unicode abuse

	addHits(scoreContentAll(in.Content))

	if earlyExit() {
		return result
	}

	// O(n) regex: blocklist pattern matching

	if in.Blocklist != nil && in.Blocklist.MatchesPattern(in.Content) {
		add(WeightBlockedPattern, RuleBlockedPattern)
	}

	if earlyExit() {
		return result
	}

	// O(n) tokenization + AVL: Bayes and keyword rules

	// Tokenize only when needed by Bayes or keyword rules.
	var tokens []string
	if in.Corpus != nil || in.Keywords != nil {
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
	if s, r := scoreKeywordsTokens(tokens, in.Keywords); s > 0 {
		add(s, r)
	}

	if earlyExit() {
		return result
	}

	// O(n) hashing: fingerprint near-duplicate detection

	if s, r := ScoreFingerprint(in.Content, in.Fingerprints); s > 0 {
		add(s, r)
	}

	return result
}
