package main

import (
	"bytes"
	"encoding/binary"
)

func main() {
	buf := new(bytes.Buffer)

	str := []byte("Hello World")

	binary.Write(buf, binary.BigEndian, str)

	println(buf.String())
}

// Output:
// Hello World
