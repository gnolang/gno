package expect_test

import (
	"testing"

	"gno.land/p/jeronimoalbi/expect"
)

func TestStringChecker(t *testing.T) {
	t.Run("to equal", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewStringChecker(ctx, "foo").ToEqual("foo")
	})

	t.Run("not to equal", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewStringChecker(ctx, "foo").Not().ToEqual("bar")
	})

	t.Run("to be empty", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewStringChecker(ctx, "").ToBeEmpty()
	})

	t.Run("not to be empty", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewStringChecker(ctx, "foo").Not().ToBeEmpty()
	})

	t.Run("same length", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewStringChecker(ctx, "foo").ToHaveLength(3)
	})

	t.Run("different length", func(t *testing.T) {
		t.Parallel()

		ctx := expect.NewContext(t)
		expect.NewStringChecker(ctx, "foo").Not().ToHaveLength(1)
	})
}
