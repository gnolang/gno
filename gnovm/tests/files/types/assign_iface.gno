package main

type MyStruct struct {
	doSomething int // only name matches the interface
}

type MyInterface interface {
	doSomething(string) string
}

func main() {
	var x MyInterface = MyStruct{}
	println("ok", x == nil)
}

// Error:
// main/assign_iface.gno:12:6-32: main.MyStruct does not implement main.MyInterface (wrong type for method doSomething)

// TypeCheckError:
// main/assign_iface.gno:12:22: cannot use MyStruct{} (value of struct type MyStruct) as MyInterface value in variable declaration: MyStruct does not implement MyInterface (MyStruct.doSomething is a field, not a method)
