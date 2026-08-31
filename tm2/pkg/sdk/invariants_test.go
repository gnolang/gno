package sdk

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Guard's contract has three parts, and each is here because deleting it leaves
// every other test in the tree passing: the clean result, the finding cap that
// keeps a corrupt store from OOMing the check that exists to diagnose it, and the
// recover that turns a buggy check into a report instead of a node crash.
func TestGuard(t *testing.T) {
	t.Parallel()

	t.Run("no findings is not broken", func(t *testing.T) {
		t.Parallel()

		msg, broken := Guard("mod", "check", func(Context, *InvariantReport) {})(Context{})
		require.False(t, broken)
		require.Contains(t, msg, "no violations found")
	})

	t.Run("under the cap every finding is shown", func(t *testing.T) {
		t.Parallel()

		msg, broken := Guard("mod", "check", func(_ Context, rep *InvariantReport) {
			for i := range reportCap {
				rep.Addf("finding %d", i)
			}
		})(Context{})
		require.True(t, broken)
		require.Contains(t, msg, "10 violation(s) found")
		require.NotContains(t, msg, "shown", "nothing was withheld, so do not claim a subset")
		require.Equal(t, reportCap, strings.Count(msg, "finding "))
	})

	t.Run("over the cap the count is full but the message is bounded", func(t *testing.T) {
		t.Parallel()

		const found = reportCap * 100
		msg, broken := Guard("mod", "check", func(_ Context, rep *InvariantReport) {
			for i := range found {
				rep.Addf("finding %d", i)
			}
		})(Context{})
		require.True(t, broken)
		require.Contains(t, msg, "1000 violation(s) found", "the count must be exact even when the detail is not")
		require.Contains(t, msg, "first 10 shown")
		require.Equal(t, reportCap, strings.Count(msg, "finding "),
			"a corrupt store must not be able to grow the message it produces")
		require.NotContains(t, msg, "finding 10", "the withheld findings are the later ones")
	})

	t.Run("a panicking check reports instead of escaping", func(t *testing.T) {
		t.Parallel()

		var msg string
		var broken bool
		require.NotPanics(t, func() {
			msg, broken = Guard("mod", "check", func(Context, *InvariantReport) {
				panic("read a key that is not there")
			})(Context{})
		})
		require.True(t, broken, "a bug in a check must be louder than a violation, not quieter")
		require.Contains(t, msg, "read a key that is not there")
	})
}
