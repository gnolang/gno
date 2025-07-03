package main

type MyStruct struct {
	arr [2]int
}

const a = len(MyStruct{arr: [2]int{1, 2}}.arr)

func main() {
	println("ok")
}

// Output:
// ok
