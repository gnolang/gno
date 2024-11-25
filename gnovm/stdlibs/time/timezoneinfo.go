package time

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_GetChainTz(m *gno.Machine) string {
	if m == nil || m.Context == nil {
		return "UTC"
	}

	return GetContext(m).ChainTz
}
