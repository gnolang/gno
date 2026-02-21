package antispam

import (
	"testing"
)

// ---- Blocklist type tests ----

func TestBlocklistAddress(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Blocklist
		addr    string
		blocked bool
	}{
		{
			name:    "success: empty blocklist not blocked",
			setup:   NewBlocklist,
			addr:    "g1user123",
			blocked: false,
		},
		{
			name: "success: blocked address is blocked",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.BlockAddress("g1spammer")
				return bl
			},
			addr:    "g1spammer",
			blocked: true,
		},
		{
			name: "success: other address not blocked",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.BlockAddress("g1spammer")
				return bl
			},
			addr:    "g1legit",
			blocked: false,
		},
		{
			name: "success: unblock removes block",
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
			name: "success: multiple addresses",
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
			name: "edge: unblock non-existent address",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.UnblockAddress("g1nobody")
				return bl
			},
			addr:    "g1nobody",
			blocked: false,
		},
		{
			name: "edge: double block same address",
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
			name:    "edge: empty address",
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
			name:    "success: empty allowlist not allowed",
			setup:   NewBlocklist,
			addr:    "g1user",
			allowed: false,
		},
		{
			name: "success: allowed address is allowed",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AllowAddress("g1trusted")
				return bl
			},
			addr:    "g1trusted",
			allowed: true,
		},
		{
			name: "success: remove allow",
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
			name: "success: allow takes priority over block",
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
			name: "success: multiple allowed addresses",
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
		name        string
		setup       func() *Blocklist
		content     string
		wantMatch   bool
		wantPattern string
	}{
		{
			name:      "success: no patterns, no match",
			setup:     NewBlocklist,
			content:   "Hello world",
			wantMatch: false,
		},
		{
			name: "success: simple word pattern matches",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)free\s*airdrop`)
				return bl
			},
			content:     "Get your FREE AIRDROP now",
			wantMatch:   true,
			wantPattern: `(?i)free\s*airdrop`,
		},
		{
			name: "success: URL pattern matches",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`https?://(www\.)?scam\.com`)
				return bl
			},
			content:     "Visit https://www.scam.com for details",
			wantMatch:   true,
			wantPattern: `https?://(www\.)?scam\.com`,
		},
		{
			name: "success: pattern does not match clean content",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)free\s*airdrop`)
				return bl
			},
			content:   "This is a legitimate discussion about governance",
			wantMatch: false,
		},
		{
			name: "success: case insensitive pattern",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)casino\s*bonus`)
				return bl
			},
			content:     "Get your CASINO BONUS today",
			wantMatch:   true,
			wantPattern: `(?i)casino\s*bonus`,
		},
		{
			name: "success: multiple patterns, first match returned",
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
			name: "success: remove pattern",
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
			name: "success: crypto scam pattern",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)send\s+\d+\s*gnot`)
				return bl
			},
			content:     "Send 1000 GNOT to this address and get 5000 back",
			wantMatch:   true,
			wantPattern: `(?i)send\s+\d+\s*gnot`,
		},
		{
			name: "edge: empty content",
			setup: func() *Blocklist {
				bl := NewBlocklist()
				bl.AddPattern(`(?i)spam`)
				return bl
			},
			content:   "",
			wantMatch: false,
		},
		{
			name: "edge: unicode content with ascii pattern",
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
			matched, pattern := bl.MatchesPattern(tt.content)
			if matched != tt.wantMatch {
				t.Errorf("MatchesPattern() matched = %v, want %v", matched, tt.wantMatch)
			}
			if tt.wantPattern != "" && pattern != tt.wantPattern {
				t.Errorf("MatchesPattern() pattern = %q, want %q", pattern, tt.wantPattern)
			}
		})
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
