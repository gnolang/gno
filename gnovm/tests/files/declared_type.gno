package main

type mySuperFuncType func() // define a typed function

func funcAndInt() (mySuperFuncType, int) { // return this type and another crap
	return func() { println("hey") }, 42
}

func main() {
	var f mySuperFuncType
	f, i := funcAndInt() // try to get both: panic
	_ = i
	_ = f
	f()
	println(i)
}

// Output:
// hey
// 42
