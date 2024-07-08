package blockchain

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	btypes "github.com/gnolang/gno/tm2/pkg/bft/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bft/blockchain",
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
