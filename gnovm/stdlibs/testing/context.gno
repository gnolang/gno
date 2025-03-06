package testing

import (
	"time"

	"std"
)

type Context struct {
	CurrentRealm     std.Realm
	OriginCaller     std.Address
	OriginPkgAddress std.Address
	OriginSend       std.Coins
	OriginSpend      std.Coins
	ChainID          string
	Height           int64
	Time             time.Time
}
