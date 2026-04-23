package main

import (
	"strings"
	"testing"
)

func TestParseValsetList(t *testing.T) {
	// Three distinct pubkeys so Address() derivations are distinct too. Values
	// are real bech32 pubkeys from gnoland ed25519 keys; they're not
	// round-tripped here, just parsed.
	in := `
# comment line, ignored
node-1 1 gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0wau58zgeg7g9z5hn9k9p4emkjckpnfxhg5h30s7h08yza4dffwxqc8fqd

node-2 5 gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0wau58zgeg7g9z5hn9k9p4emkjckpnfxhg5h30s7h08yza4dffwxqc8fqd
node-3 10 gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0wau58zgeg7g9z5hn9k9p4emkjckpnfxhg5h30s7h08yza4dffwxqc8fqd
`
	got, err := parseValsetList(strings.NewReader(in))
	if err != nil {
		t.Fatalf("parseValsetList: unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 validators, got %d", len(got))
	}
	wantNames := []string{"node-1", "node-2", "node-3"}
	wantPowers := []int64{1, 5, 10}
	for i, v := range got {
		if v.Name != wantNames[i] {
			t.Errorf("[%d] Name = %q, want %q", i, v.Name, wantNames[i])
		}
		if v.Power != wantPowers[i] {
			t.Errorf("[%d] Power = %d, want %d", i, v.Power, wantPowers[i])
		}
		if v.PubKey == nil {
			t.Errorf("[%d] PubKey is nil", i)
			continue
		}
		if v.Address != v.PubKey.Address() {
			t.Errorf("[%d] Address %s != derived %s", i, v.Address, v.PubKey.Address())
		}
	}
}

func TestParseValsetList_BadLines(t *testing.T) {
	cases := map[string]string{
		"too-few-fields": "node-1 1\n",
		"bad-power":      "node-1 NaN gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0wau58zgeg7g9z5hn9k9p4emkjckpnfxhg5h30s7h08yza4dffwxqc8fqd\n",
		"bad-pubkey":     "node-1 1 not-a-pubkey\n",
		"empty":          "\n\n# just comments\n",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseValsetList(strings.NewReader(in)); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
