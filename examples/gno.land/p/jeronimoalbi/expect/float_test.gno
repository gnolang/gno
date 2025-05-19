package expect

import "testing"

func TestFloatChecker(t *testing.T) {
	t.Run("to equal", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().ToEqual(1.2)
	})

	t.Run("not to equal", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().Not().ToEqual(3.4)
	})

	t.Run("to be greater than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().ToBeGreaterThan(1)
	})

	t.Run("not to be greater than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().Not().ToBeGreaterThan(1.3)
	})

	t.Run("to be greater or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().ToBeGreaterOrEqualThan(1.2)
		Value(t, 1.2).AsFloat().ToBeGreaterOrEqualThan(1.1)
	})

	t.Run("not to be greater or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().Not().ToBeGreaterOrEqualThan(1.3)
	})

	t.Run("to be lower than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().ToBeLowerThan(1.3)
	})

	t.Run("not to be lower than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().Not().ToBeLowerThan(1)
	})

	t.Run("to be lower or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().ToBeLowerOrEqualThan(1.2)
		Value(t, 1.2).AsFloat().ToBeLowerOrEqualThan(1.3)
	})

	t.Run("not to be lower or equal than", func(t *testing.T) {
		t.Parallel()
		Value(t, 1.2).AsFloat().Not().ToBeLowerOrEqualThan(1.1)
	})
}
