package crashing

import (
	"github.com/cespare/xxhash/v2"
)

func X_UseSum64String(stVal string) uint64 {
	return xxhash.Sum64String(stVal)
}
