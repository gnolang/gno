package expect

import "testing"

func TestIntChecker(t *testing.T) {
	t.Run("to equal", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsInt().ToEqual(1)
	})

	t.Run("not to equal", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsInt().Not().ToEqual(2)
	})

	t.Run("to be greater than", func(t *testing.T) {
		t.Parallel()
		Value(t, 2).AsInt().ToBeGreaterThan(1)
	})

	t.Run("not to be greater than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsInt().Not().ToBeGreaterThan(2)
	})

	t.Run("to be greater or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 2).AsInt().ToBeGreaterOrEqualThan(2)
		Value(t, 2).AsInt().ToBeGreaterOrEqualThan(1)
	})

	t.Run("not to be greater or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsInt().Not().ToBeGreaterOrEqualThan(2)
	})

	t.Run("to be lower than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsInt().ToBeLowerThan(2)
	})

	t.Run("not to be lower than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsInt().Not().ToBeLowerThan(1)
	})

	t.Run("to be lower or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1).AsInt().ToBeLowerOrEqualThan(1)
		Value(t, 1).AsInt().ToBeLowerOrEqualThan(2)
	})

	t.Run("not to be lower or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 2).AsInt().Not().ToBeLowerOrEqualThan(1)
	})
}
