// Package antispam provides a shared on-chain anti-spam scoring realm.
//
// It wraps p/gnoland/antispam (the scoring engine) with shared state:
// a Bayesian corpus, a fingerprint store, a keyword dictionary, and
// a blocklist with pre-loaded regex patterns for common crypto scams.
//
// Any realm can call Score() to evaluate content.
// Only the admin can modify the shared state (patterns, blocklist, keywords, training data).
//
// Usage from another realm:
//
//	import antispamr "gno.land/r/gnoland/antispam"
//	import engine "gno.land/p/gnoland/antispam"
//
//	result := antispamr.Score(author, content, rate, rep, nil, nil)
//	if result.Total >= engine.ThresholdHide {
//	    // hide the post
//	}
package antispam

import (
	"chain/runtime"
	"errors"

	engine "gno.land/p/gnoland/antispam"
	"gno.land/p/nt/ufmt"
)

// ---- Shared state

var (
	adminAddr address
	corpus    *engine.Corpus
	fps       *engine.FingerprintStore
	bl        *engine.Blocklist
	keywords  *engine.KeywordDict

	errUnauthorized = errors.New("unauthorized: not admin")
)

func init() {
	adminAddr = runtime.OriginCaller()
	corpus = engine.NewCorpus()
	fps = engine.NewFingerprintStore()
	bl = engine.NewBlocklist()
	keywords = engine.NewKeywordDict()

	// Pre-load common crypto/web3 spam patterns.
	// Admin can add/remove patterns at any time.
	defaultPatterns := []string{
		// Crypto scams: "send X GNOT/tokens"
		`(?i)send\s+\d+\s*(gnot|token|coin)`,
		// Free airdrop / free token scams
		`(?i)free\s*(airdrop|token|nft|mint)`,
		// Gambling spam
		`(?i)(casino|poker|gambling|betting)\s*(bonus|jackpot|win|payout)`,
		// Social redirect spam (telegram, discord, whatsapp group links)
		`(?i)(telegram|discord|whatsapp)\s*(group|channel|chat)\s*.*https?://`,
		// "Double your tokens" scams
		`(?i)double\s+your\s+(gnot|token|coin|crypto|money)`,
		// "Send X get Y back" scams
		`(?i)send\s+\d+.*get\s+\d+.*back`,
		// Wallet connect / seed phrase phishing
		`(?i)(connect|enter|verify)\s+(your\s+)?(wallet|seed\s*phrase|private\s*key|mnemonic)`,
		// Pump-and-dump language
		`(?i)(100|1000)x\s*(gem|token|coin|guaranteed|profit)`,
	}

	for _, p := range defaultPatterns {
		bl.AddPattern(p)
	}

	// Pre-load spam keyword dictionary (~180 words, multi-language).
	// See default_keywords.gno for the full list and weight definitions.
	keywords.BulkAdd(defaultKeywords)
}

func assertAdmin() {
	if runtime.OriginCaller() != adminAddr {
		panic(errUnauthorized)
	}
}

// ---- Public scoring

// Score evaluates content using the shared corpus, fingerprint store, keywords, and blocklist.
// Callers can optionally pass their own corpus/fps to use instead of the shared ones.
func Score(
	author, content string,
	rate engine.RateState,
	rep engine.ReputationData,
	callerCorpus *engine.Corpus,
	callerFps *engine.FingerprintStore,
) engine.SpamScore {
	c := corpus
	if callerCorpus != nil {
		c = callerCorpus
	}
	f := fps
	if callerFps != nil {
		f = callerFps
	}
	return engine.Score(engine.ScoreInput{
		Author:  author,
		Content: content,
		Rate:    rate,
		Rep:     rep,
		Corpus:  c,
		Fps:     f,
		Bl:      bl,
		Dict:    keywords,
	})
}

// ---- Admin: blocklist management

// AdminBlockAddress adds an address to the blocklist.
func AdminBlockAddress(_ realm, addr string) {
	assertAdmin()
	bl.BlockAddress(addr)
}

// AdminUnblockAddress removes an address from the blocklist.
func AdminUnblockAddress(_ realm, addr string) {
	assertAdmin()
	bl.UnblockAddress(addr)
}

// AdminAllowAddress adds an address to the allowlist (bypass all scoring).
func AdminAllowAddress(_ realm, addr string) {
	assertAdmin()
	bl.AllowAddress(addr)
}

// AdminRemoveAllow removes an address from the allowlist.
func AdminRemoveAllow(_ realm, addr string) {
	assertAdmin()
	bl.RemoveAllow(addr)
}

// ---- Admin: pattern management

// AdminAddPattern adds a regex pattern to the blocklist.
// Panics if the pattern is invalid regex.
func AdminAddPattern(_ realm, pattern string) {
	assertAdmin()
	if err := bl.AddPattern(pattern); err != nil {
		panic("invalid pattern: " + err.Error())
	}
}

// AdminRemovePattern removes a regex pattern from the blocklist.
func AdminRemovePattern(_ realm, pattern string) {
	assertAdmin()
	bl.RemovePattern(pattern)
}

// ---- Admin: keyword management

// AdminAddKeyword adds a single keyword with a weight (1-3).
func AdminAddKeyword(_ realm, word string, weight int) {
	assertAdmin()
	keywords.Add(word, weight)
}

// AdminRemoveKeyword removes a keyword from the dictionary.
func AdminRemoveKeyword(_ realm, word string) {
	assertAdmin()
	keywords.Remove(word)
}

// AdminBulkAddKeywords adds keywords from a newline-separated "word:weight" list.
// Words without :weight get default weight 1.
func AdminBulkAddKeywords(_ realm, data string) {
	assertAdmin()
	keywords.BulkAdd(data)
}

// ---- Admin: Bayesian training

// AdminTrain feeds content into the shared Bayesian corpus.
// isSpam=true for spam examples, false for legitimate (ham) examples.
func AdminTrain(_ realm, content string, isSpam bool) {
	assertAdmin()
	corpus.Train(content, isSpam)
}

// ---- Admin: fingerprint management

// AdminAddSpamFingerprint registers a known spam content fingerprint.
// Future content similar to this will be flagged as NEAR_DUPLICATE.
func AdminAddSpamFingerprint(_ realm, content string) {
	assertAdmin()
	fps.Add(content)
}

// ---- Admin: ownership

// AdminSetAdmin transfers admin rights to a new address.
func AdminSetAdmin(_ realm, newAdmin address) {
	assertAdmin()
	adminAddr = newAdmin
}

// ---- Render

// Render displays the realm status and active patterns.
func Render(_ string) string {
	out := "# Anti-Spam Scoring Realm\n\n"
	out += ufmt.Sprintf("**Admin:** `%s`\n\n", adminAddr.String())
	out += ufmt.Sprintf("**Corpus size:** %d tokens\n\n", corpus.Size())
	out += ufmt.Sprintf("**Fingerprint store:** %d entries\n\n", fps.Size())
	out += ufmt.Sprintf("**Keyword dictionary:** %d words\n\n", keywords.Size())
	out += ufmt.Sprintf("**Active patterns:** %d\n\n", bl.PatternCount())
	out += "## Patterns\n\n"
	bl.IteratePatterns(func(pattern string) bool {
		out += "- `" + pattern + "`\n"
		return false
	})
	out += "\n## Usage\n\n"
	out += "```gno\n"
	out += "import antispamr \"gno.land/r/gnoland/antispam\"\n"
	out += "import engine \"gno.land/p/gnoland/antispam\"\n\n"
	out += "result := antispamr.Score(author, content, rate, rep, nil, nil)\n"
	out += "if result.Total >= engine.ThresholdHide {\n"
	out += "    // hide the post\n"
	out += "}\n"
	out += "```\n"
	return out
}
