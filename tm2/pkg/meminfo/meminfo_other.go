//go:build !linux && !darwin

package meminfo

// Read has no reading on this platform, which leaves callers on their
// static fallback. On the BSDs, the figure within reach, hw.usermem, is
// physical memory less wired pages: it does not fall as a heap grows, so a
// budget fed by it would ramp to GOMAXPROCS and never shrink.
func Read() (Info, bool) {
	return Info{}, false
}
