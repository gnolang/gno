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
	Type       string // type of event
	PkgPath    string // event occurred package path
	Identifier string // event occurred function identifier
	Timestamp  int64
	Attributes []GnoEventAttribute // list of event attributes (comma separated key-value pairs)
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

func (e GnoEvent) String() string {
	result, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("Error marshalling event: %v", err)
	}
	return string(result)
}

type GnoEventAttribute struct {
	Key   string
	Value string
}
