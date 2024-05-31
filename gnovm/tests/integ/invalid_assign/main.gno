// This should result in a compilation error for assigning a bool to a custom bool type without conversion

package main

func main() {
	b := true
	var v C = b
	fmt.Println(v)
}

type C bool

const (
	F C = false
	T C = true
)