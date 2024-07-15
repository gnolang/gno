package executor

import (
	"errors"
	"std"
	"testing"

	"github.com/gnolang/gno/examples/gno.land/r/sys/vars"
	"gno.land/p/demo/context"
	"gno.land/p/demo/uassert"
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
		e := NewCallbackExecutor(cb)

		// Execute as not the /r/gov/dao caller
		uassert.PanicsWithMessage(t, errNotGovDAO.Error(), func() {
			_ = e.Execute()
		})

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
		e := NewCallbackExecutor(cb)

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(vars.GetStringValue(vars.GovdaoRealm))
		std.TestSetRealm(r)

		uassert.NotPanics(t, func() {
			err := e.Execute()

			uassert.NoError(t, err)
		})

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
		e := NewCallbackExecutor(cb)

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(vars.GetStringValue(vars.GovdaoRealm))
		std.TestSetRealm(r)

		uassert.NotPanics(t, func() {
			err := e.Execute()

			uassert.ErrorIs(t, err, expectedErr)
		})

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
		e := NewCtxExecutor(cb)

		// Execute as not the /r/gov/dao caller
		uassert.PanicsWithMessage(t, errNotGovDAO.Error(), func() {
			_ = e.Execute()
		})

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
		e := NewCtxExecutor(cb)

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(vars.GetStringValue(vars.GovdaoRealm))
		std.TestSetRealm(r)

		uassert.NotPanics(t, func() {
			err := e.Execute()

			uassert.NoError(t, err)
		})

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
		e := NewCtxExecutor(cb)

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(vars.GetStringValue(vars.GovdaoRealm))
		std.TestSetRealm(r)

		uassert.NotPanics(t, func() {
			err := e.Execute()

			uassert.ErrorIs(t, err, expectedErr)
		})

		uassert.True(t, called, "expected proposal to execute")
	})
}
