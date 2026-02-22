package antispam

import (
	"strings"

	"gno.land/p/nt/avl"
)

const (
	// keywordMinMatches is the minimum number of distinct keyword hits
	// needed to trigger the rule. Prevents single-word false positives.
	keywordMinMatches = 2

	// keywordSumThreshold is the minimum combined weight of matched keywords
	// needed to trigger the rule. Higher-weight words reach it faster.
	keywordSumThreshold = 4

	// keywordMaxSize caps the number of keywords in the dictionary.
	// Existing keywords still get their weights updated; only new
	// keywords are rejected when the cap is reached.
	keywordMaxSize = 10000
)

// ---- KeywordDict

// KeywordDict stores spam keywords with per-word weights.
type KeywordDict struct {
	tree *avl.Tree
}

// NewKeywordDict creates an empty KeywordDict.
func NewKeywordDict() *KeywordDict {
	return &KeywordDict{tree: avl.NewTree()}
}

// Add inserts a keyword with a weight. The word is lowercased.
// Weights are clamped to [1, 3]; values <=0 are ignored, values >3 become 3.
// When the dictionary reaches keywordMaxSize, new keywords are ignored
// but existing keywords still get their weights updated.
func (kd *KeywordDict) Add(word string, weight int) {
	w := strings.ToLower(strings.TrimSpace(word))
	if w == "" || weight <= 0 {
		return
	}
	if weight > 3 {
		weight = 3
	}
	if kd.tree.Size() >= keywordMaxSize && !kd.tree.Has(w) {
		return
	}
	kd.tree.Set(w, weight)
}

// Remove deletes a keyword from the dictionary.
func (kd *KeywordDict) Remove(word string) {
	kd.tree.Remove(strings.ToLower(strings.TrimSpace(word)))
}

// GetWeight returns the weight for a keyword and whether it exists.
// Input is normalized (lowercased, trimmed) before lookup.
func (kd *KeywordDict) GetWeight(word string) (int, bool) {
	return kd.getWeightDirect(strings.ToLower(strings.TrimSpace(word)))
}

// getWeightDirect looks up a pre-normalized keyword without additional
// string processing. Used by internal scoring paths where tokens are
// already lowercase and trimmed from Tokenize().
func (kd *KeywordDict) getWeightDirect(word string) (int, bool) {
	v, ok := kd.tree.Get(word)
	if !ok {
		return 0, false
	}
	return v.(int), true
}

// Size returns the number of keywords in the dictionary.
func (kd *KeywordDict) Size() int {
	return kd.tree.Size()
}

// BulkAdd parses a newline-separated list of "word:weight" pairs.
// If :weight is omitted, default weight 1 is used.
// Empty lines and whitespace are ignored.
func (kd *KeywordDict) BulkAdd(data string) {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		idx := strings.LastIndex(line, ":")
		if idx == -1 {
			kd.Add(line, 1)
			continue
		}
		if idx == len(line)-1 {
			// Trailing colon with no weight -> strip colon, default weight 1
			kd.Add(line[:idx], 1)
			continue
		}

		word := line[:idx]
		weightStr := line[idx+1:]
		weight := parseSimpleInt(weightStr)
		if weight <= 0 {
			weight = 1
		}
		kd.Add(word, weight)
	}
}

// parseSimpleInt parses a small positive integer from a string.
// Returns early once the value exceeds 9 (sufficient for keyword weights
// which are clamped to [1, 3]). Returns 0 on any error.
func parseSimpleInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
		if n > 9 {
			return n
		}
	}
	return n
}

// ---- Leet speak normalization

// normalizeLeet converts common leet speak digit substitutions back to letters.
// Only handles digit-based leet (0->o, 1->i, 3->e, 4->a, 5->s, 7->t) since
// symbol-based leet (@, $) is stripped during tokenization.
func normalizeLeet(s string) string {
	if s == "" {
		return s
	}
	hasDigit := false
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '0':
			b.WriteByte('o')
		case '1':
			b.WriteByte('i')
		case '3':
			b.WriteByte('e')
		case '4':
			b.WriteByte('a')
		case '5':
			b.WriteByte('s')
		case '7':
			b.WriteByte('t')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// ---- Scoring

// ScoreKeywords checks content tokens against a keyword dictionary.
// Uses co-occurrence: triggers only when multiple keywords are found
// AND their combined weight exceeds the threshold.
// Also checks leet-speak normalized variants of each token.
func ScoreKeywords(content string, dict *KeywordDict) (int, string) {
	return scoreKeywordsTokens(dedup(Tokenize(content)), dict)
}

// scoreKeywordsTokens scores pre-deduplicated tokens against the dictionary.
// For each token, checks both the original and leet-normalized variant.
// Uses getWeightDirect since tokens from Tokenize() are already lowercase.
// Exits early once both co-occurrence thresholds are met.
func scoreKeywordsTokens(tokens []string, dict *KeywordDict) (int, string) {
	if dict == nil || dict.Size() == 0 {
		return 0, ""
	}
	if len(tokens) == 0 {
		return 0, ""
	}

	weightSum := 0
	matches := 0

	for _, tok := range tokens {
		w, ok := dict.getWeightDirect(tok)
		if !ok {
			// Try leet-normalized variant.
			norm := normalizeLeet(tok)
			if norm != tok {
				w, ok = dict.getWeightDirect(norm)
			}
		}
		if ok {
			weightSum += w
			matches++
			if matches >= keywordMinMatches && weightSum >= keywordSumThreshold {
				return WeightKeyword, RuleKeywordSpam
			}
		}
	}

	return 0, ""
}
