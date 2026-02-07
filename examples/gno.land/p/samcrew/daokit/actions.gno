package daokit

// Defines executable operations that proposals can perform.
// Each action has a type and handler functions that implements the actual logic.
//
// Example: Adding a member, transferring funds, or updating configuration.

import (
	"errors"

	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt"
)

// Interface storing Action's data.
type Action interface {
	// Returns human-readable description of the action.
	String() string
	// Returns unique identifier for the action type.
	Type() string
}

// Interface storing executable function that processes Action's data.
type ActionHandler interface {
	// Executes the given action's logic.
	Execute(action Action)
	// Returns the action type this handler processes.
	Type() string
}

// Generic action implementation that stores a type and a payload.
// Payload contains the data parameters that will be processed by the handler.
// Example: {kind="/p/demo/transfer.Transfer", payload={amount: 1000, recipient: "user123"}}
type genericAction struct {
	kind    string
	payload interface{}
}

// Generic action handler that executes a function with a given payload.
// Example: A transfer handler that processes {amount, recipient} to move funds.
// {kind="transfer", executor=transferFunds}
type genericActionHandler struct {
	kind     string
	executor func(payload interface{})
}

func NewAction(kind string, payload interface{}) Action {
	return &genericAction{kind: kind, payload: payload}
}

// Returns string representation of the action's payload.
func (g *genericAction) String() string {
	return ufmt.Sprintf("%v", g.payload)
}

// Returns the action's type identifier.
func (g *genericAction) Type() string {
	return g.kind
}

func NewActionHandler(kind string, executor func(interface{})) ActionHandler {
	return &genericActionHandler{kind: kind, executor: executor}
}

// Executes the action by calling the executor function with the payload.
func (g *genericActionHandler) Execute(iaction Action) {
	action, ok := iaction.(*genericAction)
	if !ok {
		panic(errors.New("invalid action type"))
	}
	g.executor(action.payload)
}

// Returns the action's type identifier.
func (g *genericActionHandler) Type() string {
	return g.kind
}

// Creates a new empty action instance.
func (g *genericActionHandler) Instantiate() Action {
	return &genericAction{
		kind: g.kind,
	}
}

// //////////////////////////////////////////////////////////
// ActionExecuteLambdaKind
//
// Built-in action type for executing lambda functions.
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

// //////////////////////////////////////////////////////////
// ActionInstantExecuteKind
//
// Built-in action type for instantly executing sub-proposals.
const ActionInstantExecuteKind = "gno.land/p/samcrew/daokit.InstantExecute"

type actionInstantExecute struct {
	dao DAO             // Target DAO to execute on
	req ProposalRequest // Proposal to execute instantly
}

func (a *actionInstantExecute) String() string {
	// XXX: find a way to be explicit about the subdao
	s := ""
	s += md.Paragraph(md.Blockquote(a.req.Action.Type()))
	s += md.Paragraph(a.req.Action.String())
	return s
}

func NewInstantExecuteHandler() ActionHandler {
	return NewActionHandler(ActionInstantExecuteKind, func(i interface{}) {
		action, ok := i.(*actionInstantExecute)
		if !ok {
			panic(errors.New("invalid action type"))
		}
		InstantExecute(action.dao, action.req)
	})
}

func NewInstantExecuteAction(dao DAO, req ProposalRequest) Action {
	return NewAction(ActionInstantExecuteKind, &actionInstantExecute{dao: dao, req: req})
}
