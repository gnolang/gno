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

	// bayesSpamThresholdPct is the spam ratio threshold (percentage).
	// Tokens appearing in spam more than this% of the time are considered spam indicators.
	bayesSpamThresholdPct = 70

	// bayesMinSpamTokens is the minimum number of spam-indicating tokens
	// needed to trigger the BAYES_SPAM rule.
	bayesMinSpamTokens = 3

	// bayesMinTokenOccurrences is the minimum total observations (spam + ham)
	// for a token to be considered reliable. Tokens seen fewer times are
	// ignored to prevent overfitting on small training sets.
	bayesMinTokenOccurrences = 2

	// corpusMaxSize caps the number of unique tokens in the corpus.
	// Existing tokens still get their counts updated; only new tokens
	// are rejected when the cap is reached. Prevents unbounded storage growth.
	corpusMaxSize = 10000

	// tokenMaxTotal caps the total (spam + ham) count per token.
	// When reached, both counts are halved before the new observation
	// is added. This gives recent observations more weight and makes
	// corpus poisoning attacks temporary rather than permanent -
	// critical when auto-training from moderation actions.
	tokenMaxTotal = 100
)

// ---- Tokenizer

// Tokenize splits text into lowercase tokens, filtering short words and URLs.
func Tokenize(text string) []string {
	var tokens []string
	var current strings.Builder
	inURL := false

	for i := 0; i < len(text); {
		if !inURL && hasURLPrefix(text, i) {
			if current.Len() >= tokenMinLen {
				tokens = append(tokens, current.String())
			}
			current.Reset()
			inURL = true
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

// ---- Corpus

// tokenStats holds spam and ham counts for a single token.
type tokenStats struct {
	spam int
	ham  int
}

// Corpus stores token frequency data for Bayesian classification.
// It is owned by the calling realm and persists across transactions.
type Corpus struct {
	tree *avl.Tree
}

// NewCorpus creates an empty Corpus.
func NewCorpus() *Corpus {
	return &Corpus{tree: avl.NewTree()}
}

// Train updates the corpus with tokens from a text sample.
// If isSpam is true, tokens are counted as spam indicators; otherwise as ham.
// When the corpus reaches corpusMaxSize, existing tokens still get updated
// but new tokens are ignored to prevent unbounded storage growth.
//
// Per-token counts are capped at tokenMaxTotal. When the total (spam + ham)
// for a token reaches the cap, both counts are halved before adding the
// new observation. This decay gives recent observations more weight and
// prevents corpus poisoning from being permanent - essential for
// auto-training scenarios where moderation actions feed the corpus.
func (c *Corpus) Train(content string, isSpam bool) {
	tokens := dedup(Tokenize(content))
	for _, tok := range tokens {
		var stats *tokenStats
		v, ok := c.tree.Get(tok)
		if ok {
			stats = v.(*tokenStats)
		} else {
			if c.tree.Size() >= corpusMaxSize {
				continue
			}
			stats = &tokenStats{}
			c.tree.Set(tok, stats)
		}
		// Decay: halve both counts when total reaches cap.
		// Preserves the ratio while bounding growth and giving
		// recent observations more influence.
		if stats.spam+stats.ham >= tokenMaxTotal {
			stats.spam /= 2
			stats.ham /= 2
		}
		if isSpam {
			stats.spam++
		} else {
			stats.ham++
		}
	}
}

// dedup removes duplicates from a string slice.
// It sorts and modifies the input slice in-place; callers must not
// retain references to the original ordering.
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
// Triggers BAYES_SPAM if enough spam-indicating tokens are found.
func ScoreBayes(content string, corpus *Corpus) (int, string) {
	return scoreBayesTokens(dedup(Tokenize(content)), corpus)
}

// scoreBayesTokens scores tokens against the corpus.
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
		if total < bayesMinTokenOccurrences {
			continue
		}
		if spam*100 > bayesSpamThresholdPct*total {
			spamIndicators++
		} else if ham*100 > bayesSpamThresholdPct*total {
			hamIndicators++
		}
	}

	if spamIndicators >= bayesMinSpamTokens && spamIndicators > hamIndicators {
		return WeightBayesSpam, RuleBayesSpam
	}
	return 0, ""
}
