package txindexer

import "context"

var _ Manager = (*NoOp)(nil)

// NoOp is a no-operation implementation of the tx-indexer manager.
// // It is used when tx-indexer is not enabled.
type NoOp struct{}

// NewNoOp creates a new NoOp instance of a tx-indexer manager.
func NewNoOp() *NoOp { return new(NoOp) }

func (n *NoOp) Start(_ context.Context) error  { return nil }
func (n *NoOp) Stop(_ context.Context) error   { return nil }
func (n *NoOp) Reload(_ context.Context) error { return nil }
