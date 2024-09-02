package bank

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// RouterKey is they name of the bank module
const RouterKey = ModuleName

// MsgSend - high level transaction of the coin module
type MsgSend struct {
	FromAddress crypto.Address `json:"from_address" yaml:"from_address"`
	ToAddress   crypto.Address `json:"to_address" yaml:"to_address"`
	Amount      std.Coins      `json:"amount" yaml:"amount"`
}

var _ std.Msg = MsgSend{}

// NewMsgSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgSend(fromAddr, toAddr crypto.Address, amount std.Coins) MsgSend {
	return MsgSend{FromAddress: fromAddr, ToAddress: toAddr, Amount: amount}
}

// Route Implements Msg.
func (msg MsgSend) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSend) Type() string { return "send" }

// ValidateBasic Implements Msg.
func (msg MsgSend) ValidateBasic() error {
	if msg.FromAddress.IsZero() {
		return std.ErrInvalidAddress("missing sender address")
	}
	if msg.ToAddress.IsZero() {
		return std.ErrInvalidAddress("missing recipient address")
	}
	if !msg.Amount.IsValid() {
		return std.ErrInvalidCoins(msg.Amount.String())
	}
	if !msg.Amount.IsAllPositive() {
		return std.ErrInsufficientCoins("send amount must be positive")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgSend) GetSigners() []crypto.Address {
	return []crypto.Address{msg.FromAddress}
}

// MsgMultiSend - high level transaction of the coin module
type MsgMultiSend struct {
	Inputs  []Input  `json:"inputs" yaml:"inputs"`
	Outputs []Output `json:"outputs" yaml:"outputs"`
}

var _ std.Msg = MsgMultiSend{}

// NewMsgMultiSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgMultiSend(in []Input, out []Output) MsgMultiSend {
	return MsgMultiSend{Inputs: in, Outputs: out}
}

// Route Implements Msg
func (msg MsgMultiSend) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgMultiSend) Type() string { return "multisend" }

// ValidateBasic Implements Msg.
func (msg MsgMultiSend) ValidateBasic() error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs()
	}
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs()
	}

	return ValidateInputsOutputs(msg.Inputs, msg.Outputs)
}

// GetSignBytes Implements Msg.
func (msg MsgMultiSend) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgMultiSend) GetSigners() []crypto.Address {
	addrs := make([]crypto.Address, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

// Input models transaction input
type Input struct {
	Address crypto.Address `json:"address" yaml:"address"`
	Coins   std.Coins      `json:"coins" yaml:"coins"`
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() error {
	if len(in.Address) == 0 {
		return std.ErrInvalidAddress(in.Address.String())
	}
	if !in.Coins.IsValid() {
		return std.ErrInvalidCoins(in.Coins.String())
	}
	if !in.Coins.IsAllPositive() {
		return std.ErrInvalidCoins(in.Coins.String())
	}
	return nil
}

// NewInput - create a transaction input, used with MsgMultiSend
func NewInput(addr crypto.Address, coins std.Coins) Input {
	return Input{
		Address: addr,
		Coins:   coins,
	}
}

// Output models transaction outputs
type Output struct {
	Address crypto.Address `json:"address" yaml:"address"`
	Coins   std.Coins      `json:"coins" yaml:"coins"`
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() error {
	if len(out.Address) == 0 {
		return std.ErrInvalidAddress(out.Address.String())
	}
	if !out.Coins.IsValid() {
		return std.ErrInvalidCoins(out.Coins.String())
	}
	if !out.Coins.IsAllPositive() {
		return std.ErrInvalidCoins(out.Coins.String())
	}
	return nil
}

// NewOutput - create a transaction output, used with MsgMultiSend
func NewOutput(addr crypto.Address, coins std.Coins) Output {
	return Output{
		Address: addr,
		Coins:   coins,
	}
}

// ValidateInputsOutputs validates that each respective input and output is
// valid and that the sum of inputs is equal to the sum of outputs.
func ValidateInputsOutputs(inputs []Input, outputs []Output) error {
	var totalIn, totalOut std.Coins

	for _, in := range inputs {
		if err := in.ValidateBasic(); err != nil {
			return errors.Wrap(err, "")
		}
		totalIn = totalIn.Add(in.Coins)
	}

	for _, out := range outputs {
		if err := out.ValidateBasic(); err != nil {
			return errors.Wrap(err, "")
		}
		totalOut = totalOut.Add(out.Coins)
	}

	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return ErrInputOutputMismatch()
	}

	return nil
}
