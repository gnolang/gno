// PKGPATH: gno.land/r/test
package test

type MyStruct struct{}

func (ms MyStruct) Panic() {
	panic("panic")
}

func main(cur realm) {
	ms := MyStruct{}
	ms.Panic()
}

// Error:
// panic

// Stacktrace:
// panic: panic
// ms<VPBlock(1,1)>.Panic()
//     gno.land/r/test/zrealm_panic.gno:7
// main<VPBlock(1,1)>(struct{( .uverse.address),("gno.land/r/test" string),(nil .uverse..grealm)})
//     gno.land/r/test/zrealm_panic.gno:12
