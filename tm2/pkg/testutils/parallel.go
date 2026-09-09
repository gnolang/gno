package testutils

import (
	"os"
	"runtime"
	"strconv"
)

// MaxParallelEnv pins the worker counts derived in this file. Setting it opts
// out of scaling to fit the machine, which is occasionally useful when
// profiling or bisecting a memory problem.
const MaxParallelEnv = "GNO_TEST_MAX_PARALLEL"

// defaultMaxParallel is the worker count used when the machine's memory can't
// be read. It matches the CPU count of the GitHub-hosted runners the suites are
// tuned on, which is why CI stays within its memory budget while a workstation
// with four times the cores does not.
const defaultMaxParallel = 4

// MaxParallelOverride reports the [MaxParallelEnv] setting, if it holds a
// positive integer.
func MaxParallelOverride() (int, bool) {
	n, err := strconv.Atoi(os.Getenv(MaxParallelEnv))
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// MaxParallel reports how many memory-heavy workers a test suite may run
// concurrently, for suites that must fix the count up front — a pool of GnoVM
// stores, say, as opposed to workers started one at a time, which can instead
// ramp against [MemInfo].
//
// Such workers cost hundreds of megabytes of live heap each, so sizing their
// concurrency off GOMAXPROCS — the default for `go test` parallelism — makes
// peak memory scale with the core count. That is the wrong axis: these suites
// are bound by memory, not CPU, and stop getting faster well before they stop
// getting bigger.
func MaxParallel() int {
	if n, ok := MaxParallelOverride(); ok {
		return n
	}
	return min(runtime.GOMAXPROCS(0), defaultMaxParallel)
}

// MemInfo is a snapshot of the machine's memory, in bytes.
type MemInfo struct {
	// Total is the machine's physical memory.
	Total uint64

	// Available is the kernel's estimate of what can still be allocated
	// without pushing the system into swap. Unlike free memory it accounts for
	// reclaimable caches, and it falls as this process's heap grows — which is
	// what makes it usable as a feedback signal for admitting more workers.
	Available uint64
}
