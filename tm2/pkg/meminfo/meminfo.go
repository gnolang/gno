// Package meminfo reads how much memory this process may still use, for code
// that sizes its concurrency to the machine rather than to a fixed count.
package meminfo

// Info is a snapshot of the machine's memory, in bytes.
type Info struct {
	// Total is the machine's physical memory.
	Total uint64

	// Available is the kernel's estimate of what can still be allocated
	// without pushing the system into swap. Unlike free memory it accounts for
	// reclaimable caches, and it falls as this process's heap grows — which is
	// what makes it usable as a feedback signal for admitting more workers.
	Available uint64
}
