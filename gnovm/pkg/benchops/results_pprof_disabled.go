//go:build !gnobench

package benchops

import "io"

// PprofOption configures pprof output generation.
// This is a no-op type when profiling is disabled.
type PprofOption func(*pprofConfig)

type pprofConfig struct{}

// WithDuration is a no-op when profiling is disabled.
func WithDuration() PprofOption { return func(*pprofConfig) {} }

// WithCount is a no-op when profiling is disabled.
func WithCount() PprofOption { return func(*pprofConfig) {} }

// WithLabels is a no-op when profiling is disabled.
func WithLabels() PprofOption { return func(*pprofConfig) {} }

// WritePprof is a no-op when profiling is disabled.
func (r *Results) WritePprof(w io.Writer) error {
	return nil
}

// WritePprofWithOptions is a no-op when profiling is disabled.
func (r *Results) WritePprofWithOptions(w io.Writer, opts ...PprofOption) error {
	return nil
}
