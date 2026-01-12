package valopers

import (
	"chain"
	"strings"
	"testing"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable/exts/authorizable"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
)

func validValidatorInfo(t *testing.T) struct {
	Moniker     string
	Description string
	ServerType  string
	Address     address
	PubKey      string
} {
	t.Helper()

	return struct {
		Moniker     string
		Description string
		ServerType  string
		Address     address
		PubKey      string
	}{
		Moniker:     "test-1",
		Description: "test-1's description",
		ServerType:  ServerTypeOnPrem,
		Address:     address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h"),
		PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p",
	}
}

func TestValopers_Register(t *testing.T) {
	test1 := testutils.TestAddress("test1")
	testing.SetRealm(testing.NewUserRealm(test1))

	t.Run("already a valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		v := Valoper{
			Moniker:     info.Moniker,
			Description: info.Description,
			ServerType:  info.ServerType,
			Address:     info.Address,
			PubKey:      info.PubKey,
			KeepRunning: true,
		}

		// Add the valoper
		valopers.Set(v.Address.String(), v)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.AbortsWithMessage(t, ErrValoperExists.Error(), func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})
	})

	t.Run("no coins deposited", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send no coins
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", 0)})

		uassert.AbortsWithMessage(t, ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom), func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})
	})

	t.Run("insufficient coins amount deposited", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send invalid coins
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", minFee.Amount-1)})

		uassert.AbortsWithMessage(t, ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom), func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})
	})

	t.Run("coin amount deposited is not ugnot", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send invalid coins
		testing.SetOriginSend(chain.Coins{chain.NewCoin("gnogno", minFee.Amount)})

		uassert.AbortsWithMessage(t, "incompatible coin denominations: gnogno, ugnot", func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})
	})

	t.Run("successful registration", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotAborts(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, func() {
			valoper := GetByAddr(info.Address)

			uassert.Equal(t, info.Moniker, valoper.Moniker)
			uassert.Equal(t, info.Description, valoper.Description)
			uassert.Equal(t, info.ServerType, valoper.ServerType)
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
		testing.SetOriginSend(chain.Coins{minFee})

		testing.SetRealm(testing.NewUserRealm(test1Address))

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		testing.SetRealm(testing.NewUserRealm(info.Address))

		// try to add member without being authorized
		uassert.AbortsWithMessage(t, authorizable.ErrNotSuperuser.Error(), func() {
			AddToAuthList(cross, info.Address, test2Address)
		})
	})

	t.Run("unauthorized member deletes member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		testing.SetRealm(testing.NewUserRealm(test1Address))

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, func() {
			// XXX this panics.
			AddToAuthList(cross, info.Address, test2Address)
		})

		testing.SetRealm(testing.NewUserRealm(info.Address))

		// try to add member without being authorized
		uassert.AbortsWithMessage(t, authorizable.ErrNotSuperuser.Error(), func() {
			DeleteFromAuthList(cross, info.Address, test2Address)
		})
	})

	t.Run("authorized member adds member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		testing.SetRealm(testing.NewUserRealm(test1Address))

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, func() {
			AddToAuthList(cross, info.Address, test2Address)
		})

		testing.SetRealm(testing.NewUserRealm(test2Address))

		newMoniker := "new moniker"
		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateMoniker(cross, info.Address, newMoniker)
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
			UpdateMoniker(cross, info.Address, "new moniker")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Change the origin caller
		testing.SetOriginCaller(test2Address)

		// Update the valoper
		uassert.AbortsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateMoniker(cross, info.Address, "new moniker")
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
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		for _, invalidMoniker := range invalidMonikers {
			// Update the valoper
			uassert.AbortsWithMessage(t, ErrInvalidMoniker.Error(), func() {
				UpdateMoniker(cross, info.Address, invalidMoniker)
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
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Update the valoper
		uassert.AbortsWithMessage(t, ErrInvalidMoniker.Error(), func() {
			UpdateMoniker(cross, info.Address, strings.Repeat("a", MonikerMaxLength+1))
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		newMoniker := "new moniker"
		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateMoniker(cross, info.Address, newMoniker)
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
			UpdateDescription(cross, validValidatorInfo(t).Address, "new description")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		info := validValidatorInfo(t)

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Change the origin caller
		testing.SetOriginCaller(test2Address)

		// Update the valoper
		uassert.AbortsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateDescription(cross, info.Address, "new description")
		})
	})

	t.Run("empty description", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		emptyDescription := ""
		// Update the valoper
		uassert.AbortsWithMessage(t, ErrInvalidDescription.Error(), func() {
			UpdateDescription(cross, info.Address, emptyDescription)
		})
	})

	t.Run("too long description", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Update the valoper
		uassert.AbortsWithMessage(t, ErrInvalidDescription.Error(), func() {
			UpdateDescription(cross, info.Address, strings.Repeat("a", DescriptionMaxLength+1))
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		newDescription := "new description"
		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateDescription(cross, info.Address, newDescription)
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
			UpdateKeepRunning(cross, validValidatorInfo(t).Address, false)
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Change the origin caller
		testing.SetOriginCaller(test2Address)

		// Update the valoper
		uassert.AbortsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateKeepRunning(cross, info.Address, false)
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateKeepRunning(cross, info.Address, false)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(info.Address)

			uassert.Equal(t, false, valoper.KeepRunning)
		})
	})
}

func TestValopers_UpdateServerType(t *testing.T) {
	test1Address := testutils.TestAddress("test1")
	test2Address := testutils.TestAddress("test2")

	t.Run("non-existing valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.AbortsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateServerType(cross, validValidatorInfo(t).Address, ServerTypeCloud)
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Change the origin caller
		testing.SetOriginCaller(test2Address)

		// Update the valoper
		uassert.AbortsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateServerType(cross, info.Address, ServerTypeCloud)
		})
	})

	t.Run("invalid server type", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		invalidServerTypes := []string{
			"",
			"invalid",
			"Cloud",      // case sensitive
			"ON-PREM",    // case sensitive
			"datacenter", // wrong format
		}

		for _, invalidType := range invalidServerTypes {
			// Update the valoper with invalid server type
			uassert.AbortsWithMessage(t, ErrInvalidServerType.Error(), func() {
				UpdateServerType(cross, info.Address, invalidType)
			})
		}
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		info := validValidatorInfo(t)

		// Set the origin caller
		testing.SetOriginCaller(test1Address)

		// Send coins
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper with on-prem server type
		uassert.NotPanics(t, func() {
			Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Update the valoper to cloud
		uassert.NotPanics(t, func() {
			UpdateServerType(cross, info.Address, ServerTypeCloud)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(info.Address)
			uassert.Equal(t, ServerTypeCloud, valoper.ServerType)
		})

		// Update the valoper to data-center
		uassert.NotPanics(t, func() {
			UpdateServerType(cross, info.Address, ServerTypeDataCenter)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(info.Address)
			uassert.Equal(t, ServerTypeDataCenter, valoper.ServerType)
		})
	})
}
