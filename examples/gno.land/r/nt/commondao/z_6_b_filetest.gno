package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(commondao.Execute)(404, 1)
}

// Error:
// DAO not found
