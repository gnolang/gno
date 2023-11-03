package book

import "std"

//----------------------------------------
// Action

// A actor (subject) acts upon an object with a verb and arguments.
// As if actor is calling object.verb(args...).
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

func (a Action) setPanic() {
	a.status = ActionStatusPanic
}

type ActionStatus int

const (
	ActionStatusNew = iota
	ActionStatusPending
	ActionStatusComplete
	ActionStatusPanic
)

//----------------------------------------
// Actor

type Actor interface {
	std.Addressable
}

//----------------------------------------
// Object

type Object interface {
	std.Addressable
	// Receive performs action on self.
	// Panics if action is not permitted.
	// NOTE: no return value, like a one way channel.
	ReceiveFrom(*Action, Actor)
}

//----------------------------------------
// Main entry methods

// XXX Well maybe it should be a method on a directory.
// XXX otherwise how do we know if actor/object are true.
func Perform(action Action, actor Actor, object Object) {
	// Defer to set panic upon panic.
	defer func() {
		if r := recover(); r != nil {
			action.setPanic()
			panic(r)
		}
	}()

	// Check actor matches.
	// XXX well but maybe the actor is fake
	// Check object matches.
	// XXX well but maybe the object is fake

	// XXX ensure that the actor has authorized
	// XXX if actor wants signature verification
	// XXX if actor has a privactionbook

	// Make action pending.
	// This is the only place where this happens.
	if action.Status() != ActionStatusNew {
		panic("action not new")
	} else {
		action.setPending()
	}
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
