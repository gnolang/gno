package precompile

import (
	"fmt"
	"net/http"
	"os"
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

type debuggingPrecompile bool

// using a const is probably faster.
// const debug debugging = true // or flip
var debugPrecompile debuggingPrecompile = false

func init() {
	debugPrecompile = os.Getenv("DEBUG_PC") == "1"
	if debugPrecompile {
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

func (d debuggingPrecompile) Println(args ...interface{}) {
	if d {
		if enabled {
			fmt.Println(append([]interface{}{"DEBUG:"}, args...)...)
		}
	}
}

func (d debuggingPrecompile) Printf(format string, args ...interface{}) {
	if d {
		if enabled {
			fmt.Printf("DEBUG: "+format, args...)
		}
	}
}
