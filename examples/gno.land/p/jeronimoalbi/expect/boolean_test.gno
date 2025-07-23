package expect_test

import (
	"testing"

	"gno.land/p/jeronimoalbi/expect"
)

func TestBooleanChecker(t *testing.T) {
	t.Run("to be truthy", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewBooleanChecker(ctx, true).ToBeTruthy()
	})

	t.Run("not to be truthy", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewBooleanChecker(ctx, false).Not().ToBeTruthy()
	})

	t.Run("to be falsy", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewBooleanChecker(ctx, false).ToBeFalsy()
	})

	t.Run("not to be falsy", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewBooleanChecker(ctx, true).Not().ToBeFalsy()
	})
}
