package antispam

// RateState holds caller-provided posting rate data for an author.
type RateState struct {
	PostCount     int
	WindowSeconds int64
}

const (
	// rateMaxPerHour is the maximum posts per hour before scoring.
	// Above this -> RATE_BURST triggers.
	rateMaxPerHour = 10
)

// ScoreRate evaluates posting frequency.
// It computes a posts-per-hour rate from the caller-provided RateState.
// Returns score and rule name.
func ScoreRate(rate RateState) (int, string) {
	if rate.PostCount <= 0 {
		return 0, ""
	}

	const maxSafePostCount = (1 << 50) / 3600
	if rate.PostCount > maxSafePostCount {
		return WeightRateBurst, "RATE_BURST"
	}

	var postsPerHour int
	if rate.WindowSeconds <= 0 {
		postsPerHour = rate.PostCount * 3600
	} else {
		postsPerHour = int(int64(rate.PostCount) * 3600 / rate.WindowSeconds)
	}

	if postsPerHour <= rateMaxPerHour {
		return 0, ""
	}

	if postsPerHour <= rateMaxPerHour*2 {
		return WeightRateBurst / 2, "RATE_BURST"
	}
	return WeightRateBurst, "RATE_BURST"
}
