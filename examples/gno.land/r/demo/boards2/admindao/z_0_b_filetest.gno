package main // <-- Caller realm is main instead of a realm within admindao parent namespace

import (
	"std"

	"gno.land/r/demo/boards2/admindao"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOrigCaller(owner)
}

func main() {
	admindao.New("foo")
}

// Error:
// unauthorized
