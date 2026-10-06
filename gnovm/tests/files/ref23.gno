package main

import "fmt"

// ref of a []byte element (DataByteType path).
// The static type of s[i] is uint8, so &s[i] must be *uint8.
// Previously doOpRef had explicit DataByteType handling; now the
// static-type path (getTypeOf) already returns uint8.

func main() {
	s := []byte("hello")
	p := &s[0]
	fmt.Printf("%T\n", p)
	*p = 'H'
	fmt.Println(string(s))
}

// Output:
// *uint8
// Hello
