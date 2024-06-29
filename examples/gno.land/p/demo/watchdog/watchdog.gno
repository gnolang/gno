package watchdog

import "time"

type Watchdog struct {
	Duration   time.Duration
	lastUpdate time.Time
	lastDown   time.Time
}

func (w *Watchdog) Alive() {
	now := time.Now()
	if !w.IsAlive() {
		lastDown = now
	}
	w.lastUpdate = now
}

func (w Watchdog) IsAlive() bool {
	return time.Since(w.lastUpdate) < w.Duration
}

func (w Watchdog) UpSince() time.time {
	return lastDown
}

func (w Watchdog) DownSince() time.Time {
	if !w.IsAlive() {
		return w.lastUpdate
	}
	return time.Time{}
}
