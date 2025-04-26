package main

type mySuperFuncType func() // define a typed function

func funcAndInt() (mySuperFuncType, int) {
	return func() { println("hey") }, 42
}

func main() {
	var f mySuperFuncType
	f, i := funcAndInt()
	_ = i
	_ = f
	f()
	println(i)
}

// Output:
// hey
// 42
