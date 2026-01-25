package benchops

import "time"

// ---- Gas-only stats (default, minimal overhead ~5ns per op)

// opStat tracks statistics for a single opcode (gas-only by default).
type opStat struct {
	count int64
	gas   int64
}

func (s *opStat) record(gas int64) {
	s.count++
	s.gas += gas
}

// storeStat tracks statistics for a single store operation (gas-only by default).
type storeStat struct {
	count     int64
	totalSize int64
}

func (s *storeStat) record(size int) {
	s.count++
	s.totalSize += int64(size)
}

// nativeStat tracks statistics for a single native operation (gas-only by default).
type nativeStat struct {
	count int64
}

func (s *nativeStat) record() {
	s.count++
}

// ---- Stats with timing (opt-in via WithTiming())

// timingStats holds min/max/total duration tracking, embedded in timed stat types.
type timingStats struct {
	totalDur time.Duration
	minDur   time.Duration
	maxDur   time.Duration
}

func (t *timingStats) recordTiming(dur time.Duration) {
	t.totalDur += dur
	if t.minDur == 0 || dur < t.minDur {
		t.minDur = dur
	}
	if dur > t.maxDur {
		t.maxDur = dur
	}
}

// opStatTimed extends opStat with timing information.
type opStatTimed struct {
	opStat
	timingStats
}

func (s *opStatTimed) recordTimed(gas int64, dur time.Duration) {
	s.count++
	s.gas += gas
	s.recordTiming(dur)
}

// storeStatTimed extends storeStat with timing information.
type storeStatTimed struct {
	storeStat
	timingStats
}

func (s *storeStatTimed) recordTimed(size int, dur time.Duration) {
	s.count++
	s.totalSize += int64(size)
	s.recordTiming(dur)
}

// nativeStatTimed extends nativeStat with timing information.
type nativeStatTimed struct {
	nativeStat
	timingStats
}

func (s *nativeStatTimed) recordTimed(dur time.Duration) {
	s.count++
	s.recordTiming(dur)
}

// ---- Stack entries for in-progress measurements

// opStackEntry tracks an in-progress opcode measurement that was paused.
type opStackEntry struct {
	op        Op
	startTime time.Time // only used if timing enabled
	elapsed   time.Duration
	ctx       OpContext // source location context
}

// storeStackEntry tracks an in-progress store operation for nested calls.
type storeStackEntry struct {
	op        StoreOp
	startTime time.Time // only used if timing enabled
}

// nativeEntry tracks an in-progress native operation.
type nativeEntry struct {
	op        NativeOp
	startTime time.Time // only used if timing enabled
}

// ---- Location stats (for hot spots analysis)

// locationStat tracks statistics for a source location (file:line).
type locationStat struct {
	file     string
	line     int
	funcName string
	pkgPath  string
	count    int64
	totalDur time.Duration // only populated if timing enabled
	gasTotal int64
}
