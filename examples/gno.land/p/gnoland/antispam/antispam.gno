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
// Check the filetest.gno files for examples of how to use the package
// and the scenarios it covers.
package antispam

// ---- Scoring thresholds (recommended defaults, callers can use their own)

const (
	ThresholdHide   = 5
	ThresholdReject = 8
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

// RuleHit represents a single triggered anti-spam rule with its score.
type RuleHit struct {
	Score int
	Rule  string
}

// ScoreInput holds all parameters for spam scoring.
// State fields (Corpus, Fps, Bl, Dict) can be nil if not available.
//
// NOTE: Zero-value fields penalize. With Rep at zero-value, NEW_ACCOUNT(2) +
// NO_USERNAME(1) + LOW_BALANCE(1) = 4 points. Callers must fill Rep accurately.
type ScoreInput struct {
	Author  string
	Content string
	Rate    RateState
	Rep     ReputationData
	Corpus  *Corpus
	Fps     *FingerprintStore
	Bl      *Blocklist
	Dict    *KeywordDict
}

// SpamScore is the result of scoring content.
type SpamScore struct {
	Total     int      // sum of all triggered rule weights
	Triggered []string // names of rules that fired
	TopRule   string   // highest-scoring rule
	TopScore  int      // score of the highest-scoring rule
}

// Score evaluates content against all anti-spam rules.
// Returns a SpamScore with the combined result.
func Score(in ScoreInput) SpamScore {
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

	// Blocklist (address + patterns)
	if s, r := ScoreBlocklist(in.Author, in.Content, in.Bl); s > 0 {
		add(s, r)
		// If address is blocked, no need to check further
		if r == "BLOCKED_ADDRESS" {
			return result
		}
	}

	tokens := dedup(Tokenize(in.Content))

	// Content heuristics + unicode abuse
	addHits(scoreContentAll(in.Content))

	// Rate limiting
	if s, r := ScoreRate(in.Rate); s > 0 {
		add(s, r)
	}

	// Reputation
	addHits(ScoreReputation(in.Rep))

	// Bayesian filter
	if s, r := scoreBayesTokens(tokens, in.Corpus); s > 0 {
		add(s, r)
	}

	// Fingerprint near-duplicate detection
	if s, r := ScoreFingerprint(in.Content, in.Fps); s > 0 {
		add(s, r)
	}

	// Keyword dictionary
	if s, r := scoreKeywordsTokens(tokens, in.Dict); s > 0 {
		add(s, r)
	}

	return result
}
