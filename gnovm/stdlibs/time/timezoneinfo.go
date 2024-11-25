package time

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

func GetChainTz(m *gno.Machine) string {
	if m == nil || m.Context == nil {
		return "UTC"
	}

	return std.GetContext(m).ChainTz
}
