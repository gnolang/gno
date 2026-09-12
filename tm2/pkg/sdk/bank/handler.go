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

	return sdk.Result{}
}

//----------------------------------------
// Query

// Query paths.
const (
	QueryBalance   = "balances"
	QuerySupply    = "supply"
	QuerySpendable = "spendable"
)

func (bh bankHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	switch secondPart(req.Path) {
	case QueryBalance:
		return bh.queryBalance(ctx, req)
	case QuerySupply:
		return bh.querySupply(ctx, req)
	case QuerySpendable:
		return bh.querySpendable(ctx, req)
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

// AccountSpendable is the response of /bank/spendable/{addr}.
//
// Coins is the full cross-tier balance, the same set queryBalance returns, so
// Locked, Spendable and Coins all speak about the same denoms -- which is why
// this lives in bank and not beside the schedule in auth: only this module can
// read a denom that lives outside the account object, and a schedule may name
// one.
//
// BlockTime is echoed because the answer is a function of it: a continuous
// schedule releases coins every second, so the same account queried a moment
// later reports a different Spendable. Without the timestamp a caller cannot
// tell a stale answer from a current one, nor reproduce the arithmetic. Note
// amino renders an int64 as a JSON string, so it is on the wire as
// "block_time":"1789228670" -- the same trap querySupply documents.
type AccountSpendable struct {
	Coins     std.Coins `json:"coins"`
	Locked    std.Coins `json:"locked"`
	Spendable std.Coins `json:"spendable"`
	BlockTime int64     `json:"block_time"`
}

// querySpendable reports how much of an account's balance its vesting schedule
// currently permits to be transferred out.
//
// It exists because the two facts a caller needs sit in different places and
// neither of them is the answer: /auth/accounts/{addr} carries the schedule but
// not what it evaluates to, and queryBalance carries the total with no knowledge
// that any of it is locked. Computing it off-chain means restating
// std.VestingSchedule's curve -- which has two variants, short-circuits
// differently for each, and needs big.Int for grants whose amount*elapsed
// overflows int64 -- so every reimplementation is somewhere the answer can drift
// from what this module enforces. This calls the same LockedCoins that
// SubtractCoins calls, and clamps the same way.
//
// Cost note: GetCoins walks the split-tier prefix for a caller-supplied address,
// so its cost rises with the number of denoms that address holds -- which a third
// party can grow by sending it new ones. That is the same exposure queryBalance
// already accepts on this same unmetered path, and the price of reporting a set
// that agrees with what transfers actually enforce.
func (bh bankHandler) querySpendable(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	b32addr := thirdPart(req.Path)
	addr, err := crypto.AddressFromBech32(b32addr)
	if err != nil {
		return sdk.ABCIResponseQueryFromError(
			std.ErrInvalidAddress("invalid query address " + b32addr))
	}

	blockTime := ctx.BlockTime()
	coins := bh.bank.GetCoins(ctx, addr)

	// An address that has never transacted has no account, so no schedule and
	// nothing locked. Reporting that beats an error: it is the honest answer and
	// it spares every caller a special case.
	var locked std.Coins
	if acc := bh.bank.acck.GetAccount(ctx, addr); acc != nil {
		locked = acc.LockedCoins(blockTime)
	}

	// Clamped per denom, the same arithmetic SubtractCoins performs before it
	// permits a transfer: a schedule can lock more than the account still holds,
	// once an unrestricted transfer has spent into the locked portion. Without
	// the clamp this serves a negative rather than failing -- Coins.MarshalAmino
	// is Coins.String and never errors, so "-990ugnot" goes out on the wire and
	// blows up in the client's decoder.
	var spendable std.Coins
	for _, coin := range coins {
		amount := max(coin.Amount-locked.AmountOf(coin.Denom), 0)
		if amount == 0 {
			continue // Coins carries no zero entries
		}
		// coins is sorted and appended in order, so spendable stays sorted --
		// which Coins.AmountOf's binary search relies on downstream.
		spendable = append(spendable, std.Coin{Denom: coin.Denom, Amount: amount})
	}

	bz, err := amino.MarshalJSONIndent(AccountSpendable{
		Coins:     coins,
		Locked:    locked,
		Spendable: spendable,
		BlockTime: blockTime.Unix(),
	}, "", "  ")
	if err != nil {
		return sdk.ABCIResponseQueryFromError(
			std.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", err.Error())))
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
