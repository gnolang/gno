package blockchain

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/bft/blockchain",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	&bcBlockRequestMessage{}, "BlockRequest",
	&bcBlockResponseMessage{}, "BlockResponse",
	&bcNoBlockResponseMessage{}, "NoBlockResponse",
	&bcStatusRequestMessage{}, "StatusRequest",
	&bcStatusResponseMessage{}, "StatusResponse",
))
