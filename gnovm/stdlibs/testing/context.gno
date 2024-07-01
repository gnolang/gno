package testing

import (
	"chain"
)

// Context contains execution context variables which are used primarily by the
// [runtime] package to give information about the current GnoVM execution.
type Context struct {
	IsOrigin     bool
	CurrentRealm chain.Realm
	Deposit      chain.Coins
	ChainID      string
	Height       uint64
	Time         time.Time
	Banker       chain.Banker
	Logger       bool // TODO
}

func (c Context) Run(f func()) {
}
