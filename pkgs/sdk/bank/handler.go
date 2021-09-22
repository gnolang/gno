package bank

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/std"
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
	switch queryPath(req.Path) {
	case QueryBalance:
		return bh.queryBalance(ctx, req)
	default:
		res.Error = sdk.ABCIError(
			std.ErrUnknownRequest("unknown bank query endpoint"))
		return
	}
}

// QueryBalanceParams defines the params for querying an account balance.
type QueryBalanceParams struct {
	Address crypto.Address
}

// NewQueryBalanceParams creates a new instance of QueryBalanceParams.
func NewQueryBalanceParams(addr crypto.Address) QueryBalanceParams {
	return QueryBalanceParams{Address: addr}
}

// queryBalance fetch an account's balance for the supplied height.
// Height and account address are passed as first and second path components respectively.
func (bh bankHandler) queryBalance(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	var params QueryBalanceParams

	if err := amino.UnmarshalJSON(req.Data, &params); err != nil {
		res.Error = sdk.ABCIError(
			std.ErrInternal(fmt.Sprintf("failed to pare params: %s", err.Error())))
		return
	}

	bz, err := amino.MarshalJSONIndent(bh.bank.GetCoins(ctx, params.Address), "", "  ")
	if err != nil {
		res.Error = sdk.ABCIError(
			std.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", err.Error())))
		return
	}

	res.Data = bz
	return
}

//----------------------------------------
// misc

func abciResult(err error) sdk.Result {
	return sdk.ABCIResultFromError(err)
}

// returns the second component of a query path.
func queryPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return ""
	} else {
		return parts[1]
	}
}
