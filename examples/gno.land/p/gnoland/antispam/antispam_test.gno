package antispam

import (
	"testing"
)

func TestScore(t *testing.T) {
	// Build shared state for tests
	emptyCorpus := NewCorpus()
	emptyFps := NewFingerprintStore()
	emptyBl := NewBlocklist()

	trainedCorpus := NewCorpus()
	trainedCorpus.Train("free airdrop bonus click claim tokens", true)
	trainedCorpus.Train("free money send gnot prize winner", true)
	trainedCorpus.Train("casino bonus jackpot free spin win", true)
	trainedCorpus.Train("airdrop claim free tokens hurry limited", true)
	trainedCorpus.Train("click here free prize bonus amazing", true)
	trainedCorpus.Train("governance proposal voting community discussion", false)
	trainedCorpus.Train("validator staking delegation reward system", false)
	trainedCorpus.Train("development update release version changelog", false)
	trainedCorpus.Train("question about using gnokey transaction", false)
	trainedCorpus.Train("review this pull request changes code", false)
	trainedCorpus.Train("interesting discussion about protocol design", false)
	trainedCorpus.Train("thanks for the helpful explanation today", false)

	spamFps := NewFingerprintStore()
	spamFps.Add("free airdrop bonus click claim prize tokens hurry")

	blocklist := NewBlocklist()
	blocklist.BlockAddress("g1spammer")
	blocklist.AllowAddress("g1admin")
	blocklist.AddPattern(`(?i)send\s+\d+\s*gnot`)

	goodRep := ReputationData{
		AccountAgeDays: 90,
		Balance:        10000000000,
		FlaggedCount:   0,
		TotalPosts:     50,
		HasUsername:     true,
		BanCount:       0,
	}

	newAccountRep := ReputationData{
		AccountAgeDays: 0,
		Balance:        3000000000,
		FlaggedCount:   0,
		TotalPosts:     0,
		HasUsername:     false,
		BanCount:       0,
	}

	badRep := ReputationData{
		AccountAgeDays: 30,
		Balance:        3000000000,
		FlaggedCount:   15,
		TotalPosts:     20,
		HasUsername:     true,
		BanCount:       2,
	}

	zeroRate := RateState{PostCount: 0, WindowSeconds: 3600}
	normalRate := RateState{PostCount: 5, WindowSeconds: 3600}
	burstRate := RateState{PostCount: 30, WindowSeconds: 3600}

	tests := []struct {
		name     string
		input    ScoreInput
		wantMin  int
		wantMax  int
		wantZero bool
	}{
		// ---- Clean content, good user ----

		{
			name: "clean: normal post from good user",
			input: ScoreInput{
				Author: "g1gooduser", Content: "I think we should consider adjusting the staking rewards for validators.",
				Rate: normalRate, Rep: goodRep, Corpus: emptyCorpus, Fps: emptyFps, Bl: emptyBl,
			},
			wantZero: true,
		},
		{
			name: "clean: all nil state",
			input: ScoreInput{
				Author: "g1user", Content: "Just saying hello to the community.",
				Rate: zeroRate, Rep: goodRep,
			},
			wantZero: true,
		},

		// ---- Allowlisted user bypasses everything ----

		{
			name: "allowed: admin posts spam-like content",
			input: ScoreInput{
				Author: "g1admin", Content: "FREE AIRDROP BONUS for everyone!!! https://a.com https://b.com https://c.com https://d.com",
				Rate: burstRate, Rep: badRep, Corpus: trainedCorpus, Fps: spamFps, Bl: blocklist,
			},
			wantZero: true,
		},

		// ---- Blocked user ----

		{
			name: "blocked: known spammer",
			input: ScoreInput{
				Author: "g1spammer", Content: "Totally normal message.",
				Rate: normalRate, Rep: goodRep, Corpus: emptyCorpus, Fps: emptyFps, Bl: blocklist,
			},
			wantMin: 99,
		},

		// ---- Content spam ----

		{
			name: "spam: all caps flood",
			input: ScoreInput{
				Author: "g1user", Content: "THIS IS ALL CAPS SPAM MESSAGE THAT SHOULD BE DETECTED BY CONTENT RULES EASILY",
				Rate: normalRate, Rep: goodRep, Corpus: emptyCorpus, Fps: emptyFps, Bl: emptyBl,
			},
			wantMin: 3,
		},

		// ---- Rate burst ----

		{
			name: "spam: rate burst from good user",
			input: ScoreInput{
				Author: "g1user", Content: "Normal content but posted too fast.",
				Rate: burstRate, Rep: goodRep, Corpus: emptyCorpus, Fps: emptyFps, Bl: emptyBl,
			},
			wantMin: 4,
		},

		// ---- Bad reputation ----

		{
			name: "suspicious: new account no username",
			input: ScoreInput{
				Author: "g1newbie", Content: "Hello everyone I just joined.",
				Rate: zeroRate, Rep: newAccountRep, Corpus: emptyCorpus, Fps: emptyFps, Bl: emptyBl,
			},
			wantMin: 2,
		},

		// ---- Pattern match ----

		{
			name: "spam: pattern match crypto scam",
			input: ScoreInput{
				Author: "g1user", Content: "Send 1000 GNOT to this address and get 5000 back guaranteed",
				Rate: normalRate, Rep: goodRep, Corpus: emptyCorpus, Fps: emptyFps, Bl: blocklist,
			},
			wantMin: 5,
		},

		// ---- Near duplicate ----

		{
			name: "spam: near duplicate of known spam",
			input: ScoreInput{
				Author: "g1user", Content: "free airdrop bonus click claim prize tokens quick",
				Rate: normalRate, Rep: goodRep, Corpus: emptyCorpus, Fps: spamFps, Bl: emptyBl,
			},
			wantMin: 4,
		},

		// ---- Multi-signal stacking ----

		{
			name: "combined: new account + caps + rate burst",
			input: ScoreInput{
				Author: "g1newspammer", Content: "BUY NOW THIS IS THE BEST DEAL EVER YOU WILL NOT REGRET THIS AMAZING OPPORTUNITY",
				Rate: burstRate, Rep: newAccountRep, Corpus: emptyCorpus, Fps: emptyFps, Bl: emptyBl,
			},
			wantMin: 9,
		},
		{
			name: "combined: bad rep + pattern + bayes",
			input: ScoreInput{
				Author: "g1baduser", Content: "Send 5000 GNOT and get free airdrop bonus click claim prize tokens",
				Rate: normalRate, Rep: badRep, Corpus: trainedCorpus, Fps: emptyFps, Bl: blocklist,
			},
			wantMin: 10,
		},

		// ---- Edge cases ----

		{
			name:    "edge: empty everything",
			input:   ScoreInput{},
			wantMin: 2, // NEW_ACCOUNT + NO_USERNAME
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Score(tt.input)

			if tt.wantZero {
				if result.Total != 0 {
					t.Errorf("expected total 0, got %d (triggered: %v)", result.Total, result.Triggered)
				}
				return
			}

			if result.Total < tt.wantMin {
				t.Errorf("expected total >= %d, got %d (triggered: %v)", tt.wantMin, result.Total, result.Triggered)
			}
			if tt.wantMax > 0 && result.Total > tt.wantMax {
				t.Errorf("expected total <= %d, got %d (triggered: %v)", tt.wantMax, result.Total, result.Triggered)
			}
		})
	}
}

// ---- False positive test suite ----

func TestFalsePositives(t *testing.T) {
	// Good reputation, normal rate, trained corpus, no blocklist
	goodRep := ReputationData{
		AccountAgeDays: 60,
		Balance:        5000000000,
		FlaggedCount:   0,
		TotalPosts:     20,
		HasUsername:     true,
		BanCount:       0,
	}
	normalRate := RateState{PostCount: 3, WindowSeconds: 3600}

	corpus := NewCorpus()
	corpus.Train("free airdrop bonus click claim tokens", true)
	corpus.Train("governance proposal voting community discussion", false)
	corpus.Train("validator staking delegation reward system", false)
	corpus.Train("development update release version changelog", false)
	corpus.Train("question about using gnokey transaction", false)
	corpus.Train("review this pull request changes code", false)

	fps := NewFingerprintStore()
	fps.Add("free airdrop bonus click claim prize tokens hurry")

	bl := NewBlocklist()

	// These messages MUST NOT be flagged as spam
	legitimateMessages := []struct {
		name    string
		content string
	}{
		{"governance discussion", "I think we should consider adjusting the staking rewards for validators from 80% to 70% to incentivize broader participation."},
		{"technical question", "I'm trying to send a transaction with gnokey but getting an error. Can someone help me understand the command syntax?"},
		{"code review", "Looking at the PR, I think the approach is correct but we should add more test coverage for edge cases."},
		{"community announcement", "We're organizing a Gno developer meetup in Paris next month. Check the details at https://meetup.example.com and RSVP."},
		{"bug report", "When I try to call a realm function with invalid arguments, the transaction fails but no error message is shown. Steps to reproduce: 1. Open gnokey 2. Run the command 3. Notice no output"},
		{"feature request", "It would be great if gnoweb had a dark mode option. Many developers work at night and the bright white background is hard on the eyes."},
		{"disagreement", "While I respect the proposal, I think there are better alternatives. The current approach has significant drawbacks that we should discuss."},
		{"emotional post", "I've spent HOURS trying to get this working!!! The documentation is outdated and the error messages aren't helpful at all. Can someone PLEASE help?"},
		{"link sharing", "Here are some links I found helpful: https://docs.gno.land for official docs, https://github.com/gnolang/gno for source, and https://gno.land for the main site."},
		{"short reply", "That worked perfectly. Appreciate the help!"},
		{"multilingual", "Je pense que cette proposition est intéressante. Nous devrions en discuter plus en détail lors de la prochaine réunion."},
		{"numbers in text", "I just staked 5000 GNOT on the validator. The APR is around 15% which seems reasonable compared to other chains."},
		{"excited user", "Just tried the new feature and it works great! The team has done an incredible job. Looking forward to the next update!"},
	}

	for _, msg := range legitimateMessages {
		t.Run(msg.name, func(t *testing.T) {
			result := Score(ScoreInput{
				Author: "g1legitimateuser", Content: msg.content,
				Rate: normalRate, Rep: goodRep, Corpus: corpus, Fps: fps, Bl: bl,
			})
			if result.Total >= 5 {
				t.Errorf("FALSE POSITIVE: %q scored %d (triggered: %v)", msg.name, result.Total, result.Triggered)
			}
		})
	}
}
