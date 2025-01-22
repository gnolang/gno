package testing

import (
	"std"
	"time"
)

type Context struct {
	IsOrigin         bool
	OriginCaller     std.Address
	OriginPkgAddress string
	OrigSend         std.Coins
	ChainID          string
	Height           int64
	Time             time.Time
	Banker           bool // TODO
	Logger           bool // TODO
}
