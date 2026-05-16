package cla

import (
	"testing"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/uassert/v0"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
const (
	testHash1 = "abc123def456"
	testHash2 = "xyz789uvw012"
	testUser1 = "g1user1address1234567890"
	testUser2 = "g1user2address0987654321"
)

func resetState() {
	signatures = addrset.Set{}
	requiredHash = ""
	claURL = ""
}

func TestSign(cur realm, t *testing.T) {
	resetState()

	setRequiredHash(testHash1)

	testing.SetRealm(testing.NewUserRealm(testUser1))
	Sign(cross2(cur), testHash1)

	uassert.True(t, HasValidSignature(address(testUser1)))
}

func TestSign_WrongHash(cur realm, t *testing.T) {
	resetState()

	setRequiredHash(testHash1)

	testing.SetRealm(testing.NewUserRealm(testUser1))
	uassert.AbortsWithMessage(t, cur, "hash does not match required CLA hash", func() {
		Sign(cross2(cur), testHash2)
	})

	uassert.False(t, HasValidSignature(address(testUser1)))
}

func TestHasValidSignature_Disabled(t *testing.T) {
	resetState()

	uassert.Equal(t, "", requiredHash)
	uassert.True(t, HasValidSignature(address(testUser1)))
	uassert.True(t, HasValidSignature(address(testUser2)))
}

func TestHasValidSignature_Valid(cur realm, t *testing.T) {
	resetState()

	setRequiredHash(testHash1)

	testing.SetRealm(testing.NewUserRealm(testUser1))
	Sign(cross2(cur), testHash1)

	uassert.True(t, HasValidSignature(address(testUser1)))
}

func TestHasValidSignature_NotSigned(t *testing.T) {
	resetState()

	setRequiredHash(testHash1)

	uassert.False(t, HasValidSignature(address(testUser1)))
}

func TestSetRequiredHash_ResetsSignatures(cur realm, t *testing.T) {
	resetState()

	setRequiredHash(testHash1)

	testing.SetRealm(testing.NewUserRealm(testUser1))
	Sign(cross2(cur), testHash1)
	uassert.True(t, HasValidSignature(address(testUser1)))
	uassert.Equal(t, 1, signatures.Size())

	// Update hash - should reset signatures
	setRequiredHash(testHash2)

	uassert.False(t, HasValidSignature(address(testUser1)))
	uassert.Equal(t, 0, signatures.Size())
}
