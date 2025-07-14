package bridge

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestCallbackRegistration(t *testing.T) {
	callbackTree = avl.NewTree()

	err := RegisterCallback(cross, adminAddr, "test", "add", func(a, b int) int {
		return a + b
	})
	uassert.NoError(t, err)

	cb, exists := LookupCallback("test", "add")
	uassert.True(t, exists)

	addFn, ok := cb.(func(int, int) int)
	uassert.True(t, ok)
	result := addFn(5, 3)
	uassert.Equal(t, result, 8)
}

func TestCallbackUpdate(t *testing.T) {
	callbackTree = avl.NewTree()

	err := RegisterCallback(cross, adminAddr, "test", "multiply", func(a, b int) int {
		return a * b
	})
	uassert.NoError(t, err)

	cb, exists := LookupCallback("test", "multiply")
	uassert.True(t, exists)

	multiplyFn, ok := cb.(func(int, int) int)
	uassert.True(t, ok)
	result := multiplyFn(3, 4)
	uassert.Equal(t, result, 12) // 3 * 4 = 12

	// update callback
	err = UpdateCallback(cross, adminAddr, "test", "multiply", func(a, b int) int {
		return a * b * 2
	})
	uassert.NoError(t, err)

	// get callback and check the result
	cb, exists = LookupCallback("test", "multiply")
	uassert.True(t, exists)

	multiplyFn, ok = cb.(func(int, int) int)
	uassert.True(t, ok)
	result = multiplyFn(3, 4)
	uassert.Equal(t, result, 24) // 3 * 4 * 2 = 24
}

func TestCallbackRegistrationErrors(t *testing.T) {
	callbackTree = avl.NewTree()
	userAddr := testutils.TestAddress("user")

	// duplicate registration test
	err := RegisterCallback(cross, adminAddr, "test", "duplicate", func() {})
	uassert.NoError(t, err)

	err = RegisterCallback(cross, adminAddr, "test", "duplicate", func() {})
	uassert.Error(t, err)

	// unauthorized user registration test
	err = RegisterCallback(cross, userAddr, "test", "unauthorized", func() {})
	uassert.Error(t, err)
}

func TestCallbackUpdateErrors(t *testing.T) {
	callbackTree = avl.NewTree()
	userAddr := testutils.TestAddress("user2")

	err := UpdateCallback(cross, adminAddr, "test", "nonexistent", func() {})
	uassert.Error(t, err)

	err = RegisterCallback(cross, adminAddr, "test", "update", func() {})
	uassert.NoError(t, err)

	err = UpdateCallback(cross, userAddr, "test", "update", func() {})
	uassert.Error(t, err)
}

func TestUpgradablePattern(t *testing.T) {
	t.Run("version Management", func(t *testing.T) {
		callbackTree = avl.NewTree()

		// v1
		err := RegisterCallback(cross, adminAddr, "emission", "MintAndDistributeGns", func() uint64 {
			return 100
		})
		uassert.NoError(t, err)

		// v2
		err = UpdateCallback(cross, adminAddr, "emission", "MintAndDistributeGns", func() uint64 {
			return 200
		})
		uassert.NoError(t, err)

		// check version
		cb, exists := LookupCallback("emission", "MintAndDistributeGns")
		uassert.True(t, exists)
		fn, ok := cb.(func() uint64)
		uassert.True(t, ok)
		result := fn()
		uassert.Equal(t, result, uint64(200))
	})

	t.Run("Multiple Contract Callbacks", func(t *testing.T) {
		callbackTree = avl.NewTree()

		// Emission contract callback
		err := RegisterCallback(cross, adminAddr, "emission", "MintAndDistributeGns", func() uint64 {
			return 100
		})
		uassert.NoError(t, err)

		// Position contract callback
		err = RegisterCallback(cross, adminAddr, "position", "GetPosition", func(id uint64) string {
			return "position"
		})
		uassert.NoError(t, err)

		// check each callback is independent
		cb1, exists := LookupCallback("emission", "MintAndDistributeGns")
		uassert.True(t, exists)
		cb2, exists := LookupCallback("position", "GetPosition")
		uassert.True(t, exists)

		fn1, ok := cb1.(func() uint64)
		uassert.True(t, ok)
		fn2, ok := cb2.(func(uint64) string)
		uassert.True(t, ok)

		uassert.Equal(t, fn1(), uint64(100))
		uassert.Equal(t, fn2(1), "position")
	})

	t.Run("Rollback Scenario", func(t *testing.T) {
		callbackTree = avl.NewTree()

		// v1
		originalCallback := func() uint64 {
			return 100
		}
		err := RegisterCallback(cross, adminAddr, "emission", "MintAndDistributeGns", originalCallback)
		uassert.NoError(t, err)

		// buggy version
		err = UpdateCallback(cross, adminAddr, "emission", "MintAndDistributeGns", func() uint64 {
			panic("buggy version")
		})
		uassert.NoError(t, err)

		// rollback
		err = UpdateCallback(cross, adminAddr, "emission", "MintAndDistributeGns", originalCallback)
		uassert.NoError(t, err)

		// check rollback
		cb, exists := LookupCallback("emission", "MintAndDistributeGns")
		uassert.True(t, exists)
		fn, ok := cb.(func() uint64)
		uassert.True(t, ok)
		result := fn()
		uassert.Equal(t, result, uint64(100))
	})

	t.Run("Namespace Collision", func(t *testing.T) {
		callbackTree = avl.NewTree()

		// same namespace, different function
		err := RegisterCallback(cross, adminAddr, "emission", "func1", func() int { return 1 })
		uassert.NoError(t, err)
		err = RegisterCallback(cross, adminAddr, "emission", "func2", func() int { return 2 })
		uassert.NoError(t, err)

		// check each function is independent
		cb1, exists := LookupCallback("emission", "func1")
		uassert.True(t, exists)
		cb2, exists := LookupCallback("emission", "func2")
		uassert.True(t, exists)

		fn1, ok := cb1.(func() int)
		uassert.True(t, ok)
		fn2, ok := cb2.(func() int)
		uassert.True(t, ok)

		uassert.Equal(t, fn1(), 1)
		uassert.Equal(t, fn2(), 2)
	})
}
