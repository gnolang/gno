package gnolang

import (
	"strings"
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

func TestStatementCoverage_CountMode(t *testing.T) {
	cov := NewStatementCoverageWithMode(CoverModeCount)
	n := S(Return(X("1")))
	cov.TrackNode(n)

	cov.MarkExecuted(n)
	cov.MarkExecuted(n)
	cov.MarkExecuted(n)

	require.Equal(t, 3, cov.HitCount(n))
}

func TestStatementCoverage_SetMode(t *testing.T) {
	cov := NewStatementCoverageWithMode(CoverModeSet)
	n := S(Return(X("1")))
	cov.TrackNode(n)

	cov.MarkExecuted(n)
	cov.MarkExecuted(n)
	cov.MarkExecuted(n)

	// Set mode always records 1.
	require.Equal(t, 1, cov.HitCount(n))
}

func TestStatementCoverage_AtomicMode(t *testing.T) {
	cov := NewStatementCoverageWithMode(CoverModeAtomic)
	n := S(Return(X("1")))
	cov.TrackNode(n)

	cov.MarkExecuted(n)
	cov.MarkExecuted(n)

	// Atomic behaves like count in single-threaded gno.
	require.Equal(t, 2, cov.HitCount(n))
}

func TestParseCoverMode(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  CoverMode
		err   bool
	}{
		{"set", CoverModeSet, false},
		{"count", CoverModeCount, false},
		{"atomic", CoverModeAtomic, false},
		{"invalid", "", true},
	} {
		got, err := ParseCoverMode(tc.input)
		if tc.err {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		}
	}
}

func TestStatementCoverage_Profile(t *testing.T) {
	cov := NewStatementCoverageWithMode(CoverModeCount)
	n := S(Return(X("1")))
	cov.TrackStmtWithFile(n, "foo.gno")

	cov.MarkExecuted(n)
	cov.MarkExecuted(n)

	profile := cov.Profile("gno.land/r/demo/foo")
	require.True(t, strings.HasPrefix(profile, "mode: count\n"))
	require.Contains(t, profile, "gno.land/r/demo/foo/foo.gno:")
}
