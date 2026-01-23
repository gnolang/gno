package benchops

import "time"

// opStat tracks statistics for a single opcode.
type opStat struct {
	count    int64
	totalDur time.Duration
	minDur   time.Duration
	maxDur   time.Duration
}

func (s *opStat) record(dur time.Duration) {
	s.count++
	s.totalDur += dur
	if s.minDur == 0 || dur < s.minDur {
		s.minDur = dur
	}
	if dur > s.maxDur {
		s.maxDur = dur
	}
}

// storeStat tracks statistics for a single store operation.
type storeStat struct {
	count     int64
	totalDur  time.Duration
	totalSize int64
	minDur    time.Duration
	maxDur    time.Duration
}

func (s *storeStat) record(dur time.Duration, size int) {
	s.count++
	s.totalDur += dur
	s.totalSize += int64(size)
	if s.minDur == 0 || dur < s.minDur {
		s.minDur = dur
	}
	if dur > s.maxDur {
		s.maxDur = dur
	}
}

// nativeStat tracks statistics for a single native operation.
type nativeStat struct {
	count    int64
	totalDur time.Duration
	minDur   time.Duration
	maxDur   time.Duration
}

func (s *nativeStat) record(dur time.Duration) {
	s.count++
	s.totalDur += dur
	if s.minDur == 0 || dur < s.minDur {
		s.minDur = dur
	}
	if dur > s.maxDur {
		s.maxDur = dur
	}
}

// opStackEntry tracks an in-progress opcode measurement that was paused.
type opStackEntry struct {
	op        Op
	startTime time.Time
	elapsed   time.Duration
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
