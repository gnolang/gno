package main

type MyStruct struct {
	arr []int
}

const a = len(MyStruct{arr: []int{1, 2}}.arr)

func main() {
	println("ok")
}

// Error:
// main/files/const45_b.gno:7:7: unexpected selector expression type *gnolang.DeclaredType
