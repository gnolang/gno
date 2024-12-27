package vault

import "errors"

var (
	ErrTooEarlyToRedeem = errors.New("too early to redeem")
	ErrNoSuchVault      = errors.New("no such vault")
)
