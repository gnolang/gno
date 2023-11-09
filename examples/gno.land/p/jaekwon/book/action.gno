package book

import (
	"errors"
	"std"
)

//----------------------------------------
// Action

// A actor (subject) acts upon an object with a verb and
// arguments.  As if actor is calling object.verb(args...).
//
// The sequence is usually an incrementing number,
// but could be something else tracked by an action book.
//
// The actor and object are denoted by their respective
// paths (or URIs?).
type Action struct {
	sequence string       // typically an incrementing number
	actor    std.Address  // actor (subject) id
	object   std.Address  // object id
	verb     string       // verb name
	args     []string     // legible args
	status   ActionStatus // new, pending, etc
}

func (a Action) Sequence() string     { return a.sequence }
func (a Action) Actor() std.Address   { return a.actor }
func (a Action) Object() std.Address  { return a.object }
func (a Action) Verb() string         { return a.verb }
func (a Action) NumArgs() int         { return len(a.args) }
func (a action) Arg(n int) string     { return a.args[n] }
func (a Action) Status() ActionStatus { return a.status }

func (a Action) setPending() {
	if a.status != ActionStatusNew {
		panic("should not happen")
	}
	a.status = ActionStatusPending
}

func (a Action) setComplete() {
	if a.status != ActionStatusPending {
		panic("should not happen")
	}
	a.status = ActionStatusComplete
}

func (a Action) setError() {
	a.status = ActionStatusError
}

type ActionStatus int

const (
	ActionStatusNew = iota
	ActionStatusPending
	ActionStatusReceived
	ActionStatusComplete
	ActionStatusError
)

//----------------------------------------
// Actor

type Actor interface {
	std.Addressable
	// CheckAction verifies that action is
	// valid as the next action for Actor.
	// It should have no side effect.
	CheckAction(action) error
}

//----------------------------------------
// Object

type Object interface {
	std.Addressable
	// Receive performs action on self.
	// Returns an error if action is not permitted.
	ReceiveFrom(*Action, Actor) error
}

//----------------------------------------
// Main entry methods

// Perform the action on the actor/object in directory.
// How an action is authenticated is determined by the
// implementation of the actor; one could use cryptographic
// private key to authenticate an action, or, one could use
// a object (capabilities) private key, or any other method.
//
// In the implementations provided here, the action is
// first either added to a CryptoActionBook or a
// PrivateActionBook, each backed by an ActionBook (but an
// actor is not required to have an ActionBook).
//
// Case 1 w/ signatures
//  1. caller adds Action to CryptoActionBook.
//  2. CryptoActionBook checks signature & sequence.
//  3. CryptoActionBook calls Perform().
//  4. Perform asks Actor to verify Action.
//  5. Actor's ActionBook says it is good.
//  6. Perform makes object receive action.
//
// Case 2 w/o signatures
//  1. caller adds Action to PrivActionBook.
//  2. PrivActionBook checks sequence.
//  3. PrivActionBook calls Perform().
//  4. Perform asks Actor to verify Action.
//  5. Actor's ActionBook says it is good.
//  6. Perform makes object receive action.
func Perform(dir Directory, action Action) (err error) {
	// Defer to set panic upon panic.
	defer func() {
		if r := recover(); r != nil {
			action.setError()
			panic(r)
		} else if err != nil {
			action.setError()
		}
	}()

	// Validate what can be validated.
	err = action.ValidateBasic()
	if err != nil {
		return
	}
	// Check that status is new.
	if action.Status() != ActionStatusNew {
		return errors.New("expected a new action")
	}
	// Get actor and object.
	actor := dir.Get(action.actor)
	object := dir.Get(action.object)
	// Ask Actor to verify action.
	err = actor.CheckAction(action)
	if err != nil {
		return
	}
	// Set status as pending.
	action.setPending()
	// Let object handle action.
	// (action.status == pending)
	err = object.ReceiveFrom(action, actor)
	if err != nil {
		return
	}
	// (action.status == pending|received)
	// Set status as complete.
	action.setComplete()
}

//----------------------------------------
// Authorization

// Authorization struct is only needed for Actions that
// require cryptographic authorization, where the Action's
// actor has a pubkey to verify signatures with.
//
// Presumably once Authorization is validated (signature
// checked) the Action becomes committed, and given a index
// number.
type Authorization struct {
	action     Action
	signatures []Signature
}

type Signature struct {
	account   number // or address with some extra data unknown
	sequence  number // or alternative to sequence
	signature []byte
}
