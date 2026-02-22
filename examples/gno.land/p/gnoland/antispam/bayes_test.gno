package antispam

import (
	"testing"
)

// Tokenizer tests

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantLen  int
		contains []string
		excludes []string
	}{
		{
			name:     "simple sentence",
			text:     "Hello world this is a test",
			contains: []string{"hello", "world", "test"},
			excludes: []string{"is", "a"}, // short words filtered
		},
		{
			name:     "mixed case normalized",
			text:     "HELLO World TeSt",
			contains: []string{"hello", "world", "test"},
		},
		{
			name:    "empty string",
			text:    "",
			wantLen: 0,
		},
		{
			name:    "only short words",
			text:    "I am a go to do",
			wantLen: 0,
		},
		{
			name:     "with punctuation",
			text:     "Hello, world! This is great.",
			contains: []string{"hello", "world", "this", "great"},
		},
		{
			name:     "with URLs stripped",
			text:     "Visit https://example.com for more info",
			contains: []string{"visit", "more", "info"},
			excludes: []string{"https", "example", "com"},
		},
		{
			name:     "numbers preserved",
			text:     "Send 1000 GNOT to address",
			contains: []string{"send", "1000", "gnot", "address"},
		},
		{
			name:     "duplicates preserved",
			text:     "spam spam spam hello spam",
			contains: []string{"spam", "hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.text)

			if tt.wantLen > 0 && len(tokens) != tt.wantLen {
				t.Errorf("expected %d tokens, got %d: %v", tt.wantLen, len(tokens), tokens)
			}
			if tt.wantLen == 0 && len(tt.contains) == 0 && len(tokens) != 0 {
				t.Errorf("expected 0 tokens, got %d: %v", len(tokens), tokens)
			}

			for _, want := range tt.contains {
				found := false
				for _, tok := range tokens {
					if tok == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected token %q not found in %v", want, tokens)
				}
			}

			for _, exclude := range tt.excludes {
				for _, tok := range tokens {
					if tok == exclude {
						t.Errorf("unexpected token %q found in %v", exclude, tokens)
						break
					}
				}
			}
		})
	}
}

// Corpus tests

func TestCorpus(t *testing.T) {
	t.Run("success: new corpus is empty", func(t *testing.T) {
		c := NewCorpus()
		if c.Size() != 0 {
			t.Errorf("expected size 0, got %d", c.Size())
		}
		spam, ham := c.GetTokenStats("hello")
		if spam != 0 || ham != 0 {
			t.Errorf("expected (0, 0), got (%d, %d)", spam, ham)
		}
	})

	t.Run("success: train spam increases spam count", func(t *testing.T) {
		c := NewCorpus()
		c.Train("free airdrop bonus click now", true)
		spam, ham := c.GetTokenStats("free")
		if spam != 1 {
			t.Errorf("expected spam=1, got %d", spam)
		}
		if ham != 0 {
			t.Errorf("expected ham=0, got %d", ham)
		}
	})

	t.Run("success: train ham increases ham count", func(t *testing.T) {
		c := NewCorpus()
		c.Train("legitimate governance proposal discussion", false)
		spam, ham := c.GetTokenStats("governance")
		if spam != 0 {
			t.Errorf("expected spam=0, got %d", spam)
		}
		if ham != 1 {
			t.Errorf("expected ham=1, got %d", ham)
		}
	})

	t.Run("success: multiple trainings accumulate", func(t *testing.T) {
		c := NewCorpus()
		c.Train("free bonus airdrop click here", true)
		c.Train("free tokens airdrop claim now", true)
		c.Train("free speech is important topic", false)

		spam, ham := c.GetTokenStats("free")
		if spam != 2 {
			t.Errorf("expected spam=2, got %d", spam)
		}
		if ham != 1 {
			t.Errorf("expected ham=1, got %d", ham)
		}
	})

	t.Run("success: size reflects unique tokens", func(t *testing.T) {
		c := NewCorpus()
		c.Train("hello world test", true)
		c.Train("hello again more words", false)
		if c.Size() < 4 {
			t.Errorf("expected size >= 4, got %d", c.Size())
		}
	})

	t.Run("edge: train empty string", func(t *testing.T) {
		c := NewCorpus()
		c.Train("", true)
		if c.Size() != 0 {
			t.Errorf("expected size 0 after empty train, got %d", c.Size())
		}
	})
}

// ScoreBayes tests

func TestScoreBayes(t *testing.T) {
	// Build a trained corpus
	trainedCorpus := func() *Corpus {
		c := NewCorpus()
		// Train with spam
		c.Train("free airdrop bonus click claim tokens", true)
		c.Train("free money send gnot prize winner", true)
		c.Train("casino bonus jackpot free spin win", true)
		c.Train("airdrop claim free tokens hurry limited", true)
		c.Train("click here free prize bonus amazing", true)
		// Train with ham
		c.Train("governance proposal voting community discussion", false)
		c.Train("validator staking delegation reward", false)
		c.Train("development update release version changelog", false)
		c.Train("question about using gnokey transaction", false)
		c.Train("review this pull request changes code", false)
		c.Train("interesting discussion about protocol design", false)
		c.Train("thanks for the helpful explanation today", false)
		return c
	}

	tests := []struct {
		name     string
		body     string
		corpus   func() *Corpus
		wantMin  int
		wantRule string
		wantZero bool
	}{
		{
			name:     "nil corpus scores zero",
			body:     "free airdrop bonus",
			corpus:   nil,
			wantZero: true,
		},
		{
			name:     "empty corpus scores zero",
			body:     "free airdrop bonus",
			corpus:   func() *Corpus { return NewCorpus() },
			wantZero: true,
		},
		{
			name:     "obvious spam words",
			body:     "free airdrop bonus click claim prize",
			corpus:   trainedCorpus,
			wantMin:  3,
			wantRule: "BAYES_SPAM",
		},
		{
			name:     "ham: legitimate discussion",
			body:     "governance proposal for community voting process",
			corpus:   trainedCorpus,
			wantZero: true,
		},
		{
			name:     "ham: development talk",
			body:     "development update and release changelog review",
			corpus:   trainedCorpus,
			wantZero: true,
		},
		{
			name:     "mixed: some spam words in normal context",
			body:     "this proposal is free for the community discussion",
			corpus:   trainedCorpus,
			wantZero: true,
		},
		{
			name:     "empty body",
			body:     "",
			corpus:   trainedCorpus,
			wantZero: true,
		},
		{
			name:     "unknown words only",
			body:     "xylophone zebra platypus kumquat",
			corpus:   trainedCorpus,
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c *Corpus
			if tt.corpus != nil {
				c = tt.corpus()
			}
			score, rule := ScoreBayes(tt.body, c)

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
				t.Errorf("expected rule %q, got %q", tt.wantRule, rule)
			}
		})
	}
}

// Bayes threshold boundary tests

func TestBayesThresholdBoundaries(t *testing.T) {
	t.Run("boundary: corpus with 9 tokens does NOT score", func(t *testing.T) {
		c := NewCorpus()
		// 9 unique tokens: below bayesMinCorpusSize=10
		c.Train("aaa bbb ccc ddd eee fff ggg hhh iii", true)
		if c.Size() != 9 {
			t.Fatalf("test setup: expected corpus size 9, got %d", c.Size())
		}
		score, _ := ScoreBayes("aaa bbb ccc ddd eee", c)
		if score != 0 {
			t.Errorf("corpus with 9 tokens should not score, got %d", score)
		}
	})

	t.Run("boundary: corpus with 10 tokens DOES score spam", func(t *testing.T) {
		c := NewCorpus()
		// Spam: 5 tokens trained 3x (high spam ratio)
		c.Train("aaa bbb ccc ddd eee", true)
		c.Train("aaa bbb ccc ddd eee", true)
		c.Train("aaa bbb ccc ddd eee", true)
		// Ham: 5 different tokens (total unique: 10)
		c.Train("fff ggg hhh iii jjj", false)
		c.Train("fff ggg hhh iii jjj", false)
		c.Train("fff ggg hhh iii jjj", false)
		if c.Size() != 10 {
			t.Fatalf("test setup: expected corpus size 10, got %d", c.Size())
		}
		// Content with 3+ spam tokens: should trigger
		score, rule := ScoreBayes("aaa bbb ccc xxx yyy", c)
		if score == 0 {
			t.Error("corpus with 10 tokens should score spam content")
		}
		if rule != "BAYES_SPAM" {
			t.Errorf("expected rule BAYES_SPAM, got %q", rule)
		}
	})

	t.Run("boundary: only 2 spam tokens does NOT trigger", func(t *testing.T) {
		c := NewCorpus()
		c.Train("aaa bbb ccc ddd eee", true)
		c.Train("aaa bbb ccc ddd eee", true)
		c.Train("fff ggg hhh iii jjj", false)
		c.Train("fff ggg hhh iii jjj", false)
		if c.Size() < 10 {
			t.Fatalf("test setup: expected corpus size >= 10, got %d", c.Size())
		}
		// Only 2 spam tokens (< bayesMinSpamTokens=3)
		score, _ := ScoreBayes("aaa bbb xxx yyy zzz", c)
		if score != 0 {
			t.Errorf("only 2 spam tokens should not trigger, got score %d", score)
		}
	})

	t.Run("boundary: exactly 3 spam tokens triggers", func(t *testing.T) {
		c := NewCorpus()
		c.Train("aaa bbb ccc ddd eee", true)
		c.Train("aaa bbb ccc ddd eee", true)
		c.Train("fff ggg hhh iii jjj", false)
		c.Train("fff ggg hhh iii jjj", false)
		if c.Size() < 10 {
			t.Fatalf("test setup: expected corpus size >= 10, got %d", c.Size())
		}
		// Exactly 3 spam tokens (= bayesMinSpamTokens)
		score, rule := ScoreBayes("aaa bbb ccc xxx yyy", c)
		if score == 0 {
			t.Error("exactly 3 spam tokens should trigger BAYES_SPAM")
		}
		if rule != "BAYES_SPAM" {
			t.Errorf("expected rule BAYES_SPAM, got %q", rule)
		}
	})

	t.Run("boundary: single-occurrence tokens are unreliable", func(t *testing.T) {
		c := NewCorpus()
		// Each spam token appears only once - insufficient evidence
		c.Train("aaa bbb ccc ddd eee", true)
		// Ham tokens to reach bayesMinCorpusSize=10
		c.Train("fff ggg hhh iii jjj", false)
		if c.Size() < 10 {
			t.Fatalf("test setup: expected corpus size >= 10, got %d", c.Size())
		}
		// All tokens have total=1 (seen once). Despite 100% spam ratio,
		// single observations are unreliable and should be ignored.
		score, _ := ScoreBayes("aaa bbb ccc ddd eee", c)
		if score != 0 {
			t.Errorf("single-occurrence tokens should not trigger Bayes, got score %d", score)
		}
	})
}

// Corpus cap tests

func TestCorpusMaxSize(t *testing.T) {
	t.Run("cap: existing tokens update beyond capacity", func(t *testing.T) {
		c := NewCorpus()
		c.Train("hello world test message here", true)
		initialSpam, _ := c.GetTokenStats("hello")
		if initialSpam != 1 {
			t.Fatalf("expected initial spam=1, got %d", initialSpam)
		}

		// Fill corpus to capacity with many unique tokens.
		for i := 0; i < 2500; i++ {
			c.Train("tok"+intToKey(i)+"abc def"+intToKey(i+10000)+"ghi", true)
		}

		if c.Size() > corpusMaxSize {
			t.Errorf("corpus exceeded max size: %d > %d", c.Size(), corpusMaxSize)
		}

		// Existing token "hello" should still be updatable
		c.Train("hello world test message here", true)
		updatedSpam, _ := c.GetTokenStats("hello")
		if updatedSpam <= initialSpam {
			t.Errorf("expected existing token to update beyond cap: initial=%d, updated=%d", initialSpam, updatedSpam)
		}
	})
}

// Token decay tests

func TestCorpusTokenDecay(t *testing.T) {
	t.Run("decay: counts stay bounded under heavy training", func(t *testing.T) {
		c := NewCorpus()
		n := 200
		// Train a token as spam 200 times.
		// Without decay, spam count would be 200.
		for i := 0; i < n; i++ {
			c.Train("target word here", true)
		}
		spam, _ := c.GetTokenStats("target")
		// Decay halves counts at tokenMaxTotal, so spam stays
		// at most tokenMaxTotal (can reach it, never exceeds).
		if spam > tokenMaxTotal {
			t.Errorf("expected spam <= %d (bounded by decay), got %d", tokenMaxTotal, spam)
		}
		// Verify decay actually happened: 200 trainings should NOT
		// produce spam=200.
		if spam >= n {
			t.Errorf("expected spam < %d (decay should reduce), got %d", n, spam)
		}
	})

	t.Run("decay: counter-training recovers from poisoning", func(t *testing.T) {
		c := NewCorpus()
		// Simulate poisoning attack: train "governance" as spam 100 times.
		for i := 0; i < 100; i++ {
			c.Train("governance target poison", true)
		}
		// Legitimate counter-training: 40 ham observations.
		// Without decay: spam=100, ham=40 -> ratio=71.4% (above 70%, still poisoned).
		// With decay: halving at total=100 reduces spam to 50, then ham=40
		//   -> ratio=50/90=55.5% (below 70%, recovered).
		for i := 0; i < 40; i++ {
			c.Train("governance target poison", false)
		}
		spam, ham := c.GetTokenStats("governance")
		total := spam + ham
		if total == 0 {
			t.Fatal("expected non-zero total for governance token")
		}
		ratio := spam * 100 / total
		if ratio > bayesSpamThresholdPct {
			t.Errorf("expected counter-training to reduce ratio below %d%%, got %d%% (spam=%d, ham=%d)",
				bayesSpamThresholdPct, ratio, spam, ham)
		}
	})

	t.Run("decay: small counts unaffected", func(t *testing.T) {
		c := NewCorpus()
		// Normal usage: a few trainings. Decay should not trigger.
		c.Train("hello world test", true)
		c.Train("hello world test", true)
		c.Train("hello world test", false)
		spam, ham := c.GetTokenStats("hello")
		if spam != 2 || ham != 1 {
			t.Errorf("expected (spam=2, ham=1) for small counts, got (%d, %d)", spam, ham)
		}
	})
}
