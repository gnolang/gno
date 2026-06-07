package gnolang

import (
	"fmt"
	"net/http"
	"path"
	"runtime"
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

// Build tags for zero-cost debug toggles:
//   -tags debug       → enables debug logging (debug.Printf/Println + pprof server)
//   -tags debugAssert → enables runtime invariant checks that panic on violation
// debugAssert sites currently exist in realm.go and store.go; remaining
// if debug { panic } sites across other files are candidates for migration.

type debugging bool

// using a const is probably faster.
// const debug debugging = true // or flip

func init() {
	if debug {
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
}

// runtime debugging flag.

var enabled bool = true

func (debugging) Println(args ...any) {
	if debug {
		if enabled {
			_, file, line, _ := runtime.Caller(2)
			caller := fmt.Sprintf("%-.12s:%-4d", path.Base(file), line)
			prefix := fmt.Sprintf("DEBUG: %17s: ", caller)
			fmt.Println(append([]any{prefix}, args...)...)
		}
	}
}

func (debugging) Printf(format string, args ...any) {
	if debug {
		if enabled {
			_, file, line, _ := runtime.Caller(2)
			caller := fmt.Sprintf("%.12s:%-4d", path.Base(file), line)
			prefix := fmt.Sprintf("DEBUG: %17s: ", caller)
			fmt.Printf(prefix+format, args...)
		}
	}
}

var derrors []string = nil

// Instead of actually panic'ing, which messes with tests, errors are sometimes
// collected onto `var derrors`.  tests/file_test.go checks derrors after each
// test, and the file test fails if any unexpected debug errors were found.
func (debugging) Errorf(format string, args ...any) {
	if debug {
		if enabled {
			derrors = append(derrors, fmt.Sprintf(format, args...))
		}
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
//
// Each frame is rendered as "<kind> <pkg-path>/<file>:<line:col-line:col>",
// not as the BlockNode's String() (which dumps the partially preprocessed
// AST including const-folded values, potentially megabytes per decl).
//
// Frames are walked from top (innermost = last pushed) downward. The
// walk stops AFTER printing the first frame whose location has a
// non-zero span. Earlier frames with zero spans (e.g., synthetic
// blocks) are still printed; later frames are dropped because the
// remaining lexical context is recoverable by reading the source at
// the printed location.
//
// Bounding by the first non-zero-span frame keeps output O(stack
// depth × ~70B) at worst (synthetic frames at the top are rare) and
// typically O(1) frame, regardless of how big the values being
// processed at panic time are.
func (p *PreprocessError) Stack() string {
	var stacktrace strings.Builder
	for i := len(p.stack) - 1; i >= 0; i-- {
		sbn := p.stack[i]
		loc := sbn.GetLocation()
		fmt.Fprintf(&stacktrace, "stack %d: %s %s\n", i, blockNodeKind(sbn), loc)
		if !loc.Span.IsZero() {
			break
		}
	}
	return stacktrace.String()
}

// blockNodeKind returns a short label for the BlockNode type. Used
// only by PreprocessError.Stack() to produce stack frames without
// dumping the BlockNode's contents.
func blockNodeKind(bn BlockNode) string {
	switch bn.(type) {
	case *FileNode:
		return "file"
	case *PackageNode:
		return "package"
	case *FuncDecl:
		return "func"
	case *FuncLitExpr:
		return "func-lit"
	case *BlockStmt:
		return "block"
	case *IfStmt:
		return "if"
	case *IfCaseStmt:
		return "if-case"
	case *ForStmt:
		return "for"
	case *RangeStmt:
		return "range"
	case *SwitchStmt:
		return "switch"
	case *SwitchClauseStmt:
		return "switch-case"
	case *SelectCaseStmt:
		return "select-case"
	default:
		return fmt.Sprintf("%T", bn)
	}
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

func HasDebugErrors() bool {
	return len(derrors) > 0
}

func GetDebugErrors() []string {
	return derrors
}

func ClearDebugErrors() {
	derrors = nil
}

func IsDebug() bool {
	return bool(debug)
}

func IsDebugEnabled() bool {
	return bool(debug) && enabled
}

func DisableDebug() {
	enabled = false
}

func EnableDebug() {
	enabled = true
}

func PrintCaller(start, end int) {
	for i := start; i < end; i++ {
		_, file, line, _ := runtime.Caller(i)
		caller := fmt.Sprintf("%-.12s:%-4d", path.Base(file), line)
		prefix := fmt.Sprintf("DEBUG: %17s: ", caller)
		fmt.Println(append([]any{prefix}, "")...)
	}
}
