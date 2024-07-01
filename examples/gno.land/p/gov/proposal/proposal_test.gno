package proposal

import (
	"errors"
	"testing"

	"std"
)

func TestExecutor(t *testing.T) {
	t.Parallel()

	verifyProposalFailed := func(e Executor) {
		if !e.IsDone() {
			t.Fatal("expected proposal to be done")
		}

		if e.IsSuccessful() {
			t.Fatal("expected proposal to fail")
		}

		if e.GetStatus() != Failed {
			t.Fatal("expected status to be Failed")
		}
	}

	verifyProposalSucceeded := func(e Executor) {
		if !e.IsDone() {
			t.Fatal("expected proposal to be done")
		}

		if !e.IsSuccessful() {
			t.Fatal("expected proposal to be successful")
		}

		if e.GetStatus() != Succeeded {
			t.Fatal("expected status to be Succeeded")
		}
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

		if e.GetStatus() != NotExecuted {
			t.Fatal("expected status to be NotExecuted")
		}

		// Execute as not the /r/gov/dao caller
		testing.PanicsWithError(t, errNotGovDAO.Error(), func() {
			e.Execute()
		})

		if called {
			t.Fatal("unexpected proposal execution")
		}
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

		if e.GetStatus() != NotExecuted {
			t.Fatal("expected status to be NotExecuted")
		}

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		testing.NotPanics(t, func() {
			testing.NoError(t, e.Execute())
		})

		if !called {
			t.Fatal("expected proposal to execute")
		}

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

		if e.GetStatus() != NotExecuted {
			t.Fatal("expected status to be NotExecuted")
		}

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		testing.NotPanics(t, func() {
			err := e.Execute()

			testing.ErrorIs(t, err, expectedErr)
		})

		if !called {
			t.Fatal("expected proposal to execute")
		}

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

		if e.GetStatus() != NotExecuted {
			t.Fatal("expected status to be NotExecuted")
		}

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		testing.NotPanics(t, func() {
			testing.NoError(t, e.Execute())
		})

		if !called {
			t.Fatal("expected proposal to execute")
		}

		// Make sure the execution params are correct
		verifyProposalSucceeded(e)

		// Attempt to execute the proposal again
		testing.NotPanics(t, func() {
			err := e.Execute()

			testing.ErrorIs(t, err, ErrAlreadyDone)
		})
	})
}
