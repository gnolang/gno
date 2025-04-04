package testing

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"time"
)

type Context struct {
	CurrentRealm runtime.Realm
	OriginCaller chain.Address
	OriginSend   banker.Coins
	OriginSpend  banker.Coins
	ChainID      string
	Height       int64
	Time         time.Time
}
