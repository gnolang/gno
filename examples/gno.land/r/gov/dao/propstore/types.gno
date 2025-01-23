package propstore

import (
	"errors"
	"std"

	"gno.land/r/gov/dao/ruler"
)

type DAOCallback interface {
	// BeforeNewProposal is called before a Proposal is created, allowing to the actual 
	// GovDAO implementation the opportunity of validate it.
	BeforeNewProposal(caller std.Address,title string, meta Metadata) error

	// AfterNewProposal is called after the creation of a new Proposal. 
	// It gives information about the proposal ID, so it can be paired with 
	// the GovDAO proposal voting status, also allowing future queries to get
	// the Proposal object.
	AfterNewProposal(caller std.Address,title string, meta Metadata, pid int)
}

type Proposal struct {
	Author std.Address

	Title       string
	Description string

	Executor *executor

	RulerOnCreation string
}

// TODO it should depends on the GovDAO
// func (p *Proposal) String() string {
// 	return ufmt.Sprintf(`
// Title: %s

// Proposed by: %s

// Proposed by GovDAO: %s

// %s

// This proposal contains the following metadata:

// %s`, p.Title, p.Author, p.RulerOnCreation, p.Description, p.Executor.String())
// }

// Executor is in charge of execute a proposal when it is approved by the members.
type executor struct {
	metadata   Metadata
	callbackFn func(m Metadata) error
}

func (e *executor) String() string {
	if e.metadata == nil {
		return "No metadata available."
	}

	return e.metadata.String()
}

func newExecutor(fn func(m Metadata) error, metadata Metadata) *executor {
	return &Executor{
		metadata:   metadata,
		callbackFn: securityWrapper(fn),
	}
}

func securityWrapper(fn func(m Metadata) error) func(m Metadata) error {
	return func (m Metadata) error {
		if std.PrevRealm().PkgPath() != ruler.CurrentDAO {
			return ErrInvalidExecutorCaller
		}

		return fn(m)
	}
}

var ErrInvalidExecutorCaller error = errors.New("the realm calling the proposal executor was not actually in charge")

// Metadata contains any kind of information that will be stored on the executor
// and will be sent to the callback function when the proposal is validated.
// It can contains any kind of data.
type Metadata interface {
	IsMetadata()
	String() string
}
