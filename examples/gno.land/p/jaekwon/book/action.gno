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
	sequence string      // typically an incrementing number
	actor    std.Address // actor (subject) id
	object   std.Address // object id
	verb     string      // verb name
	args     []string    // legible args
}

func (a Action) Sequence() string    { return a.sequence }
func (a Action) Actor() std.Address  { return a.actor }
func (a Action) Object() std.Address { return a.object }
func (a Action) Verb() string        { return a.verb }
func (a Action) NumArgs() int        { return len(a.args) }
func (a action) Arg(n int) string    { return a.args[n] }

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
