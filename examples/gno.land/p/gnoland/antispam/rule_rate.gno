package antispam

// RateState holds caller-provided posting rate data for an author.
type RateState struct {
	PostCount     int
	WindowSeconds int64
}

const (
	// rateMaxPerHour is the maximum posts per hour before scoring.
	// Above this → RATE_BURST triggers.
	rateMaxPerHour = 10
)

// ScoreRate evaluates posting frequency.
// It computes a posts-per-hour rate from the caller-provided RateState.
// Returns score and rule name.
func ScoreRate(rate RateState) (int, string) {
	if rate.PostCount <= 0 {
		return 0, ""
	}

	// Guard against integer overflow: PostCount * 3600 can overflow int64.
	// Any PostCount above this threshold is absurdly high → definitely spam.
	const maxSafePostCount = (1 << 50) / 3600
	if rate.PostCount > maxSafePostCount {
		return WeightRateBurst, "RATE_BURST"
	}

	// Compute posts per hour.
	// If window is zero or negative, treat as "all posts happened instantly" → spam.
	var postsPerHour int
	if rate.WindowSeconds <= 0 {
		postsPerHour = rate.PostCount * 3600
	} else {
		// postsPerHour = postCount * 3600 / windowSeconds
		postsPerHour = int(int64(rate.PostCount) * 3600 / rate.WindowSeconds)
	}

	if postsPerHour <= rateMaxPerHour {
		return 0, ""
	}

	// Score scales with how far over the limit:
	// 11-20/hr → base weight (2), 21+/hr → double weight (4)
	if postsPerHour <= rateMaxPerHour*2 {
		return WeightRateBurst / 2, "RATE_BURST"
	}
	return WeightRateBurst, "RATE_BURST"
}
