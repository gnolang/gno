package gnolang

// CoverageTracker is an interface for tracking code coverage during VM execution.
// This interface allows decoupling the coverage implementation from the VM core.
type CoverageTracker interface {
	// TrackExecution records that a line in a file has been executed.
	TrackExecution(pkgPath, fileName string, line int)

	// TrackStatement records statement execution with additional context.
	TrackStatement(stmt Stmt)

	// TrackExpression records expression evaluation.
	TrackExpression(expr Expr)

	// IsEnabled returns whether coverage tracking is currently enabled.
	IsEnabled() bool

	// SetEnabled enables or disables coverage tracking.
	SetEnabled(enabled bool)
}

// NopCoverageTracker is a no-op implementation of CoverageTracker.
// Used when coverage is not enabled to avoid nil checks.
type NopCoverageTracker struct{}

func (n *NopCoverageTracker) TrackExecution(pkgPath, fileName string, line int) {}
func (n *NopCoverageTracker) TrackStatement(stmt Stmt)                          {}
func (n *NopCoverageTracker) TrackExpression(expr Expr)                         {}
func (n *NopCoverageTracker) IsEnabled() bool                                   { return false }
func (n *NopCoverageTracker) SetEnabled(enabled bool)                           {}

// DefaultCoverageTracker returns a default no-op coverage tracker.
func DefaultCoverageTracker() CoverageTracker {
	return &NopCoverageTracker{}
}
