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

// Query paths.
const (
	QueryBalance = "balances"
	QuerySupply  = "supply"
)

func (bh bankHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	switch secondPart(req.Path) {
	case QueryBalance:
		return bh.queryBalance(ctx, req)
	case QuerySupply:
		return bh.querySupply(ctx, req)
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
		// Must return: otherwise execution falls through to the success path, which
		// sets res.Data from a GetCoins on the zero address. The error survives, so
		// a caller checking it is fine — but the response carries both an error and
		// a balance, and a caller reading Data first sees an empty balance for what
		// is actually a malformed request.
		return sdk.ABCIResponseQueryFromError(
			std.ErrInvalidAddress("invalid query address " + b32addr))
	}

	// get coins from addr.
	bz, err := amino.MarshalJSONIndent(bh.bank.GetCoins(ctx, addr), "", "  ")
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", err.Error())))
		return
	}

	res.Data = bz
	return
}

// querySupply returns the total supply of one denom, as an amount.
//
// Deliberately an amount rather than a std.Coin, which would have named its own denom
// and matched the balances query's shape: a zero-amount Coin marshals to the empty
// string, because Coins may not carry zeros — and zero is a legitimate answer here. The
// caller supplied the denom, so nothing is lost. Note amino renders an int64 as a
// quoted string, so the response is "1000", not 1000.
//
// A denom nobody holds reports zero, which is also what an unknown denom reports; the
// two are indistinguishable by design, exactly as for a balance.
//
// The denom is the whole path remainder, not a split component: a realm-issued denom
// is "/pkgPath:base" and contains slashes, so splitting on "/" the way the balance
// query does would truncate it at the first one. So both of these work:
//
//	bank/supply/ugnot
//	bank/supply//gno.land/r/demo/foo:gold
func (bh bankHandler) querySupply(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	denom := pathRemainder(req.Path, 2)
	if denom == "" {
		return sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest("bank/supply requires a denom, e.g. bank/supply/ugnot"))
	}
	// Validated even though TotalSupply bounds the length itself: this is an
	// unauthenticated query carrying a caller-supplied string toward a store key, and
	// naming the problem beats returning a silent zero for a typo.
	if err := std.ValidateDenom(denom); err != nil {
		return sdk.ABCIResponseQueryFromError(
			std.ErrInvalidCoins(fmt.Sprintf("invalid query denom %q: %v", denom, err)))
	}

	bz, err := amino.MarshalJSONIndent(bh.bank.TotalSupply(ctx, denom), "", "  ")
	if err != nil {
		return sdk.ABCIResponseQueryFromError(
			std.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", err.Error())))
	}
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

// pathRemainder returns everything after the first n path components, slashes
// included. Use this rather than a component split when the tail can itself contain
// slashes, as a realm denom does.
func pathRemainder(path string, n int) string {
	for range n {
		i := strings.IndexByte(path, '/')
		if i < 0 {
			return ""
		}
		path = path[i+1:]
	}
	return path
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
