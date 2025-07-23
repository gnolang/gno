package expect_test

import (
	"testing"

	"gno.land/p/jeronimoalbi/expect"
)

func TestIntChecker(t *testing.T) {
	t.Run("to equal", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 1).ToEqual(1)
	})

	t.Run("not to equal", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 1).Not().ToEqual(2)
	})

	t.Run("to be greater than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 2).ToBeGreaterThan(1)
	})

	t.Run("not to be greater than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 1).Not().ToBeGreaterThan(2)
	})

	t.Run("to be greater or equal than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 2).ToBeGreaterOrEqualThan(2)
		expect.NewIntChecker(ctx, 2).ToBeGreaterOrEqualThan(1)
	})

	t.Run("not to be greater or equal than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 1).Not().ToBeGreaterOrEqualThan(2)
	})

	t.Run("to be lower than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 1).ToBeLowerThan(2)
	})

	t.Run("not to be lower than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 1).Not().ToBeLowerThan(1)
	})

	t.Run("to be lower or equal than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 1).ToBeLowerOrEqualThan(1)
		expect.NewIntChecker(ctx, 1).ToBeLowerOrEqualThan(2)
	})

	t.Run("not to be lower or equal than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewIntChecker(ctx, 2).Not().ToBeLowerOrEqualThan(1)
	})
}
