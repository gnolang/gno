package evmnet

import (
	"encoding/hex"
	"testing"
)

// A live evmnet beacon, fetched from api.drand.sh.
const (
	liveRound      = 21167346
	liveSig        = "05ac314484cc91d6a4fdac301d34949526af8ca11ccd3f028357b9596a07ee8504c39404fbd9e4d730a700c144aa05bf0a0963b897fb2e9ea9be9bd47382e9b2"
	liveRandomness = "1941955a14cb54f5a48cb42bdb54068fb8c3338b80a5aa433e109e8ca534a62c"
)

func TestVerifyLiveBeacon(t *testing.T) {
	sig, _ := hex.DecodeString(liveSig)
	if err := Verify(liveRound, sig); err != nil {
		t.Fatalf("valid beacon rejected: %v", err)
	}
	r := Randomness(sig)
	if got := hex.EncodeToString(r[:]); got != liveRandomness {
		t.Fatalf("randomness = %s, want %s", got, liveRandomness)
	}
}

func TestVerifyRejects(t *testing.T) {
	sig, _ := hex.DecodeString(liveSig)
	if err := Verify(liveRound+1, sig); err == nil {
		t.Fatal("signature accepted for the wrong round")
	}
	bad := append([]byte(nil), sig...)
	bad[10] ^= 1
	if err := Verify(liveRound, bad); err == nil {
		t.Fatal("tampered signature accepted")
	}
	if err := Verify(liveRound, sig[:63]); err == nil {
		t.Fatal("short signature accepted")
	}
}

func TestRoundTime(t *testing.T) {
	if RoundAt(GenesisTime) != 1 || TimeOf(1) != GenesisTime {
		t.Fatal("round 1 must be published at genesis")
	}
	for _, r := range []uint64{1, 2, 1000, liveRound} {
		if got := RoundAt(TimeOf(r)); got != r {
			t.Fatalf("RoundAt(TimeOf(%d)) = %d", r, got)
		}
		if got := RoundAt(TimeOf(r) + Period - 1); got != r {
			t.Fatalf("RoundAt just before next round = %d, want %d", got, r)
		}
	}
	if RoundAt(GenesisTime-1) != 0 {
		t.Fatal("no round before genesis")
	}
}
