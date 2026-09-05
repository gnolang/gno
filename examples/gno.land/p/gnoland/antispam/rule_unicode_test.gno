package antispam

import (
	"testing"
)

func TestScoreUnicode(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantMin  int
		wantRule string
		wantZero bool
	}{
		// Success: legitimate content scores zero

		{
			name:     "normal english text",
			content:  "Hello world, this is a normal post about governance.",
			wantZero: true,
		},
		{
			name:     "accented french text",
			content:  "Café résumé naïve über straße. C'est très important.",
			wantZero: true,
		},
		{
			name:     "CJK text",
			content:  "你好世界，这是关于治理的讨论。",
			wantZero: true,
		},
		{
			name:     "pure cyrillic text",
			content:  "Привет мир, это обсуждение.",
			wantZero: true,
		},
		{
			name:     "pure greek text",
			content:  "Γεια σου κόσμε, αυτό είναι ένα μήνυμα.",
			wantZero: true,
		},
		{
			name:     "empty content",
			content:  "",
			wantZero: true,
		},
		{
			name:     "normal text with single accent",
			content:  "The word café has one accent and that is fine.",
			wantZero: true,
		},
		{
			name:     "numbers and ascii",
			content:  "Version 2.0 released on 2024-01-15 with 42 improvements.",
			wantZero: true,
		},

		// ZALGO_TEXT detection

		{
			name:     "zalgo: heavy combining marks on single char",
			content:  "H\u0300\u0301\u0302\u0303\u0304ello world",
			wantMin:  WeightZalgoText,
			wantRule: "ZALGO_TEXT",
		},
		{
			name:     "zalgo: multiple chars with stacked marks",
			content:  "T\u0300\u0301\u0302h\u0303\u0304\u0305i\u0306\u0307\u0308s is zalgo",
			wantMin:  WeightZalgoText,
			wantRule: "ZALGO_TEXT",
		},
		{
			name:     "zalgo: extended combining marks range",
			content:  "A\u1DC0\u1DC1\u1DC2\u1DC3normal text",
			wantMin:  WeightZalgoText,
			wantRule: "ZALGO_TEXT",
		},
		{
			name:     "zalgo: boundary 2 combining marks is ok",
			content:  "n\u0303\u0301 is a valid character cluster",
			wantZero: true,
		},

		// INVISIBLE_CHARS detection

		{
			name:     "invisible: zero-width spaces",
			content:  "He\u200Bllo wo\u200Brld",
			wantMin:  WeightInvisibleChars,
			wantRule: "INVISIBLE_CHARS",
		},
		{
			name:     "invisible: zero-width joiners",
			content:  "te\u200Dst\u200D message",
			wantMin:  WeightInvisibleChars,
			wantRule: "INVISIBLE_CHARS",
		},
		{
			name:     "invisible: directional overrides",
			content:  "Normal \u202Ehidden\u202C text",
			wantMin:  WeightInvisibleChars,
			wantRule: "INVISIBLE_CHARS",
		},
		{
			name:     "invisible: mixed invisible chars",
			content:  "a\u200Bb\u200Cc\u200Dd",
			wantMin:  WeightInvisibleChars,
			wantRule: "INVISIBLE_CHARS",
		},
		{
			name:     "invisible: single zero-width space is ok",
			content:  "Hello\u200Bworld",
			wantZero: true,
		},

		// HOMOGLYPH_MIX detection

		{
			name:     "homoglyph: cyrillic a in latin word",
			content:  "p\u0430yment received",
			wantMin:  WeightHomoglyphMix,
			wantRule: "HOMOGLYPH_MIX",
		},
		{
			name:     "homoglyph: cyrillic o in latin word",
			content:  "hell\u043E world",
			wantMin:  WeightHomoglyphMix,
			wantRule: "HOMOGLYPH_MIX",
		},
		{
			name:     "homoglyph: cyrillic e in latin word",
			content:  "s\u0435cure your account now",
			wantMin:  WeightHomoglyphMix,
			wantRule: "HOMOGLYPH_MIX",
		},
		{
			name:     "homoglyph: greek omicron in latin word",
			content:  "hell\u03BF world",
			wantMin:  WeightHomoglyphMix,
			wantRule: "HOMOGLYPH_MIX",
		},
		{
			name:     "homoglyph: separate script words is ok",
			content:  "Hello Привет",
			wantZero: true,
		},
		{
			name:     "homoglyph: all latin is ok",
			content:  "Hello world payment secure",
			wantZero: true,
		},

		// Combined unicode abuse

		{
			name:    "combined: zalgo + invisible chars",
			content: "H\u0300\u0301\u0302\u0303ello\u200B\u200Bworld",
			wantMin: WeightZalgoText + WeightInvisibleChars,
		},
		{
			name:    "combined: homoglyph + invisible",
			content: "p\u0430yment\u200B\u200Breceived",
			wantMin: WeightHomoglyphMix + WeightInvisibleChars,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := ScoreUnicode(tt.content)
			score, rule := hitsTotal(hits)

			if tt.wantZero {
				if score != 0 {
					t.Errorf("expected score 0, got %d (rule: %s)", score, rule)
				}
				return
			}

			if score < tt.wantMin {
				t.Errorf("expected score >= %d, got %d (rule: %s)", tt.wantMin, score, rule)
			}
			if tt.wantRule != "" && rule != tt.wantRule {
				t.Errorf("expected top rule %q, got %q (score: %d)", tt.wantRule, rule, score)
			}
		})
	}
}

func TestIsInvisibleChar(t *testing.T) {
	t.Parallel()

	invisible := []rune{
		0x200B, 0x200C, 0x200D, 0x200E, 0x200F,
		0x202A, 0x202B, 0x202C, 0x202D, 0x202E,
		0x2060, 0xFEFF, 0x00AD, 0x034F,
	}
	for _, r := range invisible {
		if !isInvisibleChar(r) {
			t.Errorf("expected rune U+%04X to be invisible", r)
		}
	}

	visible := []rune{'a', 'Z', '0', ' ', '.', 'é', 0x4E16}
	for _, r := range visible {
		if isInvisibleChar(r) {
			t.Errorf("expected rune U+%04X to NOT be invisible", r)
		}
	}
}

func TestIsCombiningMark(t *testing.T) {
	t.Parallel()

	combining := []rune{0x0300, 0x0301, 0x036F, 0x1DC0, 0xFE20}
	for _, r := range combining {
		if !isCombiningMark(r) {
			t.Errorf("expected rune U+%04X to be combining mark", r)
		}
	}

	notCombining := []rune{'a', 'A', '0', ' ', 0x0041, 0x00E9}
	for _, r := range notCombining {
		if isCombiningMark(r) {
			t.Errorf("expected rune U+%04X to NOT be combining mark", r)
		}
	}
}

func TestUnicodeThresholdBoundaries(t *testing.T) {
	t.Run("zalgo: exactly 3 combining marks triggers", func(t *testing.T) {
		// zalgoMinCombining=3, exactly 3 consecutive combining marks
		content := "A\u0300\u0301\u0302 normal text here"
		hits := ScoreUnicode(content)
		found := false
		for _, h := range hits {
			if h.Rule == "ZALGO_TEXT" {
				found = true
			}
		}
		if !found {
			t.Error("exactly 3 combining marks should trigger ZALGO_TEXT")
		}
	})

	t.Run("invisible: exactly 2 invisible chars triggers", func(t *testing.T) {
		// invisibleMinCount=2, exactly 2 invisible characters
		content := "Hello\u200B world\u200B normal text"
		hits := ScoreUnicode(content)
		found := false
		for _, h := range hits {
			if h.Rule == "INVISIBLE_CHARS" {
				found = true
			}
		}
		if !found {
			t.Error("exactly 2 invisible chars should trigger INVISIBLE_CHARS")
		}
	})
}

func TestScriptDetection(t *testing.T) {
	t.Parallel()

	// Latin characters
	latinChars := []rune{'a', 'z', 'A', 'Z', 0x00C0, 0x024F}
	for _, r := range latinChars {
		if !isLatin(r) {
			t.Errorf("expected rune U+%04X to be Latin", r)
		}
	}

	// Cyrillic characters
	cyrillicChars := []rune{0x0410, 0x0430, 0x044F, 0x04FF}
	for _, r := range cyrillicChars {
		if !isCyrillic(r) {
			t.Errorf("expected rune U+%04X to be Cyrillic", r)
		}
	}

	// Greek characters
	greekChars := []rune{0x0391, 0x03B1, 0x03C9, 0x03FF}
	for _, r := range greekChars {
		if !isGreek(r) {
			t.Errorf("expected rune U+%04X to be Greek", r)
		}
	}

	// Cross-checks: cyrillic is not latin
	if isLatin(0x0430) {
		t.Error("cyrillic а (U+0430) should not be detected as Latin")
	}
	if isCyrillic('a') {
		t.Error("latin a should not be detected as Cyrillic")
	}
}
