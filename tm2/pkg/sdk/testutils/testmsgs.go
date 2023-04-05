package testutils

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	RouteMsgCounter  = "MsgCounter"
	RouteMsgCounter2 = "MsgCounter2"
)

// ValidateBasic() fails on negative counters.
// Otherwise it's up to the handlers
type MsgCounter struct {
	Counter       int64
	FailOnHandler bool
}

// Implements Msg
func (msg MsgCounter) Route() string                { return RouteMsgCounter }
func (msg MsgCounter) Type() string                 { return "counter1" }
func (msg MsgCounter) GetSignBytes() []byte         { return nil }
func (msg MsgCounter) GetSigners() []crypto.Address { return nil }
func (msg MsgCounter) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return std.ErrInvalidSequence("counter should be a non-negative integer.")
}

// a msg we dont know how to route
type MsgNoRoute struct {
	MsgCounter
}

func (tx MsgNoRoute) Route() string { return "noroute" }

// Another counter msg. Duplicate of MsgCounter
type MsgCounter2 struct {
	Counter int64
}

// Implements Msg
func (msg MsgCounter2) Route() string                { return RouteMsgCounter2 }
func (msg MsgCounter2) Type() string                 { return "counter2" }
func (msg MsgCounter2) GetSignBytes() []byte         { return nil }
func (msg MsgCounter2) GetSigners() []crypto.Address { return nil }
func (msg MsgCounter2) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return std.ErrInvalidSequence("counter should be a non-negative integer.")
}
