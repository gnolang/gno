package chain

// ref: https://github.com/gnolang/gno/pull/575
// ref: https://github.com/gnolang/gno/pull/1833

import (
	"errors"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errInvalidGnoEventAttrs = errors.New("cannot pair attributes due to odd count")

func X_emit(m *gno.Machine, typ string, attrs []string) {
	eventAttrs, err := attrKeysAndValues(attrs)
	if err != nil {
		m.PanicString(err.Error())
	}

	pkgPath := currentPkgPath(m)

	ctx := execctx.GetContext(m)

	evt := Event{
		Type:       typ,
		Attributes: eventAttrs,
		PkgPath:    pkgPath,
	}

	ctx.EventLogger.EmitEvent(evt)
}

// currentPkgPath retrieves the current package's pkgPath.
// It's not a native binding; but is used within this package to clarify usage.
func currentPkgPath(m *gno.Machine) (pkgPath string) {
	return m.MustPeekCallFrame(2).LastPackage.PkgPath
}

func attrKeysAndValues(attrs []string) ([]EventAttribute, error) {
	attrLen := len(attrs)
	if attrLen%2 != 0 {
		return nil, errInvalidGnoEventAttrs
	}
	eventAttrs := make([]EventAttribute, attrLen/2)
	for i := 0; i < attrLen-1; i += 2 {
		eventAttrs[i/2] = EventAttribute{
			Key:   attrs[i],
			Value: attrs[i+1],
		}
	}
	return eventAttrs, nil
}

type Event struct {
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attrs"`
	PkgPath    string           `json:"pkg_path"`
}

func (e Event) AssertABCIEvent() {}

type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// StorageDepositEvent is emitted when a storage deposit fee is locked.
type StorageDepositEvent struct {
	BytesDelta int64    `json:"bytes_delta"`
	FeeDelta   std.Coin `json:"fee_delta"`
	PkgPath    string   `json:"pkg_path"`
}

func (e StorageDepositEvent) AssertABCIEvent() {}

// StorageUnlockEvent is emitted when a storage deposit fee is unlocked.
type StorageUnlockEvent struct {
	// For unlock, BytesDelta is negative
	BytesDelta int64    `json:"bytes_delta"`
	FeeRefund  std.Coin `json:"fee_refund"`
	PkgPath    string   `json:"pkg_path"`
	// RefundWithheld is true if the refund was retained because of token lock
	RefundWithheld bool `json:"refund_withheld"`
}

func (e StorageUnlockEvent) AssertABCIEvent() {}
