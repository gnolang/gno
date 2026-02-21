package main

// Demonstration: Comprehensive Multi-Layer Spam Detection
// ========================================================
// This test exercises ALL antispam algorithms in combination, showing how
// the multi-layer system catches spam that any single layer would miss.
//
// Every rule in the engine triggers at least once across these scenarios:
//
//   BLOCKED_ADDRESS (99)  BLOCKED_PATTERN (5)   NEAR_DUPLICATE (4)
//   RATE_BURST (4)        ALL_CAPS (3)          SHORT_WITH_LINK (3)
//   BAYES_SPAM (3)        BAD_REPUTATION (3)    KEYWORD_SPAM (3)
//   ZALGO_TEXT (3)        LINK_HEAVY (2)        NEW_ACCOUNT (2)
//   INVISIBLE_CHARS (2)   HOMOGLYPH_MIX (2)    EXCESSIVE_PUNCT (1)
//   NO_USERNAME (1)       LOW_BALANCE (1)        BANNED_BEFORE (1-3)
//
// Scenarios:
// 1. Legitimate power user -> score 0
// 2. New account spammer -> reputation + content + keywords + bayes + punct
// 3. Spam flood from flagged user -> caps + links + rate + keywords + bayes + reputation + ban history
// 4. Near-duplicate + pattern match -> fingerprint + regex
// 5. Short link spam -> short body with URL
// 6. Blocked address -> instant rejection
// 7. Allowlisted moderator -> bypass all rules
// 8. Unicode abuse -> zalgo + invisible chars + homoglyph mixing

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
	println("=== COMPREHENSIVE MULTI-LAYER SPAM DETECTION ===\n")

	// ---- Shared setup

	// Bayesian corpus: trained with spam and ham vocabulary
	corpus := antispam.NewCorpus()
	corpus.Train("free airdrop bonus click claim tokens prize win", true)
	corpus.Train("casino gambling jackpot money prize winner", true)
	corpus.Train("limited time offer act now exclusive deal", true)
	corpus.Train("amazing opportunity earn passive income guaranteed", true)
	corpus.Train("governance proposal voting community discussion", false)
	corpus.Train("validator staking delegation reward system", false)
	corpus.Train("development update release version changelog", false)
	corpus.Train("pull request review code changes implementation", false)

	// Keyword dictionary
	dict := antispam.NewKeywordDict()
	dict.BulkAdd("free:3\nairdrop:3\nbonus:3\nclaim:2\nprize:3\nmoney:2\nclick:2\ncasino:3\njackpot:3")

	// Fingerprint store with a known spam fingerprint
	fps := antispam.NewFingerprintStore()
	fps.Add("Send 100 GNOT to receive 1000 GNOT back! Guaranteed returns! Act now!")

	// Blocklist with addresses and patterns
	bl := antispam.NewBlocklist()
	bl.BlockAddress("g1blocked")
	bl.AllowAddress("g1admin")
	bl.AddPattern(`(?i)send\s+\d+\s+\w+\s+get\s+\d+`)

	// ========================================
	// SCENARIO 1: Legitimate Power User
	// ========================================

	println("SCENARIO 1: Legitimate Power User")
	println("══════════════════════════════════\n")

	goodRep := antispam.ReputationData{
		AccountAgeDays: 365,
		Balance:        50000000000,
		FlaggedCount:   0,
		TotalPosts:     500,
		HasUsername:    true,
		BanCount:       0,
	}

	legitMsg := "The governance proposal for validator delegation has been submitted. Please review and vote before the deadline."

	result1 := antispam.Score(antispam.ScoreInput{
		Author:  "g1veteran",
		Content: legitMsg,
		Rate:    antispam.RateState{PostCount: 2, WindowSeconds: 3600},
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Author:  g1veteran (365 days, 500 posts, good standing)")
	println("Message: \"" + legitMsg + "\"")
	println("Score:   " + intToString(result1.Total) + " | Triggered: " + sliceToString(result1.Triggered))
	println("Status:  ✓ APPROVED\n")

	// ========================================
	// SCENARIO 2: New Account Spammer
	// ========================================

	println("SCENARIO 2: New Account Spammer")
	println("════════════════════════════════\n")

	newRep := antispam.ReputationData{
		AccountAgeDays: 0,
		Balance:        500000000, // 500 GNOT (< 1000 threshold)
		FlaggedCount:   0,
		TotalPosts:     0,
		HasUsername:    false,
		BanCount:       0,
	}

	// Message with spam keywords AND >20% punctuation (EXCESSIVE_PUNCT)
	spamPunct := "FREE!!! AIRDROP!!! Claim your bonus prize now!!!"

	result2 := antispam.Score(antispam.ScoreInput{
		Author:  "g1newspam",
		Content: spamPunct,
		Rate:    antispam.RateState{PostCount: 5, WindowSeconds: 3600},
		Rep:     newRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Author:  g1newspam (0 days, no username, low balance)")
	println("Message: \"" + spamPunct + "\"")
	println("Score:   " + intToString(result2.Total) + " | Triggered: " + sliceToString(result2.Triggered))
	println("Rules breakdown:")
	println("  NEW_ACCOUNT(2) + NO_USERNAME(1) + LOW_BALANCE(1) = reputation penalties")
	println("  KEYWORD_SPAM(3) + BAYES_SPAM(3) = content detection")
	println("  EXCESSIVE_PUNCT(1) = punctuation spam (>20% of characters)")
	println("Status:  🚫 REJECTED (score >= " + intToString(antispam.ThresholdReject) + ")\n")

	// ========================================
	// SCENARIO 3: Spam Flood from Flagged User
	// ========================================

	println("SCENARIO 3: Spam Flood from Flagged User")
	println("═════════════════════════════════════════\n")

	flaggedRep := antispam.ReputationData{
		AccountAgeDays: 30,
		Balance:        5000000000,
		FlaggedCount:   20,
		TotalPosts:     40, // 50% flag ratio (> 30% threshold)
		HasUsername:    true,
		BanCount:       2,
	}

	floodMsg := "FREE MONEY BONUS PRIZE AIRDROP CLICK NOW!!! http://spam1.com http://spam2.com http://spam3.com http://spam4.com"

	result3 := antispam.Score(antispam.ScoreInput{
		Author:  "g1flooder",
		Content: floodMsg,
		Rate:    antispam.RateState{PostCount: 30, WindowSeconds: 3600},
		Rep:     flaggedRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Author:  g1flooder (50% flag ratio, 2 past bans, 30 posts/hour)")
	println("Message: \"" + floodMsg + "\"")
	println("Score:   " + intToString(result3.Total) + " | Triggered: " + sliceToString(result3.Triggered))
	println("Rules breakdown:")
	println("  ALL_CAPS(3) + LINK_HEAVY(2) = content heuristics")
	println("  RATE_BURST(4) = posting frequency")
	println("  BAD_REPUTATION(3) + BANNED_BEFORE(2) = reputation")
	println("  KEYWORD_SPAM(3) + BAYES_SPAM(3) = vocabulary analysis")
	println("Status:  🚫 REJECTED\n")

	// ========================================
	// SCENARIO 4: Near-Duplicate + Pattern Match
	// ========================================

	println("SCENARIO 4: Near-Duplicate with Pattern Match")
	println("══════════════════════════════════════════════\n")

	// This message matches both the stored fingerprint AND a regex pattern
	// Pattern: send \d+ \w+ get \d+ - requires no extra words between GNOT and get
	dupPatternMsg := "Send 100 GNOT get 1000 GNOT back! Guaranteed returns! Act now!"

	result4 := antispam.Score(antispam.ScoreInput{
		Author:  "g1scammer",
		Content: dupPatternMsg,
		Rate:    antispam.RateState{PostCount: 3, WindowSeconds: 3600},
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Author:  g1scammer (good reputation, normal rate)")
	println("Message: \"" + dupPatternMsg + "\"")
	println("Score:   " + intToString(result4.Total) + " | Triggered: " + sliceToString(result4.Triggered))
	println("Rules breakdown:")
	println("  BLOCKED_PATTERN(5) = regex matched 'send X get Y' scam pattern")
	println("  NEAR_DUPLICATE(4) = fingerprint matched known spam")
	println("  BAYES_SPAM(3) = corpus recognizes spam vocabulary")
	println("  Even a trusted user gets caught when content matches multiple layers")
	println("Status:  🚫 REJECTED\n")

	// ========================================
	// SCENARIO 5: Short Link Spam
	// ========================================

	println("SCENARIO 5: Short Link Spam")
	println("═══════════════════════════\n")

	shortLink := "Check this! http://x.co"

	result5 := antispam.Score(antispam.ScoreInput{
		Author:  "g1shortlink",
		Content: shortLink,
		Rate:    antispam.RateState{PostCount: 3, WindowSeconds: 3600},
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Author:  g1shortlink (good reputation)")
	println("Message: \"" + shortLink + "\"")
	println("Score:   " + intToString(result5.Total) + " | Triggered: " + sliceToString(result5.Triggered))
	println("Rules breakdown:")
	println("  SHORT_WITH_LINK(3) = body <=30 chars with a URL")
	println("  EXCESSIVE_PUNCT(1) = URL punctuation (://) inflates the ratio")
	println("  Minimal content with a link is a common spam pattern")
	println("Status:  ⚠ SUSPICIOUS (below reject, above normal)\n")

	// ========================================
	// SCENARIO 6: Blocked Address
	// ========================================

	println("SCENARIO 6: Blocked Address")
	println("═══════════════════════════\n")

	result6 := antispam.Score(antispam.ScoreInput{
		Author:  "g1blocked",
		Content: legitMsg, // clean content doesn't matter
		Rate:    antispam.RateState{PostCount: 1, WindowSeconds: 3600},
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Author:  g1blocked (BLOCKLISTED)")
	println("Message: \"" + legitMsg + "\"")
	println("Score:   " + intToString(result6.Total) + " | Triggered: " + sliceToString(result6.Triggered))
	println("Rules breakdown:")
	println("  BLOCKED_ADDRESS(99) = instant rejection, no further rules evaluated")
	println("  Clean content and good reputation are irrelevant")
	println("Status:  🚫 BLOCKED\n")

	// ========================================
	// SCENARIO 7: Allowlisted Moderator
	// ========================================

	println("SCENARIO 7: Allowlisted Moderator")
	println("══════════════════════════════════\n")

	result7 := antispam.Score(antispam.ScoreInput{
		Author:  "g1admin",
		Content: floodMsg, // obvious spam content
		Rate:    antispam.RateState{PostCount: 50, WindowSeconds: 3600},
		Rep: antispam.ReputationData{
			AccountAgeDays: 0,
			Balance:        0,
			FlaggedCount:   100,
			TotalPosts:     0,
			HasUsername:    false,
			BanCount:       10,
		},
		Corpus: corpus,
		Fps:    fps,
		Bl:     bl,
		Dict:   dict,
	})

	println("Author:  g1admin (ALLOWLISTED)")
	println("Message: \"" + floodMsg + "\"")
	println("Score:   " + intToString(result7.Total) + " | Triggered: " + sliceToString(result7.Triggered))
	println("Rules breakdown:")
	println("  Allowlist bypasses ALL rules - score is always 0")
	println("  Use case: moderators, admins, emergency announcements")
	println("Status:  ✓ BYPASSED\n")

	// ========================================
	// SCENARIO 8: Unicode Abuse
	// ========================================

	println("SCENARIO 8: Unicode Abuse (Zalgo + Invisible + Homoglyph)")
	println("═════════════════════════════════════════════════════════\n")

	// Zalgo (4 combining marks on S) + homoglyph (cyrillic а in "pаyment") + invisible (2x zero-width space)
	unicodeMsg := "S\u0300\u0301\u0302\u0303end p\u0430yment\u200B\u200Bnow"

	result8 := antispam.Score(antispam.ScoreInput{
		Author:  "g1griefer",
		Content: unicodeMsg,
		Rate:    antispam.RateState{PostCount: 1, WindowSeconds: 3600},
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Author:  g1griefer (good reputation, normal rate)")
	println("Message: contains zalgo text, cyrillic homoglyph, zero-width spaces")
	println("Score:   " + intToString(result8.Total) + " | Triggered: " + sliceToString(result8.Triggered))
	println("Rules breakdown:")
	println("  ZALGO_TEXT(3) = stacked combining marks distort rendering")
	println("  INVISIBLE_CHARS(2) = zero-width spaces for obfuscation")
	println("  HOMOGLYPH_MIX(2) = cyrillic 'а' mixed with latin in same word")
	println("  Good reputation irrelevant - unicode abuse is a content signal")
	println("Status:  ⚠ HIDDEN (score >= " + intToString(antispam.ThresholdHide) + ")\n")

	// ========================================
	// SUMMARY
	// ========================================

	println("RULE COVERAGE MATRIX")
	println("════════════════════\n")

	println("Scenario  Score  Rules Triggered")
	println("────────  ─────  ───────────────")
	println("1. Legit     " + intToString(result1.Total) + "  " + sliceToString(result1.Triggered))
	println("2. New acc  " + intToString(result2.Total) + "  " + sliceToString(result2.Triggered))
	println("3. Flood   " + intToString(result3.Total) + "  " + sliceToString(result3.Triggered))
	println("4. Dup+Pat  " + intToString(result4.Total) + "  " + sliceToString(result4.Triggered))
	println("5. Short    " + intToString(result5.Total) + "  " + sliceToString(result5.Triggered))
	println("6. Blocked " + intToString(result6.Total) + "  " + sliceToString(result6.Triggered))
	println("7. Allowed   " + intToString(result7.Total) + "  " + sliceToString(result7.Triggered))
	println("8. Unicode   " + intToString(result8.Total) + "  " + sliceToString(result8.Triggered))
	println("")

	println("All 18 rules triggered at least once across these scenarios.")
	println("No single rule is sufficient - multi-layer detection is the key.")
}

// Output:
// === COMPREHENSIVE MULTI-LAYER SPAM DETECTION ===
//
// SCENARIO 1: Legitimate Power User
// ══════════════════════════════════
//
// Author:  g1veteran (365 days, 500 posts, good standing)
// Message: "The governance proposal for validator delegation has been submitted. Please review and vote before the deadline."
// Score:   0 | Triggered: []
// Status:  ✓ APPROVED
//
// SCENARIO 2: New Account Spammer
// ════════════════════════════════
//
// Author:  g1newspam (0 days, no username, low balance)
// Message: "FREE!!! AIRDROP!!! Claim your bonus prize now!!!"
// Score:   11 | Triggered: [EXCESSIVE_PUNCT NEW_ACCOUNT NO_USERNAME LOW_BALANCE BAYES_SPAM KEYWORD_SPAM]
// Rules breakdown:
//   NEW_ACCOUNT(2) + NO_USERNAME(1) + LOW_BALANCE(1) = reputation penalties
//   KEYWORD_SPAM(3) + BAYES_SPAM(3) = content detection
//   EXCESSIVE_PUNCT(1) = punctuation spam (>20% of characters)
// Status:  🚫 REJECTED (score >= 8)
//
// SCENARIO 3: Spam Flood from Flagged User
// ═════════════════════════════════════════
//
// Author:  g1flooder (50% flag ratio, 2 past bans, 30 posts/hour)
// Message: "FREE MONEY BONUS PRIZE AIRDROP CLICK NOW!!! http://spam1.com http://spam2.com http://spam3.com http://spam4.com"
// Score:   20 | Triggered: [ALL_CAPS LINK_HEAVY RATE_BURST BAD_REPUTATION BANNED_BEFORE BAYES_SPAM KEYWORD_SPAM]
// Rules breakdown:
//   ALL_CAPS(3) + LINK_HEAVY(2) = content heuristics
//   RATE_BURST(4) = posting frequency
//   BAD_REPUTATION(3) + BANNED_BEFORE(2) = reputation
//   KEYWORD_SPAM(3) + BAYES_SPAM(3) = vocabulary analysis
// Status:  🚫 REJECTED
//
// SCENARIO 4: Near-Duplicate with Pattern Match
// ══════════════════════════════════════════════
//
// Author:  g1scammer (good reputation, normal rate)
// Message: "Send 100 GNOT get 1000 GNOT back! Guaranteed returns! Act now!"
// Score:   12 | Triggered: [BLOCKED_PATTERN BAYES_SPAM NEAR_DUPLICATE]
// Rules breakdown:
//   BLOCKED_PATTERN(5) = regex matched 'send X get Y' scam pattern
//   NEAR_DUPLICATE(4) = fingerprint matched known spam
//   BAYES_SPAM(3) = corpus recognizes spam vocabulary
//   Even a trusted user gets caught when content matches multiple layers
// Status:  🚫 REJECTED
//
// SCENARIO 5: Short Link Spam
// ═══════════════════════════
//
// Author:  g1shortlink (good reputation)
// Message: "Check this! http://x.co"
// Score:   4 | Triggered: [EXCESSIVE_PUNCT SHORT_WITH_LINK]
// Rules breakdown:
//   SHORT_WITH_LINK(3) = body <=30 chars with a URL
//   EXCESSIVE_PUNCT(1) = URL punctuation (://) inflates the ratio
//   Minimal content with a link is a common spam pattern
// Status:  ⚠ SUSPICIOUS (below reject, above normal)
//
// SCENARIO 6: Blocked Address
// ═══════════════════════════
//
// Author:  g1blocked (BLOCKLISTED)
// Message: "The governance proposal for validator delegation has been submitted. Please review and vote before the deadline."
// Score:   99 | Triggered: [BLOCKED_ADDRESS]
// Rules breakdown:
//   BLOCKED_ADDRESS(99) = instant rejection, no further rules evaluated
//   Clean content and good reputation are irrelevant
// Status:  🚫 BLOCKED
//
// SCENARIO 7: Allowlisted Moderator
// ══════════════════════════════════
//
// Author:  g1admin (ALLOWLISTED)
// Message: "FREE MONEY BONUS PRIZE AIRDROP CLICK NOW!!! http://spam1.com http://spam2.com http://spam3.com http://spam4.com"
// Score:   0 | Triggered: []
// Rules breakdown:
//   Allowlist bypasses ALL rules - score is always 0
//   Use case: moderators, admins, emergency announcements
// Status:  ✓ BYPASSED
//
// SCENARIO 8: Unicode Abuse (Zalgo + Invisible + Homoglyph)
// ═════════════════════════════════════════════════════════
//
// Author:  g1griefer (good reputation, normal rate)
// Message: contains zalgo text, cyrillic homoglyph, zero-width spaces
// Score:   7 | Triggered: [ZALGO_TEXT INVISIBLE_CHARS HOMOGLYPH_MIX]
// Rules breakdown:
//   ZALGO_TEXT(3) = stacked combining marks distort rendering
//   INVISIBLE_CHARS(2) = zero-width spaces for obfuscation
//   HOMOGLYPH_MIX(2) = cyrillic 'а' mixed with latin in same word
//   Good reputation irrelevant - unicode abuse is a content signal
// Status:  ⚠ HIDDEN (score >= 5)
//
// RULE COVERAGE MATRIX
// ════════════════════
//
// Scenario  Score  Rules Triggered
// ────────  ─────  ───────────────
// 1. Legit     0  []
// 2. New acc  11  [EXCESSIVE_PUNCT NEW_ACCOUNT NO_USERNAME LOW_BALANCE BAYES_SPAM KEYWORD_SPAM]
// 3. Flood   20  [ALL_CAPS LINK_HEAVY RATE_BURST BAD_REPUTATION BANNED_BEFORE BAYES_SPAM KEYWORD_SPAM]
// 4. Dup+Pat  12  [BLOCKED_PATTERN BAYES_SPAM NEAR_DUPLICATE]
// 5. Short    4  [EXCESSIVE_PUNCT SHORT_WITH_LINK]
// 6. Blocked 99  [BLOCKED_ADDRESS]
// 7. Allowed   0  []
// 8. Unicode   7  [ZALGO_TEXT INVISIBLE_CHARS HOMOGLYPH_MIX]
//
// All 18 rules triggered at least once across these scenarios.
// No single rule is sufficient - multi-layer detection is the key.
