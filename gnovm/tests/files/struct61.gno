package main

type MyStruct struct {
	FieldA string
	FieldB string
}

func myStruct(a, b string) MyStruct {
	return MyStruct{
		FieldA: a,
		FieldB: b,
	}
}

func main() {
	s := myStruct("aaa", "bbb")
	print(s.FieldA + "," + s.FieldB)
}

// Output:
// aaa,bbb
