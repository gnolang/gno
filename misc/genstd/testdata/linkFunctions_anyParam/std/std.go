package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func AnyParam(m *gno.Machine, n any) string { return "" }

func AnyParamLong(m *gno.Machine, n interface{}) string { return "" }
