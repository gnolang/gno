package std

// ref: https://github.com/gnolang/gno/pull/853

import (
	"encoding/json"
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func X_emit(m *gno.Machine, typ string, attrs []string) {
	attrLen := len(attrs)
	eventAttrs := make([]GnoEventAttribute, attrLen/2)
	pkgPath := CurrentRealmPath(m)
	fnIdent := GetFuncNameFromCallStack(m)
	timestamp := GetTimestamp(m)

	for i := 0; i < attrLen-1; i += 2 {
		eventAttrs[i/2] = GnoEventAttribute{
			Key:   attrs[i],
			Value: attrs[i+1],
		}
	}

	event := NewGnoEvent(typ, pkgPath, fnIdent, timestamp, eventAttrs...)

	ctx := m.Context.(ExecContext)
	ctx.EventLogger.EmitEvent(event)
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

func (e GnoEvent) MarshalJSON() ([]byte, error) {
	attributesMap := make(map[string]string)
	for _, attr := range e.Attributes {
		attributesMap[attr.Key] = attr.Value
	}
	data := map[string]interface{}{
		"pkg_path":   e.PkgPath,
		"identifier": e.Identifier,
		"timestamp":  e.Timestamp,
		"attributes": attributesMap,
	}
	wrapper := map[string]interface{}{
		e.Type: data,
	}
	res, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("error marshalling event: %v", err)
	}
	return res, nil
}

type GnoEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
