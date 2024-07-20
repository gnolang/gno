package main

func main() {
	defer recover()

	panic("ahhhhh")
}

// Stacktrace:
// panic ahhhhh
// panic((const ("ahhhhh" <untyped> string)))
//     main/files/recover10.gno:6
// main()
//     main/files/recover10.gno:3

// Error:
// ahhhhh
