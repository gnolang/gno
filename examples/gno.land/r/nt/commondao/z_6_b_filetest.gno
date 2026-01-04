package main

import (
	"testing"

	"gno.land/r/nt/commondao"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	commondao.Execute(cross, 404, 1)
}

// Error:
// DAO not found
