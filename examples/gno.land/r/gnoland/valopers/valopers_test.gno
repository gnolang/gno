package valopers

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

const testPubKey = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqg8dnu7u3qwe952azxe2sw0zvu45cclk523m9qfjq8zzuw620vjeyaqyv3a"

func TestValopers_Register(t *testing.T) {
	t.Parallel()

	t.Run("already a valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		v := Valoper{
			Address: testutils.TestAddress("valoper"),
			PubKey:  testPubKey,
		}

		// Add the valoper
		valopers.Set(v.Address.String(), v)

		uassert.PanicsWithMessage(t, ErrValoperExists, func() {
			Register(v)
		})
	})

	t.Run("successful registration", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		v := Valoper{
			Address: testutils.TestAddress("valoper"),
			Name:    "new valoper",
			PubKey:  testPubKey,
		}

		uassert.NotPanics(t, func() {
			Register(v)
		})

		uassert.NotPanics(t, func() {
			valoper := GetByAddr(v.Address)

			uassert.Equal(t, v.Address, valoper.Address)
			uassert.Equal(t, v.Name, valoper.Name)
			uassert.Equal(t, v.PubKey, valoper.PubKey)
		})
	})
}

func TestValopers_Update(t *testing.T) {
	t.Parallel()

	t.Run("non-existing valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		v := Valoper{
			Address: testutils.TestAddress("missing"),
			PubKey:  testPubKey,
		}

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrValoperMissing, func() {
			Update(v.Address, v)
		})
	})

	t.Run("overwrite valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		one := Valoper{
			Address: testutils.TestAddress("valoper 1"),
			PubKey:  testPubKey,
		}

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(one)
		})

		initialAddress := testutils.TestAddress("valoper 2")
		two := Valoper{
			Address: initialAddress,
			PubKey:  testPubKey,
		}

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(two)
		})

		// Update the valoper address
		// so it overlaps
		two = Valoper{
			Address: one.Address,
			PubKey:  testPubKey,
		}

		// Update the valoper
		uassert.PanicsWithMessage(t, ErrInvalidAddressUpdate, func() {
			Update(initialAddress, two)
		})
	})

	t.Run("successful update", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		var (
			name = "new valoper"
			v    = Valoper{
				Address: testutils.TestAddress("valoper"),
				Name:    name,
				PubKey:  testPubKey,
			}
		)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(v)
		})

		// Update the valoper name
		v.Name = "new name"
		v.Active = false

		// Update the valoper
		uassert.NotPanics(t, func() {
			Update(v.Address, v)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := GetByAddr(v.Address)

			uassert.Equal(t, v.Name, valoper.Name)
			uassert.Equal(t, v.Active, valoper.Active)
		})
	})
}
