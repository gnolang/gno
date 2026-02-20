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

	// Re-load default patterns
	for _, p := range []string{
		`(?i)send\s+\d+\s*(gnot|token|coin)`,
		`(?i)free\s*(airdrop|token|nft|mint)`,
		`(?i)(casino|poker|gambling|betting)\s*(bonus|jackpot|win|payout)`,
		`(?i)(telegram|discord|whatsapp)\s*(group|channel|chat)\s*.*https?://`,
		`(?i)double\s+your\s+(gnot|token|coin|crypto|money)`,
		`(?i)send\s+\d+.*get\s+\d+.*back`,
		`(?i)(connect|enter|verify)\s+(your\s+)?(wallet|seed\s*phrase|private\s*key|mnemonic)`,
		`(?i)(100|1000)x\s*(gem|token|coin|guaranteed|profit)`,
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
	setAdmin(t)
	AdminAddPattern(cross, `(?i)test\s+specific\s+pattern`)

	result := Score("g1user", "this is a test specific pattern here", engine.RateState{}, goodRep(), nil, nil)
	if result.Total < 5 {
		t.Errorf("expected blocked pattern score >= 5, got %d", result.Total)
	}

	// Cleanup
	AdminRemovePattern(cross, `(?i)test\s+specific\s+pattern`)
}

func TestAdminRemovePattern(t *testing.T) {
	setAdmin(t)
	AdminAddPattern(cross, `(?i)removable\s+pattern`)
	AdminRemovePattern(cross, `(?i)removable\s+pattern`)

	result := Score("g1user", "removable pattern here with enough length for content analysis", engine.RateState{}, goodRep(), nil, nil)
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
	setAdmin(t)
	AdminBlockAddress(cross, "g1blocktest")

	result := Score("g1blocktest", "Normal message.", engine.RateState{}, goodRep(), nil, nil)
	if result.Total < 99 {
		t.Errorf("expected blocked address score >= 99, got %d", result.Total)
	}

	// Cleanup
	AdminUnblockAddress(cross, "g1blocktest")
}

func TestAdminUnblockAddress(t *testing.T) {
	setAdmin(t)
	AdminBlockAddress(cross, "g1temp")
	AdminUnblockAddress(cross, "g1temp")

	result := Score("g1temp", "Normal message from unblocked user with reasonable length.", engine.RateState{}, goodRep(), nil, nil)
	if result.Total > 0 {
		t.Errorf("expected score 0 after unblock, got %d (triggered: %v)", result.Total, result.Triggered)
	}
}

func TestAdminAllowAddress(t *testing.T) {
	setAdmin(t)
	AdminAllowAddress(cross, "g1trusted")

	// Allowed user should score 0 even with spam content
	result := Score("g1trusted", "FREE AIRDROP BONUS CLICK NOW!!!", engine.RateState{PostCount: 50, WindowSeconds: 60}, engine.ReputationData{BanCount: 5}, nil, nil)
	if result.Total != 0 {
		t.Errorf("expected score 0 for allowed address, got %d (triggered: %v)", result.Total, result.Triggered)
	}

	// Cleanup
	AdminRemoveAllow(cross, "g1trusted")
}

// ---- admin training

func TestAdminTrain(t *testing.T) {
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
	result := Score("g1user", "free airdrop bonus click claim prize", engine.RateState{}, goodRep(), nil, nil)
	if result.Total < 3 {
		t.Errorf("expected bayes score >= 3 for spam content, got %d (triggered: %v)", result.Total, result.Triggered)
	}
}

// ---- admin fingerprint management

func TestAdminAddSpamFingerprint(t *testing.T) {
	setAdmin(t)
	AdminAddSpamFingerprint(cross, "free airdrop bonus click claim prize tokens hurry now limited")

	// Near-duplicate should be detected
	result := Score("g1user", "free airdrop bonus click claim prize tokens quick now fast", engine.RateState{}, goodRep(), nil, nil)
	if result.Total < 4 {
		t.Errorf("expected fingerprint score >= 4, got %d (triggered: %v)", result.Total, result.Triggered)
	}
}

// ---- default patterns

func TestDefaultPatternsLoaded(t *testing.T) {
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
		{"success: wallet phishing", "Connect your wallet to claim rewards", 5},
		{"success: pump and dump", "This 100x gem token is guaranteed profit", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Score("g1user", tt.content, engine.RateState{}, goodRep(), nil, nil)
			if result.Total < tt.wantMin {
				t.Errorf("expected score >= %d, got %d (triggered: %v)", tt.wantMin, result.Total, result.Triggered)
			}
		})
	}
}

// ---- default keywords

func TestDefaultKeywordsLoaded(t *testing.T) {
	if keywords.Size() < 100 {
		t.Errorf("expected at least 100 default keywords, got %d", keywords.Size())
	}

	// Content with multiple default spam keywords should trigger
	result := Score("g1user", "This is a phishing scam arnaque with ponzi scheme and ransomware", engine.RateState{}, goodRep(), nil, nil)
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
	setAdmin(t)
	AdminAddKeyword(cross, "customword1", 3)
	AdminAddKeyword(cross, "customword2", 3)
	AdminAddKeyword(cross, "customword3", 2)

	// Content with multiple custom keywords should trigger
	result := Score("g1user", "This has customword1 and customword2 plus customword3 inside it", engine.RateState{}, goodRep(), nil, nil)
	hasKeyword := false
	for _, r := range result.Triggered {
		if r == "KEYWORD_SPAM" {
			hasKeyword = true
		}
	}
	if !hasKeyword {
		t.Errorf("expected KEYWORD_SPAM rule, got triggered: %v", result.Triggered)
	}

	// Cleanup
	AdminRemoveKeyword(cross, "customword1")
	AdminRemoveKeyword(cross, "customword2")
	AdminRemoveKeyword(cross, "customword3")
}

func TestAdminBulkAddKeywords(t *testing.T) {
	setAdmin(t)
	sizeBefore := keywords.Size()
	AdminBulkAddKeywords(cross, "bulktest1:3\nbulktest2:2\nbulktest3:2\nbulktest4:3")

	if keywords.Size() != sizeBefore+4 {
		t.Errorf("expected size %d after bulk add, got %d", sizeBefore+4, keywords.Size())
	}

	// Content with bulk-added keywords should trigger
	result := Score("g1user", "This bulktest1 message has bulktest2 and bulktest3 words inside", engine.RateState{}, goodRep(), nil, nil)
	hasKeyword := false
	for _, r := range result.Triggered {
		if r == "KEYWORD_SPAM" {
			hasKeyword = true
		}
	}
	if !hasKeyword {
		t.Errorf("expected KEYWORD_SPAM after bulk add, got triggered: %v", result.Triggered)
	}

	// Cleanup
	AdminRemoveKeyword(cross, "bulktest1")
	AdminRemoveKeyword(cross, "bulktest2")
	AdminRemoveKeyword(cross, "bulktest3")
	AdminRemoveKeyword(cross, "bulktest4")
}

func TestAdminRemoveKeyword(t *testing.T) {
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
	setAdmin(t) // ensure admin is g1admin

	other := testutils.TestAddress("other")
	testing.SetOriginCaller(other)

	uassert.AbortsWithMessage(t, errUnauthorized.Error(), func() {
		AdminBlockAddress(cross, "g1someone")
	})
}

// ---- scoring

func TestScoreCleanContent(t *testing.T) {
	rep := engine.ReputationData{
		AccountAgeDays: 60,
		Balance:        5000000000,
		FlaggedCount:   0,
		TotalPosts:     20,
		HasUsername:    true,
		BanCount:       0,
	}

	result := Score("g1legit", "I think we should consider adjusting the staking rewards.", engine.RateState{PostCount: 3, WindowSeconds: 3600}, rep, nil, nil)
	if result.Total != 0 {
		t.Errorf("expected score 0 for clean content, got %d (triggered: %v)", result.Total, result.Triggered)
	}
}

// ---- admin transfer

func TestAdminTransfer(t *testing.T) {
	setAdmin(t)

	newAdmin := testutils.TestAddress("newadmin")
	AdminSetAdmin(cross, newAdmin)

	// New admin can operate
	testing.SetOriginCaller(newAdmin)
	AdminBlockAddress(cross, "g1transfertest")
	AdminUnblockAddress(cross, "g1transfertest")

	// Old admin should fail
	testing.SetOriginCaller(address("g1admin"))
	uassert.AbortsWithMessage(t, errUnauthorized.Error(), func() {
		AdminBlockAddress(cross, "g1shouldfail")
	})

	// Restore for other tests
	testing.SetOriginCaller(newAdmin)
	AdminSetAdmin(cross, address("g1admin"))
}

// ---- render

func TestRender(t *testing.T) {
	setAdmin(t)
	out := Render("")
	if out == "" {
		t.Error("expected non-empty Render output")
	}
}
