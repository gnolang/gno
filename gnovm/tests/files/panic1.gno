package main

type Test struct {
	next map[string]*Test
}

func getNewTest() *Test {
	curr := Test{
		next: map[string]*Test{
			"test": {},
		},
	}

	// Assigning a reference to the struct itself within the map
	curr.next["test"] = &curr

	return &curr
}

func main() {
	test := getNewTest()
	panic("here")
}

// Stacktrace:
// panic here
// panic((const ("here" <untyped> string)))
//     main/files/panic1.gno:22
// main()
//     main/files/panic1.gno:20

// Error:
// here
