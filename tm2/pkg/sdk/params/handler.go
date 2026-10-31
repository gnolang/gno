package params

import (
	"fmt"
	"strings"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
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

// ----------------------------------------
// Query:
// - params/prefix:key for a prefixed module parameter key.
// - params/key for an arbitrary parameter key.
func (bh paramsHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	parts := strings.SplitN(req.Path, "/", 2)
	var path, rest string
	if len(parts) == 0 {
		// return helpful instructions.
	} else if len(parts) == 1 {
		path = parts[0]
		rest = ""
	} else {
		path = parts[0]
		rest = parts[1]
	}
	switch path {
	case "params":
		module, err := moduleOf(rest)
		if err != nil {
			res = sdk.ABCIResponseQueryFromError(err)
			return
		}
		if !bh.params.ModuleExists(module) {
			res = sdk.ABCIResponseQueryFromError(
				std.ErrUnknownRequest(fmt.Sprintf("module not registered: %q", module)))
			return
		}
		var val []byte
		bh.params.GetBytes(ctx, rest, &val)
		res.Height = req.Height
		res.Data = val
		return

	default:
		res = sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest(fmt.Sprintf("unknown params query endpoint %q", path)))
		return
	}
}

//----------------------------------------
// misc

func abciResult(err error) sdk.Result {
	return sdk.ABCIResultFromError(err)
}

// extracts the module portion of a key
// of the format <module>:<submodule>:<name>
func moduleOf(key string) (module string, err error) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) < 2 {
		return "", errors.New("expected param key format <module>:<rest>, but got %q", key)
	}
	module = parts[0]
	if len(module) == 0 {
		return "", errors.New("expected param key format <module>:<rest>, but got %q", key)
	}
	return module, nil
}
