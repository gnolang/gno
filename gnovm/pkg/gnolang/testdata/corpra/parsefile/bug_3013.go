package main

import "testing"

func TestDummy(t *testing.T) {
	testTable := []struct {
		name string
	}{
		{
			"one",
		},
		{
			"two",
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		println(testCase.name)
	}
}
