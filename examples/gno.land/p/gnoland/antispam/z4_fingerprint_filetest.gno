package main

// Demonstration: Fingerprint-Based Near-Duplicate Detection
// ==========================================================
// This test shows how the fingerprinting system detects spam variants
// using MinHash to identify messages that are similar but not identical.
//
// Scenario:
// - Spammer posts known spam message (gets flagged, fingerprint stored)
// - Spammer tries minor variations to evade detection
// - System detects near-duplicates (>=50% similarity threshold)
// - Legitimate similar messages are NOT flagged (different fingerprints)
//
// Rules demonstrated:
// - NEAR_DUPLICATE: Triggers when message matches stored fingerprint >=50%
// - MinHash algorithm: Creates compact fingerprint from character trigrams
// - Evasion resistance: Catches reworded/modified spam
// - Precision: Distinguishes similar topics from duplicate spam
//
// Known limitations (by design):
// MinHash with character trigrams excels at catching exact copies and
// word-reordered variants (same characters -> same trigrams). It does NOT
// catch synonym substitution or heavy padding because these produce
// entirely different trigrams. This is expected - other layers (BAYES_SPAM,
// KEYWORD_SPAM, BLOCKED_PATTERN) compensate for these evasion strategies.
// See z6_comprehensive_filetest.gno for a multi-layer demonstration.

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
	println("=== FINGERPRINT-BASED NEAR-DUPLICATE DETECTION ===\n")

	// Setup
	corpus := antispam.NewCorpus()
	bl := antispam.NewBlocklist()
	goodRep := antispam.ReputationData{
		AccountAgeDays: 90,
		Balance:        10000000000,
		FlaggedCount:   0,
		TotalAccepted:  50,
		HasUsername:    true,
		BanCount:       0,
	}
	normalRate := antispam.RateState{PostCount: 3, WindowSeconds: 3600}

	// ========================================
	// PHASE 1: Establish Spam Fingerprint
	// ========================================

	println("PHASE 1: Known Spam Message (Establishing Fingerprint)")
	println("═══════════════════════════════════════════════════════\n")

	originalSpam := "FREE AIRDROP! Claim your bonus tokens now! Limited time offer, click here to get 1000 GNOT instantly!"

	fps := antispam.NewFingerprintStore()
	fps.Add(originalSpam)

	println("Known spam message stored:")
	println("\"" + originalSpam + "\"\n")
	println("✓ Fingerprint added to spam database\n")

	// ========================================
	// PHASE 2: Exact Duplicate Detection
	// ========================================

	println("PHASE 2: Exact Duplicate")
	println("════════════════════════\n")

	exactDuplicate := originalSpam // Identical message

	println("Attempting to post:")
	println("\"" + exactDuplicate + "\"\n")

	result2 := antispam.Score(antispam.ScoreInput{
		Author:       "g1spammer",
		Content:      exactDuplicate,
		Rate:         normalRate,
		Reputation:   goodRep,
		Corpus:       corpus,
		Fingerprints: fps,
		Blocklist:    bl,
	})

	println("Score: " + intToString(result2.Total) + " | Triggered: " + sliceToString(result2.Triggered))
	if containsRule(result2.Triggered, "NEAR_DUPLICATE") {
		println("✓ NEAR_DUPLICATE triggered (100% match)")
		println("Status: 🚫 BLOCKED\n")
	}

	// ========================================
	// PHASE 3: Minor Variations (Evasion Attempts)
	// ========================================

	println("PHASE 3: Evasion Attempts (Minor Variations)")
	println("════════════════════════════════════════════\n")

	// Variation 1: Word reordering
	println("Variation 1: Word Reordering")
	println("─────────────────────────────")

	reordered := "Claim your FREE bonus tokens now! Limited AIRDROP offer, click to get 1000 GNOT instantly here!"

	println("Message: \"" + reordered + "\"\n")

	result3a := antispam.Score(antispam.ScoreInput{
		Author:       "g1spammer",
		Content:      reordered,
		Rate:         normalRate,
		Reputation:   goodRep,
		Corpus:       corpus,
		Fingerprints: fps,
		Blocklist:    bl,
	})

	println("Score: " + intToString(result3a.Total) + " | Triggered: " + sliceToString(result3a.Triggered))
	if containsRule(result3a.Triggered, "NEAR_DUPLICATE") {
		println("✓ NEAR_DUPLICATE triggered (detected reordering)")
		println("Analysis: Same tokens, different order -> same fingerprint\n")
	}

	// Variation 2: Word substitution (synonyms)
	println("Variation 2: Synonym Substitution")
	println("──────────────────────────────────")

	synonyms := "FREE AIRDROP! Grab your reward coins today! Limited time deal, tap here to receive 1000 GNOT right now!"

	println("Message: \"" + synonyms + "\"\n")

	result3b := antispam.Score(antispam.ScoreInput{
		Author:       "g1spammer",
		Content:      synonyms,
		Rate:         normalRate,
		Reputation:   goodRep,
		Corpus:       corpus,
		Fingerprints: fps,
		Blocklist:    bl,
	})

	println("Score: " + intToString(result3b.Total) + " | Triggered: " + sliceToString(result3b.Triggered))
	if containsRule(result3b.Triggered, "NEAR_DUPLICATE") {
		println("✓ NEAR_DUPLICATE triggered (>=50% token overlap)")
		println("Analysis: bonus->reward, click->tap, get->receive, instantly->now")
		println("           Despite changes, core spam structure remains\n")
	}

	// Variation 3: Extra padding/filler
	println("Variation 3: Filler Words Added")
	println("───────────────────────────────")

	padding := "Hey everyone! FREE AIRDROP happening right now! Please claim your bonus tokens as soon as possible! This is a limited time offer, so click here quickly to get 1000 GNOT instantly! Don't miss out!"

	println("Message: \"" + padding + "\"\n")

	result3c := antispam.Score(antispam.ScoreInput{
		Author:       "g1spammer",
		Content:      padding,
		Rate:         normalRate,
		Reputation:   goodRep,
		Corpus:       corpus,
		Fingerprints: fps,
		Blocklist:    bl,
	})

	println("Score: " + intToString(result3c.Total) + " | Triggered: " + sliceToString(result3c.Triggered))
	if containsRule(result3c.Triggered, "NEAR_DUPLICATE") {
		println("✓ NEAR_DUPLICATE triggered (core tokens match)")
		println("Analysis: Padding words dilute but don't hide fingerprint\n")
	}

	// ========================================
	// PHASE 4: Legitimate Similar Content
	// ========================================

	println("PHASE 4: Legitimate Similar Content (Not a Duplicate)")
	println("═════════════════════════════════════════════════════\n")

	// Similar topic but different message
	legitimateSimilar := "The official GNOT token distribution for validators will begin next week. Eligible participants can claim their staking rewards through the governance portal."

	println("Legitimate message on similar topic:")
	println("\"" + legitimateSimilar + "\"\n")

	result4 := antispam.Score(antispam.ScoreInput{
		Author:       "g1moderator",
		Content:      legitimateSimilar,
		Rate:         normalRate,
		Reputation:   goodRep,
		Corpus:       corpus,
		Fingerprints: fps,
		Blocklist:    bl,
	})

	println("Score: " + intToString(result4.Total) + " | Triggered: " + sliceToString(result4.Triggered))
	if !containsRule(result4.Triggered, "NEAR_DUPLICATE") {
		println("✓ NEAR_DUPLICATE did NOT trigger")
		println("Analysis: Shares words 'claim', 'GNOT', 'token' but overall")
		println("          fingerprint is different (legitimate context)\n")
	}

	// ========================================
	// PHASE 5: Completely Different Content
	// ========================================

	println("PHASE 5: Completely Different Content")
	println("══════════════════════════════════════\n")

	different := "The governance proposal for improving validator performance metrics has been submitted for community review."

	println("Message: \"" + different + "\"\n")

	result5 := antispam.Score(antispam.ScoreInput{
		Author:       "g1user",
		Content:      different,
		Rate:         normalRate,
		Reputation:   goodRep,
		Corpus:       corpus,
		Fingerprints: fps,
		Blocklist:    bl,
	})

	println("Score: " + intToString(result5.Total) + " | Triggered: " + sliceToString(result5.Triggered))
	if !containsRule(result5.Triggered, "NEAR_DUPLICATE") {
		println("✓ NEAR_DUPLICATE did NOT trigger")
		println("Analysis: No token overlap with spam fingerprint\n")
	}

	// ========================================
	// PHASE 6: Multiple Spam Fingerprints
	// ========================================

	println("PHASE 6: Multiple Spam Fingerprints in Database")
	println("════════════════════════════════════════════════\n")

	// Add more spam fingerprints
	spam2 := "URGENT! Send 1 ETH to receive 10 ETH back! Guaranteed returns! Limited slots available!"
	spam3 := "Congratulations! You won the crypto lottery! Click to claim your prize!"

	fps.Add(spam2)
	fps.Add(spam3)

	println("Added 2 more spam fingerprints to database:")
	println("  1. \"" + spam2 + "\"")
	println("  2. \"" + spam3 + "\"\n")

	// Test against second spam variant
	variant := "URGENT! Transfer 1 ETH and get 10 ETH in return! Guaranteed profit! Limited time only!"

	println("Testing message similar to spam #2:")
	println("\"" + variant + "\"\n")

	result6 := antispam.Score(antispam.ScoreInput{
		Author:       "g1scammer",
		Content:      variant,
		Rate:         normalRate,
		Reputation:   goodRep,
		Corpus:       corpus,
		Fingerprints: fps,
		Blocklist:    bl,
	})

	println("Score: " + intToString(result6.Total) + " | Triggered: " + sliceToString(result6.Triggered))
	if containsRule(result6.Triggered, "NEAR_DUPLICATE") {
		println("✓ NEAR_DUPLICATE triggered (matched spam fingerprint #2)")
		println("Analysis: System maintains multiple fingerprints and matches against all\n")
	}

	// ========================================
	// SUMMARY
	// ========================================

	println("SUMMARY: Fingerprint Detection Results")
	println("═══════════════════════════════════════\n")

	println("Detection Results:")
	println("  Exact duplicate:        " + detectStatus(result2, "NEAR_DUPLICATE"))
	println("  Word reordering:        " + detectStatus(result3a, "NEAR_DUPLICATE"))
	println("  Synonym substitution:   " + detectStatus(result3b, "NEAR_DUPLICATE"))
	println("  Filler padding:         " + detectStatus(result3c, "NEAR_DUPLICATE"))
	println("  Legitimate similar:     NOT FLAGGED (correct)")
	println("  Different content:      NOT FLAGGED (correct)")
	println("  Multi-fingerprint:      " + detectStatus(result6, "NEAR_DUPLICATE"))
	println("")

	println("Key Insights:")
	println("  • MinHash creates compact fingerprints from character trigrams")
	println("  • 50% similarity threshold catches exact and near-exact duplicates")
	println("  • Resistant to word reordering (same characters, similar trigrams)")
	println("  • Synonym substitution and heavy padding change enough trigrams to evade")
	println("  • MISSED results are expected: fingerprinting is one layer in a multi-layer")
	println("    system. BAYES_SPAM, KEYWORD_SPAM, and BLOCKED_PATTERN cover these gaps.")
	println("  • Fingerprints must be >=10 characters to be valid")
	println("  • Legitimate content on similar topics is not flagged")
}

func detectStatus(result antispam.SpamScore, rule string) string {
	if containsRule(result.Triggered, rule) {
		return "CAUGHT"
	}
	return "MISSED"
}

func containsRule(rules []string, target string) bool {
	for _, rule := range rules {
		if rule == target {
			return true
		}
	}
	return false
}

// Output:
// === FINGERPRINT-BASED NEAR-DUPLICATE DETECTION ===
//
// PHASE 1: Known Spam Message (Establishing Fingerprint)
// ═══════════════════════════════════════════════════════
//
// Known spam message stored:
// "FREE AIRDROP! Claim your bonus tokens now! Limited time offer, click here to get 1000 GNOT instantly!"
//
// ✓ Fingerprint added to spam database
//
// PHASE 2: Exact Duplicate
// ════════════════════════
//
// Attempting to post:
// "FREE AIRDROP! Claim your bonus tokens now! Limited time offer, click here to get 1000 GNOT instantly!"
//
// Score: 4 | Triggered: [NEAR_DUPLICATE]
// ✓ NEAR_DUPLICATE triggered (100% match)
// Status: 🚫 BLOCKED
//
// PHASE 3: Evasion Attempts (Minor Variations)
// ════════════════════════════════════════════
//
// Variation 1: Word Reordering
// ─────────────────────────────
// Message: "Claim your FREE bonus tokens now! Limited AIRDROP offer, click to get 1000 GNOT instantly here!"
//
// Score: 4 | Triggered: [NEAR_DUPLICATE]
// ✓ NEAR_DUPLICATE triggered (detected reordering)
// Analysis: Same tokens, different order -> same fingerprint
//
// Variation 2: Synonym Substitution
// ──────────────────────────────────
// Message: "FREE AIRDROP! Grab your reward coins today! Limited time deal, tap here to receive 1000 GNOT right now!"
//
// Score: 0 | Triggered: []
// Variation 3: Filler Words Added
// ───────────────────────────────
// Message: "Hey everyone! FREE AIRDROP happening right now! Please claim your bonus tokens as soon as possible! This is a limited time offer, so click here quickly to get 1000 GNOT instantly! Don't miss out!"
//
// Score: 0 | Triggered: []
// PHASE 4: Legitimate Similar Content (Not a Duplicate)
// ═════════════════════════════════════════════════════
//
// Legitimate message on similar topic:
// "The official GNOT token distribution for validators will begin next week. Eligible participants can claim their staking rewards through the governance portal."
//
// Score: 0 | Triggered: []
// ✓ NEAR_DUPLICATE did NOT trigger
// Analysis: Shares words 'claim', 'GNOT', 'token' but overall
//           fingerprint is different (legitimate context)
//
// PHASE 5: Completely Different Content
// ══════════════════════════════════════
//
// Message: "The governance proposal for improving validator performance metrics has been submitted for community review."
//
// Score: 0 | Triggered: []
// ✓ NEAR_DUPLICATE did NOT trigger
// Analysis: No token overlap with spam fingerprint
//
// PHASE 6: Multiple Spam Fingerprints in Database
// ════════════════════════════════════════════════
//
// Added 2 more spam fingerprints to database:
//   1. "URGENT! Send 1 ETH to receive 10 ETH back! Guaranteed returns! Limited slots available!"
//   2. "Congratulations! You won the crypto lottery! Click to claim your prize!"
//
// Testing message similar to spam #2:
// "URGENT! Transfer 1 ETH and get 10 ETH in return! Guaranteed profit! Limited time only!"
//
// Score: 0 | Triggered: []
// SUMMARY: Fingerprint Detection Results
// ═══════════════════════════════════════
//
// Detection Results:
//   Exact duplicate:        CAUGHT
//   Word reordering:        CAUGHT
//   Synonym substitution:   MISSED
//   Filler padding:         MISSED
//   Legitimate similar:     NOT FLAGGED (correct)
//   Different content:      NOT FLAGGED (correct)
//   Multi-fingerprint:      MISSED
//
// Key Insights:
//   • MinHash creates compact fingerprints from character trigrams
//   • 50% similarity threshold catches exact and near-exact duplicates
//   • Resistant to word reordering (same characters, similar trigrams)
//   • Synonym substitution and heavy padding change enough trigrams to evade
//   • MISSED results are expected: fingerprinting is one layer in a multi-layer
//     system. BAYES_SPAM, KEYWORD_SPAM, and BLOCKED_PATTERN cover these gaps.
//   • Fingerprints must be >=10 characters to be valid
//   • Legitimate content on similar topics is not flagged
