package main

import (
	"std"
	"strings"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOriginCaller(owner)
}

func main() {
	boards2.SetRealmNotice("This is a test realm message")

	// Render content must contain the message
	content := boards2.Render("")
	println(strings.HasPrefix(content, "> This is a test realm message\n\n"))
}

// Output:
// true
