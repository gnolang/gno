package simpledao

import "testing"

func TestSimpleDAO_GetMembStore(t *testing.T) {
	t.Parallel()

	t.Run("empty store", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("store with members", func(t *testing.T) {
		t.Parallel()

	})
}

func TestSimpleDAO_GetPropStore(t *testing.T) {
	t.Parallel()

	t.Run("empty store", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("store with proposals", func(t *testing.T) {
		t.Parallel()

	})
}

func TestSimpleDAO_Propose(t *testing.T) {
	t.Parallel()

	t.Run("invalid executor", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("caller cannot cover fee", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("proposal added", func(t *testing.T) {
		t.Parallel()

	})
}

func TestSimpleDAO_VoteOnProposal(t *testing.T) {
	t.Parallel()

	t.Run("not govdao member", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("missing proposal", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("proposal executed", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("proposal expired", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("double vote on proposal", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("majority accepted", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("majority rejected", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("majority undecided", func(t *testing.T) {
		t.Parallel()

	})
}

func TestSimpleDAO_ExecuteProposal(t *testing.T) {
	t.Parallel()

	t.Run("caller cannot cover fee", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("missing proposal", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("proposal not accepted", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("proposal executed", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("proposal expired", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("execution error", func(t *testing.T) {
		t.Parallel()

	})

	t.Run("successful execution", func(t *testing.T) {
		t.Parallel()

	})
}
