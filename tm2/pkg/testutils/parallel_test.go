package testutils

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxParallelOverride(t *testing.T) {
	for _, tt := range []struct {
		set  string
		want int
		ok   bool
	}{
		{set: "", want: 0, ok: false},
		{set: "8", want: 8, ok: true},
		{set: "1", want: 1, ok: true},
		{set: "0", want: 0, ok: false},    // meaningless: would admit nothing
		{set: "-2", want: 0, ok: false},   // ditto
		{set: "many", want: 0, ok: false}, // typo shouldn't silently pin a count
		{set: "4.5", want: 0, ok: false},
	} {
		t.Run("set="+tt.set, func(t *testing.T) {
			t.Setenv(MaxParallelEnv, tt.set)
			n, ok := MaxParallelOverride()
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, n)
		})
	}
}

func TestMaxParallel(t *testing.T) {
	t.Run("override wins", func(t *testing.T) {
		t.Setenv(MaxParallelEnv, "11")
		assert.Equal(t, 11, MaxParallel())
	})

	t.Run("otherwise capped", func(t *testing.T) {
		t.Setenv(MaxParallelEnv, "")
		n := MaxParallel()
		assert.GreaterOrEqual(t, n, 1)
		assert.LessOrEqual(t, n, defaultMaxParallel)
		assert.LessOrEqual(t, n, runtime.GOMAXPROCS(0), "never more workers than cores")
	})
}

func TestReadMemInfo(t *testing.T) {
	t.Parallel()

	mi, ok := ReadMemInfo()
	if !ok {
		t.Skip("not implemented on this platform")
	}
	require.NotZero(t, mi.Total)
	assert.LessOrEqual(t, mi.Available, mi.Total, "more available than exists")
	// A machine running this test has some memory going spare, and the units
	// had better be bytes: a kB/byte mix-up would show up as an absurd total.
	assert.Greater(t, mi.Total, uint64(256<<20), "total looks like it is not in bytes")
}
