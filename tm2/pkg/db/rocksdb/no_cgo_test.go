//go:build !cgo

package rocksdb

import "testing"

func TestSkip(t *testing.T) {
	t.Skip("This package requires cgo to compile and test")
}
