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
	switch msg := msg.(type) {
	case MsgCreateSession:
		return ah.handleMsgCreateSession(ctx, msg)
	case MsgRevokeSession:
		return ah.handleMsgRevokeSession(ctx, msg)
	case MsgRevokeAllSessions:
		return ah.handleMsgRevokeAllSessions(ctx, msg)
	default:
		errMsg := fmt.Sprintf("unrecognized auth message type: %T", msg)
		return abciResult(std.ErrUnknownRequest(errMsg))
	}
}

func (ah authHandler) handleMsgCreateSession(ctx sdk.Context, msg MsgCreateSession) sdk.Result {
	acc := ah.acck.GetAccount(ctx, msg.Creator)
	if acc == nil {
		return abciResult(std.ErrUnknownAddress("account not found"))
	}

	blockTime := ctx.BlockTime().Unix()
	if msg.ExpiresAt != 0 {
		if msg.ExpiresAt <= blockTime {
			return abciResult(std.ErrUnauthorized(fmt.Sprintf(
				"session already expired: expires_at=%d, block_time=%d",
				msg.ExpiresAt, blockTime)))
		}
		if msg.ExpiresAt > blockTime+std.MaxSessionDuration {
			return abciResult(std.ErrUnauthorized(fmt.Sprintf(
				"session duration exceeds maximum: expires_at=%d, max=%d (block_time+%ds)",
				msg.ExpiresAt, blockTime+std.MaxSessionDuration, std.MaxSessionDuration)))
		}
	}

	sessionAddr := msg.SessionKey.Address()

	// Check collision with existing regular account.
	if ah.acck.GetAccount(ctx, sessionAddr) != nil {
		return abciResult(std.ErrUnauthorized("session key address collides with existing account"))
	}
	// Check duplicate session.
	if ah.acck.GetSessionAccount(ctx, msg.Creator, sessionAddr) != nil {
		return abciResult(std.ErrUnauthorized("session key already exists"))
	}
	// Check session count.
	count := 0
	ah.acck.IterateSessions(ctx, msg.Creator, func(_ std.Account) bool {
		count++
		return count >= std.MaxSessionsPerAccount
	})
	if count >= std.MaxSessionsPerAccount {
		return abciResult(std.ErrSessionLimit(fmt.Sprintf(
			"too many sessions: count=%d, max=%d", count, std.MaxSessionsPerAccount)))
	}
	// Check SpendPeriod max.
	if msg.SpendPeriod > std.MaxSessionDuration {
		return abciResult(std.ErrUnauthorized(fmt.Sprintf(
			"spend_period exceeds maximum: got=%d, max=%d",
			msg.SpendPeriod, std.MaxSessionDuration)))
	}
	// Check AllowPaths count.
	if len(msg.AllowPaths) > std.MaxAllowPathsPerSession {
		return abciResult(std.ErrUnauthorized(fmt.Sprintf(
			"too many allow paths: count=%d, max=%d",
			len(msg.AllowPaths), std.MaxAllowPathsPerSession)))
	}
	for i, path := range msg.AllowPaths {
		if path == "" {
			return abciResult(std.ErrUnauthorized(fmt.Sprintf(
				"empty allow_path entry at index %d", i)))
		}
		if strings.HasSuffix(path, "/") {
			return abciResult(std.ErrUnauthorized(fmt.Sprintf(
				"allow_path must not end with /: got %q at index %d", path, i)))
		}
	}

	// Create session account via prototype.
	sa := ah.acck.NewSessionAccount(ctx, msg.Creator, msg.SessionKey)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(msg.ExpiresAt)
	da.SetSpendLimit(msg.SpendLimit)
	da.SetSpendPeriod(msg.SpendPeriod)
	da.SetSpendReset(blockTime)

	// Set AllowPaths via local interface — concrete type is GnoSessionAccount.
	// This is the CREATION-time writer. The READ-side interface (pathRestricted,
	// which exposes GetAllowPaths) lives in gno.land/pkg/gnoland/app.go and
	// is called at tx-time by checkSessionRestrictions. The two interfaces are
	// deliberately separated — tm2 must not import gno.land types, so each
	// layer defines its own local interface against the same concrete methods
	// on *GnoSessionAccount.
	type allowPathsSetter interface{ SetAllowPaths([]string) }
	if ps, ok := sa.(allowPathsSetter); ok && len(msg.AllowPaths) > 0 {
		ps.SetAllowPaths(msg.AllowPaths)
	}

	ah.acck.SetSessionAccount(ctx, msg.Creator, sa)
	return sdk.Result{}
}

func (ah authHandler) handleMsgRevokeSession(ctx sdk.Context, msg MsgRevokeSession) sdk.Result {
	sessionAddr := msg.SessionKey.Address()
	if ah.acck.GetSessionAccount(ctx, msg.Creator, sessionAddr) == nil {
		return abciResult(std.ErrSessionNotFound("session not found"))
	}
	ah.acck.RemoveSessionAccount(ctx, msg.Creator, sessionAddr)
	return sdk.Result{}
}

func (ah authHandler) handleMsgRevokeAllSessions(ctx sdk.Context, msg MsgRevokeAllSessions) sdk.Result {
	ah.acck.RemoveAllSessions(ctx, msg.Creator)
	return sdk.Result{}
}

//----------------------------------------
// Query

// query path
const (
	QueryAccount        = "accounts"
	QueryGasPrice       = "gasprice"
	QuerySessions       = "sessions" // /auth/accounts/{addr}/sessions
	QuerySessionAccount = "session"  // /auth/accounts/{addr}/session/{sessionAddr}
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
// Sub-paths: .../sessions lists all sessions, .../session/<addr> gets one.
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

	subQuery := fourthPart(req.Path)
	switch subQuery {
	case "":
		// get account from addr.
		bz, err := amino.MarshalJSONIndent(
			ah.acck.GetAccount(ctx, addr),
			"", "  ")
		if err != nil {
			res = sdk.ABCIResponseQueryFromError(
				std.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", err.Error())))
			return
		}
		res.Data = bz
		return
	case QuerySessions:
		return ah.querySessions(ctx, addr)
	case QuerySessionAccount:
		sessionB32 := fifthPart(req.Path)
		sessionAddr, err := crypto.AddressFromBech32(sessionB32)
		if err != nil {
			res = sdk.ABCIResponseQueryFromError(
				std.ErrInvalidAddress(
					"invalid session query address " + sessionB32))
			return
		}
		return ah.querySession(ctx, addr, sessionAddr)
	default:
		res = sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest("unknown account sub-query: " + subQuery))
		return
	}
}

// querySessions returns all session accounts for a master address.
func (ah authHandler) querySessions(ctx sdk.Context, master crypto.Address) (res abci.ResponseQuery) {
	var sessions []std.Account
	ah.acck.IterateSessions(ctx, master, func(acc std.Account) bool {
		sessions = append(sessions, acc)
		return false
	})
	bz, err := amino.MarshalJSONIndent(sessions, "", "  ")
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", err.Error())))
		return
	}
	res.Data = bz
	return
}

// querySession returns a specific session account for a master address.
func (ah authHandler) querySession(ctx sdk.Context, master, session crypto.Address) (res abci.ResponseQuery) {
	acc := ah.acck.GetSessionAccount(ctx, master, session)
	if acc == nil {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrSessionNotFound("session not found"))
		return
	}
	bz, err := amino.MarshalJSONIndent(acc, "", "  ")
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrInternal(fmt.Sprintf("could not marshal result to JSON: %s", err.Error())))
		return
	}
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

// returns the fourth component of a path.
func fourthPart(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		return ""
	}
	return parts[3]
}

// returns the fifth component of a path.
func fifthPart(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		return ""
	}
	return parts[4]
}
