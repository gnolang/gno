package antispam

import (
	"strings"
	"unicode"
)

// ---- Thresholds

const (
	capsThresholdPct  = 50 // % of letters that are uppercase
	punctThresholdPct = 20 // % of characters that are punctuation
	linkMaxCount      = 3  // links above this count trigger LINK_HEAVY
	shortBodyMaxLen   = 30 // body length below this with a link triggers SHORT_WITH_LINK
	minLettersForCaps = 10 // minimum letters to evaluate ALL_CAPS (avoid false positives on short text)
	minCharsForPunct  = 10 // minimum characters to evaluate punctuation
)

// ScoreContent checks content for spam heuristics.
// Returns individual RuleHit entries for each triggered rule.
func ScoreContent(content string) []RuleHit {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil
	}

	var hits []RuleHit

	// Single-pass: count upper/lower letters, punctuation, and URL occurrences,
	// skipping URL ranges for letter counting (avoids allocating a stripped copy).
	upper, lower, punct, nonSpace, urlCount := countContentStats(content)

	// ALL_CAPS check
	totalLetters := upper + lower
	if totalLetters >= minLettersForCaps && upper*100 > capsThresholdPct*totalLetters {
		hits = append(hits, RuleHit{WeightAllCaps, "ALL_CAPS"})
	}

	// Excessive punctuation check
	if nonSpace >= minCharsForPunct && punct*100 > punctThresholdPct*nonSpace {
		hits = append(hits, RuleHit{WeightExcessivePunct, "EXCESSIVE_PUNCT"})
	}

	// Link-heavy check (uses pre-computed URL count)
	if urlCount > linkMaxCount {
		hits = append(hits, RuleHit{WeightLinkHeavy, "LINK_HEAVY"})
	}

	// Short body with link
	if len(content) <= shortBodyMaxLen && urlCount > 0 {
		hits = append(hits, RuleHit{WeightShortWithLink, "SHORT_WITH_LINK"})
	}

	return hits
}

// countContentStats performs a single pass over the content, counting:
// - upper: uppercase letters (excluding URL ranges)
// - lower: lowercase letters (excluding URL ranges)
// - punct: punctuation characters
// - nonSpace: all non-space characters
// - urls: number of http:// and https:// URLs
// URLs are skipped for letter counting to avoid skewing the caps ratio.
func countContentStats(content string) (upper, lower, punct, nonSpace, urls int) {
	s := content
	pos := 0
	inURL := false

	for pos < len(s) {
		// Check if we're at a URL start (guarded by cheap byte check)
		if !inURL && s[pos] == 'h' && pos+7 < len(s) {
			rest := s[pos:]
			if strings.HasPrefix(rest, "https://") || strings.HasPrefix(rest, "http://") {
				inURL = true
				urls++
			}
		}

		r := rune(s[pos])
		advance := 1
		// Handle multi-byte runes
		if r >= 0x80 {
			for _, rr := range s[pos:] {
				r = rr
				advance = runeLen(rr)
				break
			}
		}

		if unicode.IsSpace(r) {
			if inURL {
				inURL = false // URL ends at whitespace
			}
		} else {
			nonSpace++
			if isPunct(r) {
				punct++
			}
			// Count letters only outside URLs
			if !inURL {
				if unicode.IsUpper(r) {
					upper++
				} else if unicode.IsLower(r) {
					lower++
				}
			}
		}
		pos += advance
	}
	return
}

// runeLen returns the UTF-8 byte length of a rune without allocating.
func runeLen(r rune) int {
	switch {
	case r < 0x80:
		return 1
	case r < 0x800:
		return 2
	case r < 0x10000:
		return 3
	default:
		return 4
	}
}

// ---- Helpers

// isPunct returns true if the rune is a common punctuation mark used in spam.
func isPunct(r rune) bool {
	switch r {
	case '!', '?', '.', ',', ';', ':', '-', '(', ')', '[', ']', '{', '}', '"', '\'', '`', '*', '~', '|', '/', '\\', '#', '@', '$', '%', '^', '&', '+', '=', '<', '>':
		return true
	}
	return false
}
