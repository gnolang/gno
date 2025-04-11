package chain

// ref: https://github.com/gnolang/gno/pull/575
// ref: https://github.com/gnolang/gno/pull/1833

import (
	"errors"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
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

	evt := Event{
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
	Func       string           `json:"func"`
}

func (e Event) AssertABCIEvent() {}

type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
