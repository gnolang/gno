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
// panic: here
// main<VPBlock(1,2)>()
//     main/files/panic1.gno:22

// Error:
// here

// TypeCheckError:
// main/files/panic1.gno:21:2: declared and not used: test
