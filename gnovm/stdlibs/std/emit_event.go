package std

// ref: https://github.com/gnolang/gno/pull/575
// ref: https://github.com/gnolang/gno/pull/1833

import (
	"encoding/json"
	"errors"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

func X_emit(m *gno.Machine, typ string, attrs []string) abci.EventString {
	eventAttrs, err := attrKeysAndValues(attrs)
	if err != nil {
		m.Panic(typedString(err.Error()))
	}

	pkgPath := CurrentRealmPath(m)
	fnIdent := getPrevFunctionNameFromTarget(m, "Emit")

	evt := NewGnoEvent(typ, pkgPath, fnIdent, eventAttrs...)
	ctx := m.Context.(ExecContext)
	ctx.EventLogger.EmitEvent(evt)

	bb, err := json.Marshal(ctx.EventLogger.Events())
	if err != nil {
		m.Panic(typedString(err.Error()))
	}

	return abci.EventString(bb)
}

func attrKeysAndValues(attrs []string) ([]gnoEventAttribute, error) {
	attrLen := len(attrs)
	if attrLen%2 != 0 {
		return nil, errors.New("cannot pair attributes due to odd count")
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

func NewGnoEvent(eventType, pkgPath, ident string, attrs ...gnoEventAttribute) *gnoEvent {
	return newGnoEvent(eventType, pkgPath, ident, attrs...)
}

type gnoEvent struct {
	Type       string              `json:"type"`
	PkgPath    string              `json:"pkg_path"`
	Identifier string              `json:"identifier"`
	Attributes []gnoEventAttribute `json:"attributes"`
}

func (e gnoEvent) AssertABCIEvent() {}

func newGnoEvent(eventType string, pkgPath string, ident string, attrs ...gnoEventAttribute) *gnoEvent {
	return &gnoEvent{
		Type:       eventType,
		PkgPath:    pkgPath,
		Identifier: ident,
		Attributes: attrs,
	}
}

type gnoEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
