package gnolang

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	// Ignore pprof import, as the server does not
	// handle http requests if the user doesn't enable them
	// outright by using environment variables (starts serving)
	//nolint:gosec
	_ "net/http/pprof"
)

// NOTE: the golang compiler doesn't seem to be intelligent
// enough to remove steps when const debug is True,
// so it is still faster to first check the truth value
// before calling debug.Println or debug.Printf.

type Debugging struct {
	enabled bool
	derrors []string
}

func NewDebugging(d bool) *Debugging {
	if d {
		go func() {
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

	return &Debugging{
		enabled: d,
		derrors: nil,
	}
}

func (d *Debugging) DeepCopy() *Debugging {
	if d == nil {
		return nil
	}

	deers := make([]string, len(d.derrors))

	copy(deers, d.derrors)

	return &Debugging{
		enabled: d.enabled,
		derrors: deers,
	}
}

func (d *Debugging) Println(args ...interface{}) {
	if d != nil && d.enabled {
		fmt.Println(append([]interface{}{"DEBUG:"}, args...)...)
	}
}

func (d *Debugging) Printf(format string, args ...interface{}) {
	if d != nil && d.enabled {
		fmt.Printf("DEBUG: "+format, args...)
	}
}

// Instead of actually panic'ing, which messes with tests, errors are sometimes
// collected onto `var derrors`.  tests/file_test.go checks derrors after each
// test, and the file test fails if any unexpected debug errors were found.
func (d *Debugging) Errorf(format string, args ...interface{}) {
	if d != nil && d.enabled {
		d.derrors = append(d.derrors, fmt.Sprintf(format, args...))
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

// ----------------------------------------
// Exposed errors accessors
// File tests may access debug errors.
func (d *Debugging) HasDebugErrors() bool {
	return d != nil && len(d.derrors) > 0
}

func (d *Debugging) GetDebugErrors() []string {
	if d == nil {
		return nil
	}
	return d.derrors
}

func (d *Debugging) ClearDebugErrors() {
	if d != nil {
		d.derrors = nil
	}
}

func (d *Debugging) IsDebug() bool {
	return d != nil
}

func (d *Debugging) IsDebugEnabled() bool {
	return d != nil && d.enabled
}

func (d *Debugging) DisableDebug() {
	if d != nil {
		d.enabled = false
	}
}

func (d *Debugging) EnableDebug() {
	if d != nil {
		d.enabled = true
	}
}
