// PKGPATH: gno.land/r/test/paygas
package paygas

import "chain/runtime"

func try(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			println(name, "panic:", r)
		}
	}()
	f()
	println(name, "ok")
}

// Outside a sponsored transaction PayGas and PayStorage are no-ops, so a
// sponsor realm can be exercised under gno test; misuse still panics.
func main(cur realm) {
	try("PayGas", func() { runtime.PayGas(1000, cur) })
	try("PayStorage", func() { runtime.PayStorage(1000, cur) })
	try("zero fee", func() { runtime.PayGas(0, cur) })
	try("previous realm", func() { runtime.PayGas(1000, cur.Previous()) })
	try("sub-realm token", func() { runtime.PayStorage(1000, cur.Sub("x")) })
}

// Output:
// PayGas ok
// PayStorage ok
// zero fee panic: PayGas: maxFee must be positive
// previous realm panic: PayGas: rlm must be the current realm
// sub-realm token panic: PayStorage: rlm is not a realm
