package coverage

import (
	"sync"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// Tracker implements the gnolang.CoverageTracker interface
// to track code execution during VM runtime.
type Tracker struct {
	enabled bool
	mu      sync.RWMutex

	// Package -> File -> Line -> Count
	coverage map[string]map[string]map[int]int64

	// Set of all executable lines for coverage calculation
	executableLines map[string]map[string]map[int]bool
}

// NewTracker creates a new VM coverage tracker.
func NewTracker() *Tracker {
	return &Tracker{
		enabled:         false,
		coverage:        make(map[string]map[string]map[int]int64),
		executableLines: make(map[string]map[string]map[int]bool),
	}
}

// TrackExecution records that a line in a file has been executed.
func (t *Tracker) TrackExecution(pkgPath, fileName string, line int) {
	if !t.enabled || line <= 0 {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.coverage[pkgPath] == nil {
		t.coverage[pkgPath] = make(map[string]map[int]int64)
	}
	if t.coverage[pkgPath][fileName] == nil {
		t.coverage[pkgPath][fileName] = make(map[int]int64)
	}

	t.coverage[pkgPath][fileName][line]++
}

// TrackStatement records statement execution with additional context.
func (t *Tracker) TrackStatement(stmt gnolang.Stmt) {
	if !t.enabled || stmt == nil {
		return
	}

	// For now, we can't get full location info from statements
	// The machine should call TrackExecution directly with package/file context
	// This is a no-op implementation to satisfy the interface
}

// TrackExpression records expression evaluation.
func (t *Tracker) TrackExpression(expr gnolang.Expr) {
	if !t.enabled || expr == nil {
		return
	}

	// For now, we can't get full location info from expressions
	// The machine should call TrackExecution directly with package/file context
	// This is a no-op implementation to satisfy the interface
}

// IsEnabled returns whether coverage tracking is currently enabled.
func (t *Tracker) IsEnabled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.enabled
}

// SetEnabled enables or disables coverage tracking.
func (t *Tracker) SetEnabled(enabled bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = enabled
}

// RegisterExecutableLine marks a line as executable for coverage calculation.
func (t *Tracker) RegisterExecutableLine(pkgPath, fileName string, line int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.executableLines[pkgPath] == nil {
		t.executableLines[pkgPath] = make(map[string]map[int]bool)
	}
	if t.executableLines[pkgPath][fileName] == nil {
		t.executableLines[pkgPath][fileName] = make(map[int]bool)
	}

	t.executableLines[pkgPath][fileName][line] = true
}

// GetCoverageData returns a copy of the current coverage data.
func (t *Tracker) GetCoverageData() map[string]map[string]map[int]int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Deep copy the coverage data
	result := make(map[string]map[string]map[int]int64)
	for pkg, files := range t.coverage {
		result[pkg] = make(map[string]map[int]int64)
		for file, lines := range files {
			result[pkg][file] = make(map[int]int64)
			for line, count := range lines {
				result[pkg][file][line] = count
			}
		}
	}

	return result
}

// GetExecutableLines returns a copy of the executable lines data.
func (t *Tracker) GetExecutableLines() map[string]map[string]map[int]bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Deep copy the executable lines data
	result := make(map[string]map[string]map[int]bool)
	for pkg, files := range t.executableLines {
		result[pkg] = make(map[string]map[int]bool)
		for file, lines := range files {
			result[pkg][file] = make(map[int]bool)
			for line := range lines {
				result[pkg][file][line] = true
			}
		}
	}

	return result
}

// Clear resets all coverage data.
func (t *Tracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.coverage = make(map[string]map[string]map[int]int64)
	// Note: we keep executable lines as they don't change
}

// SetCoverageData sets the coverage data directly.
// This is useful for loading cached coverage data.
func (t *Tracker) SetCoverageData(data map[string]map[string]map[int]int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.coverage = data
}

// SetExecutableLines sets the executable lines data directly.
// This is useful for loading cached coverage data.
func (t *Tracker) SetExecutableLines(data map[string]map[string]map[int]bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.executableLines = data
}
