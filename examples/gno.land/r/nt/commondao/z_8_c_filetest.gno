package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const caller = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(std.NewUserRealm(caller))

	commondao.GetSubDAO(1, "foo/bar")
}

// Error:
// caller is not a realm
