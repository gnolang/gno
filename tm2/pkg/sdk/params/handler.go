package params

import (
	"errors"
	"fmt"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	errInvalidMsgType       = errors.New("unrecognized params message type")
	errUnknownQueryEndpoint = errors.New("unknown params query endpoint")
)

type paramsHandler struct {
	params Keeper
}

func NewHandler(params Keeper) paramsHandler {
	return paramsHandler{
		params: params,
	}
}

func (ph paramsHandler) Process(_ sdk.Context, msg std.Msg) sdk.Result {
	errMsg := fmt.Sprintf("%s: %T", errInvalidMsgType, msg)

	return sdk.ABCIResultFromError(std.ErrUnknownRequest(errMsg))
}

func (ph paramsHandler) Query(ctx sdk.Context, req abci.RequestQuery) abci.ResponseQuery {
	// Locate the first and second slashes
	firstSlash, secondSlash := findFirstTwoSlashes(req.Path)
	if firstSlash == -1 {
		return sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest(errUnknownQueryEndpoint.Error()),
		)
	}

	// Extract prefix and key directly from the path
	var (
		prefix = req.Path[firstSlash+1 : secondSlash]
		key    = req.Path[secondSlash+1:]
	)

	// Check prefix and key validity
	if prefix != ph.params.prefix || key == "" {
		return sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest(errUnknownQueryEndpoint.Error()),
		)
	}

	// Fetch the data and prepare the response
	var res abci.ResponseQuery

	res.Data = ph.params.GetRaw(ctx, key)

	return res
}

// findSlashPositions finds the positions of the first and second slashes in a path.
// Returns the positions of the slashes or (-1, -1) if less than two slashes are found.
func findFirstTwoSlashes(s string) (int, int) {
	first := -1

	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			if first == -1 {
				first = i
			} else {
				return first, i
			}
		}
	}
	return -1, -1
}
