# PR #5313 — autofile: halt writes on disk space exhaustion with auto-recovery

## Context

`tm2/pkg/autofile.Group` is a rotating-file writer that backs the consensus WAL
and the file event store. Writes go through a `bufio.Writer` (`headBuf`), so the
actual disk I/O is deferred until `FlushAndSync` or `rotateFile`. On a node with
a filling disk, the buffered `Write` calls succeeded silently and the operator
only discovered the problem when a flush returned `ENOSPC`, which the consensus
WAL turns into a panic.

Issue #5061 asks the node to halt cleanly when space is unavailable and to emit
warnings before the critical threshold is reached.

## Decision

Add a disk-space gate to `Group`:

- Before each buffered write, `ensureDiskSpace` queries available space via
  `statfs` on `g.Dir`. The syscall is throttled to once every
  `diskSpaceCheckInterval` (100) writes to keep it off the latency-sensitive hot
  path; while halted the check runs on every write so recovery is prompt.
- Below `minDiskSpaceLimit` (16 MB) the group sets `halted=true` and returns
  `ErrDiskSpaceUnavailable`. Below `minDiskSpaceLimit * diskSpaceWarningMultiplier`
  (64 MB) a single warning is logged on entering the low-space band.
- Recovery is automatic: once a re-check sees sufficient space the group clears
  `halted` and resumes.
- `FlushAndSync` and `rotateFile` are the load-bearing `ENOSPC` sites because the
  buffer defers I/O. Both now route their errors through `handleIOErr`, which on
  `ENOSPC` halts the group and wraps the error in `ErrDiskSpaceUnavailable`.
  `rotateFile` no longer panics on `ENOSPC`; it halts and returns the wrapped
  error instead. Non-`ENOSPC` errors in `rotateFile` still panic, preserving
  prior behavior.
- Windows/wasm are stubbed: `availableDiskSpace` returns the
  `diskSpaceUnsupported` (`math.MaxUint64`) sentinel and checks are skipped.
- `Halted() bool` is a read-only getter so halt state can be surfaced in a status
  endpoint or metric.

## Alternatives considered

- **Configurable threshold (`GroupMinDiskSpaceLimit`).** An earlier revision made
  the limit a functional option; it was reduced to a fixed constant for
  simplicity. Re-introducing operator control over this liveness knob is a
  reasonable follow-up but is out of scope here.
- **Time-based or buffer-fill-based throttle.** A fixed write-count throttle was
  chosen for simplicity and determinism of the hot-path cost.
- **`Resume()` manual-recovery API.** Removed because auto-recovery is the right
  default; a manual resume added surface area with no production caller.

## Consequences

- The halt is a per-node liveness gate, not a chain-level state change. A node
  with less free disk than its peers halts first; under correlated disk pressure
  multiple validators could halt together. The 16 MB threshold is a fixed
  heuristic and is not deployment-aware.
- Disk-full now surfaces as a wrapped `ErrDiskSpaceUnavailable` from
  `FlushAndSync`/`rotateFile`, so callers using `errors.Is` can disambiguate.
  The consensus WAL still panics on `WriteSync` failure; draining consensus
  gracefully is a separate, larger change.
- `availableDiskSpaceFn` is an unexported package-level variable so tests can
  stub disk space without real syscalls; production code uses the platform
  `availableDiskSpace`.
