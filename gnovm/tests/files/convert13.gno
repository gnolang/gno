package main

func main() {
	b := struct{ name string }([]byte("hello"))
}

// Error:
// main/convert13.gno:4:7-45: cannot convert [](const-type uint8)((const ("hello" <untyped> string))) (of type []uint8) to type struct{name string}

// TypeCheckError:
// main/convert13.gno:4:29: cannot convert []byte("hello") (value of type []byte) to type struct{name string}; main/convert13.gno:4:2: declared and not used: b
