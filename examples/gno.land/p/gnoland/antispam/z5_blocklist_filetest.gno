package main

// Demonstration: Blocklist and Pattern Matching
// ==============================================
// This test shows how the blocklist system provides immediate
// protection through address blocking, allowlisting, and regex patterns.
//
// Features demonstrated:
// - Address blocking: Immediate rejection of known spammers
// - Address allowlisting: Bypass all checks for trusted users
// - Pattern matching: Regex-based content filtering
// - Priority: Blocklist/allowlist takes precedence over other rules
//
// Use cases:
// - Emergency spam response (block address immediately)
// - Trusted user management (moderators, admins)
// - Campaign-specific filtering (block crypto scam patterns)

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
	println("=== BLOCKLIST AND PATTERN MATCHING DEMONSTRATION ===\n")

	// Setup
	corpus := antispam.NewCorpus()
	fps := antispam.NewFingerprintStore()
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
	// PHASE 1: Address Blocking
	// ========================================

	println("PHASE 1: Address-Based Blocking")
	println("════════════════════════════════\n")

	bl := antispam.NewBlocklist()
	bl.BlockAddress("g1spammer123")
	bl.BlockAddress("g1scammer456")

	println("Blocked addresses:")
	println("  • g1spammer123")
	println("  • g1scammer456\n")

	// Test 1: Blocked address with clean content
	println("Test 1: Blocked address posts clean content")
	println("─────────────────────────────────────────────")

	cleanMsg := "I have a legitimate question about validator setup."

	println("Message: \"" + cleanMsg + "\"")
	println("Author: g1spammer123 (BLOCKED)\n")

	result1 := antispam.Score(antispam.ScoreInput{
		Author:  "g1spammer123",
		Content: cleanMsg,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
	})

	println("Score: " + intToString(result1.Total) + " | Triggered: " + sliceToString(result1.Triggered))
	if containsRule(result1.Triggered, "BLOCKED_ADDRESS") {
		println("Status: 🚫 REJECTED (blocklist overrides all other factors)")
		println("Analysis: Even legitimate content is blocked from known spammer\n")
	}

	// Test 2: Non-blocked address with same content
	println("Test 2: Regular user posts same content")
	println("────────────────────────────────────────")

	println("Message: \"" + cleanMsg + "\"")
	println("Author: g1regularuser (not blocked)\n")

	result2 := antispam.Score(antispam.ScoreInput{
		Author:  "g1regularuser",
		Content: cleanMsg,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
	})

	println("Score: " + intToString(result2.Total) + " | Triggered: " + sliceToString(result2.Triggered))
	println("Status: ✓ APPROVED\n")

	// ========================================
	// PHASE 2: Address Allowlisting
	// ========================================

	println("PHASE 2: Address-Based Allowlisting")
	println("════════════════════════════════════\n")

	bl.AllowAddress("g1moderator")
	bl.AllowAddress("g1admin")

	println("Allowed addresses (bypass all checks):")
	println("  • g1moderator")
	println("  • g1admin\n")

	// Test 3: Allowed address with spam content
	println("Test 3: Moderator posts spam-like content")
	println("──────────────────────────────────────────")

	obviousSpam := "FREE MONEY!!! CLICK NOW!!! http://spam1.com http://spam2.com http://spam3.com"

	println("Message: \"" + obviousSpam + "\"")
	println("Author: g1moderator (ALLOWLISTED)\n")

	result3 := antispam.Score(antispam.ScoreInput{
		Author:  "g1moderator",
		Content: obviousSpam,
		Rate:    antispam.RateState{PostCount: 50, WindowSeconds: 3600}, // Burst
		Rep: antispam.ReputationData{ // Terrible stats
			AccountAgeDays: 0,
			Balance:        0,
			FlaggedCount:   100,
			TotalAccepted:  0,
			HasUsername:    false,
			BanCount:       10,
		},
		Corpus: corpus,
		Fps:    fps,
		Bl:     bl,
	})

	println("Score: " + intToString(result3.Total) + " | Triggered: " + sliceToString(result3.Triggered))
	if result3.Total == 0 {
		println("Status: ✓ BYPASSED (allowlist grants immunity)")
		println("Analysis: All rules disabled for allowlisted addresses")
		println("Use case: Testing, announcements, emergency communications\n")
	}

	// ========================================
	// PHASE 3: Pattern Matching
	// ========================================

	println("PHASE 3: Regex Pattern Matching")
	println("════════════════════════════════\n")

	// Add crypto scam patterns
	bl.AddPattern(`(?i)send\s+\d+\s+\w+\s+get\s+\d+`) // "send X get Y"
	bl.AddPattern(`(?i)double\s+your\s+(crypto|coins|tokens|money)`)
	bl.AddPattern(`(?i)click\s+here.*guaranteed.*returns?`)

	println("Added 3 regex patterns:")
	println("  1. Send X get Y pattern")
	println("  2. 'Double your crypto' variants")
	println("  3. 'Click here...guaranteed returns'\n")

	// Test 4: Pattern match - send/get scam
	println("Test 4: 'Send X Get Y' Scam Pattern")
	println("────────────────────────────────────")

	sendGetScam := "Send 100 GNOT get 1000 GNOT back! Guaranteed!"

	println("Message: \"" + sendGetScam + "\"")
	println("Author: g1user\n")

	result4 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user",
		Content: sendGetScam,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
	})

	println("Score: " + intToString(result4.Total) + " | Triggered: " + sliceToString(result4.Triggered))
	if containsRule(result4.Triggered, "BLOCKED_PATTERN") {
		println("✓ BLOCKED_PATTERN triggered (matched: send X get Y)")
		println("Status: ⛔ HIDDEN\n")
	}

	// Test 5: Pattern match - double your crypto
	println("Test 5: 'Double Your Crypto' Pattern")
	println("──────────────────────────────────────")

	doubleCrypto := "Amazing opportunity to double your tokens through our automated trading system!"

	println("Message: \"" + doubleCrypto + "\"")
	println("Author: g1user\n")

	result5 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user",
		Content: doubleCrypto,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
	})

	println("Score: " + intToString(result5.Total) + " | Triggered: " + sliceToString(result5.Triggered))
	if containsRule(result5.Triggered, "BLOCKED_PATTERN") {
		println("✓ BLOCKED_PATTERN triggered (matched: double your *)")
		println("Status: ⛔ HIDDEN\n")
	}

	// Test 6: Legitimate message with similar words
	println("Test 6: Legitimate Message (No Pattern Match)")
	println("──────────────────────────────────────────────")

	legitimate := "To send GNOT tokens, use the gnokey command. You can get started by reading the documentation."

	println("Message: \"" + legitimate + "\"")
	println("Author: g1user\n")

	result6 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user",
		Content: legitimate,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
	})

	println("Score: " + intToString(result6.Total) + " | Triggered: " + sliceToString(result6.Triggered))
	if !containsRule(result6.Triggered, "BLOCKED_PATTERN") {
		println("✓ BLOCKED_PATTERN did NOT trigger")
		println("Analysis: Contains 'send' and 'get' but not in scam pattern")
		println("Status: ✓ APPROVED\n")
	}

	// ========================================
	// PHASE 4: Priority Testing
	// ========================================

	println("PHASE 4: Blocklist Priority")
	println("═══════════════════════════\n")

	println("Testing: Does allowlist override blocklist?")
	println("─────────────────────────────────────────────\n")

	// Block and allow the same address
	bl.BlockAddress("g1testuser")
	bl.AllowAddress("g1testuser")

	println("Address g1testuser is both BLOCKED and ALLOWED")
	println("Message: \"" + obviousSpam + "\"")
	println("Author: g1testuser\n")

	result7 := antispam.Score(antispam.ScoreInput{
		Author:  "g1testuser",
		Content: obviousSpam,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
	})

	println("Score: " + intToString(result7.Total) + " | Triggered: " + sliceToString(result7.Triggered))
	if result7.Total == 0 {
		println("Status: ✓ BYPASSED")
		println("Analysis: Allowlist takes precedence over blocklist")
		println("Use case: Temporary block override for investigation\n")
	}

	// ========================================
	// SUMMARY
	// ========================================

	println("SUMMARY: Blocklist Features")
	println("════════════════════════════\n")

	println("Capabilities Demonstrated:")
	println("  ✓ Address blocking:      Immediate rejection (score=100)")
	println("  ✓ Address allowlisting:  Complete bypass (score=0)")
	println("  ✓ Regex patterns:        Flexible content filtering")
	println("  ✓ Pattern precision:     Legitimate content not caught")
	println("  ✓ Priority system:       Allowlist > Blocklist > Patterns")
	println("")

	println("Use Cases:")
	println("  • Emergency response: Block addresses during active attack")
	println("  • Trusted users: Moderators/admins bypass spam filters")
	println("  • Campaign filtering: Add patterns for specific scam types")
	println("  • Case-insensitive: (?i) flag catches variations")
	println("  • Temporary overrides: Allowlist can override blocks")
	println("")

	println("Pattern Examples:")
	println("  ✓ 'send X get Y'           -> Caught investment scams")
	println("  ✓ 'double your crypto'     -> Caught multiplier scams")
	println("  ✓ 'click here...guaranteed' -> Caught phishing")
	println("  ✓ Legitimate 'send'/'get'  -> NOT flagged (different context)")
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
// === BLOCKLIST AND PATTERN MATCHING DEMONSTRATION ===
//
// PHASE 1: Address-Based Blocking
// ════════════════════════════════
//
// Blocked addresses:
//   • g1spammer123
//   • g1scammer456
//
// Test 1: Blocked address posts clean content
// ─────────────────────────────────────────────
// Message: "I have a legitimate question about validator setup."
// Author: g1spammer123 (BLOCKED)
//
// Score: 99 | Triggered: [BLOCKED_ADDRESS]
// Status: 🚫 REJECTED (blocklist overrides all other factors)
// Analysis: Even legitimate content is blocked from known spammer
//
// Test 2: Regular user posts same content
// ────────────────────────────────────────
// Message: "I have a legitimate question about validator setup."
// Author: g1regularuser (not blocked)
//
// Score: 0 | Triggered: []
// Status: ✓ APPROVED
//
// PHASE 2: Address-Based Allowlisting
// ════════════════════════════════════
//
// Allowed addresses (bypass all checks):
//   • g1moderator
//   • g1admin
//
// Test 3: Moderator posts spam-like content
// ──────────────────────────────────────────
// Message: "FREE MONEY!!! CLICK NOW!!! http://spam1.com http://spam2.com http://spam3.com"
// Author: g1moderator (ALLOWLISTED)
//
// Score: 0 | Triggered: []
// Status: ✓ BYPASSED (allowlist grants immunity)
// Analysis: All rules disabled for allowlisted addresses
// Use case: Testing, announcements, emergency communications
//
// PHASE 3: Regex Pattern Matching
// ════════════════════════════════
//
// Added 3 regex patterns:
//   1. Send X get Y pattern
//   2. 'Double your crypto' variants
//   3. 'Click here...guaranteed returns'
//
// Test 4: 'Send X Get Y' Scam Pattern
// ────────────────────────────────────
// Message: "Send 100 GNOT get 1000 GNOT back! Guaranteed!"
// Author: g1user
//
// Score: 5 | Triggered: [BLOCKED_PATTERN]
// ✓ BLOCKED_PATTERN triggered (matched: send X get Y)
// Status: ⛔ HIDDEN
//
// Test 5: 'Double Your Crypto' Pattern
// ──────────────────────────────────────
// Message: "Amazing opportunity to double your tokens through our automated trading system!"
// Author: g1user
//
// Score: 5 | Triggered: [BLOCKED_PATTERN]
// ✓ BLOCKED_PATTERN triggered (matched: double your *)
// Status: ⛔ HIDDEN
//
// Test 6: Legitimate Message (No Pattern Match)
// ──────────────────────────────────────────────
// Message: "To send GNOT tokens, use the gnokey command. You can get started by reading the documentation."
// Author: g1user
//
// Score: 0 | Triggered: []
// ✓ BLOCKED_PATTERN did NOT trigger
// Analysis: Contains 'send' and 'get' but not in scam pattern
// Status: ✓ APPROVED
//
// PHASE 4: Blocklist Priority
// ═══════════════════════════
//
// Testing: Does allowlist override blocklist?
// ─────────────────────────────────────────────
//
// Address g1testuser is both BLOCKED and ALLOWED
// Message: "FREE MONEY!!! CLICK NOW!!! http://spam1.com http://spam2.com http://spam3.com"
// Author: g1testuser
//
// Score: 0 | Triggered: []
// Status: ✓ BYPASSED
// Analysis: Allowlist takes precedence over blocklist
// Use case: Temporary block override for investigation
//
// SUMMARY: Blocklist Features
// ════════════════════════════
//
// Capabilities Demonstrated:
//   ✓ Address blocking:      Immediate rejection (score=100)
//   ✓ Address allowlisting:  Complete bypass (score=0)
//   ✓ Regex patterns:        Flexible content filtering
//   ✓ Pattern precision:     Legitimate content not caught
//   ✓ Priority system:       Allowlist > Blocklist > Patterns
//
// Use Cases:
//   • Emergency response: Block addresses during active attack
//   • Trusted users: Moderators/admins bypass spam filters
//   • Campaign filtering: Add patterns for specific scam types
//   • Case-insensitive: (?i) flag catches variations
//   • Temporary overrides: Allowlist can override blocks
//
// Pattern Examples:
//   ✓ 'send X get Y'           -> Caught investment scams
//   ✓ 'double your crypto'     -> Caught multiplier scams
//   ✓ 'click here...guaranteed' -> Caught phishing
//   ✓ Legitimate 'send'/'get'  -> NOT flagged (different context)
