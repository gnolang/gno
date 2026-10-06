//go:build unix

package benchmarks

import (
	"os"
	"syscall"
)

// fileDiskBytes returns the bytes actually allocated to info on disk
// (st_blocks × 512), so sparse pre-mmapped map files (lmdb/mdbx) report real
// usage instead of their apparent map-ceiling size. Falls back to apparent
// size if the platform stat is unavailable.
func fileDiskBytes(info os.FileInfo) int64 {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return st.Blocks * 512 // st_blocks is in 512-byte units (POSIX)
	}
	return info.Size()
}
