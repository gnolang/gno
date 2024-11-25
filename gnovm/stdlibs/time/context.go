package time

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type ExecContext struct {
	ChainTz string
}

// GetContext returns the execution context.
// This is used to allow extending the exec context using interfaces,
// for instance when testing.
func (e ExecContext) GetExecContext() ExecContext {
	return e
}

var _ ExecContexter = ExecContext{}

// ExecContexter is a type capable of returning the parent [ExecContext]. When
// using these standard libraries, m.Context should always implement this
// interface. This can be obtained by embedding [ExecContext].
type ExecContexter interface {
	GetExecContext() ExecContext
}

// NOTE: In order to make this work by simply embedding ExecContext in another
// context (like TestExecContext), the method needs to be something other than
// the field name.

// GetContext returns the context from the Gno machine.
func GetContext(m *gno.Machine) ExecContext {
	return m.Context.(ExecContexter).GetExecContext()
}
