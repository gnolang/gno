package time

import (
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

func X_now(m *gno.Machine) (sec int64, nsec int32, mono int64) {
	if m == nil || m.Context == nil {
		return 0, 0, 0
	}

	ctx := m.Context.(std.ExecContext)

	if tctx, ok := ctx.(std.ExecContextTimer); ok {
		return tctx.Timestamp(), int32(tctx.TimestampNano()), tctx.Timestamp()*int64(time.Second) + tctx.TimestampNano()
	}

	return 0, 0, 0
}
