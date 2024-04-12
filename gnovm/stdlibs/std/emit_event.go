package std

// ref: https://github.com/gnolang/gno/pull/853

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func X_emitEvent(m *gno.Machine, typ string, attrs []string) {
	attrLen := len(attrs)
	eventAttrs := make([]sdk.EventAttribute, attrLen/2)
	pkgPath := CurrentRealmPath(m)
	fnIdent := GetFuncNameFromCallStack(m)

	for i := 0; i < attrLen-1; i += 2 {
		eventAttrs[i/2] = sdk.EventAttribute{
			Key:   attrs[i],
			Value: attrs[i+1],
		}
	}

	timestamp := GetTimestamp(m)

	event := sdk.NewEvent(typ, pkgPath, fnIdent, timestamp, eventAttrs...)

	ctx := m.Context.(ExecContext)
	ctx.EventLogger.EmitEvent(event)
}
