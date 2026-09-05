package secp256k1_test

import (
	"crypto/secp256k1"
	"encoding/hex"
	"testing"
)

func TestVerify(t *testing.T) {
	pub, _ := hex.DecodeString("0261c53a33848f3e4fb5cc15ff0134c9cb44510114f5eec7f307e65cb5ae2cc037")
	sig, _ := hex.DecodeString("1f500a12579d23b817f199195d36aadcef75d8545d3d93613820cd071b6f3c664f5d6181b68889e9de700dfaeef4671ab0b6c6743971acd978f592dfea7e4e53")
	msg := []byte("hello gno.land")
	if !secp256k1.Verify(pub, msg, sig) {
		t.Error("verify failed for valid signature")
	}

	// Flipping a bit in the message must fail.
	tampered := append([]byte(nil), msg...)
	tampered[0] ^= 1
	if secp256k1.Verify(pub, tampered, sig) {
		t.Error("verify accepted tampered message")
	}

	// Flipping a bit in the signature must fail.
	badSig := append([]byte(nil), sig...)
	badSig[0] ^= 1
	if secp256k1.Verify(pub, msg, badSig) {
		t.Error("verify accepted tampered signature")
	}
}

func TestVerifyBadLengths(t *testing.T) {
	pub, _ := hex.DecodeString("0261c53a33848f3e4fb5cc15ff0134c9cb44510114f5eec7f307e65cb5ae2cc037")
	sig, _ := hex.DecodeString("1f500a12579d23b817f199195d36aadcef75d8545d3d93613820cd071b6f3c664f5d6181b68889e9de700dfaeef4671ab0b6c6743971acd978f592dfea7e4e53")
	msg := []byte("hello gno.land")

	if secp256k1.Verify(pub[:32], msg, sig) {
		t.Error("verify accepted short public key")
	}
	if secp256k1.Verify(pub, msg, sig[:63]) {
		t.Error("verify accepted short signature")
	}
	if secp256k1.Verify(nil, msg, sig) {
		t.Error("verify accepted nil public key")
	}
	if secp256k1.Verify(pub, msg, nil) {
		t.Error("verify accepted nil signature")
	}
}
