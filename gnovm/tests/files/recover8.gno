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

// Error:
// do something panic
