package main

func Bar() string {
	return
}

func main() {
	println(Bar())
}

// Error:
// main/files/fun21.gno:4:2: expected 1 return values; got 0

// TypeCheckError:
// main/files/fun21.gno:4:2: not enough return values
// 	have ()
// 	want (string)
