package valopers

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable/exts/authorizable"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
)

func validValidatorInfo(t *testing.T) struct {
	Moniker     string
	Description string
	Address     std.Address
	PubKey      string
} {
	return struct {
		Moniker     string
		Description string
		Address     std.Address
		PubKey      string
	}{
		Moniker:     "test-1",
		Description: "test-1's description",
		Address:     std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h"),
		PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p",
	}
}

func TestValopers_Register(t *testing.T) {
	t.Run("already a valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		v := Valoper{
			Moniker:     validValidatorInfo(t).Moniker,
			Description: validValidatorInfo(t).Description,
			Address:     validValidatorInfo(t).Address,
			PubKey:      validValidatorInfo(t).PubKey,
			KeepRunning: true,
		}

		// Add the valoper
		valopers.Set(v.Address.String(), v)

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		uassert.PanicsWithMessage(t, ErrValoperExists.Error(), func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})
	})

	t.Run("no coins deposited", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send no coins
		std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", 0)}, nil)

		uassert.PanicsWithMessage(t, ufmt.Sprintf("payment must not be less than %d", minFee.Amount), func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})
	})

	t.Run("insufficient coins amount deposited", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send invalid coins
		std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee.Amount-1)}, nil)

		uassert.PanicsWithMessage(t, ufmt.Sprintf("payment must not be less than %d", minFee.Amount), func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})
	})

	t.Run("coin amount deposited is not ugnot", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send invalid coins
		std.TestSetOriginSend(std.Coins{std.NewCoin("gnogno", minFee.Amount)}, nil)

		uassert.PanicsWithMessage(t, "incompatible coin denominations: gnogno, ugnot", func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})
	})

	t.Run("successful registration", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validValidatorInfo(t).Address)

			uassert.Equal(t, validValidatorInfo(t).Moniker, valoper.Moniker)
			uassert.Equal(t, validValidatorInfo(t).Description, valoper.Description)
			uassert.Equal(t, validValidatorInfo(t).Address, valoper.Address)
			uassert.Equal(t, validValidatorInfo(t).PubKey, valoper.PubKey)
			uassert.Equal(t, true, valoper.KeepRunning)
		})
	})
}

func TestValopers_UpdateAuthMembers(t *testing.T) {
	test1Address := testutils.TestAddress("test1")
	test2Address := testutils.TestAddress("test2")

	t.Run("unauthorized member adds member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		std.TestSetRealm(std.NewUserRealm(test1Address))
		std.TestSetOriginCaller(test1Address) // TODO(bug, issue #2371): should not be needed

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		std.TestSetRealm(std.NewUserRealm(validValidatorInfo(t).Address))
		std.TestSetOriginCaller(validValidatorInfo(t).Address) // TODO(bug, issue #2371): should not be needed

		// try to add member without being authorized
		uassert.PanicsWithMessage(t, authorizable.ErrNotSuperuser.Error(), func() {
			AddToAuthList(validValidatorInfo(t).Address, test2Address)
		})
	})

	t.Run("unauthorized member deletes member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		std.TestSetRealm(std.NewUserRealm(test1Address))
		std.TestSetOriginCaller(test1Address) // TODO(bug, issue #2371): should not be needed

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		uassert.NotPanics(t, func() {
			AddToAuthList(validValidatorInfo(t).Address, test2Address)
		})

		std.TestSetRealm(std.NewUserRealm(validValidatorInfo(t).Address))
		std.TestSetOriginCaller(validValidatorInfo(t).Address) // TODO(bug, issue #2371): should not be needed

		// try to add member without being authorized
		uassert.PanicsWithMessage(t, authorizable.ErrNotSuperuser.Error(), func() {
			DeleteFromAuthList(validValidatorInfo(t).Address, test2Address)
		})
	})

	t.Run("authorized member adds member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		std.TestSetRealm(std.NewUserRealm(test1Address))
		std.TestSetOriginCaller(test1Address) // TODO(bug, issue #2371): should not be needed

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		uassert.NotPanics(t, func() {
			AddToAuthList(validValidatorInfo(t).Address, test2Address)
		})

		std.TestSetRealm(std.NewUserRealm(test2Address))
		std.TestSetOriginCaller(test2Address) // TODO(bug, issue #2371): should not be needed

		newMoniker := "new moniker"
		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateMoniker(validValidatorInfo(t).Address, newMoniker)
		})

		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validValidatorInfo(t).Address)
			uassert.Equal(t, newMoniker, valoper.Moniker)
		})
	})
}

func TestValopers_UpdateMoniker(t *testing.T) {
	test1Address := testutils.TestAddress("test1")
	test2Address := testutils.TestAddress("test2")

	t.Run("non-existing valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateMoniker(validValidatorInfo(t).Address, "new moniker")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		// Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateMoniker(validValidatorInfo(t).Address, "new moniker")
		})
	})

	t.Run("invalid moniker", func(t *testing.T) {
		invalidMonikers := []string{
			"",     // Empty
			"    ", // Whitespace
			"a",    // Too short
			"a very long moniker that is longer than 32 characters", // Too long
			"!@#$%^&*()+{}|:<>?/.,;'",                               // Invalid characters
			" space in front",
			"space in back ",
		}

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		for _, invalidMoniker := range invalidMonikers {
			// Update the valoper
			uassert.PanicsWithMessage(t, ErrInvalidMoniker.Error(), func() {
				UpdateMoniker(validValidatorInfo(t).Address, invalidMoniker)
			})
		}
	})

	t.Run("too long moniker", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidMoniker.Error(), func() {
			UpdateMoniker(validValidatorInfo(t).Address, strings.Repeat("a", MonikerMaxLength+1))
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		newMoniker := "new moniker"
		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateMoniker(validValidatorInfo(t).Address, newMoniker)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validValidatorInfo(t).Address)

			uassert.Equal(t, newMoniker, valoper.Moniker)
		})
	})
}

func TestValopers_UpdateDescription(t *testing.T) {
	test1Address := testutils.TestAddress("test1")
	test2Address := testutils.TestAddress("test2")

	t.Run("non-existing valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateDescription(validValidatorInfo(t).Address, "new description")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		// Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateDescription(validValidatorInfo(t).Address, "new description")
		})
	})

	t.Run("empty description", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		emptyDescription := ""
		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidDescription.Error(), func() {
			UpdateDescription(validValidatorInfo(t).Address, emptyDescription)
		})
	})

	t.Run("too long description", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidDescription.Error(), func() {
			UpdateDescription(validValidatorInfo(t).Address, strings.Repeat("a", DescriptionMaxLength+1))
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		newDescription := "new description"
		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateDescription(validValidatorInfo(t).Address, newDescription)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validValidatorInfo(t).Address)

			uassert.Equal(t, newDescription, valoper.Description)
		})
	})
}

func TestValopers_UpdateKeepRunning(t *testing.T) {
	test1Address := testutils.TestAddress("test1")
	test2Address := testutils.TestAddress("test2")

	t.Run("non-existing valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateKeepRunning(validValidatorInfo(t).Address, false)
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		// Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateKeepRunning(validValidatorInfo(t).Address, false)
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minFee}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validValidatorInfo(t).Moniker, validValidatorInfo(t).Description, validValidatorInfo(t).Address, validValidatorInfo(t).PubKey)
		})

		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateKeepRunning(validValidatorInfo(t).Address, false)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validValidatorInfo(t).Address)

			uassert.Equal(t, false, valoper.KeepRunning)
		})
	})
}
