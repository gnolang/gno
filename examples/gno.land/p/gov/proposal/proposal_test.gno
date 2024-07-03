package proposal

import (
	"errors"
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestExecutor(t *testing.T) {
	t.Parallel()

	verifyProposalFailed := func(e Executor) {
		uassert.True(t, e.IsDone(), "expected proposal to be done")
		uassert.False(t, e.IsSuccessful(), "expected proposal to fail")
	}

	verifyProposalSucceeded := func(e Executor) {
		uassert.True(t, e.IsDone(), "expected proposal to be done")
		uassert.True(t, e.IsSuccessful(), "expected proposal to be successful")
	}

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
		e := NewExecutor(cb)

		if e.IsDone() {
			t.Fatal("expected status to be NotExecuted")
		}

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
		e := NewExecutor(cb)

		if e.IsDone() {
			t.Fatal("expected status to be NotExecuted")
		}

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		uassert.NotPanics(t, func() {
			err := e.Execute()

			uassert.NoError(t, err)
		})

		uassert.True(t, called, "expected proposal to execute")

		// Make sure the execution params are correct
		verifyProposalSucceeded(e)
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
		e := NewExecutor(cb)

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		uassert.NotPanics(t, func() {
			err := e.Execute()

			uassert.ErrorIs(t, err, expectedErr)
		})

		uassert.True(t, called, "expected proposal to execute")

		// Make sure the execution params are correct
		verifyProposalFailed(e)
	})

	t.Run("proposal already executed", func(t *testing.T) {
		t.Parallel()

		var (
			called = false

			cb = func() error {
				called = true

				return nil
			}
		)

		// Create the executor
		e := NewExecutor(cb)

		if e.IsDone() {
			t.Fatal("expected status to be NotExecuted")
		}

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		uassert.NotPanics(t, func() {
			uassert.NoError(t, e.Execute())
		})

		uassert.True(t, called, "expected proposal to execute")

		// Make sure the execution params are correct
		verifyProposalSucceeded(e)

		// Attempt to execute the proposal again
		uassert.NotPanics(t, func() {
			err := e.Execute()

			uassert.ErrorIs(t, err, ErrAlreadyDone)
		})
	})
}
