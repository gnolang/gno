package main

// Demonstration: Regex Pattern Categories
// =========================================
// This test exercises the blocklist regex patterns across all 21 scam
// categories, plus clean content that must NOT trigger false positives.
//
// The realm pre-loads 1 mega-pattern covering 21 categories in a single
// compiled regex. This test validates each category independently using
// separate patterns to confirm detection coverage.
//
// Categories tested:
//   1-10: Phishing, social redirect, airdrops, transfer scam, financial
//         promises, tech support, gambling/adult/drugs, urgency, inheritance,
//         fake jobs
//  11-21: Fake KYC, trading signals, wallet drain, loan scam, engagement
//         spam, porn, gore, fake token launch, impersonation, romance scam,
//         contact redirect (personal email)

import (
	"gno.land/p/gnoland/antispam"
)

func intToString(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	s := ""
	for i > 0 {
		s = string(rune(48+(i%10))) + s
		i /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

func sliceToString(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	s := "["
	for i, v := range arr {
		if i > 0 {
			s += " "
		}
		s += v
	}
	s += "]"
	return s
}

func containsRule(rules []string, target string) bool {
	for _, rule := range rules {
		if rule == target {
			return true
		}
	}
	return false
}

func main() {
	println("=== REGEX PATTERN CATEGORIES ===\n")

	// Setup: blocklist with patterns covering all 20 categories
	bl := antispam.NewBlocklist()

	// Categories 1-10 (original)
	bl.AddPattern(`(?i)(connect|enter|verify|confirm|submit)\s+(your\s+)?(wallet|seed\s*phrase|private\s*key|mnemonic)`)
	bl.AddPattern(`(?i)(telegram|discord|whatsapp)\s*(group|channel|chat).{0,30}https?://`)
	bl.AddPattern(`(?i)free\s+(airdrop|token|nft|mint|crypto)`)
	bl.AddPattern(`(?i)send\s+\d+.{0,30}(get\s+\d+|receive\s+\d+)`)
	bl.AddPattern(`(?i)double\s+your\s+(gnot|token|coin|crypto|money)`)
	bl.AddPattern(`(?i)(guaranteed|\d+x)\s*.{0,15}(return|profit|income|roi)`)
	bl.AddPattern(`(?i)(account|wallet)\s+(locked|suspended|hacked).{0,30}(click|restore|verify|contact)`)
	bl.AddPattern(`(?i)(casino|poker|gambling|betting)\s*(bonus|jackpot|win|payout)`)
	bl.AddPattern(`(?i)(inherit|inheritance|beneficiary).{0,30}(million|usd|\$\d+)`)
	bl.AddPattern(`(?i)(earn|make)\s+\$?\d+.{0,20}(per|a)\s+(day|hour|week).{0,15}(home|online|remote)`)

	// Categories 11-20 (new)
	bl.AddPattern(`(?i)(complete|start)\s+(your\s+)?(kyc|verification).{0,30}https?://`)
	bl.AddPattern(`(?i)(trading|crypto)\s+(signal|tip|alert)s?.{0,20}(guaranteed|accurate|free|join|vip)`)
	bl.AddPattern(`(?i)connect\s+(your\s+)?wallet\s+.{0,20}(claim|receive|mint|airdrop|reward)`)
	bl.AddPattern(`(?i)(instant|fast|quick|easy)\s+(loan|credit|cash).{0,20}(no\s+credit\s+check|approved|guaranteed)`)
	bl.AddPattern(`(?i)(buy|get|cheap)\s+(followers|likes|subscribers|views).{0,20}(cheap|fast|instant|guaranteed)`)
	bl.AddPattern(`(?i)(watch|download|stream)\s+(free\s+)?(porn|hentai|xxx|sex\s+tape)`)
	bl.AddPattern(`(?i)(watch|see|real)\s+(execution|beheading|murder|torture)\s*(video|footage|clip)`)
	bl.AddPattern(`(?i)(token|coin)\s+(just\s+)?(launched|deployed|listed).{0,20}(buy|invest|hurry|limited|moon|\d+x)`)
	bl.AddPattern(`(?i)(official|admin|team)\s+(announcement|message|notice).{0,30}(send|transfer|click|verify)`)
	bl.AddPattern(`(?i)sugar\s+(daddy|mommy|mama|baby).{0,40}(pay|send|money|gift|dm|message)`)

	// Category 21: Contact redirect (personal email addresses)
	bl.AddPattern(`(?i)[\w.+-]+@(gmail|yahoo|hotmail|outlook|proton|aol)\.(com|fr|net|org)`)

	// Shared scoring state
	goodRep := antispam.ReputationData{
		AccountAgeDays: 60,
		Balance:        5000000000,
		HasUsername:    true,
	}
	normalRate := antispam.RateState{PostCount: 2, WindowSeconds: 3600}

	score := func(content string) antispam.SpamScore {
		return antispam.Score(antispam.ScoreInput{
			Author:     "g1user",
			Content:    content,
			Rate:       normalRate,
			Reputation: goodRep,
			Blocklist:  bl,
		})
	}

	// ========================================
	// PHASE 1: Categories 1-10 (Original)
	// ========================================

	println("PHASE 1: Original Categories (1-10)")
	println("====================================\n")

	cats1 := []struct {
		num     int
		name    string
		content string
	}{
		{1, "Credential phishing", "Please verify your wallet and seed phrase to continue"},
		{2, "Social redirect", "Join our Telegram group for updates https://t.me/scamgroup"},
		{3, "Free crypto/airdrop", "Get your free airdrop tokens today"},
		{4, "Crypto transfer scam", "Send 500 GNOT and receive 5000 back guaranteed"},
		{5, "Double your crypto", "Double your GNOT in just 24 hours"},
		{6, "Financial promises", "Guaranteed 100x return on your investment"},
		{7, "Fake tech support", "Your wallet locked please click here to restore"},
		{8, "Gambling promos", "Casino bonus jackpot win big today"},
		{9, "Inheritance fraud", "You are the beneficiary of a 5 million USD inheritance"},
		{10, "Fake job", "Earn $5000 per week from home online easily"},
	}

	for _, c := range cats1 {
		r := score(c.content)
		status := "MISS"
		if containsRule(r.Triggered, "BLOCKED_PATTERN") {
			status = "HIT"
		}
		println(intToString(c.num) + ". " + c.name + ": " + status)
		println("   \"" + c.content + "\"")
		println("   Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered) + "\n")
	}

	// ========================================
	// PHASE 2: Categories 11-20 (New)
	// ========================================

	println("PHASE 2: New Categories (11-21)")
	println("================================\n")

	cats2 := []struct {
		num     int
		name    string
		content string
	}{
		{11, "Fake KYC", "Complete your KYC verification now at https://fake-kyc.example.com"},
		{12, "Trading signals", "Join our crypto signals group guaranteed accurate free tips"},
		{13, "Wallet connect drain", "Connect your wallet to claim free airdrop rewards now"},
		{14, "Loan scam", "Get instant loan approved guaranteed no credit check needed"},
		{15, "Engagement spam", "Buy cheap followers and likes instant delivery guaranteed"},
		{16, "Porn solicitation", "Watch free porn and xxx content streaming now"},
		{17, "Gore content", "See real execution video footage leaked from prison"},
		{18, "Fake token launch", "New token just launched on DEX buy now before moon"},
		{19, "Impersonation", "Official admin announcement please verify your account by clicking"},
		{20, "Romance/sugar scam", "Sugar daddy looking for a baby girl to send money and pay bills just dm me"},
		{21, "Contact redirect", "Send payment proof to winner2024@gmail.com to claim your prize now"},
	}

	for _, c := range cats2 {
		r := score(c.content)
		status := "MISS"
		if containsRule(r.Triggered, "BLOCKED_PATTERN") {
			status = "HIT"
		}
		println(intToString(c.num) + ". " + c.name + ": " + status)
		println("   \"" + c.content + "\"")
		println("   Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered) + "\n")
	}

	// ========================================
	// PHASE 3: Clean Content (no false positives)
	// ========================================

	println("PHASE 3: Clean Content (false positive check)")
	println("==============================================\n")

	clean := []struct {
		name    string
		content string
	}{
		{"Governance discussion", "The governance proposal for validator delegation has been submitted for community review."},
		{"Technical question", "How do I connect my wallet to the gno.land testnet using gnokey?"},
		{"Trading discussion", "I think the token price will go up based on the recent development updates."},
		{"Legitimate announcement", "The team announcement about the new release is available on the blog."},
		{"Financial discussion", "The annual return on staking rewards is about 7% currently."},
		{"Job posting", "We are hiring a senior Go developer to work on our remote team."},
		{"Loan discussion", "I took out a small loan to invest in my validator setup."},
		{"Email discussion", "You can reach the team via their official website or community forum."},
	}

	allClean := true
	for _, c := range clean {
		r := score(c.content)
		status := "CLEAN"
		if containsRule(r.Triggered, "BLOCKED_PATTERN") {
			status = "FALSE POSITIVE"
			allClean = false
		}
		println(c.name + ": " + status)
		println("   \"" + c.content + "\"")
		println("   Score: " + intToString(r.Total) + " | Rules: " + sliceToString(r.Triggered) + "\n")
	}

	// ========================================
	// SUMMARY
	// ========================================

	println("SUMMARY")
	println("=======\n")

	println("Categories 1-10:  All detected via BLOCKED_PATTERN")
	println("Categories 11-21: All detected via BLOCKED_PATTERN")
	println("Clean content:    No false positives")
	if allClean {
		println("\nAll 21 categories detected. Zero false positives on clean content.")
	}

	println("\n=== END ===")
}

// Output:
// === REGEX PATTERN CATEGORIES ===
//
// PHASE 1: Original Categories (1-10)
// ====================================
//
// 1. Credential phishing: HIT
//    "Please verify your wallet and seed phrase to continue"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 2. Social redirect: HIT
//    "Join our Telegram group for updates https://t.me/scamgroup"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 3. Free crypto/airdrop: HIT
//    "Get your free airdrop tokens today"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 4. Crypto transfer scam: HIT
//    "Send 500 GNOT and receive 5000 back guaranteed"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 5. Double your crypto: HIT
//    "Double your GNOT in just 24 hours"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 6. Financial promises: HIT
//    "Guaranteed 100x return on your investment"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 7. Fake tech support: HIT
//    "Your wallet locked please click here to restore"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 8. Gambling promos: HIT
//    "Casino bonus jackpot win big today"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 9. Inheritance fraud: HIT
//    "You are the beneficiary of a 5 million USD inheritance"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 10. Fake job: HIT
//    "Earn $5000 per week from home online easily"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// PHASE 2: New Categories (11-21)
// ================================
//
// 11. Fake KYC: HIT
//    "Complete your KYC verification now at https://fake-kyc.example.com"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 12. Trading signals: HIT
//    "Join our crypto signals group guaranteed accurate free tips"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 13. Wallet connect drain: HIT
//    "Connect your wallet to claim free airdrop rewards now"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 14. Loan scam: HIT
//    "Get instant loan approved guaranteed no credit check needed"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 15. Engagement spam: HIT
//    "Buy cheap followers and likes instant delivery guaranteed"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 16. Porn solicitation: HIT
//    "Watch free porn and xxx content streaming now"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 17. Gore content: HIT
//    "See real execution video footage leaked from prison"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 18. Fake token launch: HIT
//    "New token just launched on DEX buy now before moon"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 19. Impersonation: HIT
//    "Official admin announcement please verify your account by clicking"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 20. Romance/sugar scam: HIT
//    "Sugar daddy looking for a baby girl to send money and pay bills just dm me"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// 21. Contact redirect: HIT
//    "Send payment proof to winner2024@gmail.com to claim your prize now"
//    Score: 5 | Rules: [BLOCKED_PATTERN]
//
// PHASE 3: Clean Content (false positive check)
// ==============================================
//
// Governance discussion: CLEAN
//    "The governance proposal for validator delegation has been submitted for community review."
//    Score: 0 | Rules: []
//
// Technical question: CLEAN
//    "How do I connect my wallet to the gno.land testnet using gnokey?"
//    Score: 0 | Rules: []
//
// Trading discussion: CLEAN
//    "I think the token price will go up based on the recent development updates."
//    Score: 0 | Rules: []
//
// Legitimate announcement: CLEAN
//    "The team announcement about the new release is available on the blog."
//    Score: 0 | Rules: []
//
// Financial discussion: CLEAN
//    "The annual return on staking rewards is about 7% currently."
//    Score: 0 | Rules: []
//
// Job posting: CLEAN
//    "We are hiring a senior Go developer to work on our remote team."
//    Score: 0 | Rules: []
//
// Loan discussion: CLEAN
//    "I took out a small loan to invest in my validator setup."
//    Score: 0 | Rules: []
//
// Email discussion: CLEAN
//    "You can reach the team via their official website or community forum."
//    Score: 0 | Rules: []
//
// SUMMARY
// =======
//
// Categories 1-10:  All detected via BLOCKED_PATTERN
// Categories 11-21: All detected via BLOCKED_PATTERN
// Clean content:    No false positives
//
// All 21 categories detected. Zero false positives on clean content.
//
// === END ===
