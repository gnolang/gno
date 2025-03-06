package main

import (
	"std"

	"gno.land/r/nt/commondao"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	std.TestSetOriginCaller(owner)
}

func main() {
	commondao.Execute(404, 1)
}

// Error:
// DAO not found
