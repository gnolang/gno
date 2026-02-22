package antispam

import (
	"testing"
)

// ---- KeywordDict tests

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

// ---- ScoreKeywords tests

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
			name:    "success: multiple spam keywords",
			content: "Get free airdrop bonus from this spam scam site",
			wantMin: 3, // should trigger
			rule:    "KEYWORD_SPAM",
		},
		{
			name:    "success: high-weight keywords",
			content: "This is a spam phishing scam attempt against users",
			wantMin: 3,
			rule:    "KEYWORD_SPAM",
		},
		{
			name:    "clean: no keywords in normal text",
			content: "I think we should adjust the staking rewards for validators in the next upgrade.",
			wantMax: 0,
		},
		{
			name:    "clean: single keyword below threshold",
			content: "This discussion is really interesting and I enjoyed the free coffee at the meetup today.",
			wantMax: 0,
		},
		{
			name:    "edge: empty content",
			content: "",
			wantMax: 0,
		},
		{
			name:    "edge: nil dict handled in caller",
			content: "spam scam free airdrop",
			wantMax: 0, // tested separately below
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var d *KeywordDict
			if tt.name != "edge: nil dict handled in caller" {
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

// ---- Leet speak normalization tests

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

// ---- KeywordDict cap tests ----

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
