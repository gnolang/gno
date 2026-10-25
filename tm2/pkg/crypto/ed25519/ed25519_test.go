package ed25519_test

import (
	"bytes"
	stded25519 "crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"testing"

	"filippo.io/edwards25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

func TestSignAndValidateEd25519(t *testing.T) {
	t.Parallel()

	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

const (
	privKeySecretGolden = "secret_golden"
	msgGolden           = "msg_golden"
	signedGolden        = "f9d4e6a665dfb6cd7e2fedf0d46a1725472e640a5e93d654ce4caa986e5defd23c8b3af76aa6e39c24c582f0ebee860f66254b29cf6d034ce461ae2773133703"
)

func TestSignAndVerifyGolden(t *testing.T) {
	privKey := ed25519.GenPrivKeyFromSecret([]byte(privKeySecretGolden))
	// pubKey := privKey.PubKey()
	out, err := privKey.Sign([]byte(msgGolden))
	require.NoError(t, err)

	hexOut := hex.EncodeToString(out)
	require.Equal(t, signedGolden, hexOut)
}

// smallOrderCandidates derives the small-order encodings from the curve rather
// than copying the table under test. filippo.io/edwards25519 is the upstream
// of Go's internal decoder and accepts the same non-canonical encodings.
//
// An encoding is a 255-bit y and a sign bit for x, and it decodes to the point
// with y mod p. The small-order points are the eight multiples of a point of
// order 8, and an unreduced y + p fits in 255 bits only for y <= 18, so every
// small-order encoding is a multiple's canonical encoding or y + p for some
// y <= 18, with either sign bit. y in [0, 18] is added too, as points that
// IsSmallOrder must not match. small marks the candidates that decode to a
// point of small order.
func smallOrderCandidates(t *testing.T) (candidates []ed25519.PubKeyEd25519, small map[ed25519.PubKeyEd25519]bool) {
	t.Helper()

	// [L]P lies in the small-order subgroup for every P. Scalars are reduced
	// mod L, so compute it as [L-1]P + P; it generates the subgroup when it
	// has order 8, which holds for about half of all P.
	one, err := edwards25519.NewScalar().SetCanonicalBytes(append([]byte{1}, make([]byte, 31)...))
	require.NoError(t, err)
	minusOne := edwards25519.NewScalar().Subtract(edwards25519.NewScalar(), one)
	identity := edwards25519.NewIdentityPoint()
	var gen *edwards25519.Point
	for i := 0; gen == nil; i++ {
		seed := sha256.Sum256([]byte{byte(i)})
		p, err := new(edwards25519.Point).SetBytes(seed[:])
		if err != nil {
			continue
		}
		q := new(edwards25519.Point).ScalarMult(minusOne, p)
		q.Add(q, p)
		q4 := new(edwards25519.Point).Add(q, q)
		q4.Add(q4, q4)
		if q4.Equal(identity) == 0 {
			gen = q
		}
	}

	seen := map[ed25519.PubKeyEd25519]bool{}
	add := func(enc ed25519.PubKeyEd25519) {
		for _, sign := range []byte{0, 0x80} {
			enc[31] = enc[31]&0x7f | sign
			if !seen[enc] {
				seen[enc] = true
				candidates = append(candidates, enc)
			}
		}
	}
	q := edwards25519.NewIdentityPoint()
	for range 8 {
		add(ed25519.PubKeyEd25519(q.Bytes()))
		q.Add(q, gen)
	}
	for y := range byte(19) {
		var canonical, unreduced ed25519.PubKeyEd25519
		canonical[0] = y
		unreduced[0] = 0xed + y // p = 2^255 - 19
		for i := 1; i < 31; i++ {
			unreduced[i] = 0xff
		}
		unreduced[31] = 0x7f
		add(canonical)
		add(unreduced)
	}

	small = map[ed25519.PubKeyEd25519]bool{}
	for _, c := range candidates {
		p, err := new(edwards25519.Point).SetBytes(c[:])
		if err == nil && new(edwards25519.Point).MultByCofactor(p).Equal(identity) == 1 {
			small[c] = true
		}
	}
	return candidates, small
}

// smallOrderPubKeys returns every byte string that crypto/ed25519 decodes to a
// point of small order.
func smallOrderPubKeys(t *testing.T) []ed25519.PubKeyEd25519 {
	t.Helper()

	_, small := smallOrderCandidates(t)
	keys := slices.Collect(maps.Keys(small))
	slices.SortFunc(keys, func(a, b ed25519.PubKeyEd25519) int { return bytes.Compare(a[:], b[:]) })
	return keys
}

// TestPubKeyIsSmallOrder checks the predicate against the curve: it matches
// exactly the candidates that decode to a point of small order, and there are
// 14 of them.
func TestPubKeyIsSmallOrder(t *testing.T) {
	t.Parallel()

	candidates, small := smallOrderCandidates(t)
	require.Len(t, small, 14)
	for _, c := range candidates {
		assert.Equal(t, small[c], c.IsSmallOrder(), "%x", c[:])
	}

	for range 100 {
		pub := ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)
		assert.False(t, pub.IsSmallOrder(), "a key derived from a private key has prime order")
	}
}

// TestVerifyBytesRejectsSmallOrderKeys pins the gate. Go's crypto/ed25519
// accepts a small-order pubkey, and under it a forged signature with S = 0 and
// a small-order R (for the identity, R = identity, for every message). Such a
// key stands for no signer, so VerifyBytes must refuse it for every caller.
func TestVerifyBytesRejectsSmallOrderKeys(t *testing.T) {
	t.Parallel()

	for _, pub := range smallOrderPubKeys(t) {
		msg, forged := forgeUnderSmallOrderKey(t, pub)
		assert.False(t, pub.VerifyBytes(msg, forged),
			"%s verified a forged signature for %q", pub, msg)
	}

	// The gate must not touch keys derived from a private key.
	priv := ed25519.GenPrivKey()
	msg := []byte("honest signatures still verify")
	sig, err := priv.Sign(msg)
	require.NoError(t, err)
	assert.True(t, priv.PubKey().VerifyBytes(msg, sig))
}

// forgeUnderSmallOrderKey returns a message and a signature that
// crypto/ed25519 accepts under pub with no private key: S = 0 and R one of the
// small-order points. Only the identity accepts R = identity for every message;
// for the other keys the working R depends on the message, so it varies a memo
// the way an attacker would.
func forgeUnderSmallOrderKey(t *testing.T, pub ed25519.PubKeyEd25519) (msg, sig []byte) {
	t.Helper()

	rs := smallOrderPubKeys(t)
	for i := range 64 {
		msg = fmt.Appendf(nil, "memo %d", i)
		for _, r := range rs {
			sig = make([]byte, ed25519.SignatureSize)
			copy(sig, r[:])
			if stded25519.Verify(pub[:], msg, sig) {
				return msg, sig
			}
		}
	}
	t.Fatalf("premise: crypto/ed25519 accepts no S = 0 forgery under %s", pub)
	return nil, nil
}
