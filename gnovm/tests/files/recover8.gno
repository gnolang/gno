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
// panic: do something panic
// defer func(){ ... }()
//     main/files/recover8.gno:7
// doSomething<VPBlock(3,0)>()
//     main/files/recover8.gno:4
// main<VPBlock(1,2)>()
//     main/files/recover8.gno:20

// Error:
// do something panic
