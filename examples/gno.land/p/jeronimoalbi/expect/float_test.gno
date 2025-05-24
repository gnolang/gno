package expect_test

import (
	"testing"

	"gno.land/p/jeronimoalbi/expect"
)

func TestFloatChecker(t *testing.T) {
	t.Run("to equal", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).ToEqual(1.2)
	})

	t.Run("not to equal", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).Not().ToEqual(3.4)
	})

	t.Run("to be greater than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).ToBeGreaterThan(1)
	})

	t.Run("not to be greater than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).Not().ToBeGreaterThan(1.3)
	})

	t.Run("to be greater or equal than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).ToBeGreaterOrEqualThan(1.2)
		expect.NewFloatChecker(ctx, 1.2).ToBeGreaterOrEqualThan(1.1)
	})

	t.Run("not to be greater or equal than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).Not().ToBeGreaterOrEqualThan(1.3)
	})

	t.Run("to be lower than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).ToBeLowerThan(1.3)
	})

	t.Run("not to be lower than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).Not().ToBeLowerThan(1)
	})

	t.Run("to be lower or equal than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).ToBeLowerOrEqualThan(1.2)
		expect.NewFloatChecker(ctx, 1.2).ToBeLowerOrEqualThan(1.3)
	})

	t.Run("not to be lower or equal than", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewFloatChecker(ctx, 1.2).Not().ToBeLowerOrEqualThan(1.1)
	})
}
