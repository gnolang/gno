package main

var testTable = []struct {
	name string
}{
	{
		"one",
	},
	{
		"two",
	},
}

func main() {

	for _, testCase := range testTable {
		testCase := testCase

		println(testCase.name)
	}
}

// Output:
// one
// two
