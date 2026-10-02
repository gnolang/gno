package main

type S struct{ m map[string]int }

func main() {
	s := S{m: map[string]int{"a": 1}}
	_ = &s.m["a"]
}

// Error:
// main/addressable_9c_err.gno:7:6-15: invalid operation: cannot take address of s<VPBlock(1,0)>.m[(const ("a" string))] (value of type int)

// TypeCheckError:
// main/addressable_9c_err.gno:7:7: invalid operation: cannot take address of s.m["a"] (map index expression of type int)
