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
	_ = test // keep the variable used; the test is about something else
	panic("here")
}

// Stacktrace:
// panic: here
// main<VPBlock(1,2)>()
//     main/panic1.gno:23

// Error:
// here
