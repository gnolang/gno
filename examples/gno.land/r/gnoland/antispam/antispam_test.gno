package antispam

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"

	engine "gno.land/p/gnoland/antispam"
)

// ---- helpers

func resetState(t *testing.T) {
	t.Helper()
	adminAddr = address("g1admin")
	corpus = engine.NewCorpus()
	fps = engine.NewFingerprintStore()
	bl = engine.NewBlocklist()
	keywords = engine.NewKeywordDict()
	keywords.BulkAdd(defaultKeywords)

	// Re-load default patterns (simplified versions of the realm's mega-pattern categories)
	for _, p := range []string{
		`(?i)send\s+\d+\s*(gnot|token|coin)`,
		`(?i)free\s*(airdrop|token|nft|mint)`,
		`(?i)(casino|poker|gambling|betting)\s*(bonus|jackpot|win|payout)`,
		`(?i)(telegram|discord|whatsapp)\s*(group|channel|chat)\s*.*https?://`,
		`(?i)double\s+your\s+(gnot|token|coin|crypto|money)`,
		`(?i)send\s+\d+.*get\s+\d+.*back`,
		`(?i)(enter|verify)\s+(your\s+)?(wallet|seed\s*phrase|private\s*key|mnemonic)`,
		`(?i)(100|1000)x\s*(gem|token|coin|guaranteed|profit)`,
		`(?i)(complete|start)\s+(your\s+)?(kyc|verification).{0,30}https?://`,
		`(?i)(trading|crypto)\s+(signal|tip)s?.{0,20}(guaranteed|free|join|vip)`,
		`(?i)connect\s+(your\s+)?wallet\s+.{0,20}(claim|receive|mint|airdrop)`,
		`(?i)(instant|fast|quick)\s+(loan|credit|cash).{0,20}(no\s+credit\s+check|approved|guaranteed)`,
		`(?i)(buy|get|cheap)\s+(followers|likes|subscribers).{0,20}(cheap|fast|instant)`,
		`(?i)(watch|download)\s+(free\s+)?(porn|hentai|xxx|sex\s+tape)`,
		`(?i)(watch|see|real)\s+(execution|beheading|murder|torture)\s*(video|footage)`,
		`(?i)(token|coin)\s+(just\s+)?(launched|deployed|listed).{0,20}(buy|invest|hurry|limited|moon)`,
		`(?i)(official|admin|team)\s+(announcement|message|notice).{0,30}(send|transfer|click|verify)`,
		`(?i)sugar\s+(daddy|mommy|mama|baby).{0,40}(pay|send|money|gift|dm|message)`,
	} {
		bl.AddPattern(p)
	}
}

func setAdmin(t *testing.T) {
	t.Helper()
	adminAddr = address("g1admin")
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
}

func goodRep() engine.ReputationData {
	return engine.ReputationData{
		HasUsername:    true,
		AccountAgeDays: 30,
		Balance:        5000000000,
	}
}

// ---- admin pattern management

func TestAdminAddPattern(t *testing.T) {
	resetState(t)
	setAdmin(t)
	AdminAddPattern(cross, `(?i)test\s+specific\s+pattern`)

	result := Score(address("g1user"), "this is a test specific pattern here", engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	if result.Total < 5 {
		t.Errorf("expected blocked pattern score >= 5, got %d", result.Total)
	}
}

func TestAdminRemovePattern(t *testing.T) {
	resetState(t)
	setAdmin(t)
	AdminAddPattern(cross, `(?i)removable\s+pattern`)
	AdminRemovePattern(cross, `(?i)removable\s+pattern`)

	result := Score(address("g1user"), "removable pattern here with enough length for content analysis", engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	hasBlockedPattern := false
	for _, r := range result.Triggered {
		if r == "BLOCKED_PATTERN" {
			hasBlockedPattern = true
		}
	}
	if hasBlockedPattern {
		t.Errorf("expected no BLOCKED_PATTERN after removal, got triggered: %v", result.Triggered)
	}
}

// ---- admin address management

func TestAdminBlockAddress(t *testing.T) {
	resetState(t)
	setAdmin(t)
	AdminBlockAddress(cross, address("g1blocktest"))

	result := Score(address("g1blocktest"), "Normal message.", engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	if result.Total < 99 {
		t.Errorf("expected blocked address score >= 99, got %d", result.Total)
	}
}

func TestAdminUnblockAddress(t *testing.T) {
	resetState(t)
	setAdmin(t)
	AdminBlockAddress(cross, address("g1temp"))
	AdminUnblockAddress(cross, address("g1temp"))

	result := Score(address("g1temp"), "Normal message from unblocked user with reasonable length.", engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	if result.Total > 0 {
		t.Errorf("expected score 0 after unblock, got %d (triggered: %v)", result.Total, result.Triggered)
	}
}

func TestAdminAllowAddress(t *testing.T) {
	resetState(t)
	setAdmin(t)
	AdminAllowAddress(cross, address("g1trusted"))

	// Allowed user should score 0 even with spam content
	result := Score(address("g1trusted"), "FREE AIRDROP BONUS CLICK NOW!!!", engine.RateState{PostCount: 50, WindowSeconds: 60}, engine.ReputationData{BanCount: 5}, nil, nil, nil, nil, engine.EarlyExitDisabled)
	if result.Total != 0 {
		t.Errorf("expected score 0 for allowed address, got %d (triggered: %v)", result.Total, result.Triggered)
	}
}

// ---- admin training

func TestAdminTrain(t *testing.T) {
	resetState(t)
	setAdmin(t)

	// Train with spam examples
	AdminTrain(cross, "free airdrop bonus click claim tokens hurry", true)
	AdminTrain(cross, "free money send prize winner casino bonus", true)
	AdminTrain(cross, "airdrop claim free tokens limited offer now", true)
	AdminTrain(cross, "click here free prize bonus amazing deal", true)
	AdminTrain(cross, "casino jackpot winner spin bonus special", true)

	// Train with ham examples
	AdminTrain(cross, "governance proposal voting community discussion update", false)
	AdminTrain(cross, "validator staking delegation reward system node", false)
	AdminTrain(cross, "development update release version changelog fix", false)
	AdminTrain(cross, "question about using gnokey transaction send", false)
	AdminTrain(cross, "review this pull request changes code test", false)
	AdminTrain(cross, "interesting discussion about protocol design plan", false)
	AdminTrain(cross, "thanks for the helpful explanation today great", false)

	// Spam-like content should score high via bayes
	result := Score(address("g1user"), "free airdrop bonus click claim prize", engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	if result.Total < 3 {
		t.Errorf("expected bayes score >= 3 for spam content, got %d (triggered: %v)", result.Total, result.Triggered)
	}
}

// ---- admin fingerprint management

func TestAdminAddSpamFingerprint(t *testing.T) {
	resetState(t)
	setAdmin(t)
	AdminAddSpamFingerprint(cross, "free airdrop bonus click claim prize tokens hurry now limited")

	// Near-duplicate should be detected
	result := Score(address("g1user"), "free airdrop bonus click claim prize tokens quick now fast", engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	if result.Total < 4 {
		t.Errorf("expected fingerprint score >= 4, got %d (triggered: %v)", result.Total, result.Triggered)
	}
}

// ---- default patterns

func TestDefaultPatternsLoaded(t *testing.T) {
	resetState(t)

	tests := []struct {
		name    string
		content string
		wantMin int
	}{
		{"success: crypto scam send gnot", "Send 1000 GNOT to this address", 5},
		{"success: airdrop scam", "Get your FREE AIRDROP of tokens now", 5},
		{"success: gambling spam", "Casino bonus jackpot win big today", 5},
		{"success: telegram redirect", "Join our Telegram group https://t.me/scam", 5},
		{"success: double your tokens", "Double your GNOT in 24 hours guaranteed", 5},
		{"success: send get back scam", "Send 100 GNOT and get 500 back", 5},
		{"success: credential phishing", "Enter your seed phrase to verify your wallet now", 5},
		{"success: pump and dump", "This 100x gem token is guaranteed profit", 5},
		{"success: fake KYC scam", "Complete your KYC verification at https://fake-kyc.com", 5},
		{"success: trading signals", "Free crypto signals guaranteed join VIP group", 5},
		{"success: wallet connect drain", "Connect your wallet to claim free airdrop tokens", 5},
		{"success: loan scam", "Instant loan guaranteed no credit check apply now", 5},
		{"success: engagement spam", "Buy cheap followers and likes instant delivery", 5},
		{"success: porn solicitation", "Watch free porn videos and xxx content here", 5},
		{"success: gore content", "Watch real execution video footage leaked online", 5},
		{"success: fake token launch", "New token just launched on DEX buy now before moon", 5},
		{"success: impersonation scam", "Official admin announcement please verify your account by clicking here", 5},
		{"success: sugar scam", "Sugar daddy looking for baby to pay and send money weekly just dm me", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Score(address("g1user"), tt.content, engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
			if result.Total < tt.wantMin {
				t.Errorf("expected score >= %d, got %d (triggered: %v)", tt.wantMin, result.Total, result.Triggered)
			}
		})
	}
}

// ---- default keywords

func TestDefaultKeywordsLoaded(t *testing.T) {
	resetState(t)

	if keywords.Size() < 100 {
		t.Errorf("expected at least 100 default keywords, got %d", keywords.Size())
	}

	// Content with multiple default spam keywords should trigger
	result := Score(address("g1user"), "This is a phishing scam arnaque with ponzi scheme and ransomware", engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	hasKeyword := false
	for _, r := range result.Triggered {
		if r == "KEYWORD_SPAM" {
			hasKeyword = true
		}
	}
	if !hasKeyword {
		t.Errorf("expected KEYWORD_SPAM from defaults, got triggered: %v (score: %d)", result.Triggered, result.Total)
	}
}

// ---- admin keyword management

func TestAdminAddKeyword(t *testing.T) {
	resetState(t)
	setAdmin(t)
	AdminAddKeyword(cross, "customword1", 3)
	AdminAddKeyword(cross, "customword2", 3)
	AdminAddKeyword(cross, "customword3", 2)

	// Content with multiple custom keywords should trigger
	result := Score(address("g1user"), "This has customword1 and customword2 plus customword3 inside it", engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	hasKeyword := false
	for _, r := range result.Triggered {
		if r == "KEYWORD_SPAM" {
			hasKeyword = true
		}
	}
	if !hasKeyword {
		t.Errorf("expected KEYWORD_SPAM rule, got triggered: %v", result.Triggered)
	}
}

func TestAdminBulkAddKeywords(t *testing.T) {
	resetState(t)
	setAdmin(t)
	sizeBefore := keywords.Size()
	AdminBulkAddKeywords(cross, "bulktest1:3\nbulktest2:2\nbulktest3:2\nbulktest4:3")

	if keywords.Size() != sizeBefore+4 {
		t.Errorf("expected size %d after bulk add, got %d", sizeBefore+4, keywords.Size())
	}

	// Content with bulk-added keywords should trigger
	result := Score(address("g1user"), "This bulktest1 message has bulktest2 and bulktest3 words inside", engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	hasKeyword := false
	for _, r := range result.Triggered {
		if r == "KEYWORD_SPAM" {
			hasKeyword = true
		}
	}
	if !hasKeyword {
		t.Errorf("expected KEYWORD_SPAM after bulk add, got triggered: %v", result.Triggered)
	}
}

func TestAdminRemoveKeyword(t *testing.T) {
	resetState(t)
	setAdmin(t)
	sizeBefore := keywords.Size()
	AdminAddKeyword(cross, "testword", 3)
	AdminAddKeyword(cross, "testother", 3)

	if keywords.Size() != sizeBefore+2 {
		t.Errorf("expected size %d after add, got %d", sizeBefore+2, keywords.Size())
	}

	AdminRemoveKeyword(cross, "testword")
	AdminRemoveKeyword(cross, "testother")

	if keywords.Size() != sizeBefore {
		t.Errorf("expected size %d after removal, got %d", sizeBefore, keywords.Size())
	}
}

// ---- access control

func TestNonAdminCannotModify(t *testing.T) {
	resetState(t)
	setAdmin(t) // ensure admin is g1admin

	other := testutils.TestAddress("other")
	testing.SetOriginCaller(other)

	uassert.AbortsWithMessage(t, errUnauthorized.Error(), func() {
		AdminBlockAddress(cross, address("g1someone"))
	})
}

// ---- scoring

func TestScoreCleanContent(t *testing.T) {
	resetState(t)

	rep := engine.ReputationData{
		AccountAgeDays: 60,
		Balance:        5000000000,
		FlaggedCount:   0,
		TotalPosts:     20,
		HasUsername:    true,
		BanCount:       0,
	}

	result := Score(address("g1legit"), "I think we should consider adjusting the staking rewards.", engine.RateState{PostCount: 3, WindowSeconds: 3600}, rep, nil, nil, nil, nil, engine.EarlyExitDisabled)
	if result.Total != 0 {
		t.Errorf("expected score 0 for clean content, got %d (triggered: %v)", result.Total, result.Triggered)
	}
}

// ---- admin transfer

func TestAdminTransfer(t *testing.T) {
	resetState(t)
	setAdmin(t)

	newAdmin := testutils.TestAddress("newadmin")
	AdminSetAdmin(cross, newAdmin)

	// New admin can operate
	testing.SetOriginCaller(newAdmin)
	AdminBlockAddress(cross, address("g1transfertest"))
	AdminUnblockAddress(cross, address("g1transfertest"))

	// Old admin should fail
	testing.SetOriginCaller(address("g1admin"))
	uassert.AbortsWithMessage(t, errUnauthorized.Error(), func() {
		AdminBlockAddress(cross, address("g1shouldfail"))
	})

	// Restore for other tests
	testing.SetOriginCaller(newAdmin)
	AdminSetAdmin(cross, address("g1admin"))
}

// ---- caller overrides

func TestScoreCallerOverrides(t *testing.T) {
	resetState(t)

	// Create custom blocklist with a specific pattern
	customBl := engine.NewBlocklist()
	customBl.AddPattern(`(?i)custom\s+override\s+test`)

	// Message that matches the custom pattern but NOT the default patterns
	msg := "This is a custom override test message with enough length"

	// Without override: should not trigger BLOCKED_PATTERN
	result1 := Score(address("g1user"), msg, engine.RateState{}, goodRep(), nil, nil, nil, nil, engine.EarlyExitDisabled)
	for _, r := range result1.Triggered {
		if r == "BLOCKED_PATTERN" {
			t.Errorf("expected no BLOCKED_PATTERN without override, got triggered: %v", result1.Triggered)
		}
	}

	// With custom blocklist override: should trigger BLOCKED_PATTERN
	result2 := Score(address("g1user"), msg, engine.RateState{}, goodRep(), nil, nil, nil, customBl, engine.EarlyExitDisabled)
	hasPattern := false
	for _, r := range result2.Triggered {
		if r == "BLOCKED_PATTERN" {
			hasPattern = true
		}
	}
	if !hasPattern {
		t.Errorf("expected BLOCKED_PATTERN with custom blocklist, got triggered: %v", result2.Triggered)
	}
}

// ---- early exit

func TestScoreEarlyExit(t *testing.T) {
	resetState(t)

	// Bad reputation + rate burst = high Phase 1 score
	badRep := engine.ReputationData{
		AccountAgeDays: 0,
		Balance:        100000000,
		FlaggedCount:   10,
		TotalPosts:     15,
		HasUsername:    false,
		BanCount:       1,
	}
	burstRate := engine.RateState{PostCount: 25, WindowSeconds: 3600}
	msg := "FREE AIRDROP BONUS CLICK CLAIM PRIZE"

	// Full scoring: Phase 2 runs (Bayes/keywords should trigger)
	full := Score(address("g1spammer"), msg, burstRate, badRep, nil, nil, nil, nil, engine.EarlyExitDisabled)

	// Early exit at ThresholdReject: Phase 1 score >= 8, Phase 2 skipped
	early := Score(address("g1spammer"), msg, burstRate, badRep, nil, nil, nil, nil, engine.ThresholdReject)

	// Both should exceed ThresholdReject
	if full.Total < engine.ThresholdReject {
		t.Errorf("full scoring: expected score >= %d, got %d", engine.ThresholdReject, full.Total)
	}
	if early.Total < engine.ThresholdReject {
		t.Errorf("early exit: expected score >= %d, got %d", engine.ThresholdReject, early.Total)
	}

	// Early exit should have FEWER rules triggered (Phase 2 skipped)
	if len(early.Triggered) >= len(full.Triggered) {
		t.Errorf("expected early exit to trigger fewer rules: early=%v, full=%v", early.Triggered, full.Triggered)
	}
}

// ---- render

func TestRender(t *testing.T) {
	resetState(t)
	setAdmin(t)
	out := Render("")
	if out == "" {
		t.Error("expected non-empty Render output")
	}

	// Verify key sections are present
	checks := []string{
		"# Anti-Spam Scoring Realm",
		"**Admin:**",
		"**Corpus size:**",
		"**Fingerprint store:**",
		"**Keyword dictionary:**",
		"**Active patterns:**",
		"## Patterns",
		"## Usage",
	}
	for _, check := range checks {
		found := false
		for i := 0; i < len(out)-len(check)+1; i++ {
			if out[i:i+len(check)] == check {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Render output missing %q", check)
		}
	}
}
