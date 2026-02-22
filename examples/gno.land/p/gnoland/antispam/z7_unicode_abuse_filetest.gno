package main

// Demonstration: Unicode Abuse Detection
// =======================================
// This test shows how the antispam system detects three forms
// of unicode abuse used by griefers to bypass text-based filters:
//
// 1. Zalgo text: Stacking combining diacritical marks to distort rendering
// 2. Invisible characters: Zero-width and directional override chars
// 3. Homoglyph mixing: Substituting Latin chars with Cyrillic/Greek look-alikes
//
// Also demonstrates leet-speak normalization for keyword detection
// (e.g. "fr33 a1rdr0p" matching "free" and "airdrop").

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

func containsRule(rules []string, target string) bool {
	for _, r := range rules {
		if r == target {
			return true
		}
	}
	return false
}

func main() {
	println("=== UNICODE ABUSE DETECTION DEMONSTRATION ===\n")

	// Setup: trusted account so only content rules trigger
	goodRep := antispam.ReputationData{
		AccountAgeDays: 90,
		Balance:        10000000000,
		HasUsername:    true,
		TotalAccepted:  50,
	}
	normalRate := antispam.RateState{PostCount: 3, WindowSeconds: 3600}
	bl := antispam.NewBlocklist()
	dict := antispam.NewKeywordDict()
	dict.BulkAdd("free:2\nairdrop:2\nspam:3\nscam:2\nbonus:1")

	score := func(content string) antispam.SpamScore {
		return antispam.Score(antispam.ScoreInput{
			Author:  "g1testuser",
			Content: content,
			Rate:    normalRate,
			Rep:     goodRep,
			Bl:      bl,
			Dict:    dict,
		})
	}

	// ========================================
	// PHASE 1: Zalgo Text Detection
	// ========================================

	println("PHASE 1: Zalgo Text")
	println("═══════════════════\n")

	println("Test 1: Normal accented text (should pass)")
	println("───────────────────────────────────────────")
	r := score("Café résumé naïve - normal accented text is fine.")
	println("Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered))
	println("")

	println("Test 2: Zalgo text with stacked combining marks")
	println("────────────────────────────────────────────────")
	// H with 5 combining marks = obvious zalgo
	zalgo := "H\u0300\u0301\u0302\u0303\u0304ello, this looks broken"
	r = score(zalgo)
	println("Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered))
	if containsRule(r.Triggered, "ZALGO_TEXT") {
		println("ZALGO_TEXT detected: stacked combining marks distort text rendering")
	}
	println("")

	// ========================================
	// PHASE 2: Invisible Characters
	// ========================================

	println("PHASE 2: Invisible Characters")
	println("═════════════════════════════\n")

	println("Test 3: Text with zero-width spaces (obfuscation)")
	println("──────────────────────────────────────────────────")
	invisible := "He\u200Bllo wo\u200Brld - looks normal but has hidden chars"
	r = score(invisible)
	println("Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered))
	if containsRule(r.Triggered, "INVISIBLE_CHARS") {
		println("INVISIBLE_CHARS detected: zero-width characters used for obfuscation")
	}
	println("")

	println("Test 4: RTL override attack")
	println("───────────────────────────")
	rtl := "Normal text \u202Ereversed\u202C hidden content"
	r = score(rtl)
	println("Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered))
	if containsRule(r.Triggered, "INVISIBLE_CHARS") {
		println("INVISIBLE_CHARS detected: directional override chars can hide content")
	}
	println("")

	// ========================================
	// PHASE 3: Homoglyph Mixing
	// ========================================

	println("PHASE 3: Homoglyph Script Mixing")
	println("════════════════════════════════\n")

	println("Test 5: Cyrillic 'a' substituted in Latin word")
	println("───────────────────────────────────────────────")
	// "p" + cyrillic а (U+0430) + "yment" - visually identical to "payment"
	homoglyph := "Send p\u0430yment to this address now"
	r = score(homoglyph)
	println("Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered))
	if containsRule(r.Triggered, "HOMOGLYPH_MIX") {
		println("HOMOGLYPH_MIX detected: mixed Latin/Cyrillic in single word")
	}
	println("")

	println("Test 6: Separate scripts (should pass)")
	println("──────────────────────────────────────")
	r = score("Hello Привет - two scripts in separate words is fine")
	println("Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered))
	println("")

	// ========================================
	// PHASE 4: Leet Speak + Keywords
	// ========================================

	println("PHASE 4: Leet Speak Bypass Attempt")
	println("══════════════════════════════════\n")

	println("Test 7: Leet speak trying to bypass keyword filter")
	println("──────────────────────────────────────────────────")
	leet := "Get fr33 a1rdr0p b0nu5 from this exclusive event"
	r = score(leet)
	println("Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered))
	if containsRule(r.Triggered, "KEYWORD_SPAM") {
		println("KEYWORD_SPAM detected: leet speak normalized before keyword check")
		println("  fr33 -> free, a1rdr0p -> airdrop, b0nu5 -> bonus")
	}
	println("")

	println("Test 8: Normal text with digits (should pass)")
	println("─────────────────────────────────────────────")
	r = score("Version 3.0 released with 42 improvements and 7 bug fixes.")
	println("Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered))
	println("")

	// ========================================
	// PHASE 5: Combined Attack
	// ========================================

	println("PHASE 5: Combined Unicode Attack")
	println("════════════════════════════════\n")

	println("Test 9: Zalgo + invisible + homoglyph in one message")
	println("────────────────────────────────────────────────────")
	combined := "S\u0300\u0301\u0302\u0303end p\u0430yment\u200B\u200Bnow"
	r = score(combined)
	println("Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered))
	println("All three unicode abuse rules triggered simultaneously.")
	println("")

	println("=== END ===")
}

// Output:
// === UNICODE ABUSE DETECTION DEMONSTRATION ===
//
// PHASE 1: Zalgo Text
// ═══════════════════
//
// Test 1: Normal accented text (should pass)
// ───────────────────────────────────────────
// Score: 0 | Rules: []
//
// Test 2: Zalgo text with stacked combining marks
// ────────────────────────────────────────────────
// Score: 3 | Rules: [ZALGO_TEXT]
// ZALGO_TEXT detected: stacked combining marks distort text rendering
//
// PHASE 2: Invisible Characters
// ═════════════════════════════
//
// Test 3: Text with zero-width spaces (obfuscation)
// ──────────────────────────────────────────────────
// Score: 2 | Rules: [INVISIBLE_CHARS]
// INVISIBLE_CHARS detected: zero-width characters used for obfuscation
//
// Test 4: RTL override attack
// ───────────────────────────
// Score: 2 | Rules: [INVISIBLE_CHARS]
// INVISIBLE_CHARS detected: directional override chars can hide content
//
// PHASE 3: Homoglyph Script Mixing
// ════════════════════════════════
//
// Test 5: Cyrillic 'a' substituted in Latin word
// ───────────────────────────────────────────────
// Score: 2 | Rules: [HOMOGLYPH_MIX]
// HOMOGLYPH_MIX detected: mixed Latin/Cyrillic in single word
//
// Test 6: Separate scripts (should pass)
// ──────────────────────────────────────
// Score: 0 | Rules: []
//
// PHASE 4: Leet Speak Bypass Attempt
// ══════════════════════════════════
//
// Test 7: Leet speak trying to bypass keyword filter
// ──────────────────────────────────────────────────
// Score: 3 | Rules: [KEYWORD_SPAM]
// KEYWORD_SPAM detected: leet speak normalized before keyword check
//   fr33 -> free, a1rdr0p -> airdrop, b0nu5 -> bonus
//
// Test 8: Normal text with digits (should pass)
// ─────────────────────────────────────────────
// Score: 0 | Rules: []
//
// PHASE 5: Combined Unicode Attack
// ════════════════════════════════
//
// Test 9: Zalgo + invisible + homoglyph in one message
// ────────────────────────────────────────────────────
// Score: 7 | Rules: [ZALGO_TEXT INVISIBLE_CHARS HOMOGLYPH_MIX]
// All three unicode abuse rules triggered simultaneously.
//
// === END ===
