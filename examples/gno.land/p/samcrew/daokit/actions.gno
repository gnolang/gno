package daokit

import (
	"errors"

	"gno.land/p/moul/md"
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

const ActionInstantExecuteKind = "gno.land/p/samcrew/daokit.InstantExecute"

type actionInstantExecute struct {
	dao DAO
	req ProposalRequest
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
