package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_t1() int {
	return 1
}

func X_t2(m *gno.Machine) int {
	return m.NumOps
}
