package commondao

import (
	"errors"
	"time"
)

// executionKindName is the name of the arbitrary-execution kind.
const executionKindName = "execution"

var ErrExecutionFuncRequired = errors.New("execution proposal requires a non-nil Fn")

// defaultExecutionVotingPeriod is the voting period of execution proposals.
const defaultExecutionVotingPeriod = 7 * 24 * time.Hour

// ExecutionArgs are the args for the execution kind (ExecutionKind): a
// title, a body, and the ExecFunc to run on approval. The closure must be
// authored in a persistent realm so it survives Propose→Execute; a
// closure created by a `maketx run` script does not persist and cannot be
// executed later.
type ExecutionArgs struct {
	Title string
	Body  string
	Fn    ExecFunc
}

// ExecutionKind is the package's one concrete proposal kind: a stateless,
// reusable arbitrary-execution kind that runs an ExecFunc supplied by the
// proposer on approval. It is /p/-typed so any realm can register it with
// WithProposalKind(ExecutionKind{}) or RegisterKind without defining its
// own execution kind.
//
// The executor moves value only through the DAO-scoped sub the host
// passes (see ExecFunc): the host mints and passes that sub, so the
// executor receives whichever DAO's sub the host decides (its own DAO's by
// default). The closure is frozen at Propose (vote-integrity: the exact
// code is fixed before the vote).
//
// This kind applies NO policy check to the closure beyond a non-nil Fn: it
// runs the arbitrary code as-is. A realm that has treasury constraints (e.g.
// a freeze flag) should NOT catalog this kind directly; instead it should
// author its own execution kind whose definition wraps the closure with a
// Validable check that enforces those constraints (blocking execution while
// frozen, etc.), so arbitrary execution cannot bypass them. The reference
// realm gno.land/r/nt/commondao/v0 does exactly this.
type ExecutionKind struct{}

// Name returns the execution kind's registry name.
func (ExecutionKind) Name() string { return executionKindName }

// New validates ExecutionArgs and builds an execution definition. It is a
// pure factory: it receives only a readonly view and captures no mutable
// handle.
func (ExecutionKind) New(_ ReadonlyCommonDAO, args any) (ProposalDefinition, error) {
	a, ok := args.(ExecutionArgs)
	if !ok || a.Fn == nil {
		return nil, ErrExecutionFuncRequired
	}
	return executionDef{title: a.Title, body: a.Body, fn: a.Fn}, nil
}

// executionDef is the definition produced by ExecutionKind.
type executionDef struct {
	title string
	body  string
	fn    ExecFunc
}

func (d executionDef) Title() string             { return d.title }
func (d executionDef) Body() string              { return d.body }
func (executionDef) VotingPeriod() time.Duration { return defaultExecutionVotingPeriod }
func (executionDef) Threshold() Threshold        { return ThresholdSupermajority }
func (d executionDef) Executor() ExecFunc        { return d.fn }
