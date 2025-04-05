// Package gnodebug provides utility functions and methods to debug GnoVM
// execution.
//
// There are two build tags that control the behaviour of this package, and thus
// the GnoVM: [Zealous] (and related tag "zealous"), and [Debug] (and related
// tag "debug").
//
// Zealous enables additional checks in the VM, which are often redundant and
// create slowdowns, but can help to spot programming errors in the VM.
//
// Debug can be used to make logs for the GnoVM. Debug.Printf calls fmt.Fprintf
// on [Output], but only prints it if the log is enabled explicitly through the
// GNODEBUG environment variable.
//
// The GNODEBUG environment variable is parsed through [ParseFlags]. It is a
// comma-separated key=value set of pairings, where the value may be omitted to
// set it to "1". Debug.Get, Debug.Enabled and Debug.Set modify the parsed
// flags.
package gnodebug

import (
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DebugType is the type of the "Debug" constant, mostly to add methods.
type DebugType bool

// DebugFlags contains the types of debug.
type DebugFlags map[string]string

// Output is the destination of the Printf functions. It can be changed if necessary.
var Output io.Writer = os.Stderr

func (d DebugFlags) Printf(flagName, format string, args ...any) {
	if flagName != "" && d[flagName] != "1" {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	_, file, line, _ := runtime.Caller(2)
	caller := fmt.Sprintf("%-.12s:%-4d", filepath.Base(file), line)
	fmt.Fprintf(
		Output,
		"%15s: %17s: "+format,
		append([]any{flagName, caller}, args...)...,
	)
}

// ParseFlag parses a set of debug flags.
// This is similar to Go's GODEBUG: comma separated key=value options, however
// boolean values without the =value are also allowed.
//
// For instance, "a,b,c=3" evaluates a=1, b=1 and c=3.
func ParseFlags(s string) DebugFlags {
	fl := DebugFlags{}
	// TODO: replace with strings.SplitSeq with upgrade to go 1.24
	for part := range splitSeq(s, ',') {
		if part == "" {
			continue
		}
		k, v, hasV := strings.Cut(part, "=")
		if !hasV {
			v = "1"
		}
		fl[k] = v
	}
	return fl
}

func splitSeq(s string, sep byte) iter.Seq[string] {
	return func(yield func(string) bool) {
		for {
			i := strings.IndexByte(s, sep)
			if i < 0 {
				break
			}
			frag := s[:i]
			if !yield(frag) {
				return
			}
			s = s[i+1:]
		}
		yield(s)
	}
}
