package gnolang

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatementCoverage_Basic(t *testing.T) {
	cov := NewStatementCoverage()
	n := S(If(X("true"), Ss(Return(X("1"))), nil))
	cov.TrackNode(n)
	require.Greater(t, cov.Percent(), -1.0)

	// Marking an unrelated stmt should not affect tracked percent.
	cov.MarkExecuted(S(Return(X("2"))))
	require.Equal(t, 0.0, cov.Percent())

	cov.MarkExecuted(n)
	require.Greater(t, cov.Percent(), 0.0)
}
