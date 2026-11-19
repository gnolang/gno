package chain

// ref: https://github.com/gnolang/gno/pull/575
// ref: https://github.com/gnolang/gno/pull/1833

import (
	"encoding/binary"
	"errors"
	"strings"
	"sync"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

var errInvalidGnoEventAttrs = errors.New("cannot pair attributes due to odd count")

const (
	// MaxEventPairs caps the number of (key,value) pairs per emitted event.
	// `attrs` arrives as a flat slice of alternating keys/values, so the
	// flat length cap is 2× this. The bound exists so X_emit can
	// safely walk attrs to compute per-byte gas without becoming a DoS
	// vector itself, and so downstream amino+JSON encoding is bounded.
	MaxEventPairs = 64

	// MaxEventAttrLen caps each attribute's byte length. The event type,
	// every key, AND every value that exceeds this makes emit panic — it
	// fails loudly instead of silently truncating. A silent cap would be a
	// hidden invariant: each realm would have to pre-check its own value
	// lengths to guarantee protocol correctness, and a truncated value
	// (e.g. a hex-encoded IBC ack or membership proof) would corrupt
	// downstream consumers with no on-chain signal.
	//
	// Sized to hold a realistic binary payload after hex/base64 expansion
	// (~2 KB of raw bytes once hex-encoded as 0x... doubles the length) —
	// e.g. an IBC packet acknowledgement, membership proof, or packet-data
	// chunk emitted by an on-chain bridge realm — while still bounding the
	// downstream amino+JSON encoding amplification per attr at ~4 KB.
	MaxEventAttrLen = 4096
)

// emitGasPerByte prices each byte of an event's amino encoding at a store
// write's per-byte rate: the block keeps, encodes and stores it after the call.
var emitGasPerByte = store.DefaultGasConfig().WriteCostPerByte

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
		// Copied: a substring would keep its whole source string alive
		// for as long as the block keeps the event, billed for its length.
		Type:       strings.Clone(typ),
		Attributes: eventAttrs,
		PkgPath:    pkgPath,
	}

	if m.GasMeter != nil {
		m.GasMeter.ConsumeGas(emitGasPerByte*int64(eventSize(evt)), "EmitEventPerByte")
	}

	ctx.EventLogger.EmitEvent(evt)
}

var eventTypeURL = sync.OnceValue(func() string { return amino.GetTypeURL(Event{}) })

// eventSize is len(amino.MustMarshalAny(e)), computed without encoding it.
func eventSize(e Event) int {
	v := stringFieldSize(e.Type) + stringFieldSize(e.PkgPath)
	for _, a := range e.Attributes {
		n := stringFieldSize(a.Key) + stringFieldSize(a.Value)
		v += bytesFieldSize(n)
	}
	return bytesFieldSize(len(eventTypeURL())) + bytesFieldSize(v)
}

// stringFieldSize is the amino size of a string field, omitted when empty.
func stringFieldSize(s string) int {
	if s == "" {
		return 0
	}
	return bytesFieldSize(len(s))
}

// bytesFieldSize is the size of a length-prefixed field of n bytes: key, length, bytes.
func bytesFieldSize(n int) int {
	var b [binary.MaxVarintLen64]byte
	return 1 + binary.PutUvarint(b[:], uint64(n)) + n
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
		key, value := attrs[i], attrs[i+1]
		if len(key) > MaxEventAttrLen {
			m.PanicString("event attribute key is too long")
		}
		if len(value) > MaxEventAttrLen {
			m.PanicString("event attribute value is too long")
		}
		eventAttrs[i/2] = EventAttribute{
			Key:   strings.Clone(key),
			Value: strings.Clone(value),
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
	// Payer is the realm that paid the deposit through runtime.PayStorage, or
	// empty when the message's caller paid it.
	Payer crypto.Bech32Address `json:"payer,omitempty"`
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
	// SponsorRefund is the part of FeeRefund returned to the realm that
	// sponsored this storage earlier in the same transaction; the caller got
	// the rest. Nil when the caller got all of it.
	SponsorRefund *std.Coin `json:"sponsor_refund,omitempty"`
}

func (e StorageUnlockEvent) AssertABCIEvent() {}
