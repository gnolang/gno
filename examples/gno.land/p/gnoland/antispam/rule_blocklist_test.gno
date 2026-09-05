package antispam

import (
	"testing"
)

// Blocklist type tests

func TestBlocklistAddress(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Blocklist
		addr    string
		blocked bool
	}{
		{
			name:    "empty blocklist not blocked",
			setup:   NewBlocklist,
			addr:    "g1user123",
			blocked: false,
		},
		{
			name: "blocked address is blocked",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.BlockAddress("g1spammer")
				return bl
			},
			addr:    "g1spammer",
			blocked: true,
		},
		{
			name: "other address not blocked",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.BlockAddress("g1spammer")
				return bl
			},
			addr:    "g1legit",
			blocked: false,
		},
		{
			name: "unblock removes block",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.BlockAddress("g1spammer")
				bl.UnblockAddress("g1spammer")
				return bl
			},
			addr:    "g1spammer",
			blocked: false,
		},
		{
			name: "multiple addresses",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.BlockAddress("g1spam1")
				bl.BlockAddress("g1spam2")
				bl.BlockAddress("g1spam3")
				return bl
			},
			addr:    "g1spam2",
			blocked: true,
		},
		{
			name: "unblock non-existent address",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.UnblockAddress("g1nobody")
				return bl
			},
			addr:    "g1nobody",
			blocked: false,
		},
		{
			name: "double block same address",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.BlockAddress("g1spammer")
				bl.BlockAddress("g1spammer")
				return bl
			},
			addr:    "g1spammer",
			blocked: true,
		},
		{
			name:    "empty address",
			setup:   NewBlocklist,
			addr:    "",
			blocked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl := tt.setup()
			got := bl.IsBlocked(tt.addr)
			if got != tt.blocked {
				t.Errorf("IsBlocked(%q) = %v, want %v", tt.addr, got, tt.blocked)
			}
		})
	}
}

func TestBlocklistAllowlist(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Blocklist
		addr    string
		allowed bool
	}{
		{
			name:    "empty allowlist not allowed",
			setup:   NewBlocklist,
			addr:    "g1user",
			allowed: false,
		},
		{
			name: "allowed address is allowed",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AllowAddress("g1trusted")
				return bl
			},
			addr:    "g1trusted",
			allowed: true,
		},
		{
			name: "remove allow",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AllowAddress("g1trusted")
				bl.RemoveAllow("g1trusted")
				return bl
			},
			addr:    "g1trusted",
			allowed: false,
		},
		{
			name: "allow takes priority over block",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.BlockAddress("g1user")
				bl.AllowAddress("g1user")
				return bl
			},
			addr:    "g1user",
			allowed: true,
		},
		{
			name: "multiple allowed addresses",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AllowAddress("g1admin1")
				bl.AllowAddress("g1admin2")
				bl.AllowAddress("g1admin3")
				return bl
			},
			addr:    "g1admin2",
			allowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl := tt.setup()
			got := bl.IsAllowed(tt.addr)
			if got != tt.allowed {
				t.Errorf("IsAllowed(%q) = %v, want %v", tt.addr, got, tt.allowed)
			}
		})
	}
}

func TestBlocklistPatterns(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Blocklist
		content   string
		wantMatch bool
	}{
		{
			name:      "success: no patterns, no match",
			setup:     NewBlocklist,
			content:   "Hello world",
			wantMatch: false,
		},
		{
			name: "simple word pattern matches",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)free\s*airdrop`)
				return bl
			},
			content:   "Get your FREE AIRDROP now",
			wantMatch: true,
		},
		{
			name: "URL pattern matches",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`https?://(www\.)?scam\.com`)
				return bl
			},
			content:   "Visit https://www.scam.com for details",
			wantMatch: true,
		},
		{
			name: "pattern does not match clean content",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)free\s*airdrop`)
				return bl
			},
			content:   "This is a legitimate discussion about governance",
			wantMatch: false,
		},
		{
			name: "case insensitive pattern",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)casino\s*bonus`)
				return bl
			},
			content:   "Get your CASINO BONUS today",
			wantMatch: true,
		},
		{
			name: "multiple patterns via combined regex",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)spam`)
				bl.AddPattern(`(?i)scam`)
				return bl
			},
			content:   "This is a scam message",
			wantMatch: true,
		},
		{
			name: "remove pattern rebuilds combined",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)spam`)
				bl.RemovePattern(`(?i)spam`)
				return bl
			},
			content:   "This could be spam but pattern removed",
			wantMatch: false,
		},
		{
			name: "crypto scam pattern",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)send\s+\d+\s*gnot`)
				return bl
			},
			content:   "Send 1000 GNOT to this address and get 5000 back",
			wantMatch: true,
		},
		{
			name: "empty content",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)spam`)
				return bl
			},
			content:   "",
			wantMatch: false,
		},
		{
			name: "unicode content with ascii pattern",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)spam`)
				return bl
			},
			content:   "这是一条正常的中文消息",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl := tt.setup()
			if got := bl.MatchesPattern(tt.content); got != tt.wantMatch {
				t.Errorf("MatchesPattern() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

// Combined regex tests

func TestCombinedRegexBuilt(t *testing.T) {
	bl := NewBlocklist()
	bl.AddPattern(`(?i)free\s+airdrop`)
	bl.AddPattern(`(?i)send\s+\d+\s*gnot`)
	bl.AddPattern(`(?i)casino\s*bonus`)

	if bl.combined == nil {
		t.Fatal("expected combined regex to be built after AddPattern")
	}

	// Clean content: single scan, no match
	if bl.MatchesPattern("I think we should improve the staking rewards formula.") {
		t.Error("expected no match for clean content")
	}

	// Spam content: single scan, match
	if !bl.MatchesPattern("Get your free airdrop now!") {
		t.Error("expected match for spam content")
	}
}

func TestCombinedRegexRebuildOnRemove(t *testing.T) {
	bl := NewBlocklist()
	bl.AddPattern(`(?i)spam`)
	bl.AddPattern(`(?i)scam`)

	if bl.combined == nil {
		t.Fatal("expected combined regex after adding patterns")
	}

	// Remove one pattern, combined is rebuilt
	bl.RemovePattern(`(?i)spam`)
	if bl.MatchesPattern("this is spam") {
		t.Error("expected no match after removing spam pattern")
	}
	if !bl.MatchesPattern("this is a scam") {
		t.Error("expected match for remaining scam pattern")
	}

	// Remove all patterns, combined is nil
	bl.RemovePattern(`(?i)scam`)
	if bl.combined != nil {
		t.Error("expected combined to be nil after removing all patterns")
	}
	if bl.MatchesPattern("any content here") {
		t.Error("expected no match with no patterns")
	}
}

func TestBlocklistAddPatternError(t *testing.T) {
	bl := NewBlocklist()
	err := bl.AddPattern(`[invalid`)
	if err == nil {
		t.Error("expected error for invalid regex, got nil")
	}
}

func TestBlocklistPatternCap(t *testing.T) {
	bl := NewBlocklist()

	// Add exactly blMaxPatterns=30 patterns
	for i := 0; i < 30; i++ {
		err := bl.AddPattern("pattern" + intToKey(i))
		if err != nil {
			t.Fatalf("adding pattern %d failed: %v", i, err)
		}
	}

	if bl.PatternCount() != 30 {
		t.Errorf("expected 30 patterns, got %d", bl.PatternCount())
	}

	// 31st NEW pattern should fail
	err := bl.AddPattern("one_more_pattern")
	if err == nil {
		t.Error("expected error when exceeding pattern cap, got nil")
	}
	if bl.PatternCount() != 30 {
		t.Errorf("expected 30 patterns after failed add, got %d", bl.PatternCount())
	}

	// Updating an existing pattern should still work
	err = bl.AddPattern("pattern" + intToKey(0))
	if err != nil {
		t.Errorf("updating existing pattern should work, got error: %v", err)
	}
}
