package valopers

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestValopers_Register(t *testing.T) {
	t.Parallel()

	t.Run("already a valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		v := Valoper{
			Address: testutils.TestAddress("valoper"),
		}

		// Add the valoper
		valopers.Set(v.Address.String(), v)

		uassert.PanicsWithMessage(t, errValoperExists, func() {
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
			PubKey:  "pub key",
		}

		uassert.NotPanics(t, func() {
			Register(v)
		})

		uassert.NotPanics(t, func() {
			valoper := Fetch(v.Address)

			uassert.Equal(t, v.Address, valoper.Address)
			uassert.Equal(t, v.Name, valoper.Name)
			uassert.Equal(t, v.PubKey, valoper.PubKey)
		})
	})
}

func TestValopers_Unregister(t *testing.T) {
	t.Parallel()

	t.Run("non-existing valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		// Remove the valoper
		uassert.PanicsWithMessage(t, errValoperMissing, func() {
			Unregister(testutils.TestAddress("valoper"))
		})
	})

	t.Run("successful removal", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		v := Valoper{
			Address: testutils.TestAddress("valoper"),
			Name:    "new valoper",
			PubKey:  "pub key",
		}

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(v)
		})

		// Remove the valoper
		uassert.NotPanics(t, func() {
			Unregister(v.Address)
		})

		// Make sure the valoper is removed
		uassert.PanicsWithMessage(t, errValoperMissing, func() {
			_ = Fetch(v.Address)
		})
	})
}

func TestValopers_Update(t *testing.T) {
	t.Parallel()

	t.Run("non-existing valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		v := Valoper{}

		// Update the valoper
		uassert.PanicsWithMessage(t, errValoperMissing, func() {
			Update(v.Address, v)
		})
	})

	t.Run("overwrite valoper", func(t *testing.T) {
		t.Parallel()

		// Clear the set for the test
		valopers = avl.NewTree()

		one := Valoper{
			Address: testutils.TestAddress("valoper 1"),
		}

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(one)
		})

		initialAddress := testutils.TestAddress("valoper 2")
		two := Valoper{
			Address: initialAddress,
		}

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(two)
		})

		// Update the valoper address
		// so it overlaps
		two = Valoper{
			Address: one.Address,
		}

		// Update the valoper
		uassert.PanicsWithMessage(t, errInvalidAddressUpdate, func() {
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
				PubKey:  "pub key",
			}
		)

		// Add the valoper
		uassert.NotPanics(t, func() {
			Register(v)
		})

		// Update the valoper name
		v.Name = "new name"

		// Update the valoper
		uassert.NotPanics(t, func() {
			Update(v.Address, v)
		})

		// Make sure the valoper is updated
		uassert.NotPanics(t, func() {
			valoper := Fetch(v.Address)

			uassert.Equal(t, v.Name, valoper.Name)
		})
	})
}
