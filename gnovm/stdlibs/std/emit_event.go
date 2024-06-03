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

	pkgPath := CurrentRealmPath(m)
	fnIdent := getPrevFunctionNameFromTarget(m, "Emit")

	evt := gnoEvent{
		Type:       typ,
		PkgPath:    pkgPath,
		Func:       fnIdent,
		Attributes: eventAttrs,
	}
	ctx := m.Context.(ExecContext)
	ctx.EventLogger.EmitEvent(evt)
}

func attrKeysAndValues(attrs []string) ([]gnoEventAttribute, error) {
	attrLen := len(attrs)
	if attrLen%2 != 0 {
		return nil, errInvalidGnoEventAttrs
	}
	eventAttrs := make([]gnoEventAttribute, attrLen/2)
	for i := 0; i < attrLen-1; i += 2 {
		eventAttrs[i/2] = gnoEventAttribute{
			Key:   attrs[i],
			Value: attrs[i+1],
		}
	}
	return eventAttrs, nil
}

type gnoEvent struct {
	Type       string              `json:"type"`
	PkgPath    string              `json:"pkg_path"`
	Func       string              `json:"func"`
	Attributes []gnoEventAttribute `json:"attrs"`
}

func (e gnoEvent) AssertABCIEvent() {}

type gnoEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
