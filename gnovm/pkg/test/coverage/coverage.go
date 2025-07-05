package coverage

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

var globalTracker = NewTracker()

// InstrumentPackage instruments all non-test .gno files in a package
func InstrumentPackage(pkg *std.MemPackage) error {
	if pkg == nil {
		return fmt.Errorf("package is nil")
	}

	for _, file := range pkg.Files {
		if !shouldInstrumentFile(file.Name) {
			continue
		}

		engine := NewInstrumentationEngine(globalTracker, file.Name)
		instrumented, err := engine.InstrumentFile([]byte(file.Body))
		if err != nil {
			return fmt.Errorf("failed to instrument file %s: %w", file.Name, err)
		}
		file.Body = string(instrumented)
	}
	return nil
}

// shouldInstrumentFile determines if a file should be instrumented
func shouldInstrumentFile(filename string) bool {
	if strings.HasSuffix(filename, "_test.gno") || strings.HasSuffix(filename, "_filetest.gno") {
		return false
	}
	return strings.HasSuffix(filename, ".gno")
}

// GetGlobalTracker returns the global coverage tracker
func GetGlobalTracker() *Tracker {
	return globalTracker
}
