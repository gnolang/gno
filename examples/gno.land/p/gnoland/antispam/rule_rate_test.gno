package antispam

import (
	"testing"
)

func TestScoreRate(t *testing.T) {
	tests := []struct {
		name     string
		rate     RateState
		wantMin  int
		wantMax  int
		wantRule string
		wantZero bool
	}{
		// ---- Success: normal posting rates score zero ----

		{
			name:     "success: zero posts",
			rate:     RateState{PostCount: 0, WindowSeconds: 3600},
			wantZero: true,
		},
		{
			name:     "success: one post in an hour",
			rate:     RateState{PostCount: 1, WindowSeconds: 3600},
			wantZero: true,
		},
		{
			name:     "success: five posts in an hour",
			rate:     RateState{PostCount: 5, WindowSeconds: 3600},
			wantZero: true,
		},
		{
			name:     "success: ten posts in a day",
			rate:     RateState{PostCount: 10, WindowSeconds: 86400},
			wantZero: true,
		},
		{
			name:     "success: moderate activity over a week",
			rate:     RateState{PostCount: 50, WindowSeconds: 604800},
			wantZero: true,
		},

		// ---- Spam: high frequency posting ----

		{
			name:     "spam: 10+ posts in an hour",
			rate:     RateState{PostCount: 11, WindowSeconds: 3600},
			wantMin:  2,
			wantRule: "RATE_BURST",
		},
		{
			name:     "spam: 20 posts in an hour (moderate burst)",
			rate:     RateState{PostCount: 20, WindowSeconds: 3600},
			wantMin:  2,
			wantMax:  2,
			wantRule: "RATE_BURST",
		},
		{
			name:     "spam: 50 posts in an hour (heavy flood)",
			rate:     RateState{PostCount: 50, WindowSeconds: 3600},
			wantMin:  4,
			wantRule: "RATE_BURST",
		},
		{
			name:     "spam: 5 posts in one minute",
			rate:     RateState{PostCount: 5, WindowSeconds: 60},
			wantMin:  4,
			wantRule: "RATE_BURST",
		},
		{
			name:     "spam: 3 posts in 10 seconds",
			rate:     RateState{PostCount: 3, WindowSeconds: 10},
			wantMin:  4,
			wantRule: "RATE_BURST",
		},

		// ---- Edge cases ----

		{
			name:     "edge: zero window seconds",
			rate:     RateState{PostCount: 5, WindowSeconds: 0},
			wantMin:  4,
			wantRule: "RATE_BURST",
		},
		{
			name:     "edge: negative window (treat as zero)",
			rate:     RateState{PostCount: 5, WindowSeconds: -1},
			wantMin:  4,
			wantRule: "RATE_BURST",
		},
		{
			name:     "edge: negative post count",
			rate:     RateState{PostCount: -1, WindowSeconds: 3600},
			wantZero: true,
		},
		{
			name:     "edge: very large window makes rate low",
			rate:     RateState{PostCount: 100, WindowSeconds: 2592000},
			wantZero: true,
		},
		{
			name:     "edge: boundary 10 posts in an hour",
			rate:     RateState{PostCount: 10, WindowSeconds: 3600},
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, rule := ScoreRate(tt.rate)

			if tt.wantZero {
				if score != 0 {
					t.Errorf("expected score 0, got %d (rule: %s)", score, rule)
				}
				return
			}

			if score < tt.wantMin {
				t.Errorf("expected score >= %d, got %d (rule: %s)", tt.wantMin, score, rule)
			}
			if tt.wantMax > 0 && score > tt.wantMax {
				t.Errorf("expected score <= %d, got %d (rule: %s)", tt.wantMax, score, rule)
			}
			if tt.wantRule != "" && rule != tt.wantRule {
				t.Errorf("expected rule %q, got %q", tt.wantRule, rule)
			}
		})
	}
}

// ---- Overflow guard test ----

func TestRateOverflowGuard(t *testing.T) {
	// Extremely high PostCount triggers the overflow guard (maxSafePostCount)
	// instead of causing integer overflow in the PostCount*3600 multiplication.
	score, rule := ScoreRate(RateState{
		PostCount:     1 << 52,
		WindowSeconds: 1,
	})
	if score != WeightRateBurst {
		t.Errorf("expected score %d for overflow PostCount, got %d", WeightRateBurst, score)
	}
	if rule != "RATE_BURST" {
		t.Errorf("expected rule RATE_BURST, got %q", rule)
	}
}

// ---- Rate scaling boundary tests ----

func TestRateScalingBoundary(t *testing.T) {
	// 11 posts/hr → base weight (WeightRateBurst/2 = 2)
	score11, _ := ScoreRate(RateState{PostCount: 11, WindowSeconds: 3600})
	// 20 posts/hr → still base weight (rateMaxPerHour*2 = 20)
	score20, _ := ScoreRate(RateState{PostCount: 20, WindowSeconds: 3600})
	// 21 posts/hr → full weight (WeightRateBurst = 4)
	score21, _ := ScoreRate(RateState{PostCount: 21, WindowSeconds: 3600})

	if score11 != WeightRateBurst/2 {
		t.Errorf("11 posts/hr: expected score %d, got %d", WeightRateBurst/2, score11)
	}
	if score20 != WeightRateBurst/2 {
		t.Errorf("20 posts/hr: expected score %d, got %d", WeightRateBurst/2, score20)
	}
	if score21 != WeightRateBurst {
		t.Errorf("21 posts/hr: expected score %d, got %d", WeightRateBurst, score21)
	}
}
