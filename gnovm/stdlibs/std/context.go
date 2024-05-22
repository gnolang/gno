package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type ExecContext struct {
	ChainID       string
	Height        int64
	Timestamp     int64 // seconds
	TimestampNano int64 // nanoseconds, only used for testing.
	Msg           sdk.Msg
	OrigCaller    crypto.Bech32Address
	OrigPkgAddr   crypto.Bech32Address
	OrigSend      std.Coins
	OrigSendSpent *std.Coins // mutable
	Banker        BankerInterface
	EventLogger   *sdk.EventLogger
}

// ExecContext returns itself.
// This is used to allow extending the exec context using interfaces,
// for instance when testing.
func (e ExecContext) ExecContext() ExecContext {
	return e
}

// ExecContexter is a type capable of returning the parent [ExecContext]. When
// using these standard libraries, m.Context should always implement this
// interface. This can be obtained by embedding [ExecContext].
type ExecContexter interface {
	ExecContext() ExecContext
}

// GetContext returns the context from the Gno machine.
func GetContext(m *gno.Machine) ExecContext {
	return m.Context.(ExecContexter).ExecContext()
}
