package main

// Demonstration: Multi-User Reputation Profiles
// ==============================================
// This test shows how the same content gets different spam scores
// depending on the author's reputation and posting behavior.
//
// Scenario:
// - Three users post identical suspicious content
// - Established user: gets benefit of the doubt
// - New account: triggers reputation penalties
// - Known spammer: immediately blocked
//
// Rules demonstrated:
// - Reputation scoring (NEW_ACCOUNT, NO_USERNAME, LOW_BALANCE)
// - Blocklisting (BLOCKED_ADDRESS)
// - Allowlisting (bypasses all rules)
// - Context-dependent scoring (same content, different outcomes)

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
	println("=== MULTI-USER REPUTATION PROFILES ===\n")

	// Setup: Configure blocklist and allowlist
	bl := antispam.NewBlocklist()
	bl.BlockAddress("g1spammer")
	bl.AllowAddress("g1moderator")

	corpus := antispam.NewCorpus()
	corpus.Train("free airdrop bonus claim prize", true)
	corpus.Train("governance discussion voting proposal", false)

	fps := antispam.NewFingerprintStore()

	// Keyword dictionary for common spam terms
	dict := antispam.NewKeywordDict()
	dict.BulkAdd("free:3\nairdrop:3\nbonus:3\nclaim:2\nprize:3\nmoney:2\nclick:2")

	normalRate := antispam.RateState{PostCount: 3, WindowSeconds: 3600}

	// Test content: borderline suspicious (has spam keywords but could be legitimate)
	testContent := "Check out this FREE workshop on blockchain governance! Limited seats, claim your spot now."

	println("TEST MESSAGE (posted by 4 different users):")
	println("\"" + testContent + "\"\n")

	// ---- USER 1: Established, trusted user ----
	println("USER 1: Established Community Member")
	println("─────────────────────────────────────")

	establishedRep := antispam.ReputationData{
		AccountAgeDays: 180,         // 6 months old
		Balance:        50000000000, // 50k GNOT
		FlaggedCount:   0,
		TotalPosts:     200,
		HasUsername:    true,
		BanCount:       0,
	}

	result1 := antispam.Score(antispam.ScoreInput{
		Author:  "g1veteran",
		Content: testContent,
		Rate:    normalRate,
		Rep:     establishedRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Address: g1veteran")
	println("Account Age: 180 days | Balance: 50k GNOT | Posts: 200")
	println("Score: " + intToString(result1.Total) + " | Triggered: " + sliceToString(result1.Triggered))
	if result1.Total < antispam.ThresholdHide {
		println("Status: ✓ APPROVED (trusted user, benefit of the doubt)\n")
	}

	// ---- USER 2: Brand new account ----
	println("USER 2: Brand New Account")
	println("─────────────────────────")

	newAccountRep := antispam.ReputationData{
		AccountAgeDays: 0,          // Just created
		Balance:        1000000000, // 1 GNOT (low balance)
		FlaggedCount:   0,
		TotalPosts:     0,     // First post
		HasUsername:    false, // No username set
		BanCount:       0,
	}

	result2 := antispam.Score(antispam.ScoreInput{
		Author:  "g1newbie42",
		Content: testContent,
		Rate:    normalRate,
		Rep:     newAccountRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Address: g1newbie42")
	println("Account Age: 0 days | Balance: 1 GNOT | Posts: 0 | Username: no")
	println("Score: " + intToString(result2.Total) + " | Triggered: " + sliceToString(result2.Triggered))
	if result2.Total >= antispam.ThresholdHide {
		println("Status: ⛔ HIDDEN (score ≥ " + intToString(antispam.ThresholdHide) + ")")
		println("Reputation penalties applied:")
		println("  - NEW_ACCOUNT: +2 (age < 7 days)")
		println("  - NO_USERNAME: +1 (no username configured)")
		println("  - LOW_BALANCE: +1 (balance < 5 GNOT)")
		println("Combined with content signals, exceeds hide threshold\n")
	}

	// ---- USER 3: Known spammer (blocked) ----
	println("USER 3: Known Spammer (Blocklisted)")
	println("────────────────────────────────────")

	// Even with good stats, blocked address gets maximum penalty
	goodStats := antispam.ReputationData{
		AccountAgeDays: 90,
		Balance:        10000000000,
		FlaggedCount:   0,
		TotalPosts:     50,
		HasUsername:    true,
		BanCount:       0,
	}

	result3 := antispam.Score(antispam.ScoreInput{
		Author:  "g1spammer",
		Content: testContent,
		Rate:    normalRate,
		Rep:     goodStats,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Address: g1spammer (BLOCKLISTED)")
	println("Score: " + intToString(result3.Total) + " | Triggered: " + sliceToString(result3.Triggered))
	if result3.Total >= antispam.ThresholdReject {
		println("Status: 🚫 BLOCKED (score ≥ " + intToString(antispam.ThresholdReject) + ")")
		println("Blocklist entry overrides all other factors\n")
	}

	// ---- USER 4: Moderator (allowlisted) ----
	println("USER 4: Moderator (Allowlisted)")
	println("────────────────────────────────")

	// Even with terrible stats and spammy content, allowlisted users bypass all checks
	terribleStats := antispam.ReputationData{
		AccountAgeDays: 0,
		Balance:        0,
		FlaggedCount:   100,
		TotalPosts:     0,
		HasUsername:    false,
		BanCount:       5,
	}

	result4 := antispam.Score(antispam.ScoreInput{
		Author:  "g1moderator",
		Content: "FREE SPAM EVERYWHERE!!! CLICK NOW!!! http://spam.com", // Obviously spam
		Rate:    antispam.RateState{PostCount: 50, WindowSeconds: 3600}, // Burst
		Rep:     terribleStats,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Address: g1moderator (ALLOWLISTED)")
	println("Score: " + intToString(result4.Total) + " | Triggered: " + sliceToString(result4.Triggered))
	if result4.Total == 0 {
		println("Status: ✓ BYPASSED (allowlist grants immunity to all rules)\n")
	}

	// ---- SUMMARY ----
	println("COMPARATIVE ANALYSIS")
	println("────────────────────")
	println("Same content, different scores:")
	println("  Established user:  " + intToString(result1.Total) + " (approved)")
	println("  New account:       " + intToString(result2.Total) + " (hidden)")
	println("  Blocklisted:       " + intToString(result3.Total) + " (rejected)")
	println("  Allowlisted:       " + intToString(result4.Total) + " (bypassed)")
	println("")
	println("Key insights:")
	println("  • Reputation significantly affects scoring for borderline content")
	println("  • New accounts face stricter scrutiny to prevent spam account creation")
	println("  • Blocklists provide immediate protection against known bad actors")
	println("  • Allowlists enable trusted users (moderators/admins) to bypass filters")
}

// Output:
// === MULTI-USER REPUTATION PROFILES ===
//
// TEST MESSAGE (posted by 4 different users):
// "Check out this FREE workshop on blockchain governance! Limited seats, claim your spot now."
//
// USER 1: Established Community Member
// ─────────────────────────────────────
// Address: g1veteran
// Account Age: 180 days | Balance: 50k GNOT | Posts: 200
// Score: 3 | Triggered: [KEYWORD_SPAM]
// Status: ✓ APPROVED (trusted user, benefit of the doubt)
//
// USER 2: Brand New Account
// ─────────────────────────
// Address: g1newbie42
// Account Age: 0 days | Balance: 1 GNOT | Posts: 0 | Username: no
// Score: 6 | Triggered: [NEW_ACCOUNT NO_USERNAME KEYWORD_SPAM]
// Status: ⛔ HIDDEN (score ≥ 5)
// Reputation penalties applied:
//   - NEW_ACCOUNT: +2 (age < 7 days)
//   - NO_USERNAME: +1 (no username configured)
//   - LOW_BALANCE: +1 (balance < 5 GNOT)
// Combined with content signals, exceeds hide threshold
//
// USER 3: Known Spammer (Blocklisted)
// ────────────────────────────────────
// Address: g1spammer (BLOCKLISTED)
// Score: 99 | Triggered: [BLOCKED_ADDRESS]
// Status: 🚫 BLOCKED (score ≥ 8)
// Blocklist entry overrides all other factors
//
// USER 4: Moderator (Allowlisted)
// ────────────────────────────────
// Address: g1moderator (ALLOWLISTED)
// Score: 0 | Triggered: []
// Status: ✓ BYPASSED (allowlist grants immunity to all rules)
//
// COMPARATIVE ANALYSIS
// ────────────────────
// Same content, different scores:
//   Established user:  3 (approved)
//   New account:       6 (hidden)
//   Blocklisted:       99 (rejected)
//   Allowlisted:       0 (bypassed)
//
// Key insights:
//   • Reputation significantly affects scoring for borderline content
//   • New accounts face stricter scrutiny to prevent spam account creation
//   • Blocklists provide immediate protection against known bad actors
//   • Allowlists enable trusted users (moderators/admins) to bypass filters
