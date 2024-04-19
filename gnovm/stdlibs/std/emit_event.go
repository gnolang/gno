package std

// ref: https://github.com/gnolang/gno/pull/853

import (
	"encoding/json"
	"errors"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func X_emit(m *gno.Machine, typ string, attrs []string) {
	eventAttrs, err := attrKeysAndValues(attrs)
	if err != nil {
		m.Panic(typedString(err.Error()))
	}

	pkgPath := CurrentRealmPath(m)
	fnIdent := getPrevFunctionNameFromTarget(m, "Emit")
	timestamp := getTimestamp(m)

	event := NewGnoEvent(typ, pkgPath, fnIdent, timestamp, eventAttrs...)
	ctx := m.Context.(ExecContext)
	ctx.EventLogger.EmitEvent(event)
}

func attrKeysAndValues(attrs []string) ([]GnoEventAttribute, error) {
	attrLen := len(attrs)
	if attrLen%2 != 0 {
		return nil, errors.New("odd number of attributes. cannot create key-value pairs")
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
	Identifier string              `json:"identifier"`
	Timestamp  int64               `json:"timestamp"`
	Attributes []GnoEventAttribute `json:"attributes"`
}

func NewGnoEvent(eventType string, pkgPath string, ident string, timestamp int64, attrs ...GnoEventAttribute) sdk.Event {
	return GnoEvent{
		Type:       eventType,
		PkgPath:    pkgPath,
		Identifier: ident,
		Attributes: attrs,
		Timestamp:  timestamp,
	}
}

func (e GnoEvent) AssertABCIEvent() {}

type temp GnoEvent

func (e GnoEvent) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(temp(e))
	if err != nil {
		return nil, err
	}
	return res, nil
}

type GnoEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
