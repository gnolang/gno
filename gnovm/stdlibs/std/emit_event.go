package std

// ref: https://github.com/gnolang/gno/pull/575
// ref: https://github.com/gnolang/gno/pull/1833

import (
	"encoding/json"
	"errors"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

func X_emit(m *gno.Machine, typ string, attrs []string) {
	eventAttrs, err := attrKeysAndValues(attrs)
	if err != nil {
		m.Panic(typedString(err.Error()))
	}

	pkgPath := CurrentRealmPath(m)
	fnIdent := getPrevFunctionNameFromTarget(m, "Emit")
	timestamp := getTimestamp(m)

	evtstr := NewGnoEventString(typ, pkgPath, fnIdent, timestamp, eventAttrs...)
	ctx := m.Context.(ExecContext)
	ctx.EventLogger.EmitEvent(evtstr)
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

func NewGnoEventString(eventType, pkgPath, ident string, timestamp int64, attrs ...gnoEventAttribute) abci.EventString {
	evt := newGnoEvent(eventType, pkgPath, ident, timestamp, attrs...)

	res, err := json.Marshal(evt)
	if err != nil {
		panic(err)
	}

	return abci.EventString(res)
}

type gnoEvent struct {
	Type       string              `json:"type"`
	PkgPath    string              `json:"pkg_path"`
	Identifier string              `json:"identifier"`
	Timestamp  int64               `json:"timestamp"`
	Attributes []gnoEventAttribute `json:"attributes"`
}

func newGnoEvent(eventType string, pkgPath string, ident string, timestamp int64, attrs ...gnoEventAttribute) *gnoEvent {
	return &gnoEvent{
		Type:       eventType,
		PkgPath:    pkgPath,
		Identifier: ident,
		Timestamp:  timestamp,
		Attributes: attrs,
	}
}

type gnoEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
