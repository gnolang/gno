package os_test

import (
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
)

func X_sleep(m *gno.Machine, duration int64) {
	arg0 := m.LastBlock().GetParams1().TV
	d := arg0.GetInt64()
	sec := d / int64(time.Second)
	nano := d % int64(time.Second)
	ctx := m.Context.(*teststd.TestExecContext)
	ctx.Timestamp += sec
	ctx.TimestampNano += nano
	if ctx.TimestampNano >= int64(time.Second) {
		ctx.Timestamp += 1
		ctx.TimestampNano -= int64(time.Second)
	}

	m.Context = ctx
}
