package modexp

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"
)

func TestModExpKnownAnswers(t *testing.T) {
	cases := []struct {
		name string
		base []byte
		exp  []byte
		mod  []byte
		want []byte
	}{
		{"2^10 mod 1000", []byte{2}, []byte{10}, []byte{0x03, 0xe8}, []byte{0x00, 0x18}},  // 24
		{"7^2 mod 4", []byte{7}, []byte{2}, []byte{4}, []byte{0x01}},                      // 49 mod 4 = 1
		{"0^0 mod 5", []byte{0}, []byte{0}, []byte{5}, []byte{0x01}},                      // 0^0 = 1 by convention
		{"3^0 mod 7", []byte{3}, nil, []byte{7}, []byte{0x01}},                            // empty exp == 0
		{"empty base", nil, []byte{5}, []byte{7}, []byte{0}},                              // 0^5 = 0
		{"modulus 1", []byte{2}, []byte{3}, []byte{1}, []byte{0}},                         // anything mod 1 = 0
		{"modulus 0", []byte{2}, []byte{3}, []byte{0, 0, 0}, []byte{0, 0, 0}},             // m=0 => len(m) zero bytes
		{"empty modulus", []byte{2}, []byte{3}, nil, nil},                                 // len(m) == 0 => empty
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := X_modExp(tc.base, tc.exp, tc.mod)
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("got %x, want %x", got, tc.want)
			}
		})
	}
}

// TestModExpLargeModReduction exercises the cometblszk use case: reducing a
// 32-byte digest modulo (BN254_R - 1) by raising to the first power.
func TestModExpLargeModReduction(t *testing.T) {
	bn254RMinusOne, _ := hex.DecodeString(
		"30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000000",
	)

	// A value below the modulus must pass through unchanged.
	small, _ := hex.DecodeString(
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	)
	got := X_modExp(small, []byte{1}, bn254RMinusOne)
	if !bytes.Equal(got, small) {
		t.Fatalf("x < m should pass through: got %x", got)
	}

	// A value above the modulus must actually be reduced.
	big1, _ := hex.DecodeString(
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	)
	got = X_modExp(big1, []byte{1}, bn254RMinusOne)
	expected := new(big.Int).Mod(
		new(big.Int).SetBytes(big1),
		new(big.Int).SetBytes(bn254RMinusOne),
	)
	expectedBytes := make([]byte, len(bn254RMinusOne))
	expected.FillBytes(expectedBytes)
	if !bytes.Equal(got, expectedBytes) {
		t.Fatalf("reduction mismatch:\n got  %x\n want %x", got, expectedBytes)
	}
}

// TestModExpOutputLength: EIP-198 mandates output length = len(modulus),
// left-padded with zero bytes.
func TestModExpOutputLength(t *testing.T) {
	// 2 mod [0, 100]: result is 2, but padded to 2 bytes.
	got := X_modExp([]byte{2}, []byte{1}, []byte{0, 100})
	if len(got) != 2 {
		t.Fatalf("expected 2-byte output, got %d bytes", len(got))
	}
	if got[0] != 0 || got[1] != 2 {
		t.Fatalf("expected 0x0002, got %x", got)
	}
}
