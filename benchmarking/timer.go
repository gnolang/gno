package benchmarking

import "time"

type timer struct {
	startTime   time.Time
	elapsedTime time.Duration
	isStopped   bool
}

func (t *timer) start() {
	t.startTime = time.Now()
}

func (t *timer) stop() {
	if t.isStopped {
		return
	}

	t.elapsedTime += time.Since(t.startTime)
	t.isStopped = true
}
