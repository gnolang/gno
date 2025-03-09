package valopers

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
)

const (
	validMoniker   = "test-1"
	invalidMoniker = ""

	validDescription   = "test-1's description"
	invalidDescription = ""

	validAddress   = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	invalidAddress = std.Address("1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // missing prefix

	validPubKey   = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
  validPubKey2  = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpcyx3ne8m28jxyw6nc96szc2seq3w6zfdylawm5aa7mcy9q88cfn8jsm7p"
	invalidPubKey = "1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj" // missing prefix
)

var (
  test1Address = testutils.TestAddress("test1")
  test2Address = testutils.TestAddress("test2")

  validListeningAddresses = []string{
    "tcp://0.0.0.0:26656",
		"udp://192.168.1.1:8080",
		"http://example.com:80",
		"https://secure.com:443",
		"ws://localhost:8080",
    "without-protocol.com:443",
  }	

  invalidListeningAddresses = []string{
		"bad-format", 
		"ftp://127.0.0.1",         // Invalid (missing port)
		"https://:443",            // Invalid (missing host)
		":443",                    // Invalid (missing host)
	}

	validP2PAddresses   = []string{ufmt.Sprintf("%s@%s", validAddress, validListeningAddresses[0]), ufmt.Sprintf("%s@%s", validAddress, validListeningAddresses[1])}
)

func TestValopers_Register(t *testing.T) {
	t.Parallel()

	t.Run("already a valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		v := Valoper{
			Moniker:      validMoniker,
			Description:  validDescription,
			Address:      validAddress,
			PubKey:       validPubKey,
			P2PAddresses: validP2PAddresses,
			Active:       true,
		}

		// Add the valoper
		valopers.Set(v.Address.String(), v)

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		uassert.PanicsWithMessage(t, ErrValoperExists.Error(), func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})
	})

	t.Run("no coins deposit", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

         defer func() {
               if r := recover(); r == nil {
                       t.Errorf("Register without coins deposit should panic")
               }
       }()

    // Send no coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", 0)}, nil)

			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
	})

	t.Run("insufficient coins amount deposit", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

	   // Send invalid coins
	   std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee - 1)}, nil)

	        defer func() {
	              if r := recover(); r == nil {
	                      t.Errorf("Register with less than minimal coins deposit should panic")
	              }
	      }()

			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
	})

	t.Run("to much coins amount deposit", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

	   // Send invalid coins
	   std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", maxFeeMult * minFee + 1)}, nil)

	        defer func() {
	              if r := recover(); r == nil {
	                      t.Errorf("Register with more than maximal coins deposit should panic")
	              }
	      }()

			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
	})

	t.Run("successful registration", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

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
			uassert.Equal(t, true, valoper.Active)
		})
	})
}

func TestValopers_Update(t *testing.T) {
	t.Parallel()

	t.Run("non-existing valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			Update(validAddress, validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","), true)
		})
	})

	t.Run("overwrite valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		initialAddress := testutils.TestAddress("valoper 2")

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, initialAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrAddressExists.Error(), func() {
			Update(initialAddress, validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","), true)
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		t.Parallel()

    // Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

    // Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrOriginCaller.Error(), func() {
			Update(validAddress, validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","), true)
		})
	})

	t.Run("successful update same address", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		moniker := "new valoper"

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(moniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper's moniker
		uassert.NotPanics(t, func() {
			Update(validAddress, validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","), true)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validAddress)

			uassert.Equal(t, validMoniker, valoper.Moniker)
		})
	})

	t.Run("successful update different address", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		moniker := "new valoper"

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(moniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper's moniker and address
		uassert.NotPanics(t, func() {
			Update(validAddress, validMoniker, validDescription, test1Address, validPubKey, strings.Join(validP2PAddresses, ","), true)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(test1Address)

			uassert.Equal(t, validMoniker, valoper.Moniker)
		})
	})
}

func TestValopers_UpdateMoniker(t *testing.T) {
	t.Parallel()

	t.Run("non-existing valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateMoniker(validAddress, "new moniker")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		t.Parallel()

    // Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

    // Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrOriginCaller.Error(), func() {
			UpdateMoniker(validAddress, "new moniker")
		})
	})

	t.Run("invalid moniker", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidMoniker.Error(), func() {
			UpdateMoniker(validAddress, invalidMoniker)
		})
	})

	t.Run("successful update", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

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
	t.Parallel()

	t.Run("non-existing valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateDescription(validAddress, "new description")
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		t.Parallel()

    // Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

    // Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrOriginCaller.Error(), func() {
			UpdateDescription(validAddress, "new description")
		})
	})

	t.Run("invalid description", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidDescription.Error(), func() {
			UpdateDescription(validAddress, invalidDescription)
		})
	})

	t.Run("successful update", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

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

func TestValopers_UpdateAddress(t *testing.T) {
	t.Parallel()

	t.Run("non-existing valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateAddress(validAddress, test1Address)
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		t.Parallel()

    // Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

    newAddress := testutils.TestAddress("newAddress")

    // Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrOriginCaller.Error(), func() {
			UpdateAddress(validAddress, newAddress)
		})
	})

	t.Run("same address", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrSameAddress.Error(), func() {
			UpdateAddress(validAddress, validAddress)
		})
	})

	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidAddress.Error(), func() {
			UpdateAddress(validAddress, invalidAddress)
		})
	})

	t.Run("successful update", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdateAddress(validAddress, test1Address)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(test1Address)
		})
	})
}

func TestValopers_UpdatePubKey(t *testing.T) {
	t.Parallel()

	t.Run("non-existing valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdatePubKey(validAddress, validPubKey)
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		t.Parallel()

    // Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

    // Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrOriginCaller.Error(), func() {
			UpdatePubKey(validAddress, validPubKey2)
		})
	})

	t.Run("invalid PubKey", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidPubKey.Error(), func() {
			UpdatePubKey(validAddress, invalidPubKey)
		})
	})

	t.Run("successful update", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

		// Update the valoper
		uassert.NotPanics(t, func() {
			UpdatePubKey(validAddress, validPubKey2)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(validAddress)

			uassert.Equal(t, validPubKey2, valoper.PubKey)
		})
	})
}

func TestValopers_UpdateP2PAddresses(t *testing.T) {
	t.Parallel()

	t.Run("non-existing valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing.Error(), func() {
			UpdateP2PAddresses(validAddress, strings.Join(validP2PAddresses, ","))
		})
	})

	t.Run("invalid caller", func(t *testing.T) {
		t.Parallel()

    // Set the origin caller
		std.TestSetOriginCaller(test1Address)

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

    // Change the origin caller
		std.TestSetOriginCaller(test2Address)

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrOriginCaller.Error(), func() {
			UpdateP2PAddresses(validAddress, strings.Join(validP2PAddresses, ","))
		})
	})

	t.Run("invalid P2PAddresses", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
		})

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
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

    // Send coins
    std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

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
