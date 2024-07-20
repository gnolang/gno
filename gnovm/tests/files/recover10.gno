package main

func main() {
	defer recover()

	panic("ahhhhh")
}

// Stacktrace:
// panic ahhhhh
// main()
//     main/files/recover10.gno:6

// Error:
// ahhhhh
