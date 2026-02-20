package antispam

import (
	"strings"
	"unicode"
)

// ---- Rule weights

const (
	WeightAllCaps       = 3
	WeightExcessivePunct = 1
	WeightLinkHeavy     = 2
	WeightShortWithLink = 3
)

// ---- Thresholds

const (
	capsThresholdPct     = 50 // % of letters that are uppercase
	punctThresholdPct    = 20 // % of characters that are punctuation
	linkMaxCount         = 3  // links above this count trigger LINK_HEAVY
	shortBodyMaxLen      = 30 // body length below this with a link triggers SHORT_WITH_LINK
	minLettersForCaps    = 10 // minimum letters to evaluate ALL_CAPS (avoid false positives on short text)
	minCharsForPunct     = 10 // minimum characters to evaluate punctuation
)

// ScoreContent checks content for spam heuristics.
// Returns the combined score and the name of the highest-scoring rule.
func ScoreContent(content string) (int, string) {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return 0, ""
	}

	totalScore := 0
	topScore := 0
	topRule := ""

	add := func(score int, rule string) {
		totalScore += score
		if score > topScore {
			topScore = score
			topRule = rule
		}
	}

	if s := scoreAllCaps(content); s > 0 {
		add(s, "ALL_CAPS")
	}
	if s := scoreExcessivePunct(content); s > 0 {
		add(s, "EXCESSIVE_PUNCT")
	}
	if s := scoreLinkHeavy(content); s > 0 {
		add(s, "LINK_HEAVY")
	}
	if s := scoreShortWithLink(content); s > 0 {
		add(s, "SHORT_WITH_LINK")
	}

	return totalScore, topRule
}

// scoreAllCaps returns WeightAllCaps if more than capsThresholdPct of letters are uppercase.
// URLs are stripped before evaluation since they always contain lowercase.
func scoreAllCaps(content string) int {
	stripped := stripURLs(content)
	upper := 0
	lower := 0
	for _, r := range stripped {
		if unicode.IsUpper(r) {
			upper++
		} else if unicode.IsLower(r) {
			lower++
		}
	}
	total := upper + lower
	if total < minLettersForCaps {
		return 0
	}
	// upper * 100 / total > capsThresholdPct
	// rewrite to avoid division: upper * 100 > capsThresholdPct * total
	if upper*100 > capsThresholdPct*total {
		return WeightAllCaps
	}
	return 0
}

// scoreExcessivePunct returns WeightExcessivePunct if more than punctThresholdPct
// of characters are punctuation marks.
func scoreExcessivePunct(content string) int {
	punct := 0
	total := 0
	for _, r := range content {
		if unicode.IsSpace(r) {
			continue
		}
		total++
		if isPunct(r) {
			punct++
		}
	}
	if total < minCharsForPunct {
		return 0
	}
	if punct*100 > punctThresholdPct*total {
		return WeightExcessivePunct
	}
	return 0
}

// scoreLinkHeavy returns WeightLinkHeavy if the content contains more than linkMaxCount URLs.
func scoreLinkHeavy(content string) int {
	count := countURLs(content)
	if count > linkMaxCount {
		return WeightLinkHeavy
	}
	return 0
}

// scoreShortWithLink returns WeightShortWithLink if the content is short and contains a URL.
func scoreShortWithLink(content string) int {
	if len(content) > shortBodyMaxLen {
		return 0
	}
	if countURLs(content) > 0 {
		return WeightShortWithLink
	}
	return 0
}

// ---- Helpers

// countURLs counts http:// and https:// URL occurrences in text.
func countURLs(text string) int {
	count := 0
	s := text
	for {
		i := strings.Index(s, "http")
		if i == -1 {
			break
		}
		rest := s[i:]
		if strings.HasPrefix(rest, "https://") || strings.HasPrefix(rest, "http://") {
			count++
			// advance past the protocol prefix to avoid double-counting
			s = s[i+7:]
		} else {
			// "http" found but not a URL (e.g. "httpHeaderValue")
			s = s[i+4:]
		}
	}
	return count
}

// stripURLs removes http:// and https:// URLs from text to avoid
// skewing letter-ratio checks (URLs always contain lowercase).
func stripURLs(text string) string {
	var result strings.Builder
	s := text
	for {
		i := strings.Index(s, "http")
		if i == -1 {
			result.WriteString(s)
			break
		}
		rest := s[i:]
		if strings.HasPrefix(rest, "https://") || strings.HasPrefix(rest, "http://") {
			result.WriteString(s[:i])
			// skip past the URL (until whitespace or end)
			end := strings.IndexFunc(rest, unicode.IsSpace)
			if end == -1 {
				break
			}
			s = rest[end:]
		} else {
			result.WriteString(s[:i+4])
			s = s[i+4:]
		}
	}
	return result.String()
}

// isPunct returns true if the rune is a common punctuation mark used in spam.
func isPunct(r rune) bool {
	switch r {
	case '!', '?', '.', ',', ';', ':', '-', '(', ')', '[', ']', '{', '}', '"', '\'', '`', '*', '~', '|', '/', '\\', '#', '@', '$', '%', '^', '&', '+', '=', '<', '>':
		return true
	}
	return false
}
