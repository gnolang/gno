//go:build !cgo

package benchstore

import "testing"

func init() {
	if !testing.Testing() {
		panic("CGO is required for benchstore (lmdb/mdbx depend on C libraries)")
	}
}
