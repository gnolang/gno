package main

type MyStruct struct {
	Elem *MyStruct
}

func main() {
	s := &MyStruct{}
	s.Elem = s
	println(s)
}
