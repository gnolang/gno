package testing

import (
	"log"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test/coverage"
)

// coverageTracker is the global coverage tracker
var coverageTracker *coverage.CoverageTracker

func init() {
	coverageTracker = coverage.NewCoverageTracker()
}

func X_unixNano() int64 {
	// only implemented in testing stdlibs
	return 0
}

func X_matchString(pat, str string) (bool, string) {
	panic("only implemented in testing stdlibs")
}

func X_recoverWithStacktrace() (gnolang.TypedValue, string) {
	panic("only available in testing stdlibs")
}

// X_markLine marks a line as executed
func X_markLine(filename string, line int) {
	coverageTracker.MarkLine(filename, line)
}

// X_instrumentCode instruments the code for coverage
func X_instrumentCode(code string, filename string) string {
	instrumenter := coverage.NewCoverageInstrumenter(coverageTracker, filename)
	instrumented, err := instrumenter.InstrumentFile([]byte(code))
	if err != nil {
		log.Printf("Failed to instrument code: %v\n", err)
		return code
	}
	return string(instrumented)
}
