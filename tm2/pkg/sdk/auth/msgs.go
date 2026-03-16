package auth

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// MsgCreateSession creates a new session key on the creator's account.
type MsgCreateSession struct {
	Creator     crypto.Address `json:"creator" yaml:"creator"`
	SessionKey  crypto.PubKey  `json:"session_key" yaml:"session_key"`
	ExpiresAt   int64          `json:"expires_at" yaml:"expires_at"` // unix timestamp
	AllowPaths  []string       `json:"allow_paths,omitempty" yaml:"allow_paths"`
	SpendLimit  std.Coins      `json:"spend_limit,omitempty" yaml:"spend_limit"`  // max spend per period; empty = no spending allowed
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
	if len(msg.AllowPaths) > std.MaxAllowPathsPerSession {
		return std.ErrUnauthorized("too many allow_paths")
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
