package gno

import (
	"fmt"
	"net/http"

	_ "net/http/pprof"
)

// NOTE: the golang compiler doesn't seem to be intelligent
// enough to remove steps when const debug is True,
// so it is still faster to first check the truth value
// before calling debug.Println or debug.Printf.

const debug debugging = false // or flip

func init() {
	if debug {
		go func() {
			// e.g.
			// curl -sK -v http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.out
			// curl -sK -v http://localhost:8080/debug/pprof/heap > heap.out
			// curl -sK -v http://localhost:8080/debug/pprof/allocs > allocs.out
			// see https://gist.github.com/slok/33dad1d0d0bae07977e6d32bcc010188.
			http.ListenAndServe("localhost:8080", nil)
		}()
	}
}

func Debug() bool {
	return bool(debug)
}

type debugging bool

var enabled bool = true

func (d debugging) Disable() {
	enabled = false
}

func (d debugging) Enable() {
	enabled = true
}

func (d debugging) Println(args ...interface{}) {
	if d {
		if enabled {
			fmt.Println(append([]interface{}{"DEBUG:"}, args...)...)
		}
	}
}

func (d debugging) Printf(format string, args ...interface{}) {
	if d {
		if enabled {
			fmt.Printf("DEBUG: "+format, args...)
		}
	}
}

var derrors []string = nil

// Instead of actually panic'ing, which messes with tests, errors are sometimes
// collected onto `var derrors`.  tests/file_test.go checks derrors after each
// test, and the file test fails if any unexpected debug errrors were found.
func (d debugging) Errorf(format string, args ...interface{}) {
	if d {
		if enabled {
			derrors = append(derrors, fmt.Sprintf(format, args...))
		}
	}
}

//----------------------------------------
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
