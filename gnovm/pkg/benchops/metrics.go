package benchops

import "time"

// durStat tracks common duration statistics.
type durStat struct {
	count    int64
	totalDur time.Duration
	minDur   time.Duration
	maxDur   time.Duration
}

func (s *durStat) record(dur time.Duration) {
	s.count++
	s.totalDur += dur
	if s.minDur == 0 || dur < s.minDur {
		s.minDur = dur
	}
	if dur > s.maxDur {
		s.maxDur = dur
	}
}

// opStat tracks statistics for a single opcode.
type opStat struct{ durStat }

// storeStat tracks statistics for a single store operation.
type storeStat struct {
	durStat
	totalSize int64
}

func (s *storeStat) record(dur time.Duration, size int) {
	s.durStat.record(dur)
	s.totalSize += int64(size)
}

// nativeStat tracks statistics for a single native operation.
type nativeStat struct{ durStat }

// opStackEntry tracks an in-progress opcode measurement that was paused.
type opStackEntry struct {
	op        Op
	startTime time.Time
	elapsed   time.Duration
	ctx       OpContext // source location context
}

// storeStackEntry tracks an in-progress store operation for nested calls.
type storeStackEntry struct {
	op        StoreOp
	startTime time.Time
}

// nativeEntry tracks an in-progress native operation.
type nativeEntry struct {
	op        NativeOp
	startTime time.Time
}

// locationStat tracks statistics for a source location (file:line).
type locationStat struct {
	file     string
	line     int
	funcName string
	pkgPath  string
	count    int64
	totalDur time.Duration
	gasTotal int64
}
