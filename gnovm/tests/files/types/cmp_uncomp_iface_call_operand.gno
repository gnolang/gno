package main

// A single-result function call returning an interface is an interface
// comparison boundary: its dynamic map type is uncomparable and must panic,
// even though the static type is cached as a 1-element tuple.
func get() interface{} { var m map[int]int; return m }

func main() {
	defer func() { println("recover:", recover()) }()
	_ = get() == get()
	println("no panic")
}

// Output:
// recover: runtime error: comparing uncomparable type map[int]int
