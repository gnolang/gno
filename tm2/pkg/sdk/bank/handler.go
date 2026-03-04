package bank

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type bankHandler struct {
	bank BankKeeper
}

// NewHandler returns a handler for "bank" type messages.
func NewHandler(bank BankKeeper) bankHandler {
	return bankHandler{
		bank: bank,
	}
}

func (bh bankHandler) Process(ctx sdk.Context, msg std.Msg) sdk.Result {
	switch msg := msg.(type) {
	case MsgSend:
		return bh.handleMsgSend(ctx, msg)

	case MsgMultiSend:
		return bh.handleMsgMultiSend(ctx, msg)

	default:
		errMsg := fmt.Sprintf("unrecognized bank message type: %T", msg)
		return abciResult(std.ErrUnknownRequest(errMsg))
	}
}

// Handle MsgSend.
func (bh bankHandler) handleMsgSend(ctx sdk.Context, msg MsgSend) sdk.Result {
	/*
		if !bh.bank.GetSendEnabled(ctx) {
			return abciResult(ErrSendDisabled())
		}
		if bh.bank.BlacklistedAddr(msg.ToAddress) {
			return std.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", msg.ToAddress)).Result()
		}
	*/

	err := bh.bank.SendCoins(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	if err != nil {
		return abciResult(err)
	}

	/*
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			),
		)
	*/

	return sdk.Result{}
}

// Handle MsgMultiSend.
func (bh bankHandler) handleMsgMultiSend(ctx sdk.Context, msg MsgMultiSend) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked
	/*
		if !k.GetSendEnabled(ctx) {
			return abciResult(std.ErrSendDisabled())
		}
		for _, out := range msg.Outputs {
			if bh.bank.BlacklistedAddr(out.Address) {
				return abciResult(std.ErrUnauthorized(fmt.Sprintf("%s is not allowed to receive transactions", out.Address)))
			}
		}
	*/

	err := bh.bank.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return abciResult(err)
	}

	/*
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			),
		)
	*/

	return sdk.Result{}
}

//----------------------------------------
// Query

// query balance path
const QueryBalance = "balances"

func (bh bankHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	switch secondPart(req.Path) {
	case QueryBalance:
		return bh.queryBalance(ctx, req)
	default:
		res = sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest("unknown bank query endpoint"))
		return
	}
}

// queryBalance fetch an account's balance for the supplied height.
// Account address is passed as path component.
func (bh bankHandler) queryBalance(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	// parse addr from path.
	b32addr := thirdPart(req.Path)
	addr, err := crypto.AddressFromBech32(b32addr)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrInvalidAddress("invalid query address " + b32addr))
	}

	// get coins from addr.
	bz, err := amino.MarshalJSONIndent(bh.bank.GetCoins(ctx, addr), "", "  ")
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

func abciResult(err error) sdk.Result {
	return sdk.ABCIResultFromError(err)
}

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
