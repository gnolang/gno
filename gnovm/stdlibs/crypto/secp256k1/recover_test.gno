package secp256k1_test

import (
	"crypto/keccak256"
	"crypto/secp256k1"
	"encoding/hex"
	"testing"
)

// Signer 0 of the recovery corpus in testdata/: private key 1, whose public
// key is the generator. highS is the same signature with s replaced by n-s.
const (
	recDigest  = "4c358ead6f3ea252278a6c0e69069eecec31c517786eba631a79352dc8f61c56"
	recR       = "a1842bb853abc7b49fe40f5af46b0f75aa6c962219e689f5a7ccea2b81d4f4e8"
	recS       = "1091ef7e7b447cac51d460e19a242e5237810ddc1373a925bfd83f77263d0c2f"
	recHighS   = "ef6e108184bb8353ae2b9f1e65dbd1ac832dcf0a9bd4f715fffa1f15a9f93512"
	recPubkey  = "0479be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"
	recAddress = "7e5f4552091a69125d5dfcb7b8c2659029395bdf"
	order      = "fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141"
)

// sig returns a fresh r || s || id signature.
func sig(t *testing.T, r, s string, id byte) []byte {
	b, err := hex.DecodeString(r + s)
	if err != nil {
		t.Fatal(err)
	}
	return append(b, id)
}

func digest(t *testing.T) []byte {
	b, err := hex.DecodeString(recDigest)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestRecoverPubkey(t *testing.T) {
	for _, s := range [][]byte{sig(t, recR, recS, 0), sig(t, recR, recHighS, 1)} {
		key, ok := secp256k1.RecoverPubkey(digest(t), s)
		if !ok || hex.EncodeToString(key[:]) != recPubkey {
			t.Errorf("got %x, %v; want %s", key, ok, recPubkey)
		}
		h := keccak256.Sum256(key[1:])
		if got := hex.EncodeToString(h[12:]); got != recAddress {
			t.Errorf("address %s, want %s", got, recAddress)
		}
	}
}

// A wrong recovery ID or digest recovers some other key, so callers must
// compare the result with the key they expect.
func TestRecoverPubkeyOtherKey(t *testing.T) {
	key, ok := secp256k1.RecoverPubkey(digest(t), sig(t, recR, recS, 1))
	if !ok || hex.EncodeToString(key[:]) == recPubkey {
		t.Errorf("wrong recovery ID: got %x, %v", key, ok)
	}
	d := digest(t)
	d[0] ^= 1
	key, ok = secp256k1.RecoverPubkey(d, sig(t, recR, recS, 0))
	if !ok || hex.EncodeToString(key[:]) == recPubkey {
		t.Errorf("wrong digest: got %x, %v", key, ok)
	}
}

func TestRecoverPubkeyRefuses(t *testing.T) {
	zero := "0000000000000000000000000000000000000000000000000000000000000000"
	one := "0000000000000000000000000000000000000000000000000000000000000001"
	five := "0000000000000000000000000000000000000000000000000000000000000005"
	gx := "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
	digestOne, err := hex.DecodeString(one)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		digest    []byte
		signature []byte
	}{
		{"r is 0", digest(t), sig(t, zero, recS, 0)},
		{"s is n", digest(t), sig(t, recR, order, 0)},
		{"no curve point at r", digest(t), sig(t, five, recS, 0)},
		// Digest 1, r the generator's x, s 1: the key would be G - G.
		{"point at infinity", digestOne, sig(t, gx, one, 0)},
		{"recovery ID 2", digest(t), sig(t, recR, recS, 2)},
		{"recovery ID 27", digest(t), sig(t, recR, recS, 27)},
		{"nil inputs", nil, nil},
		{"31-byte digest", digest(t)[:31], sig(t, recR, recS, 0)},
		{"33-byte digest", append(digest(t), 0), sig(t, recR, recS, 0)},
		{"64-byte signature", digest(t), sig(t, recR, recS, 0)[:64]},
		{"66-byte signature", digest(t), append(sig(t, recR, recS, 0), 0)},
	}
	for _, c := range cases {
		key, ok := secp256k1.RecoverPubkey(c.digest, c.signature)
		if ok || key != [65]byte{} {
			t.Errorf("%s: got %x, %v; want a zero key and false", c.name, key, ok)
		}
	}
}

// Arguments that are windows into larger buffers give the same key, and every
// byte of those buffers is left as it was.
func TestRecoverPubkeyWindows(t *testing.T) {
	d, s := digest(t), sig(t, recR, recS, 0)
	dbuf := make([]byte, 3+len(d)+29)
	sbuf := make([]byte, 7+len(s)+57)
	for i := range dbuf {
		dbuf[i] = byte(i)
	}
	for i := range sbuf {
		sbuf[i] = byte(255 - i)
	}
	copy(dbuf[3:], d)
	copy(sbuf[7:], s)
	dbefore, sbefore := string(dbuf), string(sbuf)

	key, ok := secp256k1.RecoverPubkey(dbuf[3:3+len(d)], sbuf[7:7+len(s)])
	if !ok || hex.EncodeToString(key[:]) != recPubkey {
		t.Errorf("got %x, %v; want %s", key, ok, recPubkey)
	}
	if string(dbuf) != dbefore || string(sbuf) != sbefore {
		t.Error("RecoverPubkey changed its inputs")
	}
}
