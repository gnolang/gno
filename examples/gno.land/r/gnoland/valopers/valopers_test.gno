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

const (
	validMoniker = "test-1"
	validDescription   = "test-1's description"
	validAddress   = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	validPubKey   = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

var (
	invalidMonikers = []string{
		"",     // Empty
		"    ", // Whitespace
		"a",    // Too short
		"a very long moniker that is longer than 32 characters", // Too long
		"!@#$%^&*()+{}|:<>?/.,;'",                               // Invalid characters
		" space in front",
		"space in back ",
	}

	test1Address = testutils.TestAddress("test1")
	test2Address = testutils.TestAddress("test2")

	validListeningAddresses = []string{
		"tcp://0.0.0.0:26656",
		"udp://192.168.1.1:8080",
		"http://example.com:80",
		"https://secure.com:443",
		"ws://localhost:8080",
		"without-protocol.com:443",
    "devx-sen-1.test5.gnodevx.network:26656",
	}

	invalidListeningAddresses = []string{
		"bad-format",
		"ftp://127.0.0.1", // Invalid (missing port)
		"https://:443",    // Invalid (missing host)
		":443",            // Invalid (missing host)
	}

	validP2PAddresses = []string{ufmt.Sprintf("%s@%s", validAddress, validListeningAddresses[0]), ufmt.Sprintf("%s@%s", validAddress, validListeningAddresses[1])}
)

func TestValopers_Register(t *testing.T) {
	t.Run("already a valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		v := Valoper{
			Moniker:      validMoniker,
			Description:  validDescription,
			Address:      validAddress,
			PubKey:       validPubKey,
			P2PAddresses: validP2PAddresses,
			KeepRunning:       true,
		}

		// Add the valoper
		valopers.Set(v.Address.String(), v)

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		uassert.PanicsWithMessage(t, ErrValoperExists.Error(), func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})
	})

	t.Run("no coins deposited", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send no coins
		std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", 0)}, nil)

		uassert.PanicsWithMessage(t, ufmt.Sprintf("payment must not be less than %d", minFee), func() {
      Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})
	})

	t.Run("insufficient coins amount deposited", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send invalid coins
		std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee-1)}, nil)

		uassert.PanicsWithMessage(t, ufmt.Sprintf("payment must not be less than %d", minFee), func() {
      Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})
	})

        t.Run("coin amount deposited is not ugnot", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send invalid coins
                std.TestSetOriginSend(std.Coins{std.NewCoin("gnogno", minFee)}, nil)

		uassert.PanicsWithMessage(t, "incompatible coin denominations: gnogno, ugnot", func() {
      Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
                })
        })

	t.Run("successful registration", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validAddress)

			uassert.Equal(t, validMoniker, valoper.Moniker)
			uassert.Equal(t, validDescription, valoper.Description)
			uassert.Equal(t, validAddress, valoper.Address)
			uassert.Equal(t, validPubKey, valoper.PubKey)
			for i, p2pAddress := range validP2PAddresses {
				uassert.Equal(t, p2pAddress, valoper.P2PAddresses[i])
			}
			uassert.Equal(t, true, valoper.KeepRunning)
		})
	})
}

func TestValopers_UpdateAuthMembers(t *testing.T) {
	t.Run("unauthorized member adds member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

    std.TestSetRealm(std.NewUserRealm(test1Address))
    std.TestSetOriginCaller(test1Address) // TODO(bug, issue #2371): should not be needed

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

    std.TestSetRealm(std.NewUserRealm(validAddress))
    std.TestSetOriginCaller(validAddress) // TODO(bug, issue #2371): should not be needed

    // try to add member without being authorized
		uassert.PanicsWithMessage(t, authorizable.ErrNotSuperuser.Error(), func() {
      AddToAuthList(validAddress, test2Address)
		})
	})

	t.Run("unauthorized member deletes member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

    std.TestSetRealm(std.NewUserRealm(test1Address))
    std.TestSetOriginCaller(test1Address) // TODO(bug, issue #2371): should not be needed

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		uassert.NotPanics(t, func() {
      AddToAuthList(validAddress, test2Address)
		})

    std.TestSetRealm(std.NewUserRealm(validAddress))
    std.TestSetOriginCaller(validAddress) // TODO(bug, issue #2371): should not be needed

    // try to add member without being authorized
		uassert.PanicsWithMessage(t, authorizable.ErrNotSuperuser.Error(), func() {
      DeleteFromAuthList(validAddress, test2Address)
		})
	})

	t.Run("authorized member adds member", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

    std.TestSetRealm(std.NewUserRealm(test1Address))
    std.TestSetOriginCaller(test1Address) // TODO(bug, issue #2371): should not be needed

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		uassert.NotPanics(t, func() {
      AddToAuthList(validAddress, test2Address)
		})

    std.TestSetRealm(std.NewUserRealm(test2Address))
    std.TestSetOriginCaller(test2Address) // TODO(bug, issue #2371): should not be needed

		newMoniker := "new moniker"
		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateMoniker(validAddress, newMoniker)
		})

		uassert.NotPanics(t, func() {
      valoper := GetByAddr(validAddress)
      uassert.Equal(t, newMoniker, valoper.Moniker)
		})
	})
}

func TestValopers_UpdateMoniker(t *testing.T) {
	t.Run("non-existing valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateMoniker(validAddress, "new moniker")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateMoniker(validAddress, "new moniker")
		})
	})

	t.Run("invalid moniker", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		for _, invalidMoniker := range invalidMonikers {
			// Update the valoper
			uassert.PanicsWithMessage(t, ErrInvalidMoniker.Error(), func() {
				UpdateMoniker(validAddress, invalidMoniker)
			})
		}
	})

	t.Run("too long moniker", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidMoniker.Error(), func() {
			UpdateMoniker(validAddress, strings.Repeat("a", MonikerMaxLength+1))
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		newMoniker := "new moniker"
		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateMoniker(validAddress, newMoniker)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validAddress)

			uassert.Equal(t, newMoniker, valoper.Moniker)
		})
	})
}

func TestValopers_UpdateDescription(t *testing.T) {
	t.Run("non-existing valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateDescription(validAddress, "new description")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateDescription(validAddress, "new description")
		})
	})

	t.Run("empty description", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

    emptyDescription := ""
		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidDescription.Error(), func() {
			UpdateDescription(validAddress, emptyDescription)
		})
	})

	t.Run("too long description", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidDescription.Error(), func() {
			UpdateDescription(validAddress, strings.Repeat("a", DescriptionMaxLength+1))
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		newDescription := "new description"
		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateDescription(validAddress, newDescription)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validAddress)

			uassert.Equal(t, newDescription, valoper.Description)
		})
	})
}

func TestValopers_UpdateP2PAddresses(t *testing.T) {
	t.Run("non-existing valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateP2PAddresses(validAddress, strings.Join(validP2PAddresses, ","))
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateP2PAddresses(validAddress, strings.Join(validP2PAddresses, ","))
		})
	})

	t.Run("empty P2PAddresses", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper with empty address
		uassert.PanicsWithMessage(t, ErrP2PAddressesMissing.Error(), func() {
			UpdateP2PAddresses(validAddress, "")
		})

    // set the valoper's KeepRunning off
		uassert.NotPanics(t, func() {
			UpdateKeepRunning(validAddress, false)
		})

		// Update the valoper with empty address
		uassert.NotPanics(t, func() {
			UpdateP2PAddresses(validAddress, "")
		})
	})

	t.Run("invalid P2PAddresses", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

    invalidAddress := std.Address("1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // missing prefix
		// Update the valoper with invalid address
		invalidP2PAddress := ufmt.Sprintf("%s@%s", invalidAddress, validListeningAddresses[0])
		uassert.PanicsWithMessage(t, ErrInvalidP2PAddress.Error(), func() {
			UpdateP2PAddresses(validAddress, invalidP2PAddress)
		})

		// Update the valoper with invalid p2p address
		for _, invalidListeningAddress := range invalidListeningAddresses {
			fullAddress := ufmt.Sprintf("%s@%s", validAddress, invalidListeningAddress)
			uassert.PanicsWithMessage(t, ErrInvalidP2PAddress.Error(), func() {
				UpdateP2PAddresses(validAddress, fullAddress)
			})
		}
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		for _, validListeningAddress := range validListeningAddresses {
			fullAddress := ufmt.Sprintf("%s@%s", validAddress, validListeningAddress)

			uassert.NotPanics(t, func() {
				UpdateP2PAddresses(validAddress, fullAddress)
			})

			// Make sure the valoper is updated
			uassert.NotPanics(t, func() {
				valoper := GetByAddr(validAddress)
				uassert.Equal(t, fullAddress, valoper.P2PAddresses[0])
			})
		}
	})
}

func TestValopers_UpdateKeepRunning(t *testing.T) {
	t.Run("non-existing valoper", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateKeepRunning(validAddress, false)
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		// Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateKeepRunning(validAddress, false)
		})
	})

	t.Run("successful update", func(t *testing.T) {
		// Clear the set for the test
		valopers = avl.NewTree()

		// Send coins
		std.TestSetOriginSend(std.Coins{minCoin}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateKeepRunning(validAddress, false)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validAddress)

			uassert.Equal(t, false, valoper.KeepRunning)
		})
	})
}
