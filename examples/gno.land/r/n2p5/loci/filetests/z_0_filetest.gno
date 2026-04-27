// PKGPATH: gno.land/r/test/test
package test

import (
	"chain/runtime"

	"gno.land/r/n2p5/loci"
)

func main(cur realm) {
	caller := runtime.CurrentRealm()
	println("caller: " + string(caller.Address()))

	// test nothing being set, yet.
	r0 := loci.Get(cross, caller.Address())
	println("expect: " + "")
	println("got   : " + r0)

	// set the value, which uses the CurrentRealm as the caller.
	input1 := "aGVsbG8sIHdvcmxkCg=="
	loci.Set(cross, input1)
	println("set   : " + string(input1))
	r1 := loci.Get(cross, caller.Address())
	println("expect: " + input1)
	println("got   : " + r1)

	// change the value, which should override the previous value.
	input2 := "Z29vZGJ5ZSwgd29ybGQK"
	loci.Set(cross, input2)
	println("set   : " + string(input2))
	r2 := loci.Get(cross, caller.Address())
	println("expect: " + input2)
	println("got   : " + r2)

}

// Output:
// caller: g1z7fga7u94pdmamlvcrtvsfwxgsye0qv3rres7n
// expect:
// got   :
// set   : aGVsbG8sIHdvcmxkCg==
// expect: aGVsbG8sIHdvcmxkCg==
// got   : aGVsbG8sIHdvcmxkCg==
// set   : Z29vZGJ5ZSwgd29ybGQK
// expect: Z29vZGJ5ZSwgd29ybGQK
// got   : Z29vZGJ5ZSwgd29ybGQK
