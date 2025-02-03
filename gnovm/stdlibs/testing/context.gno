package testing

import (
	"std"
	"time"
)

type Context struct {
	IsOrigin         bool
	CurrentRealm     std.Realm
	OriginCaller     std.Address
	OriginPkgAddress std.Address
	OriginSend       std.Coins
	OriginSpend      std.Coins
	ChainID          string
	Height           int64
	Time             time.Time
	Banker           bool // TODO
	Logger           bool // TODO
}
