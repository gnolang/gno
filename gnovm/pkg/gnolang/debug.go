package gnolang

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/gnodebug"
)

const (
	zealous = gnodebug.Zealous
	dbg     = gnodebug.Debug
)

func init() {
	if dbg.Enabled("pprof") {
		go func() {
			// Start pprof server.
			// Note that inclusion of net/http/pprof is controlled by -tags debug,
			// see ./gnodebug/debug_enabled.go.

			// e.g.
			// curl -sK -v http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.out
			// curl -sK -v http://localhost:8080/debug/pprof/heap > heap.out
			// curl -sK -v http://localhost:8080/debug/pprof/allocs > allocs.out
			// see https://gist.github.com/slok/33dad1d0d0bae07977e6d32bcc010188.

			server := &http.Server{
				Addr:              "localhost:8080",
				ReadHeaderTimeout: 60 * time.Second,
			}
			server.ListenAndServe()
		}()
	}
}

// PreprocessError wraps a processing error along with its associated
// preprocessing stack for enhanced error reporting.
type PreprocessError struct {
	err   error
	stack []BlockNode
}

// Unwrap returns the encapsulated error message.
func (p *PreprocessError) Unwrap() error {
	return p.err
}

// Stack produces a string representation of the preprocessing stack
// trace that was associated with the error occurrence.
func (p *PreprocessError) Stack() string {
	var stacktrace strings.Builder
	for i := len(p.stack) - 1; i >= 0; i-- {
		sbn := p.stack[i]
		fmt.Fprintf(&stacktrace, "stack %d: %s\n", i, sbn.String())
	}
	return stacktrace.String()
}

// Error consolidates and returns the full error message, including
// the actual error followed by its associated preprocessing stack.
func (p *PreprocessError) Error() string {
	var err strings.Builder
	fmt.Fprintf(&err, "%s:\n", p.Unwrap())
	fmt.Fprintln(&err, "--- preprocess stack ---")
	fmt.Fprint(&err, p.Stack())
	fmt.Fprintf(&err, "------------------------")
	return err.String()
}
