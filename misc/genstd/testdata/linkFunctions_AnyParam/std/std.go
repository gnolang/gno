package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func AnyParam(v any) bool {
	_, ok := v.(gno.TypedValue)
	return ok
}

func AnyParamIface(v interface{}) bool {
	_, ok := v.(gno.TypedValue)
	return ok
}

func MixedParams(m *gno.Machine, s string, v any) bool {
	_, ok := v.(gno.TypedValue)
	return ok
}
