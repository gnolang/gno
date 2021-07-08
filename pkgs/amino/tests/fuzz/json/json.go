package fuzzjson

import (
	amino "github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/amino/tests"
)

//-------------------------------------
// Non-interface go-fuzz tests
// See https://github.com/dvyukov/go-fuzz
// (Test that deserialize never panics)

func Fuzz(data []byte) int {
	cdc := amino.NewCodec()
	cst := tests.ComplexSt{}
	err := cdc.UnmarshalJSON(data, &cst)
	if err != nil {
		return 0
	}
	return 1
}
