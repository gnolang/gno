package params

import (
	"fmt"
	"strings"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type paramsHandler struct {
	params ParamsKeeper
}

func NewHandler(params ParamsKeeper) paramsHandler {
	return paramsHandler{
		params: params,
	}
}

func (bh paramsHandler) Process(ctx sdk.Context, msg std.Msg) sdk.Result {
	errMsg := fmt.Sprintf("unrecognized params message type: %T", msg)
	return abciResult(std.ErrUnknownRequest(errMsg))
}

//----------------------------------------
// Query

func (bh paramsHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	prefix := secondPart(req.Path)
	if bh.params.PrefixExists(prefix) {
		return bh.queryParam(ctx, req)
	}
	res = sdk.ABCIResponseQueryFromError(
		std.ErrUnknownRequest("unknown params query endpoint"))
	return
}

// queryParam returns param for a key.
func (bh paramsHandler) queryParam(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	// parse key from path.
	key := thirdPartWithSlashes(req.Path)
	if key == "" {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest("param key is empty"))
	}

	// XXX: validate?

	val := bh.params.GetRaw(ctx, key)

	res.Data = val
	return
}

//----------------------------------------
// misc

func abciResult(err error) sdk.Result {
	return sdk.ABCIResultFromError(err)
}

// returns the second component of a path.
func secondPart(path string) string {
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 {
		return ""
	} else {
		return parts[1]
	}
}

// returns the third component of a path, including other slashes.
func thirdPartWithSlashes(path string) string {
	split := strings.SplitN(path, "/", 3)
	return split[2]
}
