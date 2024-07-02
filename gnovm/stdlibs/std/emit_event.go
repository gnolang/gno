package std

// ref: https://github.com/gnolang/gno/pull/575
// ref: https://github.com/gnolang/gno/pull/1833

import (
	"errors"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

var errInvalidGnoEventAttrs = errors.New("cannot pair attributes due to odd count")

func X_emit(m *gno.Machine, typ string, attrs []string) {
	eventAttrs, err := attrKeysAndValues(attrs)
	if err != nil {
		m.Panic(typedString(err.Error()))
	}

	_, pkgPath := currentRealm(m)
	fnIdent := getPrevFunctionNameFromTarget(m, "Emit")

	evt := GnoEvent{
		Type:       typ,
		PkgPath:    pkgPath,
		Func:       fnIdent,
		Attributes: eventAttrs,
	}
	ctx := GetContext(m)
	ctx.EventLogger.EmitEvent(evt)
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
	PkgPath    string              `json:"pkg_path"`
	Func       string              `json:"func"`
	Attributes []GnoEventAttribute `json:"attrs"`
}

func (e GnoEvent) AssertABCIEvent() {}

type GnoEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
