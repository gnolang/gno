package types

import (
	"std"

	"gno.land/p/demo/ufmt"
)

type Proposal struct {
	Author std.Address

	Title       string
	Description string

	Executor *Executor
}

func (p *Proposal) String() string {
	return ufmt.Sprintf(`
Title: %s

Proposed by: %s

%s

This proposal contains the following metadata:

%s`, p.Title, p.Author, p.Description, p.Executor.String())
}

// Executor is in charge of execute a proposal when it is approved by the members.
type Executor struct {
	metadata   Metadata
	callbackFn func(m Metadata) error
}

func (e *Executor) String() string {
	if e.metadata == nil {
		return "No metadata available."
	}

	return e.metadata.String()
}

func NewExecutor(fn func(m Metadata) error, metadata Metadata) *Executor {
	return &Executor{
		metadata:   metadata,
		callbackFn: fn,
	}
}

// Metadata contains any kind of information that will be stored on the executor
// and will be sent to the callback function when the proposal is validated.
// It can contains any kind of data.
type Metadata interface {
	IsMetadata()
	String() string
}

