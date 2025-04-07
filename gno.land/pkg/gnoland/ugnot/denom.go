package ugnot

import "strconv"

// Denom is the denomination for ugnot, gno.land's native token.
const Denom = "ugnot"

// ValueString converts `value` to a string, appends "ugnot", and returns it.
func ValueString(value int64) string {
	return strconv.FormatInt(value, 10) + Denom
}
