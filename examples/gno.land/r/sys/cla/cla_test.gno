package cla

import (
	"testing"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
)

const (
	testHash1    = "abc123def456"
	testHash2    = "xyz789uvw012"
	testUser1    = "g1user1address1234567890"
	testUser2    = "g1user2address0987654321"
	ownerAddr    = "g1replacemewithmultisig"
	nonOwnerAddr = "g1nottheowner1234567890"
)

func resetState() {
	signatures = addrset.Set{}
	requiredHash = ""
	claURL = ""
	owner = ownable.NewWithAddress(ownerAddr)
}

func TestSign(t *testing.T) {
	resetState()

	testing.SetRealm(testing.NewUserRealm(ownerAddr))
	SetRequiredHash(cross, testHash1)

	testing.SetRealm(testing.NewUserRealm(testUser1))
	Sign(cross, testHash1)

	uassert.True(t, HasValidSignature(address(testUser1)))
}

func TestSign_WrongHash(t *testing.T) {
	resetState()

	testing.SetRealm(testing.NewUserRealm(ownerAddr))
	SetRequiredHash(cross, testHash1)

	testing.SetRealm(testing.NewUserRealm(testUser1))
	uassert.AbortsWithMessage(t, "hash does not match required CLA hash", func() {
		Sign(cross, testHash2)
	})

	uassert.False(t, HasValidSignature(address(testUser1)))
}

func TestHasValidSignature_Disabled(t *testing.T) {
	resetState()

	uassert.Equal(t, "", requiredHash)
	uassert.True(t, HasValidSignature(address(testUser1)))
	uassert.True(t, HasValidSignature(address(testUser2)))
}

func TestHasValidSignature_Valid(t *testing.T) {
	resetState()

	testing.SetRealm(testing.NewUserRealm(ownerAddr))
	SetRequiredHash(cross, testHash1)

	testing.SetRealm(testing.NewUserRealm(testUser1))
	Sign(cross, testHash1)

	uassert.True(t, HasValidSignature(address(testUser1)))
}

func TestHasValidSignature_NotSigned(t *testing.T) {
	resetState()

	testing.SetRealm(testing.NewUserRealm(ownerAddr))
	SetRequiredHash(cross, testHash1)

	uassert.False(t, HasValidSignature(address(testUser1)))
}

func TestSetRequiredHash_ResetsSignatures(t *testing.T) {
	resetState()

	testing.SetRealm(testing.NewUserRealm(ownerAddr))
	SetRequiredHash(cross, testHash1)

	testing.SetRealm(testing.NewUserRealm(testUser1))
	Sign(cross, testHash1)
	uassert.True(t, HasValidSignature(address(testUser1)))
	uassert.Equal(t, 1, signatures.Size())

	// Admin updates hash - should reset signatures
	testing.SetRealm(testing.NewUserRealm(ownerAddr))
	SetRequiredHash(cross, testHash2)

	uassert.False(t, HasValidSignature(address(testUser1)))
	uassert.Equal(t, 0, signatures.Size())
}

func TestSetRequiredHash_Owner(t *testing.T) {
	resetState()

	testing.SetRealm(testing.NewUserRealm(ownerAddr))

	uassert.NotPanics(t, func() {
		SetRequiredHash(cross, testHash1)
	})

	uassert.Equal(t, testHash1, requiredHash)
}

func TestSetRequiredHash_NotOwner(t *testing.T) {
	resetState()

	testing.SetRealm(testing.NewUserRealm(nonOwnerAddr))

	uassert.AbortsWithMessage(t, ownable.ErrUnauthorized.Error(), func() {
		SetRequiredHash(cross, testHash1)
	})

	uassert.Equal(t, "", requiredHash)
}

func TestSetCLAURL_Owner(t *testing.T) {
	resetState()

	testing.SetRealm(testing.NewUserRealm(ownerAddr))

	uassert.NotPanics(t, func() {
		SetCLAURL(cross, "https://example.com/cla")
	})

	uassert.Equal(t, "https://example.com/cla", claURL)
}

func TestSetCLAURL_NotOwner(t *testing.T) {
	resetState()

	testing.SetRealm(testing.NewUserRealm(nonOwnerAddr))

	uassert.AbortsWithMessage(t, ownable.ErrUnauthorized.Error(), func() {
		SetCLAURL(cross, "https://example.com/cla")
	})

	uassert.Equal(t, "", claURL)
}

func TestTransferOwnership_Success(t *testing.T) {
	resetState()

  // newOwner should be a valid Gno address
	newOwner := testutils.TestAddress("new-owner")
	testing.SetRealm(testing.NewUserRealm(ownerAddr))

	uassert.NotPanics(t, func() {
		TransferOwnership(cross, newOwner)
	})

	uassert.Equal(t, owner.Owner(), newOwner)
}

func TestTransferOwnership_NotOwner(t *testing.T) {
	resetState()

  // newOwner should be a valid Gno address
	newOwner := testutils.TestAddress("new-owner")

	testing.SetRealm(testing.NewUserRealm(nonOwnerAddr))

	uassert.AbortsWithMessage(t, ownable.ErrUnauthorized.Error(), func() {
		TransferOwnership(cross, newOwner)
	})

	uassert.Equal(t, owner.Owner(), address(ownerAddr))
}
