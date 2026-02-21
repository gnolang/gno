package antispam

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ---- Thresholds

const (
	capsThresholdPct  = 50 // % of letters that are uppercase
	punctThresholdPct = 20 // % of characters that are punctuation
	linkMaxCount      = 3  // links above this count trigger LINK_HEAVY
	shortBodyMaxLen   = 30 // body byte length below this with a link triggers SHORT_WITH_LINK
	minLettersForCaps = 10 // minimum letters to evaluate ALL_CAPS (avoid false positives on short text)
	minCharsForPunct  = 10 // minimum characters to evaluate punctuation
)

// contentStats holds all counters from content analysis.
type contentStats struct {
	upper, lower     int // letter case counts (URLs excluded)
	punct, nonSpace  int // punctuation and non-space counts (URLs excluded)
	urlCount         int // number of http/https URLs
	invisibleCount   int // zero-width and directional override characters
	maxConsecCombine int // longest run of combining marks on one character
	mixedWords       int // words mixing Latin + Cyrillic/Greek
}

// ScoreContent checks content for spam heuristics (caps, punct, links).
func ScoreContent(content string) []RuleHit {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil
	}
	s := analyzeContent(content)
	return hitsFromContentStats(content, s)
}

// scoreContentAll combines content heuristics and unicode abuse detection.
func scoreContentAll(content string) []RuleHit {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil
	}

	s := analyzeContent(content)

	hits := hitsFromContentStats(content, s)
	hits = append(hits, hitsFromUnicodeStats(s)...)
	return hits
}

// hitsFromContentStats derives content heuristic hits from analyzed stats.
func hitsFromContentStats(content string, s contentStats) []RuleHit {
	var hits []RuleHit

	totalLetters := s.upper + s.lower
	if totalLetters >= minLettersForCaps && s.upper*100 > capsThresholdPct*totalLetters {
		hits = append(hits, RuleHit{WeightAllCaps, RuleAllCaps})
	}
	if s.nonSpace >= minCharsForPunct && s.punct*100 > punctThresholdPct*s.nonSpace {
		hits = append(hits, RuleHit{WeightExcessivePunct, RuleExcessivePunct})
	}
	if s.urlCount > linkMaxCount {
		hits = append(hits, RuleHit{WeightLinkHeavy, RuleLinkHeavy})
	}
	if len(content) <= shortBodyMaxLen && s.urlCount > 0 {
		hits = append(hits, RuleHit{WeightShortWithLink, RuleShortWithLink})
	}
	return hits
}

// hitsFromUnicodeStats derives unicode abuse hits from analyzed stats.
func hitsFromUnicodeStats(s contentStats) []RuleHit {
	var hits []RuleHit

	if s.maxConsecCombine >= zalgoMinCombining {
		hits = append(hits, RuleHit{WeightZalgoText, RuleZalgoText})
	}
	if s.invisibleCount >= invisibleMinCount {
		hits = append(hits, RuleHit{WeightInvisibleChars, RuleInvisibleChars})
	}
	if s.mixedWords > 0 {
		hits = append(hits, RuleHit{WeightHomoglyphMix, RuleHomoglyphMix})
	}
	return hits
}

// analyzeContent collects all stats for content heuristics and unicode abuse detection.
// URLs are skipped for letter, punctuation, and nonSpace counting (to avoid
// skewing caps and punct ratios with URL syntax characters like ://?=&#).
// Script tracking continues inside URLs to detect homoglyphs in IDN domains.
func analyzeContent(content string) contentStats {
	var s contentStats
	pos := 0
	inURL := false

	consecCombining := 0
	wordHasLatin := false
	wordHasOther := false

	for pos < len(content) {
		if !inURL && content[pos] == 'h' && pos+7 < len(content) {
			rest := content[pos:]
			if strings.HasPrefix(rest, "https://") || strings.HasPrefix(rest, "http://") {
				inURL = true
				s.urlCount++
			}
		}

		r, advance := utf8.DecodeRuneInString(content[pos:])

		if isCombiningMark(r) {
			consecCombining++
			if consecCombining > s.maxConsecCombine {
				s.maxConsecCombine = consecCombining
			}
			pos += advance
			continue
		}

		if isInvisibleChar(r) {
			s.invisibleCount++
			consecCombining = 0
			pos += advance
			continue
		}

		consecCombining = 0

		if unicode.IsSpace(r) {
			if inURL {
				inURL = false
			}
			if wordHasLatin && wordHasOther {
				s.mixedWords++
			}
			wordHasLatin = false
			wordHasOther = false
		} else {
			if !inURL {
				s.nonSpace++
				if isPunct(r) {
					s.punct++
				}
				if unicode.IsUpper(r) {
					s.upper++
				} else if unicode.IsLower(r) {
					s.lower++
				}
			}
			// Script tracking runs inside URLs to detect homoglyphs in IDN domains.
			if isLatin(r) {
				wordHasLatin = true
			} else if isCyrillic(r) || isGreek(r) {
				wordHasOther = true
			}
		}
		pos += advance
	}

	if wordHasLatin && wordHasOther {
		s.mixedWords++
	}

	return s
}

// ---- Helpers

// isPunct returns true if the rune is a common punctuation mark.
func isPunct(r rune) bool {
	switch r {
	case '!', '?', '.', ',', ';', ':', '-', '(', ')', '[', ']', '{', '}', '"', '\'', '`', '*', '~', '|', '/', '\\', '#', '@', '$', '%', '^', '&', '+', '=', '<', '>':
		return true
	}
	return false
}
