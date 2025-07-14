package pool

import (
	"testing"

	"gno.land/p/demo/avl"
)

func InitPoolTest(t *testing.T) {
	t.Helper()

	func(cur realm) {
		pools = avl.NewTree()
		slot0FeeProtocol = 0
		poolCreationFee = 100_000_000
		withdrawalFeeBPS = 100
	}(cross)
}
