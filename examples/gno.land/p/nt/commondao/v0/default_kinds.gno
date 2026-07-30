package commondao

import (
	"errors"
	"time"
)

// Reserved kind names for the /p/ default kinds. The register-kind name
// is reserved: DeregisterKind refuses it (self-brick guard) and the
// register-kind factory rejects re-registering it.
const (
	// executionKindName is the name of the arbitrary-execution kind.
	executionKindName = "execution"

	// registerKindName is the name of the register-kind kind, which
	// registers a new proposal kind on a DAO through governance.
	registerKindName = "register-kind"
)

var (
	ErrExecutionFuncRequired   = errors.New("execution proposal requires a non-nil Fn")
	ErrRegisterKindRequired    = errors.New("register-kind proposal requires a non-nil Kind with a name")
	ErrRegisterKindDAORequired = errors.New("register-kind proposal requires the target DAO")
)

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

// RegisterKindArgs are the args for the register-kind kind
// (RegisterKindKind): the target DAO and the ProposalKind to register on
// it when the proposal executes. Only a trusted caller can populate DAO
// (an external proposer cannot obtain a *CommonDAO), so this is the
// intended path for governance-driven registration.
type RegisterKindArgs struct {
	DAO  *CommonDAO
	Kind ProposalKind
}

// ExecutionKind is a default, stateless proposal kind that runs an
// arbitrary ExecFunc supplied by the proposer on approval. It is /p/-typed
// so any realm can seed it (WithDefaultKinds or RegisterKind) without
// defining its own execution kind.
//
// The executor moves value only through the DAO-scoped sub the host
// passes (see ExecFunc); ExecutionKind is not Funded, so it always
// receives its own DAO's sub. The closure is frozen at Propose
// (vote-integrity: the exact code is fixed before the vote).
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

// RegisterKindKind is a default, stateless proposal kind for
// governance-driven registration of a new proposal kind on a DAO. It is
// registered by default via WithDefaultKinds and its reserved name cannot
// be deregistered (self-brick guard in DeregisterKind).
//
// The target DAO arrives through RegisterKindArgs.DAO, which only a
// trusted caller can populate; the definition captures it and calls
// RegisterKind in its Executor, so registration happens only after the
// vote passes.
type RegisterKindKind struct{}

// Name returns the register-kind kind's reserved registry name.
func (RegisterKindKind) Name() string { return registerKindName }

// New validates RegisterKindArgs and builds a register-kind definition.
// It pre-checks that the DAO does not already have the kind so a doomed
// proposal is rejected up front; the executor re-checks via RegisterKind.
// The readonly view is unused: the trusted target handle comes via args.
func (RegisterKindKind) New(_ ReadonlyCommonDAO, args any) (ProposalDefinition, error) {
	a, ok := args.(RegisterKindArgs)
	if !ok || a.DAO == nil {
		return nil, ErrRegisterKindDAORequired
	}
	if a.Kind == nil || a.Kind.Name() == "" {
		return nil, ErrRegisterKindRequired
	}
	if a.DAO.HasKind(a.Kind.Name()) {
		return nil, ErrProposalKindExists
	}
	return registerKindDef{dao: a.DAO, kind: a.Kind}, nil
}

// registerKindDef is the definition produced by RegisterKindKind.
type registerKindDef struct {
	dao  *CommonDAO
	kind ProposalKind
}

func (registerKindDef) Title() string               { return "Register proposal kind" }
func (d registerKindDef) Body() string              { return d.kind.Name() }
func (registerKindDef) VotingPeriod() time.Duration { return defaultExecutionVotingPeriod }
func (registerKindDef) Threshold() Threshold        { return ThresholdSupermajority }

// Executor registers the captured kind on the captured DAO. The DAO came
// from a trusted caller via args, so this is the intended args-capture
// mutation path (the readonly view New received is unrelated).
func (d registerKindDef) Executor() ExecFunc {
	return func(_ int, _ realm) error {
		return d.dao.RegisterKind(d.kind)
	}
}
