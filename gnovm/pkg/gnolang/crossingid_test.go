package gnolang

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// A crossing id must survive ObjectInfo's 48-bit store unchanged, also past
// 2^48 calls, or a long-running node refuses writes a fresh node allows.
func TestCrossingIDWidth(t *testing.T) {
	var seq atomic.Uint64
	seq.Store(crossingIDMask - 1)
	want := []uint64{crossingIDMask, 1, 2} // 2^48 wraps to 0, which is skipped
	for _, w := range want {
		id := nextIDFrom(&seq)
		require.Equal(t, w, id)
		var oi ObjectInfo
		oi.setMadeIn(id)
		require.Equal(t, id, oi.madeIn())
	}
}
