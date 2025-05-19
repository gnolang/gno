package expect

import "testing"

func TestUintChecker(t *testing.T) {
	t.Run("to equal", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsUint().ToEqual(1)
	})

	t.Run("not to equal", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsUint().Not().ToEqual(2)
	})

	t.Run("to be greater than", func(t *testing.T) {
		t.Parallel()
		Value(t, 2).AsUint().ToBeGreaterThan(1)
	})

	t.Run("not to be greater than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsUint().Not().ToBeGreaterThan(2)
	})

	t.Run("to be greater or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 2).AsUint().ToBeGreaterOrEqualThan(2)
		Value(t, 2).AsUint().ToBeGreaterOrEqualThan(1)
	})

	t.Run("not to be greater or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsUint().Not().ToBeGreaterOrEqualThan(2)
	})

	t.Run("to be lower than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsUint().ToBeLowerThan(2)
	})

	t.Run("not to be lower than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsUint().Not().ToBeLowerThan(1)
	})

	t.Run("to be lower or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsUint().ToBeLowerOrEqualThan(1)
		Value(t, 1).AsUint().ToBeLowerOrEqualThan(2)
	})

	t.Run("not to be lower or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 2).AsUint().Not().ToBeLowerOrEqualThan(1)
	})
}
