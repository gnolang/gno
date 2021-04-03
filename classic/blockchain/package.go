package blockchain

import (
	"reflect"

	"github.com/tendermint/go-amino-x/pkg"
)

var Package = pkg.NewPackage(
	"github.com/tendermint/classic/blockchain",
	"tm",
	pkg.GetCallersDirName(),
).WithDependencies().WithTypes(
	&bcBlockRequestMessage{}, "BlockRequest",
	&bcBlockResponseMessage{}, "BlockResponse",
	&bcNoBlockResponseMessage{}, "NoBlockResponse",
	&bcStausRequestMessage{}, "StatusRequest",
	&bcStausResponseMessage{}, "StatusResponse",
)
