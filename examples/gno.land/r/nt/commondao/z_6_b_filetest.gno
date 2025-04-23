package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	testing.SetOriginCaller(owner)
}

func main() {
	cross(commondao.Execute)(404, 1)
}

// Error:
// DAO not found
