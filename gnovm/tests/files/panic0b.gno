package main

func main() {
	f()
}

func f() {
	defer func() {
		panic("third")
	}()
	defer func() {
		panic("second")
	}()
	panic("first")
}

// Stacktrace:
// panic first
// panic((const ("first" <untyped> string)))
//     main/files/panic0b.gno:14
// f<VPBlock(3,1)>()
//     main/files/panic0b.gno:4
// main()
//     main/files/panic0b.gno:3
// panic second
// panic((const ("second" <untyped> string)))
//     main/files/panic0b.gno:12
// f<VPBlock(3,1)>()
//     main/files/panic0b.gno:4
// main()
//     main/files/panic0b.gno:3
// panic third
// panic((const ("third" <untyped> string)))
//     main/files/panic0b.gno:9
// f<VPBlock(3,1)>()
//     main/files/panic0b.gno:4
// main()
//     main/files/panic0b.gno:3

// Error:
// first
// 	second
// 	third
