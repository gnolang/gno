> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `watchdog` - Liveness timer

A deadman-switch primitive: a `Watchdog` is "alive" as long as `Alive()` has been
called within the last `Duration`. Use it to monitor an off-chain process, an
oracle, or any actor that must check in regularly.

## Usage

```go
package myrealm

import (
    "time"

    "gno.land/p/nt/watchdog/v0"
)

var w = &watchdog.Watchdog{Duration: 10 * time.Minute}

// Heartbeat: the monitored party calls this on every cycle.
func Heartbeat(cur realm) {
    w.Alive()
}

// Render shows the current liveness state.
func Render(_ string) string {
    return "status: " + w.Status() // "OK" or "KO"
}
```

Note that `Watchdog` has no built-in authorization — gate `Alive()` with
`gno.land/p/nt/ownable/v0` or another check if only a specific party should be
able to reset the timer.

## API

```go
type Watchdog struct {
    Duration time.Duration // how long since the last Alive() before going down
    // unexported timestamps
}

// Heartbeat: marks the watchdog alive at time.Now().
func (w *Watchdog) Alive()

// Queries.
func (w Watchdog)  IsAlive()   bool      // true if time.Since(lastUpdate) < Duration
func (w Watchdog)  Status()    string    // "OK" if alive, "KO" otherwise
func (w Watchdog)  UpSince()   time.Time // last time the watchdog recovered from down
func (w Watchdog)  DownSince() time.Time // last update if currently down, zero value otherwise
```

## Notes

- A fresh `Watchdog{Duration: d}` starts as **down** until the first `Alive()`
  call (`lastUpdate` is the zero time).
- `Duration` is a plain field — it can be tuned at runtime, but consider gating
  any write to it.
- Liveness is measured in **block time**, not wall clock: `time.Now()` returns the
  block timestamp, and the down/up transition is observed lazily on the next read
  (`IsAlive`/`Status`). Size `Duration` against block cadence, not real seconds.
