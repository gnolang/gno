package bls12381_test

import (
	"crypto/bls12381"
	"encoding/hex"
	"testing"
)

// Three signers with secret keys 0x1234567, 0x89abcdef and 0x7654321, each
// signing msg under the Ethereum POP ciphersuite; sig is their aggregate.
const (
	vecPubKeys = "820ad0f24a42c82129fef2a137f7b7c230c2aaffb78ffd82f6cbdcd2bfbf3560435a35c62d3ff66ad696b78f8c6c6c68" +
		"b6a005f54655f0a670599f6b9de9edffc39c99799a482f62fa5e42bfe986a93ae2c52360a96139ed82c4c4d6672695e5" +
		"ac7563bb556a1687d1dfc2b5fa2d5891c0756ba960f3cfc62de348a5549ddf65121efdbc06050988704674d7820f4a3a"
	vecSig = "857cbc9c0976c88648089b0598cdc7b8c75a96f8e4f06ea0f0bbd0e2d5ac74d9dae7d4c479a7b573f0d7f7e3b875a6b9" +
		"039083a95ea6f3af6d432056337bc115ae50023fa289c320322941a3b07db7a710b438f643432a1a56f64b121931357e"
	vecMsg = "gno.land bls12381 test vector"
)

func mustDecode(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestFastAggregateVerify(t *testing.T) {
	pks := mustDecode(vecPubKeys)
	sig := mustDecode(vecSig)

	if !bls12381.FastAggregateVerify(pks, []byte(vecMsg), sig) {
		t.Fatal("valid aggregate signature rejected")
	}
	if bls12381.FastAggregateVerify(pks, []byte("tampered"), sig) {
		t.Fatal("signature over another message accepted")
	}
	if bls12381.FastAggregateVerify(pks[bls12381.PubKeySize:], []byte(vecMsg), sig) {
		t.Fatal("signature with a missing signer accepted")
	}
	if bls12381.FastAggregateVerify(pks[1:], []byte(vecMsg), sig) {
		t.Fatal("misaligned public keys accepted")
	}
	if bls12381.FastAggregateVerify(nil, []byte(vecMsg), sig) {
		t.Fatal("empty public key set accepted")
	}
}
