package vault

import "errors"

var (
	ErrNoSuchVault      = errors.New("no such vault")
	ErrTooEarlyToRedeem = errors.New("too early to redeem")
)
