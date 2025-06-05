package main

type MyStruct struct {
	arr []int
}

const a = len(MyStruct{arr: []int{1, 2}}.arr)

func main() {
	println("ok")
}

// Error:
// main/const45_b.gno:7:7-46: (const-type main.MyStruct){arr<VPField(0,0,arr)>: (const-type []int){(const (1 int)), (const (2 int))}}.arr (variable of type []int) is not constant

// TypeCheckError:
// main/const45_b.gno:7:11: len(MyStruct{…}.arr) (value of type int) is not constant
