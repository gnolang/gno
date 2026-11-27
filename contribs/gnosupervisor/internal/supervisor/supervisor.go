// Package supervisor runs a gnoland node across the coordinated upgrades a
// chain's upgrade ledger (misc/deployments/<chain>/upgrades.json) records:
// it starts the release that runs the node's next block, and when that
// release stops at its range's halt height, it starts the next one on the
// same data directory.
package supervisor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Heights reads the node's last committed block. It is only called while no
// gnoland process runs on the data directory.
type Heights interface {
	CommittedHeight() (int64, error)
}

// Artifacts prepares the release a leg runs. Prepare may be called from two
// goroutines at once (the current leg and the next one's prefetch).
type Artifacts interface {
	Prepare(ctx context.Context, leg Leg) (Artifact, error)
}

// LiveHeight reads the running node's latest block height, over RPC.
type LiveHeight func(ctx context.Context) (int64, error)

// Config is what Run needs.
type Config struct {
	Plan      *Plan
	Node      Heights
	Artifacts Artifacts

	DataDir   string   // passed to gnoland start --data-dir
	Genesis   string   // passed to gnoland start --genesis
	ExtraArgs []string // appended to gnoland start
	// LockFile is the data directory lock (see Lock). Each child inherits
	// it, so the lock outlives a killed supervisor until its node exits.
	LockFile *os.File

	// StopTimeout is how long a child gets to exit after SIGTERM before
	// it is killed.
	StopTimeout time.Duration
	// MaxRestarts is how many unexpected exits in a row, with no new block
	// committed in between, are retried before Run gives up.
	MaxRestarts int
	// BackoffMin and BackoffMax bound the wait before a retry; it doubles
	// on each consecutive failure.
	BackoffMin, BackoffMax time.Duration

	// Live, when set, lets the supervisor notice a node that reached its
	// range's halt height without exiting: a node that has left block sync
	// for consensus catches the halt panic and stays up. Once Live has
	// reported the range's end for StuckAfter, the child is stopped and the
	// switch goes on as for an exit. Nil disables the check.
	Live LiveHeight
	// HaltParam reads the chain's node:p:halt_height from the running
	// node. When the node stops moving at that height and the ledger does
	// not record a halt there, the supervisor records the stop at once (see
	// HaltMarkerFile), so that no restart runs this release past it.
	HaltParam    LiveHeight
	PollInterval time.Duration
	StuckAfter   time.Duration
	// StallAfter is how long a node may commit no block before the
	// supervisor logs an error (it does not stop it): a live halt, or a
	// node with no peers. 0 disables it.
	StallAfter time.Duration

	// ContinueAfterHalt lets the node go on, on the same release, after a
	// stop at this height that the ledger does not record as a halt (a
	// pause). 0 refuses every such stop.
	ContinueAfterHalt int64

	Stdout, Stderr io.Writer
	Log            *slog.Logger
}

// Run supervises the node until ctx is cancelled (it then stops gnoland and
// returns nil) or until it cannot go on safely (it returns why).
//
// Every iteration reads the committed height of the stopped node and asks
// the plan which release runs the next block, so a restarted supervisor
// resumes from the data directory, never from its own memory.
func Run(ctx context.Context, c Config) error {
	var prefetch sync.WaitGroup
	defer prefetch.Wait()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // stops a prefetch still running when Run returns

	failures := 0
	for {
		height, err := c.Node.CommittedHeight()
		if err != nil {
			return err
		}
		if err := c.checkHaltMarker(height); err != nil {
			return err
		}
		leg, err := c.Plan.Resolve(height + 1)
		if err != nil {
			return err
		}
		art, err := c.Artifacts.Prepare(ctx, leg)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if ctx.Err() != nil {
			return nil
		}
		// Prepare the next release while this one runs, so that the switch
		// at the halt does not wait on the network, and a registry that is
		// down does not delay the node now.
		prefetch.Go(func() {
			c.prefetchNext(ctx, leg)
		})

		c.Log.Info("starting gnoland", "version", leg.Version, "next_block", height+1,
			"range_from", leg.Range.From, "range_to", leg.Range.To, "digest", art.Digest)
		code, stuck, err := c.runChild(ctx, leg, art)
		if ctx.Err() != nil {
			c.Log.Info("gnoland stopped", "exit_code", code)
			// A halt can end the child just as the supervisor is told to
			// stop: record it all the same.
			if code == 2 {
				after, err := c.Node.CommittedHeight()
				if err != nil {
					return err
				}
				if c.unknownHalt(leg, height, after) {
					return writeHaltMarker(c.DataDir, after)
				}
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("start gnoland %s: %w", leg.Version, err)
		}

		after, err := c.Node.CommittedHeight()
		if err != nil {
			return err
		}
		switch {
		case !leg.Last() && after == leg.Range.To:
			// The ledger's halt: the next block belongs to the next release,
			// which the next iteration resolves. A node catching up exits
			// with code 2 here (the halt is an unrecovered panic in the next
			// block's BeginBlock); any other code at exactly this height is
			// worth a warning but leads to the same next step.
			if code == 2 || stuck {
				c.Log.Info("halt height reached; switching release", "version", leg.Version, "halt_height", after)
			} else {
				c.Log.Warn("gnoland exited at the halt height with an unexpected code; switching release",
					"version", leg.Version, "halt_height", after, "exit_code", code)
			}
			failures = 0
			continue
		case !leg.Last() && after > leg.Range.To:
			// The startup gates of later releases make this unreachable on a
			// healthy ledger; never guess past it.
			return fmt.Errorf("%s committed block %d, past the end of its range (%d): the ledger and the node disagree; stopping",
				leg.Version, after, leg.Range.To)
		}

		if code == 2 && c.unknownHalt(leg, height, after) {
			// Code 2 is an unrecovered panic, and a governance halt is one: a
			// halt the ledger does not record yet. It may be an upgrade whose
			// entry is not in this copy of the ledger (entries are written
			// after the halt), with or without a version floor; restarting
			// this release could then run the next release's blocks. The
			// refusal is written to the data directory, so that a restarted
			// supervisor (by hand, or by systemd) refuses too.
			if err := writeHaltMarker(c.DataDir, after); err != nil {
				return err
			}
			return unknownHaltError("gnoland "+leg.Version, after)
		}
		if after > height {
			failures = 0 // it made progress before failing
		}
		failures++
		if failures > c.MaxRestarts {
			return fmt.Errorf("gnoland %s exited %d times in a row at height %d (last exit code %d); giving up",
				leg.Version, failures, after, code)
		}
		wait := c.backoff(failures)
		c.Log.Warn("gnoland exited unexpectedly; restarting the same release",
			"version", leg.Version, "exit_code", code, "committed_height", after, "attempt", failures, "wait", wait)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
	}
}

// HaltMarkerFile records, in the data directory, a stop at a height the
// ledger does not record as a halt.
const HaltMarkerFile = "gnosupervisor.halt"

func unknownHaltError(who string, h int64) error {
	return fmt.Errorf("%s stopped with code 2 after committing block %d, which the ledger does not record as a halt: "+
		"if an upgrade happened there, update the ledger; if it was a pause, run again with --continue-after-halt %d",
		who, h, h)
}

// checkHaltMarker refuses to go on from a stop the ledger does not explain,
// until the ledger records a halt there or the operator confirms a pause. A
// marker the node has moved past is removed.
func (c Config) checkHaltMarker(height int64) error {
	path := filepath.Join(c.DataDir, HaltMarkerFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	h, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	switch {
	case height < h:
		// Not there yet (a stop between flush points can leave the state a
		// little behind): keep the record, the replay halts at h again.
		return nil
	case height == h && h != c.ContinueAfterHalt && !c.Plan.HaltAt(h):
		return unknownHaltError("the node", h) // recorded by an earlier run
	}
	// Past it, confirmed as a pause, or now recorded by the ledger.
	return os.Remove(path)
}

// unknownHalt reports whether a stop after committing block after, having
// started at height, is a halt the ledger does not record and the operator
// has not confirmed as a pause.
func (c Config) unknownHalt(leg Leg, height, after int64) bool {
	return after > height && !(!leg.Last() && after == leg.Range.To) && after != c.ContinueAfterHalt
}

// writeHaltMarker records h durably: a new file, synced, renamed over the
// marker, and the directory synced, so a reader or a power cut never sees
// a partial record.
func writeHaltMarker(dataDir string, h int64) error {
	path := filepath.Join(dataDir, HaltMarkerFile)
	if err := writeFileSync(path+".tmp", []byte(strconv.FormatInt(h, 10)+"\n")); err != nil {
		return err
	}
	if err := os.Rename(path+".tmp", path); err != nil {
		return err
	}
	return syncDir(dataDir)
}

// recordIfUnknownHalt writes the halt marker when the running node sits at
// the chain's halt height and the ledger records no halt there. A node at
// the chain's halt height has committed it and will not commit the next
// block, so there is no need to wait for it to look idle.
func (c Config) recordIfUnknownHalt(ctx context.Context, leg Leg, h int64) {
	if c.HaltParam == nil || h == c.ContinueAfterHalt || c.Plan.HaltAt(h) {
		return
	}
	hh, err := c.HaltParam(ctx)
	if err != nil || hh != h {
		return
	}
	if data, err := os.ReadFile(filepath.Join(c.DataDir, HaltMarkerFile)); err == nil && strings.TrimSpace(string(data)) == strconv.FormatInt(h, 10) {
		return // already recorded
	}
	if err := writeHaltMarker(c.DataDir, h); err != nil {
		c.Log.Error("could not record the halt", "err", err)
		return
	}
	c.Log.Error("the chain halted at a height the ledger does not record; this release will not be restarted past it",
		"version", leg.Version, "halt_height", h)
}

func (c Config) backoff(failures int) time.Duration {
	d := c.BackoffMin
	for i := 1; i < failures && d < c.BackoffMax; i++ {
		d *= 2
	}
	return min(d, c.BackoffMax)
}

// prefetchNext prepares the release after leg. A failure here is only
// logged: the switch will try again and fail then if it must.
func (c Config) prefetchNext(ctx context.Context, leg Leg) {
	if leg.Last() {
		return
	}
	next, err := c.Plan.Resolve(leg.Range.To + 1)
	if err == nil {
		_, err = c.Artifacts.Prepare(ctx, next)
	}
	if err != nil && ctx.Err() == nil {
		c.Log.Warn("could not prepare the next release ahead of its halt", "after", leg.Version, "err", err)
	}
}

// runChild runs gnoland start on the data directory until it exits, or
// until ctx is cancelled, which it forwards as SIGTERM. It returns the exit
// code (-1 if a signal ended the process), and whether the supervisor
// stopped a node it found idle at its range's halt height.
func (c Config) runChild(ctx context.Context, leg Leg, art Artifact) (int, bool, error) {
	args := append([]string{
		"start",
		"--data-dir", c.DataDir,
		"--genesis", c.Genesis,
		"--gnoroot-dir", art.GnoRoot,
	}, c.ExtraArgs...)
	cmd := exec.Command(art.Binary, args...)
	cmd.Env = append(os.Environ(), "GNOROOT="+art.GnoRoot)
	cmd.Stdout, cmd.Stderr = c.Stdout, c.Stderr
	if c.LockFile != nil {
		cmd.ExtraFiles = []*os.File{c.LockFile}
	}
	cmd.SysProcAttr = childAttrs()

	// Linux sends the parent-death signal when the thread that started the
	// child exits, not the process: keep that thread for the child's life.
	started := make(chan error, 1)
	done := make(chan error, 1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		if err := cmd.Start(); err != nil {
			started <- err
			return
		}
		started <- nil
		done <- cmd.Wait()
	}()
	if err := <-started; err != nil {
		return -1, false, err
	}

	stuck := make(chan struct{})
	watchCtx, stopWatch := context.WithCancel(ctx)
	var watcher sync.WaitGroup
	if c.Live != nil {
		watcher.Go(func() {
			c.watchHalt(watchCtx, leg, stuck)
		})
	}

	var err error
	stopped := false
	select {
	case err = <-done:
	case <-stuck:
		stopped = true
		err = c.stop(cmd, done)
	case <-ctx.Done():
		// Told to stop: if the node sits at a halt the ledger does not
		// record, record it before stopping the node.
		if c.Live != nil {
			check, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if h, herr := c.Live(check); herr == nil {
				c.recordIfUnknownHalt(check, leg, h)
			}
			cancel()
		}
		err = c.stop(cmd, done)
	}
	stopWatch()
	watcher.Wait()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0, stopped, nil
	case errors.As(err, &exitErr):
		return exitErr.ExitCode(), stopped, nil
	default:
		return -1, stopped, err
	}
}

// stop sends SIGTERM, then SIGKILL after StopTimeout, and returns Wait's error.
func (c Config) stop(cmd *exec.Cmd, done <-chan error) error {
	_ = cmd.Process.Signal(syscall.SIGTERM)
	select {
	case err := <-done:
		return err
	case <-time.After(c.StopTimeout):
		c.Log.Warn("gnoland did not stop in time; killing it", "timeout", c.StopTimeout)
		_ = cmd.Process.Kill()
		return <-done
	}
}

// watchHalt watches the running node over RPC.
//
// It closes stuck once the node has reported its range's last block,
// without moving, for StuckAfter. Only the ledger's halt heights are
// watched, so stopping there is what the ledger asks for anyway.
//
// On every range, once the node has not moved for StuckAfter, it reads the
// chain's halt height: if the node sits at a halt the ledger does not
// record, it records the stop in the data directory (HaltMarkerFile), so
// that no restart runs this release past it. After StallAfter without a new
// block it logs an error; only the operator can tell what the halt was.
func (c Config) watchHalt(ctx context.Context, leg Leg, stuck chan<- struct{}) {
	lastHeight, lastMove := int64(-1), time.Now()
	warned := false
	t := time.NewTicker(c.PollInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		h, err := c.Live(ctx)
		if err != nil {
			continue
		}
		if h != lastHeight {
			lastHeight, lastMove, warned = h, time.Now(), false
			c.recordIfUnknownHalt(ctx, leg, h)
			continue
		}
		idle := time.Since(lastMove)
		if idle < c.StuckAfter {
			continue
		}
		if !leg.Last() && h >= leg.Range.To {
			c.Log.Info("node is idle at its range's halt height without exiting; stopping it",
				"version", leg.Version, "height", h, "halt_height", leg.Range.To)
			close(stuck)
			return
		}
		if c.StallAfter > 0 && !warned && idle >= c.StallAfter {
			c.Log.Error("the node has committed no block for a while; if the chain halted at a height the ledger does not record, update the ledger and restart",
				"version", leg.Version, "height", h, "for", idle.Round(time.Second))
			warned = true
		}
	}
}
