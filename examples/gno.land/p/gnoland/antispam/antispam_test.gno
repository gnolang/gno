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
		HasUsername:    true,
		BanCount:       0,
	}

	newAccountRep := ReputationData{
		AccountAgeDays: 0,
		Balance:        3000000000,
		FlaggedCount:   0,
		TotalPosts:     0,
		HasUsername:    false,
		BanCount:       0,
	}

	badRep := ReputationData{
		AccountAgeDays: 30,
		Balance:        3000000000,
		FlaggedCount:   15,
		TotalPosts:     20,
		HasUsername:    true,
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
			wantMin: 2,
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
			wantMin: 4, // NEW_ACCOUNT(2) + NO_USERNAME(1) + LOW_BALANCE(1)
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

	// ---- Triggered accuracy: individual rules must appear ----

	t.Run("triggered: content rules listed individually", func(t *testing.T) {
		result := Score(ScoreInput{
			Author:  "g1user",
			Content: "THIS IS ALL CAPS WITH http://a.com http://b.com http://c.com http://d.com",
			Rate:    RateState{PostCount: 3, WindowSeconds: 3600},
			Rep:     goodRep,
			Corpus:  NewCorpus(),
			Fps:     NewFingerprintStore(),
			Bl:      NewBlocklist(),
		})
		wantRules := map[string]bool{"ALL_CAPS": false, "LINK_HEAVY": false}
		for _, r := range result.Triggered {
			if _, ok := wantRules[r]; ok {
				wantRules[r] = true
			}
		}
		for rule, found := range wantRules {
			if !found {
				t.Errorf("expected rule %q in Triggered, got %v", rule, result.Triggered)
			}
		}
	})

	t.Run("triggered: reputation rules listed individually", func(t *testing.T) {
		result := Score(ScoreInput{
			Author:  "g1newbie",
			Content: "Hello everyone this is a perfectly normal message with enough length.",
			Rate:    RateState{},
			Rep:     ReputationData{AccountAgeDays: 0, TotalPosts: 0, HasUsername: false, Balance: 100},
			Corpus:  NewCorpus(),
			Fps:     NewFingerprintStore(),
			Bl:      NewBlocklist(),
		})
		wantRules := map[string]bool{"NEW_ACCOUNT": false, "NO_USERNAME": false, "LOW_BALANCE": false}
		for _, r := range result.Triggered {
			if _, ok := wantRules[r]; ok {
				wantRules[r] = true
			}
		}
		for rule, found := range wantRules {
			if !found {
				t.Errorf("expected rule %q in Triggered, got %v", rule, result.Triggered)
			}
		}
	})
}

// ---- False positive test suite ----

func TestFalsePositives(t *testing.T) {
	// Good reputation, normal rate, trained corpus, no blocklist
	goodRep := ReputationData{
		AccountAgeDays: 60,
		Balance:        5000000000,
		FlaggedCount:   0,
		TotalPosts:     20,
		HasUsername:    true,
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

// ---- Single-pass optimization verification ----

func TestScoreContentAllEquivalence(t *testing.T) {
	cases := []string{
		"Normal text without any issues at all.",
		"THIS IS ALL CAPS SPAM MESSAGE THAT SHOULD BE DETECTED BY THE CONTENT RULES",
		"H\u0300\u0301\u0302\u0303\u0304ello with ze\u200Bro\u200Bwidth",
		"Send p\u0430yment now - homoglyph mixing here",
		"Check http://a.com http://b.com http://c.com http://d.com many links",
		"!!!???...!!! excessive punctuation everywhere !!!???!!!",
	}
	for _, content := range cases {
		contentHits := ScoreContent(content)
		unicodeHits := ScoreUnicode(content)
		allHits := scoreContentAll(content)

		separateTotal := 0
		for _, h := range contentHits {
			separateTotal += h.Score
		}
		for _, h := range unicodeHits {
			separateTotal += h.Score
		}

		allTotal := 0
		for _, h := range allHits {
			allTotal += h.Score
		}

		if separateTotal != allTotal {
			t.Errorf("scoreContentAll != ScoreContent+ScoreUnicode for %q: %d vs %d", content[:20], allTotal, separateTotal)
		}
	}
}

// ---- EarlyExitAt verification ----

func TestEarlyExitAt(t *testing.T) {
	// Setup: a scenario where cheap rules score high (caps + rate + bad rep)
	// and expensive rules would add more (bayes + keywords).
	corpus := NewCorpus()
	corpus.Train("free airdrop bonus click claim tokens", true)
	corpus.Train("free money bonus prize casino jackpot", true)
	corpus.Train("governance proposal voting community discussion", false)
	corpus.Train("development update release version changelog", false)

	dict := NewKeywordDict()
	dict.BulkAdd("free:3\nairdrop:3\nbonus:3\nclaim:2\nprize:3")

	bl := NewBlocklist()
	bl.AddPattern(`(?i)send\s+\d+\s*gnot`)

	badRep := ReputationData{
		AccountAgeDays: 30,
		Balance:        5000000000,
		FlaggedCount:   15,
		TotalPosts:     20,
		HasUsername:    true,
		BanCount:       2,
	}
	burstRate := RateState{PostCount: 30, WindowSeconds: 3600}

	content := "FREE AIRDROP BONUS PRIZE CLAIM CLICK NOW HURRY FAST"

	// EarlyExitDisabled: all rules evaluated, should include expensive rules
	full := Score(ScoreInput{
		Author: "g1user", Content: content,
		Rate: burstRate, Rep: badRep,
		Corpus: corpus, Dict: dict, Bl: bl,
		EarlyExitAt: EarlyExitDisabled,
	})

	// EarlyExitAt=ThresholdReject: cheap rules alone should exceed 8,
	// so expensive rules (bayes, keywords) should be skipped.
	early := Score(ScoreInput{
		Author: "g1user", Content: content,
		Rate: burstRate, Rep: badRep,
		Corpus: corpus, Dict: dict, Bl: bl,
		EarlyExitAt: ThresholdReject,
	})

	// Full score must be >= early score (early skips rules)
	if full.Total < early.Total {
		t.Errorf("full score (%d) < early exit score (%d)", full.Total, early.Total)
	}

	// Early exit should have fewer triggered rules
	if len(early.Triggered) >= len(full.Triggered) {
		t.Errorf("expected early exit to trigger fewer rules: early=%v, full=%v", early.Triggered, full.Triggered)
	}

	// Early exit score should still be >= the threshold
	if early.Total < ThresholdReject {
		t.Errorf("expected early exit score >= %d, got %d", ThresholdReject, early.Total)
	}
}

func TestEarlyExitIntraPhase2(t *testing.T) {
	// Scenario: Phase 1 scores 0 (good user), but Pattern(5) + Bayes(3) = 8.
	// With EarlyExitAt=8, Keywords and Fingerprint should be skipped.
	// With EarlyExitDisabled, Keywords also fires -> higher total.
	corpus := NewCorpus()
	corpus.Train("free airdrop bonus click claim tokens", true)
	corpus.Train("free money bonus prize casino jackpot", true)
	corpus.Train("free claim airdrop offer exclusive deal limited", true)
	corpus.Train("governance proposal voting community discussion", false)
	corpus.Train("validator staking delegation reward system", false)
	corpus.Train("development update release version changelog", false)
	corpus.Train("thanks for the helpful explanation today", false)

	dict := NewKeywordDict()
	dict.BulkAdd("free:3\nairdrop:3\nbonus:3\nclaim:2\nprize:3")

	bl := NewBlocklist()
	bl.AddPattern(`(?i)send\s+\d+\s*gnot`)

	goodRep := ReputationData{
		AccountAgeDays: 90, Balance: 10000000000,
		HasUsername: true, TotalPosts: 50,
	}
	normalRate := RateState{PostCount: 3, WindowSeconds: 3600}

	// Content matches pattern AND has Bayes/keyword spam tokens.
	content := "please send 100 gnot to claim your free airdrop bonus prize"

	full := Score(ScoreInput{
		Author: "g1user", Content: content,
		Rate: normalRate, Rep: goodRep,
		Corpus: corpus, Dict: dict, Bl: bl,
		EarlyExitAt: EarlyExitDisabled,
	})

	early := Score(ScoreInput{
		Author: "g1user", Content: content,
		Rate: normalRate, Rep: goodRep,
		Corpus: corpus, Dict: dict, Bl: bl,
		EarlyExitAt: ThresholdReject,
	})

	// Full score should be higher (Keywords fires too)
	if full.Total <= early.Total {
		t.Errorf("full score (%d) should exceed early score (%d)", full.Total, early.Total)
	}

	// Early exit must reach the threshold
	if early.Total < ThresholdReject {
		t.Errorf("expected early score >= %d, got %d (rules: %v)", ThresholdReject, early.Total, early.Triggered)
	}

	// Full should include KEYWORD_SPAM, early should not
	hasKeyword := false
	for _, r := range full.Triggered {
		if r == RuleKeywordSpam {
			hasKeyword = true
			break
		}
	}
	if !hasKeyword {
		t.Errorf("full score should include %s, got %v", RuleKeywordSpam, full.Triggered)
	}

	earlyHasKeyword := false
	for _, r := range early.Triggered {
		if r == RuleKeywordSpam {
			earlyHasKeyword = true
			break
		}
	}
	if earlyHasKeyword {
		t.Errorf("early exit should NOT include %s, got %v", RuleKeywordSpam, early.Triggered)
	}
}

// ---- MaxInputLength truncation ----

func TestMaxInputLength(t *testing.T) {
	t.Run("long input is truncated without panic", func(t *testing.T) {
		// Build input longer than MaxInputLength
		long := ""
		for len(long) < MaxInputLength+500 {
			long += "spam spam spam spam spam spam spam spam spam spam "
		}
		result := Score(ScoreInput{
			Author:  "g1user",
			Content: long,
			Rate:    RateState{PostCount: 3, WindowSeconds: 3600},
			Rep: ReputationData{
				AccountAgeDays: 90, Balance: 10000000000,
				HasUsername: true, TotalPosts: 50,
			},
			Corpus: NewCorpus(),
			Fps:    NewFingerprintStore(),
			Bl:     NewBlocklist(),
		})
		// Should not panic and should produce a valid score
		if result.Total < 0 {
			t.Errorf("expected non-negative score, got %d", result.Total)
		}
	})
}

// ---- BLOCKED_ADDRESS short-circuit verification ----

func TestBlockedAddressShortCircuit(t *testing.T) {
	bl := NewBlocklist()
	bl.BlockAddress("g1spammer")

	result := Score(ScoreInput{
		Author:  "g1spammer",
		Content: "totally normal message",
		Rate:    RateState{PostCount: 50, WindowSeconds: 60},
		Rep:     ReputationData{},
		Corpus:  NewCorpus(),
		Fps:     NewFingerprintStore(),
		Bl:      bl,
	})

	// Only BLOCKED_ADDRESS should fire - no other rules
	if len(result.Triggered) != 1 {
		t.Errorf("expected exactly 1 triggered rule, got %d: %v", len(result.Triggered), result.Triggered)
	}
	if result.Total != WeightBlockedAddress {
		t.Errorf("expected total %d, got %d", WeightBlockedAddress, result.Total)
	}
	if result.TopRule != RuleBlockedAddress {
		t.Errorf("expected top rule %q, got %q", RuleBlockedAddress, result.TopRule)
	}
}
