package daokit

import (
	"errors"

	"gno.land/p/nt/ufmt/v0"
)

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
