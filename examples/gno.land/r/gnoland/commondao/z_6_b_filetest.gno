package main

import (
	"std"
	"testing"

	"gno.land/r/gnoland/commondao"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	testing.SetOriginCaller(owner)
}

func main() {
	commondao.Execute(404, 1)
}

// Error:
// DAO not found
