package antispam

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"gno.land/p/nt/avl"
)

const (
	// tokenMinLen is the minimum token length to keep (filters "a", "is", "to", etc.)
	tokenMinLen = 3

	// bayesMinCorpusSize is the minimum number of unique tokens before Bayes scoring activates.
	// Prevents scoring with insufficient training data.
	bayesMinCorpusSize = 10

	// bayesSpamThresholdPct is the threshold for P(spam|tokens) expressed as integer percentage.
	// Tokens with spamCount/(spamCount+hamCount) > this% are considered spam indicators.
	bayesSpamThresholdPct = 70

	// bayesMinSpamTokens is the minimum number of spam-indicating tokens
	// needed to trigger the BAYES_SPAM rule.
	bayesMinSpamTokens = 3
)

// ---- Tokenizer

// Tokenize splits text into lowercase tokens, filtering short words and URLs.
// Single-pass: inlines URL skipping and lowercasing to avoid two extra string
// allocations (stripURLs + strings.ToLower).
func Tokenize(text string) []string {
	var tokens []string
	var current strings.Builder
	inURL := false

	for i := 0; i < len(text); {
		// URL start detection (byte-level guard before prefix check)
		if !inURL && text[i] == 'h' && i+7 < len(text) {
			rest := text[i:]
			if strings.HasPrefix(rest, "https://") || strings.HasPrefix(rest, "http://") {
				// Flush pending token before skipping URL
				if current.Len() >= tokenMinLen {
					tokens = append(tokens, current.String())
				}
				current.Reset()
				inURL = true
			}
		}

		r, size := utf8.DecodeRuneInString(text[i:])

		if unicode.IsSpace(r) {
			inURL = false
			if current.Len() >= tokenMinLen {
				tokens = append(tokens, current.String())
			}
			current.Reset()
		} else if !inURL {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				current.WriteRune(unicode.ToLower(r))
			} else {
				if current.Len() >= tokenMinLen {
					tokens = append(tokens, current.String())
				}
				current.Reset()
			}
		}

		i += size
	}
	if current.Len() >= tokenMinLen {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// ---- Corpus (AVL-tree-backed token frequency storage)

// tokenStats holds spam and ham counts for a single token.
type tokenStats struct {
	spam int
	ham  int
}

// Corpus stores token frequency data for Bayesian classification.
// It is owned by the calling realm and persists across transactions.
type Corpus struct {
	tree *avl.Tree // token string -> *tokenStats
}

// NewCorpus creates an empty Corpus.
func NewCorpus() *Corpus {
	return &Corpus{tree: avl.NewTree()}
}

// Train updates the corpus with tokens from a text sample.
// If isSpam is true, tokens are counted as spam indicators; otherwise as ham.
func (c *Corpus) Train(content string, isSpam bool) {
	tokens := dedup(Tokenize(content))
	for _, tok := range tokens {
		var stats *tokenStats
		v, ok := c.tree.Get(tok)
		if ok {
			stats = v.(*tokenStats)
		} else {
			stats = &tokenStats{}
			c.tree.Set(tok, stats)
		}
		if isSpam {
			stats.spam++
		} else {
			stats.ham++
		}
	}
}

// dedup sorts a string slice and removes duplicates in-place.
// Returns the deduplicated slice. Zero allocation beyond the sort.
func dedup(ss []string) []string {
	if len(ss) <= 1 {
		return ss
	}
	sort.Strings(ss)
	j := 0
	for i := 1; i < len(ss); i++ {
		if ss[i] != ss[j] {
			j++
			ss[j] = ss[i]
		}
	}
	return ss[:j+1]
}

// GetTokenStats returns the spam and ham counts for a token.
func (c *Corpus) GetTokenStats(token string) (spamCount, hamCount int) {
	v, ok := c.tree.Get(token)
	if !ok {
		return 0, 0
	}
	stats := v.(*tokenStats)
	return stats.spam, stats.ham
}

// Size returns the number of unique tokens in the corpus.
func (c *Corpus) Size() int {
	return c.tree.Size()
}

// ---- Bayesian scoring

// ScoreBayes evaluates content against the trained corpus.
// It counts tokens that are strong spam indicators (high spam ratio)
// and triggers BAYES_SPAM if enough are found.
// All math is integer-based (no floats).
func ScoreBayes(content string, corpus *Corpus) (int, string) {
	return scoreBayesTokens(dedup(Tokenize(content)), corpus)
}

// scoreBayesTokens scores pre-deduplicated tokens against the corpus.
// Tokens must already be deduplicated.
func scoreBayesTokens(tokens []string, corpus *Corpus) (int, string) {
	if corpus == nil || corpus.Size() < bayesMinCorpusSize {
		return 0, ""
	}
	if len(tokens) == 0 {
		return 0, ""
	}

	spamIndicators := 0
	hamIndicators := 0

	for _, tok := range tokens {
		spam, ham := corpus.GetTokenStats(tok)
		total := spam + ham
		if total == 0 {
			continue
		}
		if spam*100 > bayesSpamThresholdPct*total {
			spamIndicators++
		} else if ham*100 > bayesSpamThresholdPct*total {
			hamIndicators++
		}
	}

	if spamIndicators >= bayesMinSpamTokens && spamIndicators > hamIndicators {
		return WeightBayesSpam, "BAYES_SPAM"
	}
	return 0, ""
}
