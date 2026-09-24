package antispam

import "strings"

// Thresholds

const (
	// zalgoMinCombining is the minimum consecutive combining marks to trigger ZALGO_TEXT.
	// Normal accented text has 1-2 marks per character; 3+ is zalgo abuse.
	zalgoMinCombining = 3

	// invisibleMinCount is the minimum invisible characters to trigger INVISIBLE_CHARS.
	// A single invisible char could be copy-paste artifact; 2+ is suspicious.
	invisibleMinCount = 2
)

// ScoreUnicode checks content for unicode abuse: zalgo text, invisible characters,
// and homoglyph script mixing.
func ScoreUnicode(content string) []RuleHit {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil
	}
	return hitsFromUnicodeStats(analyzeContent(content))
}

// Character classification helpers

// isCombiningMark returns true for Unicode combining diacritical marks.
func isCombiningMark(r rune) bool {
	return (r >= 0x0300 && r <= 0x036F) ||
		(r >= 0x1AB0 && r <= 0x1AFF) ||
		(r >= 0x1DC0 && r <= 0x1DFF) ||
		(r >= 0xFE20 && r <= 0xFE2F)
}

// isInvisibleChar returns true for zero-width, directional override,
// and other invisible Unicode characters used for text obfuscation.
func isInvisibleChar(r rune) bool {
	switch r {
	case 0x200B, 0x200C, 0x200D, 0x200E, 0x200F,
		0x202A, 0x202B, 0x202C, 0x202D, 0x202E,
		0x2060, 0x2061, 0x2062, 0x2063, 0x2064,
		0xFEFF, 0x00AD, 0x034F:
		return true
	}
	return false
}

// isLatin returns true for Latin script characters (Basic Latin + Extended).
func isLatin(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= 0x00C0 && r <= 0x024F)
}

// isCyrillic returns true for Cyrillic script characters.
func isCyrillic(r rune) bool {
	return r >= 0x0400 && r <= 0x04FF
}

// isGreek returns true for Greek script characters.
func isGreek(r rune) bool {
	return r >= 0x0370 && r <= 0x03FF
}
