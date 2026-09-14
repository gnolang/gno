package antispam

import (
	"testing"
)

// KeywordDict tests

func TestKeywordDictAddAndSize(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	if dict.Size() != 0 {
		t.Errorf("expected size 0, got %d", dict.Size())
	}

	dict.Add("spam", 3)
	dict.Add("scam", 2)
	dict.Add("free", 1)
	if dict.Size() != 3 {
		t.Errorf("expected size 3, got %d", dict.Size())
	}
}

func TestKeywordDictCaseInsensitive(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	dict.Add("SPAM", 3)

	// "SPAM" stored as "spam"; lookup via tokenized content should match
	w, ok := dict.GetWeight("spam")
	if !ok || w != 3 {
		t.Errorf("expected (3, true), got (%d, %t)", w, ok)
	}
}

func TestKeywordDictRemove(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	dict.Add("spam", 3)
	dict.Remove("spam")
	if dict.Size() != 0 {
		t.Errorf("expected size 0 after remove, got %d", dict.Size())
	}
}

func TestKeywordDictBulkAdd(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	// BulkAdd: "word:weight" pairs, newline-separated
	dict.BulkAdd("spam:3\nscam:2\nfree:1\nairdrop:2")
	if dict.Size() != 4 {
		t.Errorf("expected size 4, got %d", dict.Size())
	}
	w, ok := dict.GetWeight("scam")
	if !ok || w != 2 {
		t.Errorf("expected (2, true), got (%d, %t)", w, ok)
	}
}

func TestKeywordDictBulkAddDefaultWeight(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	// Words without :weight get default weight 1
	dict.BulkAdd("spam\nscam\nfree:2")
	if dict.Size() != 3 {
		t.Errorf("expected size 3, got %d", dict.Size())
	}
	w, _ := dict.GetWeight("spam")
	if w != 1 {
		t.Errorf("expected default weight 1, got %d", w)
	}
	w, _ = dict.GetWeight("free")
	if w != 2 {
		t.Errorf("expected weight 2, got %d", w)
	}
}

// ScoreKeywords tests

func TestScoreKeywords(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	dict.BulkAdd("spam:3\nscam:2\nfree:1\nairdrop:2\nbonus:1\ncasino:2\njackpot:2\nphishing:3")

	tests := []struct {
		name    string
		content string
		wantMin int
		wantMax int
		rule    string
	}{
		{
			name:    "multiple spam keywords",
			content: "Get free airdrop bonus from this spam scam site",
			wantMin: 3, // should trigger
			rule:    "KEYWORD_SPAM",
		},
		{
			name:    "high-weight keywords",
			content: "This is a spam phishing scam attempt against users",
			wantMin: 3,
			rule:    "KEYWORD_SPAM",
		},
		{
			name:    "no keywords in normal text",
			content: "I think we should adjust the staking rewards for validators in the next upgrade.",
			wantMax: 0,
		},
		{
			name:    "single keyword below threshold",
			content: "This discussion is really interesting and I enjoyed the free coffee at the meetup today.",
			wantMax: 0,
		},
		{
			name:    "empty content",
			content: "",
			wantMax: 0,
		},
		{
			name:    "nil dict handled in caller",
			content: "spam scam free airdrop",
			wantMax: 0, // tested separately below
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var d *KeywordDict
			if tt.name != "nil dict handled in caller" {
				d = dict
			}
			score, rule := ScoreKeywords(tt.content, d)

			if tt.wantMin > 0 {
				if score < tt.wantMin {
					t.Errorf("expected score >= %d, got %d", tt.wantMin, score)
				}
				if tt.rule != "" && rule != tt.rule {
					t.Errorf("expected rule %q, got %q", tt.rule, rule)
				}
			}
			if tt.wantMax == 0 && tt.wantMin == 0 {
				if score != 0 {
					t.Errorf("expected score 0, got %d (rule: %s)", score, rule)
				}
			}
		})
	}
}

func TestScoreKeywordsDeduplication(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	dict.Add("spam", 3)
	dict.Add("scam", 2)
	dict.Add("free", 1)

	// Repeating "spam" many times should NOT multiply the score
	score1, _ := ScoreKeywords("spam scam free", dict)
	score2, _ := ScoreKeywords("spam spam spam spam scam free", dict)
	if score1 != score2 {
		t.Errorf("repeated keywords should not change score: %d vs %d", score1, score2)
	}
}

func TestScoreKeywordsMultilingual(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	// French spam words
	dict.BulkAdd("gratuit:2\narnaque:3\ngagnez:2\ncliquez:2\ncadeau:1")

	score, rule := ScoreKeywords("Cliquez ici pour un cadeau gratuit, gagnez une arnaque", dict)
	if score < 3 {
		t.Errorf("expected multilingual keyword score >= 3, got %d (rule: %s)", score, rule)
	}
}

// Leet speak normalization tests

func TestNormalizeLeet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"fr33", "free"},
		{"sp4m", "spam"},
		{"a1rdr0p", "airdrop"},
		{"5c4m", "scam"},
		{"b0nu5", "bonus"},
		{"hello", "hello"},
		{"normal", "normal"},
		{"", ""},
		{"1337", "ieet"},
		{"7h3", "the"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := normalizeLeet(tt.input)
			if got != tt.want {
				t.Errorf("normalizeLeet(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestKeywordWeightClamping(t *testing.T) {
	t.Parallel()

	t.Run("weight above 3 clamped to 3", func(t *testing.T) {
		t.Parallel()
		dict := NewKeywordDict()
		dict.Add("spam", 5)
		w, ok := dict.GetWeight("spam")
		if !ok {
			t.Fatal("expected keyword to exist")
		}
		if w != 3 {
			t.Errorf("expected weight 3 (clamped), got %d", w)
		}
	})

	t.Run("weight 0 is ignored", func(t *testing.T) {
		t.Parallel()
		dict := NewKeywordDict()
		dict.Add("ignore", 0)
		if dict.Size() != 0 {
			t.Errorf("expected size 0 for weight 0, got %d", dict.Size())
		}
	})

	t.Run("negative weight is ignored", func(t *testing.T) {
		t.Parallel()
		dict := NewKeywordDict()
		dict.Add("ignore", -1)
		if dict.Size() != 0 {
			t.Errorf("expected size 0 for negative weight, got %d", dict.Size())
		}
	})

	t.Run("weight 3 is accepted as-is", func(t *testing.T) {
		t.Parallel()
		dict := NewKeywordDict()
		dict.Add("word", 3)
		w, ok := dict.GetWeight("word")
		if !ok || w != 3 {
			t.Errorf("expected (3, true), got (%d, %t)", w, ok)
		}
	})

	t.Run("weight 1 is accepted as-is", func(t *testing.T) {
		t.Parallel()
		dict := NewKeywordDict()
		dict.Add("word", 1)
		w, ok := dict.GetWeight("word")
		if !ok || w != 1 {
			t.Errorf("expected (1, true), got (%d, %t)", w, ok)
		}
	})
}

// Length-scaled minMatches tests

func TestScoreKeywordsLengthScaling(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	dict.BulkAdd("airdrop:2\nclaim:2\nmint:2\npresale:2\nwhitelist:2\nbonus:2\nfree:2\nprize:2\ncasino:3\njackpot:3")

	t.Run("short spam: 2 keywords in few tokens still triggers", func(t *testing.T) {
		t.Parallel()
		// ~5 unique tokens, 2 keywords -> minMatches = 2 (unchanged)
		score, rule := ScoreKeywords("claim your free airdrop now", dict)
		if score != WeightKeyword {
			t.Errorf("short content with 2 keywords should trigger, got score %d rule %q", score, rule)
		}
	})

	t.Run("long content: 2 keywords in many tokens does not trigger", func(t *testing.T) {
		t.Parallel()
		// Build a long legitimate crypto blog post with exactly 2 keywords
		// ("airdrop" and "claim") buried in ~100+ unique tokens.
		long := "The governance proposal for increasing validator delegation limits has been " +
			"submitted for community review and discussion among all stakeholders involved " +
			"in the network operations and maintenance procedures that ensure smooth " +
			"blockchain functionality across different regions and jurisdictions worldwide " +
			"The technical specification details the exact parameters being modified " +
			"including minimum stake requirements and maximum delegation ratios for each " +
			"validator node operator participating in consensus and block production " +
			"The upcoming airdrop distribution schedule will be announced next week " +
			"Community members should review the documentation and prepare their wallets " +
			"before the snapshot date to ensure eligibility for participation " +
			"Validators who want to claim their additional staking rewards must submit " +
			"their requests through the official portal before the deadline expires"
		score, rule := ScoreKeywords(long, dict)
		if score != 0 {
			t.Errorf("long content with only 2 keywords should not trigger, got score %d rule %q", score, rule)
		}
	})

	t.Run("long content: many keywords still triggers", func(t *testing.T) {
		t.Parallel()
		// Long content saturated with spam keywords should still trigger
		long := "Amazing free airdrop happening right now where you can claim your bonus " +
			"tokens and win a huge prize from our exclusive presale and whitelist event " +
			"that gives you guaranteed casino jackpot winnings every single day " +
			"The community discussion about governance and validator staking continues " +
			"with important updates about network performance and technical improvements"
		score, rule := ScoreKeywords(long, dict)
		if score != WeightKeyword {
			t.Errorf("long content saturated with keywords should trigger, got score %d rule %q", score, rule)
		}
	})

	t.Run("medium content: 2 keywords in moderate tokens does not trigger", func(t *testing.T) {
		t.Parallel()
		// Medium-length content with exactly 2 keywords ("airdrop" + "claim")
		// and enough unique tokens to scale minMatches above 2
		medium := "The recent discussion about protocol upgrades has generated significant " +
			"interest among developers and community members who are actively building " +
			"applications on the platform and testing various features including smart " +
			"contracts realms packages and other components that make the ecosystem great " +
			"Validators should prepare their infrastructure before the upcoming airdrop " +
			"distribution while stakers need to claim their accumulated rewards promptly"
		score, rule := ScoreKeywords(medium, dict)
		if score != 0 {
			t.Errorf("medium content with only 2 keywords should not trigger, got score %d rule %q", score, rule)
		}
	})
}

// KeywordDict cap tests

func TestKeywordDictMaxSize(t *testing.T) {
	t.Run("cap: new keywords rejected at capacity", func(t *testing.T) {
		dict := NewKeywordDict()
		for i := 0; i < keywordMaxSize; i++ {
			dict.Add("word"+intToKey(i), 1)
		}
		if dict.Size() != keywordMaxSize {
			t.Fatalf("expected size %d, got %d", keywordMaxSize, dict.Size())
		}
		// New keyword should be ignored at capacity.
		dict.Add("newword", 2)
		if dict.Size() != keywordMaxSize {
			t.Errorf("expected size to stay at %d, got %d", keywordMaxSize, dict.Size())
		}
		_, ok := dict.GetWeight("newword")
		if ok {
			t.Error("expected new keyword to be rejected at capacity")
		}
	})

	t.Run("cap: existing keywords update at capacity", func(t *testing.T) {
		dict := NewKeywordDict()
		dict.Add("existing", 1)
		for i := 0; i < keywordMaxSize-1; i++ {
			dict.Add("word"+intToKey(i), 1)
		}
		if dict.Size() != keywordMaxSize {
			t.Fatalf("expected size %d, got %d", keywordMaxSize, dict.Size())
		}
		// Updating existing keyword weight should still work.
		dict.Add("existing", 3)
		w, ok := dict.GetWeight("existing")
		if !ok || w != 3 {
			t.Errorf("expected weight update to 3, got (%d, %t)", w, ok)
		}
		if dict.Size() != keywordMaxSize {
			t.Errorf("expected size unchanged at %d, got %d", keywordMaxSize, dict.Size())
		}
	})
}

func TestScoreKeywordsLeetSpeak(t *testing.T) {
	t.Parallel()

	dict := NewKeywordDict()
	dict.BulkAdd("free:2\nairdrop:2\nspam:3\nscam:2\nbonus:1")

	tests := []struct {
		name    string
		content string
		wantMin int
		rule    string
	}{
		{
			name:    "leet: fr33 a1rdr0p detected as free airdrop",
			content: "Get fr33 a1rdr0p now for your wallet",
			wantMin: 3,
			rule:    "KEYWORD_SPAM",
		},
		{
			name:    "leet: sp4m 5c4m detected",
			content: "This is sp4m and 5c4m content watch out",
			wantMin: 3,
			rule:    "KEYWORD_SPAM",
		},
		{
			name:    "leet: mixed normal and leet",
			content: "Get free b0nu5 from this airdrop event",
			wantMin: 3,
			rule:    "KEYWORD_SPAM",
		},
		{
			name:    "leet: single leet keyword not enough",
			content: "I had a fr33 lunch today at the office",
			wantMin: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			score, rule := ScoreKeywords(tt.content, dict)
			if tt.wantMin > 0 {
				if score < tt.wantMin {
					t.Errorf("expected score >= %d, got %d (rule: %s)", tt.wantMin, score, rule)
				}
				if tt.rule != "" && rule != tt.rule {
					t.Errorf("expected rule %q, got %q", tt.rule, rule)
				}
			} else {
				if score != 0 {
					t.Errorf("expected score 0, got %d (rule: %s)", score, rule)
				}
			}
		})
	}
}
