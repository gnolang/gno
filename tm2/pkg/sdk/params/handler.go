package params

import (
	"fmt"
	"strings"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"

	"errors"
)

var (
	errInvalidMsgType       = errors.New("unrecognized params message type")
	errUnknownQueryEndpoint = errors.New("unknown params query endpoint")
)

type paramsHandler struct {
	params ParamsKeeper
}

func NewHandler(params ParamsKeeper) paramsHandler {
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
	firstSlash, secondSlash, found := findSlashPositions(req.Path)
	if !found {
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
// Returns the positions of the slashes and a boolean indicating if both slashes were found
func findSlashPositions(path string) (int, int, bool) {
	// Sanity check if the path is empty
	if len(path) == 0 {
		return -1, -1, false
	}

	firstSlash := strings.Index(path, "/")

	var (
		found       = firstSlash > -1
		outOfBounds = firstSlash+1 > len(path)
	)

	if !found || outOfBounds {
		return -1, -1, false
	}

	secondSlash := strings.Index(path[firstSlash+1:], "/")

	found = secondSlash > -1
	outOfBounds = firstSlash+1+secondSlash+1 > len(path)

	if !found || outOfBounds {
		return -1, -1, false
	}

	secondSlash += firstSlash + 1

	return firstSlash, secondSlash, true
}
