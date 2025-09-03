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

	_, pkgPath := execctx.GetRealm(m, 0)
	fnIdent := getPreviousFunctionNameFromTarget(m, "Emit")

	ctx := execctx.GetContext(m)

	evt := GnoEvent{
		Type:       typ,
		Attributes: eventAttrs,
		PkgPath:    pkgPath,
		Func:       fnIdent,
	}

	ctx.EventLogger.EmitEvent(evt)
}

// getPreviousFunctionNameFromTarget returns the last called function name (identifier) from the call stack.
func getPreviousFunctionNameFromTarget(m *gno.Machine, targetFunc string) string {
	targetIndex := findTargetFunctionIndex(m, targetFunc)
	if targetIndex == -1 {
		return ""
	}
	return findPreviousFunctionName(m, targetIndex)
}

// findTargetFunctionIndex finds and returns the index of the target function in the call stack.
func findTargetFunctionIndex(m *gno.Machine, targetFunc string) int {
	for i := len(m.Frames) - 1; i >= 0; i-- {
		currFunc := m.Frames[i].Func
		if currFunc != nil && currFunc.Name == gno.Name(targetFunc) {
			return i
		}
	}
	return -1
}

// findPreviousFunctionName returns the function name before the given index in the call stack.
func findPreviousFunctionName(m *gno.Machine, targetIndex int) string {
	for i := targetIndex - 1; i >= 0; i-- {
		currFunc := m.Frames[i].Func
		if currFunc != nil {
			return string(currFunc.Name)
		}
	}

	panic("function name not found")
}

func attrKeysAndValues(attrs []string) ([]GnoEventAttribute, error) {
	attrLen := len(attrs)
	if attrLen%2 != 0 {
		return nil, errInvalidGnoEventAttrs
	}
	eventAttrs := make([]GnoEventAttribute, attrLen/2)
	for i := 0; i < attrLen-1; i += 2 {
		eventAttrs[i/2] = GnoEventAttribute{
			Key:   attrs[i],
			Value: attrs[i+1],
		}
	}
	return eventAttrs, nil
}

type GnoEvent struct {
	Type       string              `json:"type"`
	Attributes []GnoEventAttribute `json:"attrs"`
	PkgPath    string              `json:"pkg_path"`
	Func       string              `json:"func"`
}

func (e GnoEvent) AssertABCIEvent() {}

type GnoEventAttribute struct {
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
}

func (e StorageUnlockEvent) AssertABCIEvent() {}
