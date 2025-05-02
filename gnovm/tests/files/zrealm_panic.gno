// PKGPATH: gno.land/r/test
package test

type MyStruct struct{}

func (ms MyStruct) Panic() {
	panic("panic")
}

func main() {
	crossing()

	ms := MyStruct{}
	ms.Panic()
}

// Error:
// panic

// Stacktrace:
// panic: panic
// ms<VPBlock(1,0)>.Panic()
//     gno.land/r/test/files/zrealm_panic.gno:7
// main<VPBlock(1,1)>()
//     gno.land/r/test/files/zrealm_panic.gno:14
