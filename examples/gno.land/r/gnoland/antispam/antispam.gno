// Package antispam provides shared on-chain spam scoring for Gno realms.
// Wraps p/gnoland/antispam with pre-trained state (corpus, keywords, patterns) and reputation tracking.
// Anyone can score, only registered realms can record reputation, only admin can modify state.
// See README for usage.
package antispam

import (
	"chain/runtime"
	"errors"

	engine "gno.land/p/gnoland/antispam"
	"gno.land/p/nt/ufmt"
)

// Shared state

var (
	adminAddr address
	corpus    *engine.Corpus
	fps       *engine.FingerprintStore
	bl        *engine.Blocklist
	keywords  *engine.KeywordDict

	errUnauthorized = errors.New("unauthorized: not admin")
)

// defaultPattern is a combined regex covering 21 scam categories.
// Uses one pattern slot (1/30), one compiled regex, one scan per message.
// Categories:
//
//  1. Credential phishing     11. Fake KYC / verification
//  2. Social redirect + DM    12. Trading signals / pump group
//  3. Free crypto/airdrop     13. Wallet connect to claim
//  4. Crypto transfer scam    14. Loan / credit scam
//  5. Financial promises       15. Engagement spam (followers)
//  6. Fake tech support        16. Porn / adult solicitation
//  7. Gambling/adult/drugs     17. Gore / shock content
//  8. Urgency clickbait        18. Fake token launch / pump
//  9. Inheritance/wire fraud   19. Impersonation + scam CTA
//  10. Fake job / work from home 20. Romance / sugar scam
//  21. Contact redirect (personal email)
const defaultPattern = `(?i)(?:` +
	`(enter|verify|confirm|provide|share|submit|validate|restore|update|input)\s+(your\s+)?(wallet|seed\s*phrase|private\s*key|mnemonic|password|credentials|recovery\s*phrase|login)` +
	`|` +
	`(telegram|discord|whatsapp|signal)\s*(group|channel|chat|link|community).{0,30}https?://` +
	`|` +
	`(dm|message|contact)\s+me\s+.{0,15}(telegram|discord|whatsapp|signal)` +
	`|` +
	`free\s+(airdrop|token|nft|mint|crypto|gnot|coin)` +
	`|` +
	`claim\s+(your\s+)?free\s+(airdrop|token|nft|mint|crypto|gnot|coin)` +
	`|` +
	`send\s+\d+.{0,30}(get\s+\d+|receive\s+\d+|return|reward)` +
	`|` +
	`double\s+your\s+(gnot|token|coin|crypto|money)` +
	`|` +
	`(guaranteed|\d+x)\s*.{0,15}(return|profit|income|roi|gem|token|coin|gain)` +
	`|` +
	`\d+%\s*(return|profit|daily|monthly|weekly|roi|gain)` +
	`|` +
	`(account|wallet)\s+(locked|suspended|compromised|hacked).{0,30}(click\s+(here|now|below)|restore\s+now|verify\s+(your|identity)|contact\s+(us|support)\s+at)` +
	`|` +
	`(casino|poker|gambling|betting|slots)\s*(bonus|jackpot|win|payout|free)` +
	`|` +
	`(live|hot|sexy|nude|xxx)\s+(cam|girl|webcam|chat|show|pics|video).{0,30}https?://` +
	`|` +
	`(buy|order|sell)\s+(cocaine|heroin|meth|pills|xanax|fentanyl|drugs)\s+(online|cheap|discreet)` +
	`|` +
	`click\s+(here|now|below).{0,30}(before|limited|expires?|hurry|urgent|last\s+chance)` +
	`|` +
	`(use|enter|try)\s+(my\s+)?(referral|invite|promo|discount)\s*(code|link)` +
	`|` +
	`(inherit|inheritance|beneficiary|next\s+of\s+kin).{0,30}(million|usd|\$\d+|fund)` +
	`|` +
	`(wire|transfer|send)\s+(money|funds|payment)\s+(immediately|urgent|now|asap)` +
	`|` +
	`(earn|make)\s+\$?\d+.{0,20}(per|a)\s+(day|hour|week).{0,15}(home|online|remote)` +
	`|` +
	`(complete|start|begin|finish)\s+(your\s+)?(kyc|verification|identity\s+check).{0,30}https?://` +
	`|` +
	`(trading|crypto)\s+(signal|tip|alert|call)s?.{0,20}(guaranteed|accurate|free|join|vip)` +
	`|` +
	`connect\s+(your\s+)?wallet\s+.{0,20}(claim|receive|mint|airdrop|reward)` +
	`|` +
	`(instant|fast|quick|easy)\s+(loan|credit|cash).{0,20}(no\s+credit\s+check|approved|guaranteed|bad\s+credit)` +
	`|` +
	`(buy|get|cheap)\s+(followers|likes|subscribers|views|retweets).{0,20}(cheap|fast|instant|guaranteed|bulk)` +
	`|` +
	`(watch|download|stream|access)\s+(free\s+)?(porn|hentai|xxx|sex\s+tape|adult\s+video)` +
	`|` +
	`(onlyfans|chaturbate|livejasmin|pornhub|xvideos)\s*(link|free|premium|account|access)` +
	`|` +
	`(leaked|private|stolen)\s+(nude|naked|nsfw|sex)\s*(pic|photo|video|tape|content)s?` +
	`|` +
	`(watch|see|real)\s+(execution|beheading|murder|torture|death)\s*(video|footage|clip|live)` +
	`|` +
	`(gore|snuff|graphic)\s+(video|footage|content|film)` +
	`|` +
	`(token|coin)\s+(just\s+)?(launched|deployed|listed|live).{0,20}(buy|invest|hurry|limited|moon|\d+x)` +
	`|` +
	`(official|admin|team|staff|support)\s+(announcement|message|notice|update).{0,30}(send|transfer|click|visit|connect|verify)` +
	`|` +
	`sugar\s+(daddy|mommy|mama|baby).{0,40}(pay|send|money|gift|allowance|dm|message|contact)` +
	`|` +
	`(lonely|single)\s+(girl|woman|man|guy|lady).{0,40}(dm|message|contact|whatsapp|telegram)` +
	`|` +
	`[\w.+-]+@(gmail|yahoo|hotmail|outlook|proton|aol)\.(com|fr|net|org)` +
	`)`

func init() {
	adminAddr = runtime.OriginCaller()
	corpus = engine.NewCorpus()
	fps = engine.NewFingerprintStore()
	bl = engine.NewBlocklist()
	keywords = engine.NewKeywordDict()
}

// assertAdmin panics if the transaction signer is not the admin.
// Uses OriginCaller (not PreviousRealm) so admin operations are
// authorized by wallet signature regardless of the call chain.
func assertAdmin() {
	if runtime.OriginCaller() != adminAddr {
		panic(errUnauthorized)
	}
}

// Admin: load defaults

// AdminLoadDefaults loads the pre-configured spam patterns and keywords
// into the shared state. Called by admin after deployment to populate
// the blocklist pattern (1/30 slots) and keyword dictionary (~400 words).
// Safe to call multiple times (overwrites existing entries).
func AdminLoadDefaults(_ realm) {
	assertAdmin()
	bl.AddPattern(defaultPattern)
	keywords.BulkAdd(defaultKeywords)
}

// Public scoring

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
	// Auto-populate moderation data from internal reputation state.
	// Caller provides chain data (AccountAgeDays, Balance, HasUsername);
	// realm fills in FlaggedCount, BanCount, TotalAccepted.
	key := author.String()
	raw, ok := reputations.Get(key)
	if ok {
		ir := raw.(*addrReputation)
		rep.FlaggedCount = ir.flaggedCount
		rep.BanCount = ir.banCount
		rep.TotalAccepted = ir.totalAccepted
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

// Admin: blocklist management

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

// Admin: pattern management

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

// Admin: keyword management

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

// Admin: Bayesian training

// AdminTrain feeds content into the shared Bayesian corpus.
// isSpam=true for spam examples, false for legitimate (ham) examples.
func AdminTrain(_ realm, content string, isSpam bool) {
	assertAdmin()
	corpus.Train(content, isSpam)
}

// Admin: fingerprint management

// AdminAddSpamFingerprint registers a known spam content fingerprint.
// Future content similar to this will be flagged as NEAR_DUPLICATE.
func AdminAddSpamFingerprint(_ realm, content string) {
	assertAdmin()
	fps.Add(content)
}

// Admin: ownership

// AdminSetAdmin transfers admin rights to a new address.
func AdminSetAdmin(_ realm, newAdmin address) {
	assertAdmin()
	adminAddr = newAdmin
}

// Render

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
	out += ufmt.Sprintf("**Tracked addresses:** %d\n\n", reputations.Size())
	out += ufmt.Sprintf("**Trusted callers:** %d\n\n", trustedCallers.Size())
	out += "## Patterns\n\n"
	bl.IteratePatterns(func(pattern string) bool {
		out += "- `" + pattern + "`\n"
		return false
	})
	out += "\n## Usage\n\n"
	out += "```gno\n"
	out += "import antispamr \"gno.land/r/gnoland/antispam\"\n"
	out += "import engine \"gno.land/p/gnoland/antispam\"\n\n"
	out += "// Chain data only - FlaggedCount/BanCount/TotalAccepted auto-populated\n"
	out += "rep := engine.ReputationData{AccountAgeDays: 30, Balance: 5000000000, HasUsername: true}\n"
	out += "result := antispamr.Score(author, content, rate, rep, nil, nil, nil, nil, engine.EarlyExitDisabled)\n"
	out += "if result.Total >= engine.ThresholdHide {\n"
	out += "    // hide the content\n"
	out += "}\n\n"
	out += "// Record accepted content for reputation tracking\n"
	out += "antispamr.RecordAccepted(author)\n"
	out += "```\n"
	return out
}
