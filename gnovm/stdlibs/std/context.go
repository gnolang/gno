package std

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type DefaultContext struct {
	chainID       string
	height        int64
	timestamp     int64 // seconds
	timestampNano int64 // nanoseconds, only used for testing.
	msg           sdk.Msg
	origCaller    crypto.Bech32Address
	origPkgAddr   crypto.Bech32Address
	origSend      std.Coins
	origSendSpent *std.Coins // mutable
	banker        BankerInterface
	eventLogger   *sdk.EventLogger
}

func NewDefaultContext(chainID string, height int64, banker BankerInterface, eventLogger *sdk.EventLogger) *DefaultContext {
	return &DefaultContext{
		chainID:     chainID,
		height:      height,
		banker:      banker,
		eventLogger: eventLogger,
	}
}

func (d *DefaultContext) ChainID() string {
	return d.chainID
}

func (d *DefaultContext) Height() int64 {
	return d.height
}

func (d *DefaultContext) SetHeight(i int64) {
	d.height = i
}

func (d *DefaultContext) Msg() sdk.Msg {
	return d.msg
}

func (d *DefaultContext) SetMsg(m sdk.Msg) {
	d.msg = m
}

func (d *DefaultContext) OrigCaller() crypto.Bech32Address {
	return d.origCaller
}

func (d *DefaultContext) SetOrigCaller(address crypto.Bech32Address) {
	d.origCaller = address
}

func (d *DefaultContext) OrigPkgAddr() crypto.Bech32Address {
	return d.origPkgAddr
}

func (d *DefaultContext) SetOrigPkgAddr(address crypto.Bech32Address) {
	d.origPkgAddr = address
}

func (d *DefaultContext) OrigSend() std.Coins {
	return d.origSend
}

func (d *DefaultContext) SetOrigSend(coins std.Coins) {
	d.origSend = coins
}

func (d *DefaultContext) OrigSendSpent() *std.Coins {
	return d.origSendSpent
}

func (d *DefaultContext) SetOrigSendSpent(coins *std.Coins) {
	d.origSendSpent = coins
}

func (d *DefaultContext) Banker() BankerInterface {
	return d.banker
}

func (d *DefaultContext) EventLogger() *sdk.EventLogger {
	return d.eventLogger
}

func (d *DefaultContext) Timestamp() int64 {
	return d.timestamp
}

func (d *DefaultContext) SetTimestamp(t int64) {
	d.timestamp = t
}

func (d *DefaultContext) TimestampNano() int64 {
	return d.timestampNano
}

func (d *DefaultContext) SetTimestampNano(t int64) {
	d.timestampNano = t
}

type ExecContext interface{}

type ExecContextChain interface {
	ChainID() string
	Height() int64
	SetHeight(int64)
	Msg() sdk.Msg
	SetMsg(sdk.Msg)
	OrigCaller() crypto.Bech32Address
	SetOrigCaller(crypto.Bech32Address)
	OrigPkgAddr() crypto.Bech32Address
	SetOrigPkgAddr(crypto.Bech32Address)
	OrigSend() std.Coins
	SetOrigSend(std.Coins)
	OrigSendSpent() *std.Coins
	SetOrigSendSpent(*std.Coins)
	Banker() BankerInterface
}

type ExecContextLogger interface {
	EventLogger() *sdk.EventLogger
}

type ExecContextTimer interface {
	Timestamp() int64
	SetTimestamp(int64)
	TimestampNano() int64
	SetTimestampNano(int64)
}
