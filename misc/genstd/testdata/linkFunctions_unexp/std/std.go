package std

import (
	gno "github.com/gnolang/gno/gno/pkg/vm"
)

func X_t1() int {
	return 1
}

func X_t2(m *gno.Machine) int {
	return m.NumOps
}
