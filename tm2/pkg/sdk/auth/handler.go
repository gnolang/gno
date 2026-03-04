package auth

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type authHandler struct {
	acck  AccountKeeper
	gpKpr GasPriceKeeper
}

// NewHandler returns a handler for "auth" type messages.
func NewHandler(acck AccountKeeper, gpKpr GasPriceKeeper) authHandler {
	return authHandler{
		acck:  acck,
		gpKpr: gpKpr,
	}
}

func (ah authHandler) Process(ctx sdk.Context, msg std.Msg) sdk.Result {
	// no messages supported yet.
	errMsg := fmt.Sprintf("unrecognized auth message type: %T", msg)
	return abciResult(std.ErrUnknownRequest(errMsg))
}

//----------------------------------------
// Query

// query path
const (
	QueryAccount  = "accounts"
	QueryGasPrice = "gasprice"
)

func (ah authHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	switch secondPart(req.Path) {
	case QueryAccount:
		return ah.queryAccount(ctx, req)
	case QueryGasPrice:
		return ah.queryGasPrice(ctx, req)
	default:
		res = sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest("unknown auth query endpoint"))
		return
	}
}

// queryAccount fetch an account for the supplied height.
// Account address are passed as path component.
func (ah authHandler) queryAccount(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	// parse addr from path.
	b32addr := thirdPart(req.Path)
	addr, err := crypto.AddressFromBech32(b32addr)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrInvalidAddress(
				"invalid query address " + b32addr))
		return
	}

	// get account from addr.
	bz, err := amino.MarshalJSONIndent(
		ah.acck.GetAccount(ctx, addr),
		"", "  ")
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", err.Error())))
		return
	}

	res.Height = req.Height
	res.Data = bz
	return
}

// queryGasPrice fetch a gas price of the last block.
func (ah authHandler) queryGasPrice(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	// get account from addr.
	bz, err := amino.MarshalJSONIndent(
		ah.gpKpr.LastGasPrice(ctx),
		"", "  ")
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", err.Error())))
		return
	}

	res.Height = req.Height
	res.Data = bz
	return
}

//----------------------------------------
// misc

// returns the second component of a path.
func secondPart(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	} else {
		return parts[1]
	}
}

// returns the third component of a path.
func thirdPart(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return ""
	} else {
		return parts[2]
	}
}
