package std

// Both declare a concrete gno parameter type. The Go side takes the empty
// interface, so the argument reaches the implementation as the raw
// gno.TypedValue rather than a converted Go value.
func AnyParam(n int) string

func AnyParamLong(n int) string
