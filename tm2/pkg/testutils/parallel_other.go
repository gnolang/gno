//go:build !linux

package testutils

import "github.com/pbnjay/memory"

// ReadMemInfo reports the machine's memory via github.com/pbnjay/memory, which
// covers darwin, windows and the BSDs. Linux has its own implementation, using
// MemAvailable rather than the free-memory figure the library reports there.
//
// How good Available is varies by platform. Windows reports genuinely available
// physical memory, so it behaves like the Linux reading. darwin counts only
// wholly free pages — not the inactive and purgeable ones it would reclaim on
// demand — and so under-reports on any machine putting the rest to use as
// cache; callers get a conservative ramp there rather than a wrong one. On
// platforms the library does not cover it reports zero and this returns false,
// leaving callers on their static fallback.
func ReadMemInfo() (MemInfo, bool) {
	mi := MemInfo{
		Total:     memory.TotalMemory(),
		Available: memory.FreeMemory(),
	}
	return mi, mi.Total > 0 && mi.Available > 0
}
