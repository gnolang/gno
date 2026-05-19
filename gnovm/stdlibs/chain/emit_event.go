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

const (
	// MaxEventPairs caps the number of (key,value) pairs per emitted event.
	// `attrs` arrives as a flat slice of alternating keys/values, so the
	// flat length cap is 2× this. The bound exists so chargeNativeGas can
	// safely walk attrs to compute per-byte gas without becoming a DoS
	// vector itself, and so downstream amino+JSON encoding is bounded.
	MaxEventPairs = 64

	// MaxEventAttrLen caps each attribute's byte length. Strings longer
	// than this (or the type string) are truncated to MaxEventAttrLen +
	// EventTruncMarker, deterministically. Chosen to fit any realistic
	// short payload (addresses, IDs, short messages) while bounding the
	// downstream encoding amplification per attr at ~1 KB.
	MaxEventAttrLen = 1024

	// EventTruncMarker is appended to truncated strings; its 3 bytes are
	// added on top of MaxEventAttrLen, so a truncated string is exactly
	// MaxEventAttrLen + len(EventTruncMarker) bytes long.
	EventTruncMarker = "..."
)

func X_emit(m *gno.Machine, typ string, attrs []string) {
	if len(attrs)/2 > MaxEventPairs {
		m.PanicString("event has too many attribute pairs")
	}
	if len(typ) > MaxEventAttrLen {
		m.PanicString("event type is too long")
	}
	eventAttrs, err := attrKeysAndValues(m, attrs)
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

// truncateValue returns s if len(s) <= MaxEventAttrLen, otherwise the first
// MaxEventAttrLen bytes followed by EventTruncMarker. Used only for attr
// VALUES — keys and typ are hard-capped (panic on overflow) since they're
// identifiers that downstream consumers filter on; silent truncation would
// alias unrelated keys.
//
// Truncation is byte-wise, not rune-aware: an attr ending in a multi-byte
// UTF-8 sequence may be cut mid-rune. Acceptable for opaque event payloads
// consumed off-chain.
func truncateValue(s string) string {
	if len(s) <= MaxEventAttrLen {
		return s
	}
	return s[:MaxEventAttrLen] + EventTruncMarker
}

// currentPkgPath retrieves the current package's pkgPath.
// It's not a native binding; but is used within this package to clarify usage.
func currentPkgPath(m *gno.Machine) (pkgPath string) {
	return m.MustPeekCallFrame(2).LastPackage.PkgPath
}

func attrKeysAndValues(m *gno.Machine, attrs []string) ([]EventAttribute, error) {
	attrLen := len(attrs)
	if attrLen%2 != 0 {
		return nil, errInvalidGnoEventAttrs
	}
	eventAttrs := make([]EventAttribute, attrLen/2)
	for i := 0; i < attrLen-1; i += 2 {
		key := attrs[i]
		if len(key) > MaxEventAttrLen {
			m.PanicString("event attribute key is too long")
		}
		eventAttrs[i/2] = EventAttribute{
			Key:   key,
			Value: truncateValue(attrs[i+1]),
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
