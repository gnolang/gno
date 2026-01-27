//go:build !gnobench

package benchops

import "io"

// WritePprof is a no-op when profiling is disabled.
func (r *Results) WritePprof(w io.Writer) error {
	return nil
}
