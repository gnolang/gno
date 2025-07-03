package main

import (
	"fmt"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

func main() {
	// Paste an example "quoted" string from tests/fuzz/json/crashers/* here.
	// NOTE: You may want to set printLog = true.
	bz := []byte("TODO")
	cdc := amino.NewCodec()
	cst := tests.ComplexSt{}
	err := cdc.JSONUnmarshal(bz, &cst)
	fmt.Printf("Expected a panic but did not. (err: %v)", err)
}
