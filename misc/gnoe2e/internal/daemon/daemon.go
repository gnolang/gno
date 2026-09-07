// Package daemon supervises an auxiliary process that runs alongside a
// cluster: it starts the process, waits until the process is actually serving,
// and terminates it. It knows nothing about any particular daemon.
package daemon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const (
	defaultReadyWait  = 30 * time.Second
	readyPollInterval = 50 * time.Millisecond
	stopGrace         = 5 * time.Second
	// settleAfterReady is how long Start waits, after the probe first
	// succeeds, before trusting it. A probe only proves whatever it happens to
	// test; a process can still fail startup work that comes after whatever
	// made the probe pass. This is comfortably above how fast a refused TCP
	// connect fails (milliseconds) and comfortably below the 500ms-1s a test
	// suite's own retries already tick at, so it is a real but negligible cost
	// on every healthy start.
	settleAfterReady = 200 * time.Millisecond
)

// Probe reports whether the process is serving yet. It is called repeatedly
// until it returns nil, the process exits, or the deadline passes.
type Probe func(context.Context) error

// Config describes one process to supervise.
type Config struct {
	Name       string // used in error messages; defaults to the binary's base name
	BinaryPath string
	Args       []string
	Env        []string // appended to the parent environment
	Dir        string
	// Ready decides when Start may return. A nil Probe means the process is
	// considered ready as soon as it is alive.
	Ready     Probe
	ReadyWait time.Duration // defaults to defaultReadyWait
}

// Daemon is a running supervised process.
type Daemon struct {
	name    string
	cmd     *exec.Cmd
	out     *syncBuffer
	exited  chan struct{} // closed once the process has been reaped
	waitErr error

	stopOnce sync.Once
	stopErr  error
}

// Start launches the process and returns once it is ready.
//
// Readiness races three outcomes: the probe succeeds, the process exits, or
// the deadline passes. Watching for the exit matters as much as the probe: a
// daemon that opens its port and then dies would otherwise leave the caller
// blocked for the whole deadline with no explanation.
func Start(ctx context.Context, cfg Config) (*Daemon, error) {
	if cfg.BinaryPath == "" {
		return nil, errors.New("daemon: BinaryPath is required")
	}
	name := cfg.Name
	if name == "" {
		name = filepath.Base(cfg.BinaryPath)
	}
	readyWait := cfg.ReadyWait
	if readyWait <= 0 {
		readyWait = defaultReadyWait
	}

	out := &syncBuffer{}
	cmd := exec.Command(cfg.BinaryPath, cfg.Args...)
	cmd.Dir = cfg.Dir
	cmd.Env = append(os.Environ(), cfg.Env...)
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("daemon %s: start: %w", name, err)
	}

	d := &Daemon{name: name, cmd: cmd, out: out, exited: make(chan struct{})}
	go func() {
		d.waitErr = cmd.Wait()
		close(d.exited)
	}()

	if err := d.waitReady(ctx, cfg.Ready, readyWait); err != nil {
		_ = d.Stop()
		return nil, err
	}

	// The probe succeeding proves only that whatever it checks is up, not that
	// the process will keep running: startup work that runs after the probe's
	// condition is met can still kill the process. Give it a short settle
	// window to surface that before trusting the probe.
	select {
	case <-d.exited:
		_ = d.Stop()
		return nil, d.exitedAfterReadyError()
	case <-time.After(settleAfterReady):
	}
	return d, nil
}

func (d *Daemon) waitReady(ctx context.Context, probe Probe, wait time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	if probe == nil {
		select {
		case <-d.exited:
			return d.exitedError()
		case <-time.After(readyPollInterval):
			return nil
		}
	}

	ticker := time.NewTicker(readyPollInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		if err := probe(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-d.exited:
			return d.exitedError()
		case <-ctx.Done():
			return fmt.Errorf("daemon %s: not ready after %s: %w\n%s",
				d.name, wait, lastErr, d.Output())
		case <-ticker.C:
		}
	}
}

// The wait error is nil when the process exits cleanly, and %w on nil prints
// %!w(<nil>) in place of the exit status.
func (d *Daemon) exitedError() error {
	return fmt.Errorf("daemon %s: exited before it was ready (%v)\n%s",
		d.name, d.waitErr, d.Output()) //nolint:errorlint // nil on a clean exit; %w prints %!w(<nil>)
}

func (d *Daemon) exitedAfterReadyError() error {
	return fmt.Errorf("daemon %s: became ready, then exited (%v)\n%s",
		d.name, d.waitErr, d.Output()) //nolint:errorlint // nil on a clean exit; %w prints %!w(<nil>)
}

// Stop terminates the process and waits for it to be reaped. Safe to call more
// than once and safe to call on a process that has already exited.
func (d *Daemon) Stop() error {
	d.stopOnce.Do(func() { d.stopErr = d.terminate() })
	return d.stopErr
}

func (d *Daemon) terminate() error {
	if d.cmd.Process == nil {
		return nil
	}
	select {
	case <-d.exited:
		return nil
	default:
	}

	if err := d.cmd.Process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("daemon %s: signal: %w", d.name, err)
	}

	select {
	case <-d.exited:
		return nil
	case <-time.After(stopGrace):
		_ = d.cmd.Process.Kill()
		<-d.exited
		return nil
	}
}

// Output returns everything the process has written to stdout and stderr.
func (d *Daemon) Output() string { return d.out.String() }

// syncBuffer collects process output, which is written by the reaper goroutine
// and read by assertions.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
