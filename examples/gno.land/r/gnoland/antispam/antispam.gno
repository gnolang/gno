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
//	result := antispamr.Score(author, content, rate, rep, nil, nil, nil, nil, engine.EarlyExitDisabled)
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

	// Pre-load spam patterns (10/30 slots, leaving 20 for admin).
	// Each pattern targets a specific scam category with tight alternations
	// to minimize false positives on legitimate blockchain discussions.
	defaultPatterns := []string{
		// 1. Credential phishing: asks user to share wallet/key/seed
		`(?i)(connect|enter|verify|confirm|provide|share|submit)\s+(your\s+)?(wallet|seed\s*phrase|private\s*key|mnemonic|password|credentials|recovery\s*phrase|login)`,
		// 2. Social platform redirect + DM solicitation
		`(?i)(telegram|discord|whatsapp|signal)\s*(group|channel|chat|link|community).{0,30}https?://|(dm|message|contact)\s+me\s+.{0,15}(telegram|discord|whatsapp|signal)`,
		// 3. Free crypto/airdrop scam
		`(?i)free\s+(airdrop|token|nft|mint|crypto|gnot|coin)`,
		// 4. Crypto transfer scam (requires scam context: send+get back, or double)
		`(?i)send\s+\d+.{0,30}(get\s+\d+|receive\s+\d+|return|reward)|double\s+your\s+(gnot|token|coin|crypto|money)`,
		// 5. Financial promises: guaranteed returns, multiplier pump-and-dump
		`(?i)(guaranteed|10+x)\s*.{0,15}(return|profit|income|roi|gem|token|coin|gain)`,
		// 6. Fake tech support (requires scam-specific CTA after problem statement)
		`(?i)(account|wallet)\s+(locked|suspended|compromised|hacked).{0,30}(click\s+(here|now|below)|restore\s+now|verify\s+(your|identity)|contact\s+(us|support)\s+at)`,
		// 7. Illegal content: gambling promos, adult+URL, drug marketplace
		`(?i)(casino|poker|gambling|betting|slots)\s*(bonus|jackpot|win|payout|free)|(live|hot|sexy|nude|xxx)\s+(cam|girl|webcam|chat|show|pics|video).{0,30}https?://|(buy|order|sell)\s+(cocaine|heroin|meth|pills|xanax|fentanyl|drugs)\s+(online|cheap|discreet)`,
		// 8. Urgency clickbait + referral/promo spam
		`(?i)click\s+(here|now|below).{0,30}(before|limited|expires?|hurry|urgent|last\s+chance)|(use|enter|try)\s+(my\s+)?(referral|invite|promo|discount)\s*(code|link)`,
		// 9. Money scam: inheritance fraud, wire transfer with urgency
		`(?i)(inherit|inheritance|beneficiary|next\s+of\s+kin).{0,30}(million|usd|\$\d+|fund)|(wire|transfer|send)\s+(money|funds|payment)\s+(immediately|urgent|now|asap)`,
		// 10. Fake job / work from home
		`(?i)(earn|make)\s+\$?\d+.{0,20}(per|a)\s+(day|hour|week).{0,15}(home|online|remote)`,
	}

	for _, p := range defaultPatterns {
		bl.AddPattern(p)
	}

	// Pre-load spam keyword dictionary (~400 words, multi-language).
	// See default_keywords.gno for the full list and weight definitions.
	keywords.BulkAdd(defaultKeywords)
}

// assertAdmin panics if the transaction signer is not the admin.
// Uses OriginCaller (not PreviousRealm) so admin operations are
// authorized by wallet signature regardless of the call chain.
func assertAdmin() {
	if runtime.OriginCaller() != adminAddr {
		panic(errUnauthorized)
	}
}

// ---- Public scoring

// Score evaluates content using the shared corpus, fingerprint store, keywords, and blocklist.
// Callers can optionally pass their own state to override the shared defaults.
// Pass nil for any parameter to use the shared instance.
//
// earlyExitAt controls gas optimization: pass a positive threshold
// (e.g. engine.ThresholdReject) to skip expensive rules when cheap
// rules already reach it. Pass engine.EarlyExitDisabled for full scoring.
func Score(
	author address, content string,
	rate engine.RateState,
	rep engine.ReputationData,
	callerCorpus *engine.Corpus,
	callerFps *engine.FingerprintStore,
	callerDict *engine.KeywordDict,
	callerBl *engine.Blocklist,
	earlyExitAt int,
) engine.SpamScore {
	c := corpus
	if callerCorpus != nil {
		c = callerCorpus
	}
	f := fps
	if callerFps != nil {
		f = callerFps
	}
	d := keywords
	if callerDict != nil {
		d = callerDict
	}
	b := bl
	if callerBl != nil {
		b = callerBl
	}
	return engine.Score(engine.ScoreInput{
		Author:      author.String(),
		Content:     content,
		Rate:        rate,
		Rep:         rep,
		Corpus:      c,
		Fps:         f,
		Bl:          b,
		Dict:        d,
		EarlyExitAt: earlyExitAt,
	})
}

// ---- Admin: blocklist management

// AdminBlockAddress adds an address to the blocklist.
func AdminBlockAddress(_ realm, addr address) {
	assertAdmin()
	bl.BlockAddress(addr.String())
}

// AdminUnblockAddress removes an address from the blocklist.
func AdminUnblockAddress(_ realm, addr address) {
	assertAdmin()
	bl.UnblockAddress(addr.String())
}

// AdminAllowAddress adds an address to the allowlist (bypass all scoring).
func AdminAllowAddress(_ realm, addr address) {
	assertAdmin()
	bl.AllowAddress(addr.String())
}

// AdminRemoveAllow removes an address from the allowlist.
func AdminRemoveAllow(_ realm, addr address) {
	assertAdmin()
	bl.RemoveAllow(addr.String())
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
	out += ufmt.Sprintf("**Blocked addresses:** %d\n\n", bl.BlockedCount())
	out += ufmt.Sprintf("**Allowed addresses:** %d\n\n", bl.AllowedCount())
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
	out += "result := antispamr.Score(author, content, rate, rep, nil, nil, nil, nil, engine.EarlyExitDisabled)\n"
	out += "if result.Total >= engine.ThresholdHide {\n"
	out += "    // hide the post\n"
	out += "}\n"
	out += "```\n"
	return out
}
