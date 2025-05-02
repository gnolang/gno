package main

func main() {
	greeting := "hello"
	_ = &greeting[2]
}

// TypeCheckError:
// main/files/addressable_4a_err.gno:5:7: invalid operation: cannot take address of greeting[2] (value of type byte)
