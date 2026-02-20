package antispam

import (
	"testing"
)

func TestScoreReputation(t *testing.T) {
	tests := []struct {
		name     string
		rep      ReputationData
		wantMin  int
		wantMax  int
		wantRule string
		wantZero bool
	}{
		// ---- Success: good reputation scores zero ----

		{
			name: "success: established user",
			rep: ReputationData{
				AccountAgeDays: 90,
				Balance:        10000000000,
				FlaggedCount:   0,
				TotalPosts:     50,
				HasUsername:    true,
				BanCount:       0,
			},
			wantZero: true,
		},
		{
			name: "success: moderate user",
			rep: ReputationData{
				AccountAgeDays: 30,
				Balance:        5000000000,
				FlaggedCount:   0,
				TotalPosts:     10,
				HasUsername:    true,
				BanCount:       0,
			},
			wantZero: true,
		},
		{
			name: "success: new but legitimate user",
			rep: ReputationData{
				AccountAgeDays: 7,
				Balance:        3000000000,
				FlaggedCount:   0,
				TotalPosts:     2,
				HasUsername:    true,
				BanCount:       0,
			},
			wantZero: true,
		},

		// ---- Suspicious: new account signals ----

		{
			name: "suspicious: brand new account no username",
			rep: ReputationData{
				AccountAgeDays: 0,
				Balance:        3000000000,
				FlaggedCount:   0,
				TotalPosts:     0,
				HasUsername:    false,
				BanCount:       0,
			},
			wantMin:  2,
			wantRule: "NEW_ACCOUNT",
		},
		{
			name: "suspicious: new account low balance",
			rep: ReputationData{
				AccountAgeDays: 0,
				Balance:        100000000,
				FlaggedCount:   0,
				TotalPosts:     0,
				HasUsername:    false,
				BanCount:       0,
			},
			wantMin: 3,
		},
		{
			name: "suspicious: no username, old account",
			rep: ReputationData{
				AccountAgeDays: 60,
				Balance:        5000000000,
				FlaggedCount:   0,
				TotalPosts:     5,
				HasUsername:    false,
				BanCount:       0,
			},
			wantMin:  1,
			wantRule: "NO_USERNAME",
		},

		// ---- Bad reputation: flags and bans ----

		{
			name: "bad: many flags",
			rep: ReputationData{
				AccountAgeDays: 30,
				Balance:        5000000000,
				FlaggedCount:   10,
				TotalPosts:     15,
				HasUsername:    true,
				BanCount:       0,
			},
			wantMin:  3,
			wantRule: "BAD_REPUTATION",
		},
		{
			name: "bad: previously banned with high flag ratio",
			rep: ReputationData{
				AccountAgeDays: 60,
				Balance:        5000000000,
				FlaggedCount:   10,
				TotalPosts:     20,
				HasUsername:    true,
				BanCount:       1,
			},
			wantMin: 4,
		},
		{
			name: "bad: multiple bans",
			rep: ReputationData{
				AccountAgeDays: 90,
				Balance:        5000000000,
				FlaggedCount:   20,
				TotalPosts:     30,
				HasUsername:    true,
				BanCount:       3,
			},
			wantMin: 3,
		},
		{
			name: "bad: all flags no good posts",
			rep: ReputationData{
				AccountAgeDays: 10,
				Balance:        3000000000,
				FlaggedCount:   5,
				TotalPosts:     5,
				HasUsername:    true,
				BanCount:       0,
			},
			wantMin:  3,
			wantRule: "BAD_REPUTATION",
		},

		// ---- Edge cases ----

		{
			name:     "edge: all zeros",
			rep:      ReputationData{},
			wantMin:  2,
			wantRule: "NEW_ACCOUNT",
		},
		{
			name: "edge: negative balance",
			rep: ReputationData{
				AccountAgeDays: 30,
				Balance:        -1,
				FlaggedCount:   0,
				TotalPosts:     5,
				HasUsername:    true,
				BanCount:       0,
			},
			wantMin:  1,
			wantRule: "LOW_BALANCE",
		},
		{
			name: "edge: more flags than posts",
			rep: ReputationData{
				AccountAgeDays: 30,
				Balance:        5000000000,
				FlaggedCount:   10,
				TotalPosts:     5,
				HasUsername:    true,
				BanCount:       0,
			},
			wantMin:  3,
			wantRule: "BAD_REPUTATION",
		},
		{
			name: "edge: one flag out of many posts is fine",
			rep: ReputationData{
				AccountAgeDays: 90,
				Balance:        10000000000,
				FlaggedCount:   1,
				TotalPosts:     100,
				HasUsername:    true,
				BanCount:       0,
			},
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := ScoreReputation(tt.rep)
			score, rule := hitsTotal(hits)

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
				t.Errorf("expected rule %q, got %q (score: %d)", tt.wantRule, rule, score)
			}
		})
	}
}
