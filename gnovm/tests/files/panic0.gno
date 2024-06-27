package main

func main() {
	panic("wtf")
}

// Stacktrace:
// panic wtf
// panic((const ("wtf" <untyped> string)))
//     main/files/panic0.gno:4
// main()
//     main/files/panic0.gno:3

// Error:
// wtf
