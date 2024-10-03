package main

import "fmt"

type Int int

func (I Int) String() string {
	return "foo"
}

type Stringer interface {
	String() string
}

func main() {
	var i Int
	var st Stringer = i
	fmt.Println(st.String())
}

// Output:
// foo
