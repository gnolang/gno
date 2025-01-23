package testing

import (
	"std"
	"time"
)

type Context struct {
	IsOrigin         bool
	PreviousRealm    std.Realm
	OriginPkgAddress string
	OriginSend       std.Coins
	OriginSpend      std.Coins
	ChainID          string
	Height           int64
	Time             time.Time
	Banker           bool // TODO
	Logger           bool // TODO
}
