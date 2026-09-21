package daokit

import (
	"errors"

	"gno.land/p/nt/ufmt/v0"
)

// SECURITY XXX: Action, ActionHandler, and the daocond.Condition /
// daocond.Ballot interfaces are open extension points. Whatever you
// register via Core.Resources.Set runs WITH THE DAO'S AUTHORITY when
// a proposal executes — same level of trust as code inside the DAO
// package itself. A hostile handler can do anything the DAO can do
// (modify members, transfer treasury funds via an action that calls
// a token, etc.). The owning realm is responsible for only registering
// handlers/conditions from trusted code.
//
// Related re-entrancy risk: Core.Execute calls Condition.Eval before
// running Action. A malicious Condition.Eval can call back into the
// same DAO (Propose / Vote / Execute on another proposal) during the
// same crossing frame. IsCurrent still passes because we're inside
// the original caller's frame, so the re-entrant call runs as the
// original caller. Same shape as Ethereum's DAO hack.
// Fix: add a re-entrancy guard in Core.Execute — a package-level lock
// or per-proposal in-flight flag — set before any user-supplied code
// (Eval, Action.String, handler Execute) runs and cleared on return.
type Action interface {
	String() string
	Type() string
}

type ActionHandler interface {
	Execute(action Action)
	Type() string
}

func NewAction(kind string, payload interface{}) Action {
	return &genericAction{kind: kind, payload: payload}
}

type genericAction struct {
	kind    string
	payload interface{}
}

// String implements Action.
func (g *genericAction) String() string {
	return ufmt.Sprintf("%v", g.payload)
}

// Type implements Action.
func (g *genericAction) Type() string {
	return g.kind
}

func NewActionHandler(kind string, executor func(interface{})) ActionHandler {
	return &genericActionHandler{kind: kind, executor: executor}
}

type genericActionHandler struct {
	kind     string
	executor func(payload interface{})
}

// Execute implements ActionHandler.
func (g *genericActionHandler) Execute(iaction Action) {
	action, ok := iaction.(*genericAction)
	if !ok {
		panic(errors.New("invalid action type"))
	}
	g.executor(action.payload)
}

// Instantiate implements ActionHandler.
func (g *genericActionHandler) Instantiate() Action {
	return &genericAction{
		kind: g.kind,
	}
}

// Type implements ActionHandler.
func (g *genericActionHandler) Type() string {
	return g.kind
}

const ActionExecuteLambdaKind = "gno.land/p/samcrew/daokit.ExecuteLambda"

func NewExecuteLambdaHandler() ActionHandler {
	return NewActionHandler(ActionExecuteLambdaKind, func(i interface{}) {
		cb, ok := i.(func())
		if !ok {
			panic(errors.New("invalid action type"))
		}
		cb()
	})
}

func NewExecuteLambdaAction(cb func()) Action {
	return NewAction(ActionExecuteLambdaKind, cb)
}

// InstantExecute-as-an-action (NewInstantExecuteHandler /
// NewInstantExecuteAction) was removed during the rlm-based auth
// migration: the action-handler callback shape `func(i interface{})`
// has no rlm in scope, so it cannot drive Propose/Vote/Execute which
// now all require an IsCurrent-validated rlm. The free function
// InstantExecute(_ int, rlm realm, d DAO, req) remains for direct
// caller use.
