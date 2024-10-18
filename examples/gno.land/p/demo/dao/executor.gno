package dao

// Executor represents a minimal closure-oriented proposal design.
// It is intended to be used by a govdao governance proposal (v1, v2, etc)
type Executor interface {
	// Execute executes the given proposal, and returns any error encountered
	// during the execution
	Execute() error
}
