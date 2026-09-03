package cluster

import (
	"fmt"
	"log/slog"
	"os"
)

// TestingT defines the interface for cluster orchestration (building binaries,
// starting nodes, etc.). SlogTestingT.FailNow panics with FailNowPanic --
// callers must recover it. This is distinct from testscript.T which uses
// runtime.Goexit().
type TestingT interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	TempDir() string
	Cleanup(func())
	FailNow()
	Helper()
}

// SlogTestingT implements TestingT using slog.Logger for structured logging.
// It is NOT safe for concurrent use.
type SlogTestingT struct {
	logger   *slog.Logger
	cleanups []func()
}

// NewSlogTestingT creates a new SlogTestingT with the given logger and name
func NewSlogTestingT(logger *slog.Logger) (t *SlogTestingT, cleanup func()) {
	if logger == nil {
		logger = slog.Default()
	}

	t = &SlogTestingT{
		logger: logger,
	}

	return t, t.runCleanups
}

// runCleanups executes all registered cleanup functions in reverse order.
// Each cleanup runs independently; a panic in one does not prevent others.
func (s *SlogTestingT) runCleanups() {
	for i := len(s.cleanups) - 1; i >= 0; i-- {
		func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("cleanup panic", "error", r)
				}
			}()
			s.cleanups[i]()
		}()
	}
}

func (s *SlogTestingT) Errorf(format string, args ...interface{}) {
	s.logger.Error(fmt.Sprintf(format, args...))
}

// FailNowPanic is a sentinel type used by SlogTestingT.FailNow
// to distinguish intentional test failures from real panics.
type FailNowPanic struct{}

func (s *SlogTestingT) FailNow() {
	panic(FailNowPanic{})
}

func (s *SlogTestingT) Helper() {}

func (s *SlogTestingT) Log(args ...interface{}) {
	s.logger.Info(fmt.Sprint(args...))
}

func (s *SlogTestingT) Logf(format string, args ...interface{}) {
	s.logger.Info(fmt.Sprintf(format, args...))
}

func (s *SlogTestingT) Fatalf(format string, args ...interface{}) {
	s.logger.Error(fmt.Sprintf(format, args...))
	s.FailNow()
}

func (s *SlogTestingT) TempDir() string {
	tempDir, err := os.MkdirTemp("", "test*")
	if err != nil {
		s.Fatalf("failed to create temp dir: %s", err)
	}
	s.Cleanup(func() { os.RemoveAll(tempDir) })
	return tempDir
}

func (s *SlogTestingT) Cleanup(fn func()) {
	s.cleanups = append(s.cleanups, fn)
}

// RecoverFailNow runs fn and recovers from FailNowPanic, returning an error.
// Real panics are re-raised.
func RecoverFailNow(fn func()) error {
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(FailNowPanic); ok {
					err = fmt.Errorf("operation failed (see logs)")
				} else {
					panic(r)
				}
			}
		}()
		fn()
	}()
	return err
}
