package main

// switch tag is a single-result function call returning an interface; matching
// a case with an uncomparable dynamic type must panic, like `tag == case`.
func get() interface{} { var m map[int]int; return m }

func main() {
	defer func() { println("recover:", recover()) }()
	switch get() {
	case get():
		println("matched")
	}
}

// Output:
// recover: runtime error: comparing uncomparable type map[int]int
