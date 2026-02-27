package main

import (
	"fmt"
	"log/slog"
	"os"
)

// TestingT defines the interface we need for testing
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

// SlogTestingT implements TestingInterface using slog.Logger for structured logging
type SlogTestingT struct {
	logger   *slog.Logger
	cleanups func()
}

// NewSlogTestingT creates a new SlogTestingT with the given logger and name
func NewSlogTestingT(logger *slog.Logger) (t *SlogTestingT, cleanup func()) {
	if logger == nil {
		logger = slog.Default()
	}

	t = &SlogTestingT{
		logger: logger,
	}

	return t, func() { t.cleanups() }
}

func (s *SlogTestingT) Errorf(format string, args ...interface{}) {
	s.logger.Error(fmt.Sprintf(format, args...))
}

func (s *SlogTestingT) FailNow() {
	panic(fmt.Sprintf("Test failed"))
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
	old := s.cleanups
	s.cleanups = func() {
		fn()
		if old != nil {
			old()
		}
	}
}
