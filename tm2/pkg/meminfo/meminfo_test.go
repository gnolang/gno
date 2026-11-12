package meminfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	t.Parallel()

	mi, ok := Read()
	if !ok {
		t.Skip("not implemented on this platform")
	}
	// No floor on Total: a cgroup can hold it to a few hundred MiB, or less.
	// The units are checked against fixed input instead, in TestParseMemInfo.
	require.NotZero(t, mi.Total)
	assert.LessOrEqual(t, mi.Available, mi.Total, "more available than exists")
}
