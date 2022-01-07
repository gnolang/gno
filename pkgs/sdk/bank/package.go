package bank

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/sdk/bank",
	"bank",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	NoInputsError{}, "NoInputsError",
	NoOutputsError{}, "NoOutputsError",
	InputOutputMismatchError{}, "InputOutputMismatchError",
	MsgSend{}, "MsgSend",
))
