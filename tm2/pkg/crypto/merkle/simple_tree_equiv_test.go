package merkle

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleHashFromByteSlicesMatchesIterative(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewSource(1))
	counts := []int{0, 1, 2, 3, 4, 5, 7, 8, 9, 15, 16, 17, 31, 32, 33, 100, 255, 256, 257, 1000, 4097}
	// Leaf lengths cover nil, empty, short and multi-block leaves.
	lengths := []int{-1, 0, 1, 31, 32, 33, 63, 64, 65, 66, 128, 1000}
	for _, n := range counts {
		for _, l := range lengths {
			items := make([][]byte, n)
			for i := range items {
				switch {
				case l < 0:
					items[i] = nil
				case i%3 == 2:
					// Mix lengths inside one tree too.
					items[i] = make([]byte, r.Intn(130))
					r.Read(items[i])
				default:
					items[i] = make([]byte, l)
					r.Read(items[i])
				}
			}
			require.Equal(t, SimpleHashFromByteSlicesIterative(items), SimpleHashFromByteSlices(items), "n=%d len=%d", n, l)
		}
	}
}

var allocPinSink []byte

// The root is the only allocation, whatever the leaf count or length.
func TestSimpleHashFromByteSlicesAllocations(t *testing.T) {
	for _, n := range []int{1, 2, 3, 1000, 4097} {
		for _, l := range []int{-1, 0, 1, 55, 56, 63, 64, 65, 1000, 1 << 16} {
			if l == 1<<16 && n > 3 {
				continue
			}
			items := make([][]byte, n)
			for i := range items {
				if l >= 0 {
					items[i] = make([]byte, l)
				}
			}
			got := testing.AllocsPerRun(5, func() { allocPinSink = SimpleHashFromByteSlices(items) })
			require.LessOrEqual(t, got, 1.0, "n=%d len=%d", n, l)
		}
	}
}
