package antispam

// ReputationData holds caller-provided account context for an author.
type ReputationData struct {
	AccountAgeDays int
	Balance        int64
	FlaggedCount   int
	TotalPosts     int
	HasUsername    bool
	BanCount       int
}

const (
	repMinAgeDays   = 1          // accounts younger than this are "new"
	repMinBalance   = 1000000000 // 1000 GNOT in ugnot
	repFlagRatioPct = 30         // flagged/total > this% → bad reputation
	repBanPenalty   = 1          // additional score per past ban
)

// ScoreReputation evaluates account reputation signals.
// Returns individual RuleHit entries for each triggered rule.
func ScoreReputation(rep ReputationData) []RuleHit {
	var hits []RuleHit

	// New account: young + no posts
	if rep.AccountAgeDays < repMinAgeDays && rep.TotalPosts <= 0 {
		hits = append(hits, RuleHit{WeightNewAccount, "NEW_ACCOUNT"})
	}

	// No registered username
	if !rep.HasUsername {
		hits = append(hits, RuleHit{WeightNoUsername, "NO_USERNAME"})
	}

	// Low balance
	if rep.Balance < repMinBalance {
		hits = append(hits, RuleHit{WeightLowBalance, "LOW_BALANCE"})
	}

	// Bad flag ratio: flaggedCount / totalPosts > threshold
	// Rewritten as: flaggedCount * 100 > threshold * totalPosts to avoid overflow
	// on the left side, we guard large FlaggedCount.
	if rep.TotalPosts > 0 && rep.FlaggedCount > 0 {
		// Guard: if FlaggedCount exceeds TotalPosts, ratio is >100% → always bad
		flagRatio := 100
		if rep.FlaggedCount <= rep.TotalPosts {
			flagRatio = rep.FlaggedCount * 100 / rep.TotalPosts
		}
		if flagRatio > repFlagRatioPct {
			hits = append(hits, RuleHit{WeightBadReputation, "BAD_REPUTATION"})
		}
	} else if rep.TotalPosts == 0 && rep.FlaggedCount > 0 {
		// Flags but no visible posts (all deleted?) → suspicious
		hits = append(hits, RuleHit{WeightBadReputation, "BAD_REPUTATION"})
	}

	// Ban history penalty
	if rep.BanCount > 0 {
		penalty := rep.BanCount * repBanPenalty
		if penalty > WeightBadReputation {
			penalty = WeightBadReputation
		}
		hits = append(hits, RuleHit{penalty, "BANNED_BEFORE"})
	}

	return hits
}
