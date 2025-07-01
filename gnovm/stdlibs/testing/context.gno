package testing

import (
	"std"
	"time"
)

type Context struct {
	CurrentRealm std.Realm
	OriginCaller std.Address
	OriginSend   std.Coins
	OriginSpend  std.Coins
	ChainID      string
	Height       int64
	Time         time.Time
}
