package antispam

import (
	"testing"
)

// Fingerprint function tests

func TestFingerprint(t *testing.T) {
	t.Run("success: same text produces same fingerprint", func(t *testing.T) {
		fp1 := Fingerprint("Hello world this is a test message")
		fp2 := Fingerprint("Hello world this is a test message")
		if len(fp1) == 0 {
			t.Fatal("expected non-empty fingerprint")
		}
		if len(fp1) != len(fp2) {
			t.Fatalf("fingerprint lengths differ: %d vs %d", len(fp1), len(fp2))
		}
		for i := range fp1 {
			if fp1[i] != fp2[i] {
				t.Errorf("fingerprint[%d] differs: %d vs %d", i, fp1[i], fp2[i])
			}
		}
	})

	t.Run("success: different text produces different fingerprint", func(t *testing.T) {
		fp1 := Fingerprint("Hello world this is a test message")
		fp2 := Fingerprint("Completely different content about governance")
		same := 0
		for i := range fp1 {
			if i >= len(fp2) {
				break
			}
			if fp1[i] == fp2[i] {
				same++
			}
		}
		// Very different texts should have low similarity
		if len(fp1) > 0 && same*100/len(fp1) > 50 {
			t.Errorf("expected low similarity, got %d%% match", same*100/len(fp1))
		}
	})

	t.Run("success: similar text produces similar fingerprint", func(t *testing.T) {
		fp1 := Fingerprint("free airdrop bonus click claim prize now today")
		fp2 := Fingerprint("free airdrop bonus click claim prize hurry quick")
		same := 0
		for i := range fp1 {
			if i >= len(fp2) {
				break
			}
			if fp1[i] == fp2[i] {
				same++
			}
		}
		// Similar texts should have high similarity
		if len(fp1) > 0 && same*100/len(fp1) < 30 {
			t.Errorf("expected high similarity, got %d%% match", same*100/len(fp1))
		}
	})

	t.Run("edge: empty string", func(t *testing.T) {
		fp := Fingerprint("")
		if len(fp) != 0 {
			t.Errorf("expected empty fingerprint for empty string, got %d elements", len(fp))
		}
	})

	t.Run("edge: very short text", func(t *testing.T) {
		fp := Fingerprint("hi")
		if len(fp) != 0 {
			t.Errorf("expected empty fingerprint for very short text, got %d elements", len(fp))
		}
	})
}

// FingerprintStore tests

func TestFingerprintStore(t *testing.T) {
	t.Run("success: new store is empty", func(t *testing.T) {
		fs := NewFingerprintStore()
		if fs.Size() != 0 {
			t.Errorf("expected size 0, got %d", fs.Size())
		}
	})

	t.Run("success: add and find exact match", func(t *testing.T) {
		fs := NewFingerprintStore()
		fs.Add("free airdrop bonus click claim prize tokens")
		if !fs.HasSimilar("free airdrop bonus click claim prize tokens", 80) {
			t.Error("expected exact match to be found")
		}
	})

	t.Run("success: find near-duplicate", func(t *testing.T) {
		fs := NewFingerprintStore()
		fs.Add("free airdrop bonus click claim prize tokens hurry")
		if !fs.HasSimilar("free airdrop bonus click claim prize tokens quick", 50) {
			t.Error("expected near-duplicate to be found")
		}
	})

	t.Run("success: no match for different content", func(t *testing.T) {
		fs := NewFingerprintStore()
		fs.Add("free airdrop bonus click claim prize tokens")
		if fs.HasSimilar("governance proposal voting community discussion update", 50) {
			t.Error("expected no match for different content")
		}
	})

	t.Run("success: multiple fingerprints", func(t *testing.T) {
		fs := NewFingerprintStore()
		fs.Add("free airdrop bonus click claim prize tokens")
		fs.Add("casino jackpot winner spin bonus special deal")
		if !fs.HasSimilar("casino jackpot winner spin bonus amazing deal", 50) {
			t.Error("expected match for similar content")
		}
		if fs.Size() != 2 {
			t.Errorf("expected size 2, got %d", fs.Size())
		}
	})

	t.Run("success: high threshold rejects loose match", func(t *testing.T) {
		fs := NewFingerprintStore()
		fs.Add("free airdrop bonus click claim prize tokens hurry")
		if fs.HasSimilar("free discussion about tokens and community ideas", 80) {
			t.Error("expected high threshold to reject loose match")
		}
	})

	t.Run("edge: empty body", func(t *testing.T) {
		fs := NewFingerprintStore()
		fs.Add("some spam content here")
		if fs.HasSimilar("", 50) {
			t.Error("expected no match for empty body")
		}
	})

	t.Run("edge: add empty body is no-op", func(t *testing.T) {
		fs := NewFingerprintStore()
		fs.Add("")
		if fs.Size() != 0 {
			t.Errorf("expected size 0 after adding empty, got %d", fs.Size())
		}
	})

	t.Run("eviction: store caps at fpMaxStoreSize=500", func(t *testing.T) {
		fs := NewFingerprintStore()
		// Add 501 distinct fingerprints; oldest should be evicted
		for i := 0; i < 501; i++ {
			fs.Add("fingerprint test content number " + intToKey(i))
		}
		if fs.Size() != 500 {
			t.Errorf("expected store capped at 500, got %d", fs.Size())
		}
	})
}

// ScoreFingerprint tests

func TestScoreFingerprint(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		setup    func() *FingerprintStore
		wantMin  int
		wantRule string
		wantZero bool
	}{
		{
			name:     "nil store scores zero",
			body:     "some text",
			setup:    nil,
			wantZero: true,
		},
		{
			name:     "empty store scores zero",
			body:     "some text",
			setup:    func() *FingerprintStore { return NewFingerprintStore() },
			wantZero: true,
		},
		{
			name: "match: near-duplicate of known spam",
			body: "free airdrop bonus click claim prize quick",
			setup: func() *FingerprintStore {
				fs := NewFingerprintStore()
				fs.Add("free airdrop bonus click claim prize tokens")
				return fs
			},
			wantMin:  4,
			wantRule: "NEAR_DUPLICATE",
		},
		{
			name: "different content",
			body: "governance proposal for community voting",
			setup: func() *FingerprintStore {
				fs := NewFingerprintStore()
				fs.Add("free airdrop bonus click claim prize tokens")
				return fs
			},
			wantZero: true,
		},
		{
			name:     "empty body",
			body:     "",
			setup:    func() *FingerprintStore { return NewFingerprintStore() },
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fs *FingerprintStore
			if tt.setup != nil {
				fs = tt.setup()
			}
			score, rule := ScoreFingerprint(tt.body, fs)

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
