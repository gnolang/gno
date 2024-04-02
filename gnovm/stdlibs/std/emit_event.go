package std

// ref: https://github.com/gnolang/gno/pull/853

import (
	"fmt"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func X_emitEvent(m *gno.Machine, typ string, attrs []string) {
	eventAttrs := make([]sdk.EventAttribute, len(attrs)/2)
	pkgPath := CurrentRealmPath(m)

	attrLen := len(attrs)
	if attrLen%2 != 0 {
		panic(fmt.Sprintf("attributes has an odd number of elements. current length: %d", attrLen))
	}

	for i := 0; i < attrLen; i += 2 {
		eventAttrs[i/2] = sdk.EventAttribute{
			Key:   attrs[i],
			Value: attrs[i+1],
		}
	}

	timestamp := time.Now().Unix()
	height := GetHeight(m)

	event := sdk.NewEvent(typ, pkgPath, height, timestamp, eventAttrs...)

	ctx := m.Context.(ExecContext)
	ctx.EventLogger.EmitEvent(event)
}
