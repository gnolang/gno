package caller

import (
	"gno.land/r/docs/crossing/callee"
)

func CallCallee(_ realm) string {
	//return callee.GetCustom(cross).UpdateCaller()
	//return "", callee.Caller()
	r := cross

	return callee.GetCustom(r).UpdateCaller("NewString!!!")
}

func Render(_ string) string {
	return "TODO"
}
