// Package antispam provides a multi-rule spam scoring engine for Gno realms.
//
// It combines content heuristics, rate limiting, reputation scoring,
// Bayesian filtering, MinHash fingerprinting, and address/pattern blocklists
// into a single weighted score. Each algorithm lives in its own file
// for independent audit.
//
// The package provides both pure scoring functions and AVL-tree-backed
// state helpers (Corpus, FingerprintStore, Blocklist) that are owned
// and persisted by the calling realm.
//
// Usage:
//
//	var (
//	    corpus = antispam.NewCorpus()
//	    fps    = antispam.NewFingerprintStore()
//	    bl     = antispam.NewBlocklist()
//	)
//
//	result := antispam.Score(author, content, rate, rep, corpus, fps, bl)
//	if result.Total >= ThresholdHide {
//	    // hide the post
//	}
package antispam

// ---- Thresholds (recommended defaults, callers can use their own)

const (
	ThresholdHide   = 5
	ThresholdReject = 8
)

// SpamScore is the result of scoring content.
type SpamScore struct {
	Total     int      // weighted sum of all rules
	Triggered []string // names of rules that scored > 0
	TopRule   string   // rule with highest individual score
	TopScore  int      // score of the top rule
}

// Score evaluates content against all anti-spam rules.
// The content parameter is the full text to analyze (realm-agnostic).
// State parameters (corpus, fps, bl) can be nil if not available.
// Returns a SpamScore with the combined result.
func Score(
	author, content string,
	rate RateState,
	rep ReputationData,
	corpus *Corpus,
	fps *FingerprintStore,
	bl *Blocklist,
) SpamScore {
	result := SpamScore{}

	// Allowlist check: bypass all scoring
	if bl != nil && bl.IsAllowed(author) {
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

	// Blocklist (address + patterns)
	if s, r := ScoreBlocklist(author, content, bl); s > 0 {
		add(s, r)
		// If address is blocked, no need to check further
		if r == "BLOCKED_ADDRESS" {
			return result
		}
	}

	// Content heuristics
	if s, r := ScoreContent(content); s > 0 {
		add(s, r)
	}

	// Rate limiting
	if s, r := ScoreRate(rate); s > 0 {
		add(s, r)
	}

	// Reputation
	if s, r := ScoreReputation(rep); s > 0 {
		add(s, r)
	}

	// Bayesian filter
	if s, r := ScoreBayes(content, corpus); s > 0 {
		add(s, r)
	}

	// Fingerprint near-duplicate detection
	if s, r := ScoreFingerprint(content, fps); s > 0 {
		add(s, r)
	}

	return result
}
