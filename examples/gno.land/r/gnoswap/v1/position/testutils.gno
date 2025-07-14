package position

import (
	"testing"

	"gno.land/p/demo/avl"

	"gno.land/r/gnoswap/v1/gnft"
	"gno.land/r/gnoswap/v1/pool"
)

func InitPositionTest(t *testing.T) {
	t.Helper()

	func(cur realm) {
		positions = avl.NewTree()
		nextId = 1
	}(cross)
}

func initPositionTest(t *testing.T) {
	t.Helper()

	InitPositionTest(t)
	pool.InitPoolTest(t)
	gnft.InitGNFTTest(t)
}
