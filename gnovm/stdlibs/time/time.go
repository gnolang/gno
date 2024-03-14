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
	unixTime, nanoSeconds := ctx.Timestamp.Unix(), ctx.Timestamp.Nanosecond()
	return unixTime, int32(ctx.Timestamp.Nanosecond()), unixTime*int64(time.Second) + int64(nanoSeconds)
}
