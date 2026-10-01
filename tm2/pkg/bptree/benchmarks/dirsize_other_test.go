//go:build !unix

package benchmarks

import "os"

// fileDiskBytes falls back to apparent size off-Unix (no portable st_blocks).
func fileDiskBytes(info os.FileInfo) int64 { return info.Size() }
