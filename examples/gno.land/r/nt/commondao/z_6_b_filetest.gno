package main

import (
	"std"

	"gno.land/r/nt/commondao"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOriginCaller(owner)
}

func main() {
	commondao.Execute(404, 1)
}

// Error:
// DAO not found
