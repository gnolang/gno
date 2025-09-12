package main

import (
	"chain/runtime"
	"testing"

	"gno.land/r/nt/commondao"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	commondao.Execute(cross, 404, 1)
}

// Error:
// DAO not found
