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

func (d DebugFlags) Printf(kind, format string, args ...any) {
	if d[kind] != "1" {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	_, file, line, _ := runtime.Caller(2)
	caller := fmt.Sprintf("%-.12s:%-4d", filepath.Base(file), line)
	fmt.Fprintf(
		Output,
		"DEBUG[%s]: %17s: "+format,
		append([]any{kind, caller}, args...)...,
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
