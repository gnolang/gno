package ed25519_test

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"
)

func TestVerify(t *testing.T) {
	publicKey, _ := hex.DecodeString("0D853FA898A07EB91F618BB3E8B738B0E45BE9B3053799A2C42F8204F5FA3505")
	signature, _ := hex.DecodeString("2B39638983858715AD2FA059665ADFE267936B8F20C4DA01E9650958E0CA65C0255C75164360F468087FE8385140E48EE3471E332472A50AEE599F9D0EADD106")
	if !ed25519.Verify(publicKey, []byte("hello gno.land"), signature) {
		t.Error("verify failed")
	}
}

// TestVerifyRejectsSmallOrderKey covers the realm-facing native. A realm that
// authenticates a caller by a user-supplied pubkey is forgeable if the native
// accepts small-order keys: under the identity, R = identity and S = 0 verify
// for every message.
func TestVerifyRejectsSmallOrderKey(t *testing.T) {
	// libsodium's ed25519_is_small_order list; each entry also has a
	// sign-bit-set spelling, for 14 encodings of the eight small-order points.
	encodings := []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0100000000000000000000000000000000000000000000000000000000000000",
		"26e8958fc2b227b045c3f489f2ef98f0d5dfac05d3c63339b13802886d53fc05",
		"c7176a703d4dd84fba3c0b760d10670f2a2053fa2c39ccc64ec7fd7792ac037a",
		"ecffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f",
		"edffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f",
		"eeffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f",
	}

	signature := make([]byte, 64)
	signature[0] = 1 // R = identity, S = 0

	for _, h := range encodings {
		base, err := hex.DecodeString(h)
		if err != nil {
			t.Fatal("bad test vector: " + h)
		}
		for _, signBit := range []byte{0, 0x80} {
			pub := make([]byte, 32)
			copy(pub, base)
			pub[31] |= signBit
			if ed25519.Verify(pub, []byte("hello gno.land"), signature) {
				t.Error("small-order key verified a forged signature: " + hex.EncodeToString(pub))
			}
		}
	}
}
