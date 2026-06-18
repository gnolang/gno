package integration

import (
	"flag"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/pbnjay/memory"
)

// memLimitFraction is the share of available RAM the test process may use as
// its soft heap limit (GOMEMLIMIT). The remainder is headroom for non-heap
// memory (goroutine stacks, OS buffers) and other processes.
const memLimitFraction = 0.80

// perNodeMemBudgetBytes is a conservative live-memory reservation per parallel
// in-memory integration node, used only to derive a parallelism cap — it is
// not a hard limit (GOMEMLIMIT is). It is grounded in a validated run: 4 nodes
// completed under a 9 GiB heap cap with ~13 GiB available, i.e. ~3 GiB of
// budget per node once the shared baseline is amortised in.
//
// Note we deliberately do NOT divide by the SEQ_TS peak RSS (~12 GiB): that
// figure is dominated by Go heap high-watermark retention plus the shared,
// process-global stdlib/typecheck caches that accumulate across all scripts,
// none of which scales per node. Using it would force parallelism to 1.
const perNodeMemBudgetBytes = 3 << 30 // 3 GiB

// configureForAvailableMemory bounds the integration suite's memory use so it
// does not OOM on machines where the defaults would. The testing default
// (-parallel = GOMAXPROCS) boots one in-memory node per core; on a high-core
// box with limited RAM (e.g. 16 cores / ~13 GiB free) that exhausts memory.
//
// Two cooperating guards are installed (each skipped if the caller set it
// explicitly):
//
//   - GOMEMLIMIT is set to memLimitFraction of available RAM, so the Go
//     runtime caps its own heap and collects under pressure regardless of how
//     much it would otherwise retain.
//   - -test.parallel is lowered to available/perNodeMemBudget, clamped to
//     [1, GOMAXPROCS], so we neither oversubscribe CPU nor pile up more live
//     nodes than the heap budget can hold. This cap is the real OOM guard.
//
// No-op when available memory can't be determined, leaving the defaults in
// place.
func configureForAvailableMemory() {
	avail := availableMemoryBytes()
	if avail == 0 {
		return // unknown: keep defaults
	}

	// Soft heap limit. Respect an explicit GOMEMLIMIT.
	if os.Getenv("GOMEMLIMIT") == "" {
		debug.SetMemoryLimit(int64(float64(avail) * memLimitFraction))
	}

	// Parallelism cap (the OOM guard). Respect an explicit -parallel.
	if flag.Lookup("test.parallel") == nil || parallelFlagSet() {
		return
	}
	n := int(avail / perNodeMemBudgetBytes)
	if n < 1 {
		n = 1
	}
	if maxP := runtime.GOMAXPROCS(0); n > maxP {
		n = maxP
	}
	_ = flag.Set("test.parallel", strconv.Itoa(n))
}

// availableMemoryBytes returns the memory the test process can realistically
// use. On Linux it reads MemAvailable from /proc/meminfo (which accounts for
// reclaimable page cache, unlike MemFree). Elsewhere it falls back to
// memory.FreeMemory(): portable (macOS/Windows/BSD) but reporting free rather
// than available, i.e. conservative. Returns 0 if neither is determinable.
func availableMemoryBytes() uint64 {
	if n := linuxMemAvailableBytes(); n != 0 {
		return n
	}
	return memory.FreeMemory()
}

// linuxMemAvailableBytes returns MemAvailable from /proc/meminfo, or 0 when it
// can't be read (e.g. non-Linux platforms, where /proc/meminfo is absent).
func linuxMemAvailableBytes() uint64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "MemAvailable:") {
			continue
		}
		fields := strings.Fields(line) // "MemAvailable:" "<kB>" "kB"
		if len(fields) < 2 {
			return 0
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0
		}
		return kb * 1024
	}
	return 0
}

// parallelFlagSet reports whether -test.parallel was passed on the command line.
func parallelFlagSet() bool {
	set := false
	flag.Visit(func(fl *flag.Flag) {
		if fl.Name == "test.parallel" {
			set = true
		}
	})
	return set
}
