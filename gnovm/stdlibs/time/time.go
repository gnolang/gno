package time

import (
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type execContext interface {
	GetTimestamp() int64
	GetTimestampNano() int64
}

func X_now(m *gno.Machine) (sec int64, nsec int32, mono int64) {
	if m == nil || m.Context == nil {
		return 0, 0, 0
	}

	ctx := m.Context.(execContext)
	return ctx.GetTimestamp(), int32(ctx.GetTimestampNano()), ctx.GetTimestamp()*int64(time.Second) + ctx.GetTimestampNano()
}
