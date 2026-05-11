package auth

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// MsgCreateSession creates a new session key on the creator's account.
//
// ExpiresAt is a unix timestamp; 0 means no expiry (valid until revoked).
// SpendLimit caps coin spending per period (gas fees, MsgCall.Send, etc.).
// Empty SpendLimit means no spending is allowed — the session can only
// sign txs where another signer pays gas, or call functions with zero Send.
// SpendPeriod is in seconds; 0 means SpendLimit is a lifetime cap.
type MsgCreateSession struct {
	Creator     crypto.Address `json:"creator" yaml:"creator"`
	SessionKey  crypto.PubKey  `json:"session_key" yaml:"session_key"`
	ExpiresAt   int64          `json:"expires_at" yaml:"expires_at"`               // unix timestamp; 0 = no expiry
	AllowPaths  []string       `json:"allow_paths,omitempty" yaml:"allow_paths"`   // realm path prefixes; empty = unrestricted
	SpendLimit  std.Coins      `json:"spend_limit,omitempty" yaml:"spend_limit"`   // max spend per period; empty = no spending
	SpendPeriod int64          `json:"spend_period,omitempty" yaml:"spend_period"` // seconds; 0 = lifetime cap
}

var _ std.Msg = MsgCreateSession{}

func (msg MsgCreateSession) Route() string { return ModuleName }
func (msg MsgCreateSession) Type() string  { return "create_session" }

func (msg MsgCreateSession) ValidateBasic() error {
	if msg.Creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	if msg.SessionKey == nil {
		return std.ErrInvalidPubKey("missing session key")
	}
	if msg.ExpiresAt < 0 {
		return std.ErrUnauthorized("expires_at must be non-negative (0 means no expiry)")
	}
	if msg.SpendPeriod < 0 {
		return std.ErrUnauthorized("spend_period must be non-negative")
	}
	if len(msg.AllowPaths) > std.MaxAllowPathsPerSession {
		return std.ErrUnauthorized("too many allow_paths")
	}
	if !msg.SpendLimit.IsValid() {
		return std.ErrInvalidCoins("invalid spend_limit coins")
	}
	return nil
}

func (msg MsgCreateSession) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

func (msg MsgCreateSession) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Creator}
}

// MsgRevokeSession removes a specific session from the creator's account.
type MsgRevokeSession struct {
	Creator    crypto.Address `json:"creator" yaml:"creator"`
	SessionKey crypto.PubKey  `json:"session_key" yaml:"session_key"`
}

var _ std.Msg = MsgRevokeSession{}

func (msg MsgRevokeSession) Route() string { return ModuleName }
func (msg MsgRevokeSession) Type() string  { return "revoke_session" }

func (msg MsgRevokeSession) ValidateBasic() error {
	if msg.Creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	if msg.SessionKey == nil {
		return std.ErrInvalidPubKey("missing session key")
	}
	return nil
}

func (msg MsgRevokeSession) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

func (msg MsgRevokeSession) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Creator}
}

// MsgRevokeAllSessions removes all sessions from the creator's account.
type MsgRevokeAllSessions struct {
	Creator crypto.Address `json:"creator" yaml:"creator"`
}

var _ std.Msg = MsgRevokeAllSessions{}

func (msg MsgRevokeAllSessions) Route() string { return ModuleName }
func (msg MsgRevokeAllSessions) Type() string  { return "revoke_all_sessions" }

func (msg MsgRevokeAllSessions) ValidateBasic() error {
	if msg.Creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	return nil
}

func (msg MsgRevokeAllSessions) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

func (msg MsgRevokeAllSessions) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Creator}
}
