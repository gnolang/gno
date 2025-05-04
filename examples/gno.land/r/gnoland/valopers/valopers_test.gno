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
	t.Helper()

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
	test1 := testutils.TestAddress("test1")
	testing.SetRealm(std.NewUserRealm(test1))

	t.Run("already a valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		v := Valoper{
			Moniker:     info.Moniker,
			Description: info.Description,
			Address:     info.Address,
			PubKey:      info.PubKey,
			KeepRunning: true,
		}

		// Add the valoper
		valopers.Set(v.Address.String(), v)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		uassert.AbortsWithMessage(t, ErrValoperExists.Error(), func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})
	})

	t.Run("no coins deposited", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send no coins
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", 0)})

		uassert.AbortsWithMessage(t, ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom), func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})
	})

	t.Run("insufficient coins amount deposited", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send invalid coins
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", minFee.Amount-1)})

		uassert.AbortsWithMessage(t, ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom), func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})
	})

	t.Run("coin amount deposited is not ugnot", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send invalid coins
		testing.SetOriginSend(std.Coins{std.NewCoin("gnogno", minFee.Amount)})

		uassert.AbortsWithMessage(t, "incompatible coin denominations: gnogno, ugnot", func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})
	})

	t.Run("successful registration", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		uassert.NotAborts(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, func() {
			valoper := GetByAddr(info.Address)

			uassert.Equal(t, info.Moniker, valoper.Moniker)
			uassert.Equal(t, info.Description, valoper.Description)
			uassert.Equal(t, info.Address, valoper.Address)
			uassert.Equal(t, info.PubKey, valoper.PubKey)
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

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		testing.SetRealm(std.NewUserRealm(test1Address))

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		testing.SetRealm(std.NewUserRealm(info.Address))

		// try to add member without being authorized
		uassert.AbortsWithMessage(t, authorizable.ErrNotSuperuser.Error(), func() {
			cross(AddToAuthList)(info.Address, test2Address)
		})
	})

	t.Run("unauthorized member deletes member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		testing.SetRealm(std.NewUserRealm(test1Address))

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, func() {
			// XXX this panics.
			cross(AddToAuthList)(info.Address, test2Address)
		})

		testing.SetRealm(std.NewUserRealm(info.Address))

		// try to add member without being authorized
		uassert.AbortsWithMessage(t, authorizable.ErrNotSuperuser.Error(), func() {
			cross(DeleteFromAuthList)(info.Address, test2Address)
		})
	})

	t.Run("authorized member adds member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		testing.SetRealm(std.NewUserRealm(test1Address))

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, func() {
			cross(AddToAuthList)(info.Address, test2Address)
		})

		testing.SetRealm(std.NewUserRealm(test2Address))

		newMoniker := "new moniker"
		// Update the valoper
		uassert.NotPanics(t, func() {
			cross(UpdateMoniker)(info.Address, newMoniker)
		})

		uassert.NotPanics(t, func() {
			valoper := GetByAddr(info.Address)
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

		info := validValidatorInfo(t)

		// Update the valoper
		uassert.AbortsWithMessage(t, ErrValoperMissing.Error(), func() {
			cross(UpdateMoniker)(info.Address, "new moniker")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		// Change the origin caller
		testing.SetOriginCaller(test2Address)

		// Update the valoper
		uassert.AbortsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			cross(UpdateMoniker)(info.Address, "new moniker")
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

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		for _, invalidMoniker := range invalidMonikers {
			// Update the valoper
			uassert.AbortsWithMessage(t, ErrInvalidMoniker.Error(), func() {
				cross(UpdateMoniker)(info.Address, invalidMoniker)
			})
		}
	})

	t.Run("too long moniker", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		// Update the valoper
		uassert.AbortsWithMessage(t, ErrInvalidMoniker.Error(), func() {
			cross(UpdateMoniker)(info.Address, strings.Repeat("a", MonikerMaxLength+1))
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		newMoniker := "new moniker"
		// Update the valoper
		uassert.NotPanics(t, func() {
			cross(UpdateMoniker)(info.Address, newMoniker)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(info.Address)

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
		uassert.AbortsWithMessage(t, ErrValoperMissing.Error(), func() {
			cross(UpdateDescription)(validValidatorInfo(t).Address, "new description")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		info := validValidatorInfo(t)

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		// Change the origin caller
		testing.SetOriginCaller(test2Address)

		// Update the valoper
		uassert.AbortsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			cross(UpdateDescription)(info.Address, "new description")
		})
	})

	t.Run("empty description", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		emptyDescription := ""
		// Update the valoper
		uassert.AbortsWithMessage(t, ErrInvalidDescription.Error(), func() {
			cross(UpdateDescription)(info.Address, emptyDescription)
		})
	})

	t.Run("too long description", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		// Update the valoper
		uassert.AbortsWithMessage(t, ErrInvalidDescription.Error(), func() {
			cross(UpdateDescription)(info.Address, strings.Repeat("a", DescriptionMaxLength+1))
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		newDescription := "new description"
		// Update the valoper
		uassert.NotPanics(t, func() {
			cross(UpdateDescription)(info.Address, newDescription)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(info.Address)

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
		uassert.AbortsWithMessage(t, ErrValoperMissing.Error(), func() {
			cross(UpdateKeepRunning)(validValidatorInfo(t).Address, false)
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		// Change the origin caller
		testing.SetOriginCaller(test2Address)

		// Update the valoper
		uassert.AbortsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			cross(UpdateKeepRunning)(info.Address, false)
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(std.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			cross(Register)(info.Moniker, info.Description, info.Address, info.PubKey)
		})

		// Update the valoper
		uassert.NotPanics(t, func() {
			cross(UpdateKeepRunning)(info.Address, false)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(info.Address)

			uassert.Equal(t, false, valoper.KeepRunning)
		})
	})
}
