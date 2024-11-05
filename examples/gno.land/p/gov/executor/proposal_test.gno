package executor

import (
	"errors"
	"std"
	"testing"

	"gno.land/p/demo/context"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestExecutor_Callback(t *testing.T) {
	t.Parallel()

	t.Run("govdao not caller", func(t *testing.T) {
		t.Parallel()

		var (
			called = false

			cb = func() error {
				called = true

				return nil
			}
		)

		// Create the executor
		e := NewCallbackExecutor(cb, "gno.land/r/gov/dao")

		// Execute as not the /r/gov/dao caller
		uassert.ErrorIs(t, e.Execute(), errInvalidCaller)
		uassert.False(t, called, "expected proposal to not execute")
	})

	t.Run("execution successful", func(t *testing.T) {
		t.Parallel()

		var (
			called = false

			cb = func() error {
				called = true

				return nil
			}
		)

		// Create the executor
		daoPkgPath := "gno.land/r/gov/dao"
		e := NewCallbackExecutor(cb, daoPkgPath)

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		uassert.NoError(t, e.Execute())
		uassert.True(t, called, "expected proposal to execute")
	})

	t.Run("execution unsuccessful", func(t *testing.T) {
		t.Parallel()

		var (
			called      = false
			expectedErr = errors.New("unexpected")

			cb = func() error {
				called = true

				return expectedErr
			}
		)

		// Create the executor
		daoPkgPath := "gno.land/r/gov/dao"
		e := NewCallbackExecutor(cb, daoPkgPath)

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		uassert.ErrorIs(t, e.Execute(), expectedErr)
		uassert.True(t, called, "expected proposal to execute")
	})
}

func TestExecutor_Context(t *testing.T) {
	t.Parallel()

	t.Run("govdao not caller", func(t *testing.T) {
		t.Parallel()

		var (
			called = false

			cb = func(ctx context.Context) error {
				if !IsApprovedByGovdaoContext(ctx) {
					t.Fatal("not govdao caller")
				}

				called = true

				return nil
			}
		)

		// Create the executor
		e := NewCtxExecutor(cb, "gno.land/r/gov/dao")

		// Execute as not the /r/gov/dao caller
		uassert.ErrorIs(t, e.Execute(), errInvalidCaller)
		uassert.False(t, called, "expected proposal to not execute")
	})

	t.Run("execution successful", func(t *testing.T) {
		t.Parallel()

		var (
			called = false

			cb = func(ctx context.Context) error {
				if !IsApprovedByGovdaoContext(ctx) {
					t.Fatal("not govdao caller")
				}

				called = true

				return nil
			}
		)

		// Create the executor
		daoPkgPath := "gno.land/r/gov/dao"
		e := NewCtxExecutor(cb, daoPkgPath)

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		urequire.NoError(t, e.Execute())
		uassert.True(t, called, "expected proposal to execute")
	})

	t.Run("execution unsuccessful", func(t *testing.T) {
		t.Parallel()

		var (
			called      = false
			expectedErr = errors.New("unexpected")

			cb = func(ctx context.Context) error {
				if !IsApprovedByGovdaoContext(ctx) {
					t.Fatal("not govdao caller")
				}

				called = true

				return expectedErr
			}
		)

		// Create the executor
		daoPkgPath := "gno.land/r/gov/dao"
		e := NewCtxExecutor(cb, daoPkgPath)

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		uassert.NotPanics(t, func() {
			err := e.Execute()

			uassert.ErrorIs(t, err, expectedErr)
		})

		uassert.True(t, called, "expected proposal to execute")
	})
}
