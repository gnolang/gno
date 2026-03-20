package main

import (
	"chain/runtime"
	ext "gno.land/r/tests/vm/crossrealm_b"
)

func main() {
	adminBefore := ext.GetAdmin(cross)
	println("admin before:", adminBefore)

	var panicMsg string
	func() {
		defer func() {
			r := recover()
			if r != nil {
				panicMsg = r.(string)
			}
		}()
		ext.SetAdmin(runtime.CurrentRealm().Address())
	}()
	println("panic caught:", panicMsg)

	adminAfter := ext.GetAdmin(cross)
	println("admin after:", adminAfter)

	ext.Withdraw(cross)
	println("theft complete")
}