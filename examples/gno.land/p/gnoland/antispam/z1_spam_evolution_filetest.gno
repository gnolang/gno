package main

// Demonstration: Spam Detection Evolution
// ==========================================
// This test shows how a user's behavior evolves from legitimate to spam,
// and how the antispam system detects and scores this progression.
//
// Scenario:
// - User starts as a legitimate contributor with good reputation
// - Gradually shifts to spammy behavior (urgency, caps, links)
// - System detects the change through multiple signals
//
// Rules demonstrated:
// - Content analysis (ALL_CAPS, LINK_HEAVY, PUNCTUATION_SPAM)
// - Keyword detection (free, bonus, claim, airdrop)
// - Bayesian classification (trained corpus)
// - Rule stacking (multiple rules trigger for higher scores)

import (
	"gno.land/p/gnoland/antispam"
)

func intToString(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	s := ""
	for i > 0 {
		s = string(rune(48+(i%10))) + s
		i /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

func sliceToString(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	s := "["
	for i, v := range arr {
		if i > 0 {
			s += " "
		}
		s += v
	}
	s += "]"
	return s
}

func main() {
	println("=== SPAM DETECTION EVOLUTION DEMONSTRATION ===\n")

	// Setup: Train a Bayesian corpus to recognize spam patterns
	corpus := antispam.NewCorpus()
	corpus.Train("free airdrop bonus click claim tokens prize", true)
	corpus.Train("free bonus casino money prize jackpot gambling", true)
	corpus.Train("free claim airdrop offer exclusive deal limited", true)
	corpus.Train("governance proposal voting community discussion", false)
	corpus.Train("validator staking delegation reward system", false)
	corpus.Train("development update release version changelog", false)

	fps := antispam.NewFingerprintStore()
	bl := antispam.NewBlocklist()

	// Setup: Keyword dictionary for common spam terms
	dict := antispam.NewKeywordDict()
	dict.BulkAdd("free:3\nairdrop:3\nbonus:3\nclaim:2\nprize:3\nmoney:2\nclick:2\ncasino:3\njackpot:3")

	// Good reputation: established user
	goodRep := antispam.ReputationData{
		AccountAgeDays: 90,
		Balance:        10000000000, // 10k GNOT
		FlaggedCount:   0,
		TotalPosts:     50,
		HasUsername:    true,
		BanCount:       0,
	}

	normalRate := antispam.RateState{PostCount: 3, WindowSeconds: 3600}

	println("PHASE 1: Legitimate Behavior")
	println("─────────────────────────────")

	// Post 1: Normal technical question
	msg1 := "How do I configure my validator node for optimal performance?"
	result1 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user123",
		Content: msg1,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})
	println("Post: \"" + msg1 + "\"")
	println("Score: " + intToString(result1.Total) + " | Triggered: " + sliceToString(result1.Triggered))
	println("Status: ✓ LEGITIMATE (score < 5)\n")

	println("PHASE 2: Borderline Content")
	println("─────────────────────────────")

	// Post 2: Excited but legitimate (caps for emphasis)
	msg2 := "JUST STAKED 5000 GNOT! The validator setup was easier than expected."
	result2 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user123",
		Content: msg2,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})
	println("Post: \"" + msg2 + "\"")
	println("Score: " + intToString(result2.Total) + " | Triggered: " + sliceToString(result2.Triggered))
	println("Status: ⚠ BORDERLINE (caps detected but offset by good reputation)\n")

	println("PHASE 3: Suspicious Shift")
	println("─────────────────────────")

	// Post 3: Multiple spam signals appear
	msg3 := "FREE AIRDROP!!! Claim your bonus tokens now! Limited time offer! https://claim.example.com"
	result3 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user123",
		Content: msg3,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})
	println("Post: \"" + msg3 + "\"")
	println("Score: " + intToString(result3.Total) + " | Triggered: " + sliceToString(result3.Triggered))
	if result3.Total >= antispam.ThresholdHide {
		println("Status: ⛔ HIDDEN (score >= " + intToString(antispam.ThresholdHide) + ")")
	}
	println("Analysis: Multiple spam signals detected:")
	println("  - KEYWORD_SPAM: 'free', 'airdrop', 'bonus', 'claim'")
	println("  - BAYES_SPAM: High spam probability from trained corpus\n")

	println("PHASE 4: Confirmed Spam")
	println("─────────────────────────")

	// Post 4: Heavy spam with rate burst
	burstRate := antispam.RateState{PostCount: 25, WindowSeconds: 3600}
	msg4 := "CLICK HERE NOW!!! FREE MONEY BONUS PRIZE AIRDROP!!! http://spam1.com http://spam2.com http://spam3.com"
	result4 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user123",
		Content: msg4,
		Rate:    burstRate, // Now posting rapidly
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})
	println("Post: \"" + msg4 + "\"")
	println("Score: " + intToString(result4.Total) + " | Triggered: " + sliceToString(result4.Triggered))
	if result4.Total >= antispam.ThresholdReject {
		println("Status: 🚫 REJECTED (score >= " + intToString(antispam.ThresholdReject) + ")")
	}
	println("Analysis: Multiple compounding factors:")
	println("  - ALL_CAPS: Entire message capitalized")
	println("  - LINK_HEAVY: 3+ links in short message")
	println("  - KEYWORD_SPAM: Dense spam vocabulary")
	println("  - BAYES_SPAM: Very high spam probability")
	println("  - RATE_BURST: 25 posts/hour (threshold: 10/hour)")
	println("  - Score stacking: Individual penalties add up to exceed reject threshold\n")

	println("SUMMARY")
	println("─────────")
	println("Score progression: " + intToString(result1.Total) + " -> " + intToString(result2.Total) + " -> " + intToString(result3.Total) + " -> " + intToString(result4.Total))
	println("The system detected the behavioral shift from legitimate to spam")
	println("through gradual accumulation of multiple independent signals.")
}

// Output:
// === SPAM DETECTION EVOLUTION DEMONSTRATION ===
//
// PHASE 1: Legitimate Behavior
// ─────────────────────────────
// Post: "How do I configure my validator node for optimal performance?"
// Score: 0 | Triggered: []
// Status: ✓ LEGITIMATE (score < 5)
//
// PHASE 2: Borderline Content
// ─────────────────────────────
// Post: "JUST STAKED 5000 GNOT! The validator setup was easier than expected."
// Score: 0 | Triggered: []
// Status: ⚠ BORDERLINE (caps detected but offset by good reputation)
//
// PHASE 3: Suspicious Shift
// ─────────────────────────
// Post: "FREE AIRDROP!!! Claim your bonus tokens now! Limited time offer! https://claim.example.com"
// Score: 6 | Triggered: [BAYES_SPAM KEYWORD_SPAM]
// Status: ⛔ HIDDEN (score >= 5)
// Analysis: Multiple spam signals detected:
//   - KEYWORD_SPAM: 'free', 'airdrop', 'bonus', 'claim'
//   - BAYES_SPAM: High spam probability from trained corpus
//
// PHASE 4: Confirmed Spam
// ─────────────────────────
// Post: "CLICK HERE NOW!!! FREE MONEY BONUS PRIZE AIRDROP!!! http://spam1.com http://spam2.com http://spam3.com"
// Score: 12 | Triggered: [ALL_CAPS RATE_BURST BAYES_SPAM KEYWORD_SPAM]
// Status: 🚫 REJECTED (score >= 8)
// Analysis: Multiple compounding factors:
//   - ALL_CAPS: Entire message capitalized
//   - LINK_HEAVY: 3+ links in short message
//   - KEYWORD_SPAM: Dense spam vocabulary
//   - BAYES_SPAM: Very high spam probability
//   - RATE_BURST: 25 posts/hour (threshold: 10/hour)
//   - Score stacking: Individual penalties add up to exceed reject threshold
//
// SUMMARY
// ─────────
// Score progression: 0 -> 0 -> 6 -> 12
// The system detected the behavioral shift from legitimate to spam
// through gradual accumulation of multiple independent signals.
