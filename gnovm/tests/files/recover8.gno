package main

func doSomething() {
	defer func() {
		doSomethingElse()
		if r := recover(); r != nil {
			panic("do something panic")
		}
	}()
	panic("first panic")
}

func doSomethingElse() {
	if r := recover(); r != nil {
		panic("do something else panic")
	}
}

func main() {
	doSomething()
}

// Stacktrace:
// panic: first panic
// doSomething<VPBlock(3,0)>()
//     main/files/recover8.gno:10
// main<VPBlock(1,2)>()
//     main/files/recover8.gno:20
// panic: do something panic
// func(){ ... }()
//     main/:7
// doSomething<VPBlock(3,0)>()
//     main/files/recover8.gno:4
// main<VPBlock(1,2)>()
//     main/files/recover8.gno:20

// Error:
// panic: first panic[recovered]
// 	panic: do something panic
