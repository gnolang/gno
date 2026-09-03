package bank

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/sdk/bank",
	"bank",
	amino.GetCallersDirname(),
).WithDependencies(
	abci.Package,
).WithTypes(
	NoInputsError{}, "NoInputsError",
	NoOutputsError{}, "NoOutputsError",
	InputOutputMismatchError{}, "InputOutputMismatchError",
	Input{}, "Input",
	Output{}, "Output",
	MsgSend{}, "MsgSend",
	MsgMultiSend{}, "MsgMultiSend",
	GenesisState{}, "GenesisState",
	Params{}, "Params",
	TransferEvent{}, "TransferEvent",
	MultiTransferEvent{}, "MultiTransferEvent",
))
