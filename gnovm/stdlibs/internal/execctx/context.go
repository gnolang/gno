package execctx

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type BankerInterface interface {
	GetCoins(addr crypto.Bech32Address) (dst std.Coins)
	SendCoins(from, to crypto.Bech32Address, amt std.Coins)
	TotalCoin(denom string) int64
	IssueCoin(addr crypto.Bech32Address, denom string, amount int64)
	RemoveCoin(addr crypto.Bech32Address, denom string, amount int64)
}

type ParamsInterface interface {
	SetString(key, val string)
	SetBool(key string, val bool)
	SetInt64(key string, val int64)
	SetUint64(key string, val uint64)
	SetBytes(key string, val []byte)
	SetStrings(key string, val []string)
	UpdateStrings(key string, val []string, add bool)
}

type ExecContext struct {
	ChainID         string
	ChainDomain     string
	Height          int64
	Timestamp       int64 // seconds
	TimestampNano   int64 // nanoseconds, only used for testing.
	OriginCaller    crypto.Bech32Address
	OriginSend      std.Coins
	OriginSendSpent *std.Coins // mutable
	Banker          BankerInterface
	Params          ParamsInterface
	EventLogger     *sdk.EventLogger
}

// GetContext returns the execution context.
// This is used to allow extending the exec context using interfaces,
// for instance when testing.
func (e ExecContext) GetExecContext() ExecContext {
	return e
}

// GetOriginSend returns the OriginSend coins.
// This implements gno.OriginSendProvider to avoid import cycles.
func (e ExecContext) GetOriginSend() std.Coins {
	return e.OriginSend
}

var _ ExecContexter = ExecContext{}
var _ gno.OriginSendProvider = ExecContext{}

// ExecContexter is a type capable of returning the parent [ExecContext]. When
// using these standard libraries, m.Context should always implement this
// interface. This can be obtained by embedding [ExecContext].
type ExecContexter interface {
	GetExecContext() ExecContext
}

// NOTE: In order to make this work by simply embedding ExecContext in another
// context (like TestExecContext), the method needs to be something other than
// the field name.

// GetContext returns the context from the Gno machine.
func GetContext(m *gno.Machine) ExecContext {
	return m.Context.(ExecContexter).GetExecContext()
}
