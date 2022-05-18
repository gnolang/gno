package blockchain

import (
	"github.com/gnolang/gno/pkgs/amino"
	btypes "github.com/gnolang/gno/pkgs/bft/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/bft/blockchain",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies(
	btypes.Package,
).WithTypes(
	&bcBlockRequestMessage{}, "BlockRequest",
	&bcBlockResponseMessage{}, "BlockResponse",
	&bcNoBlockResponseMessage{}, "NoBlockResponse",
	&bcStatusRequestMessage{}, "StatusRequest",
	&bcStatusResponseMessage{}, "StatusResponse",
))
