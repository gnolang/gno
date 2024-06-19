package vm

import "github.com/gnolang/gno/tm2/pkg/amino"

// Public facing function signatures.
// See convertArgToGno() for supported types.
type FunctionSignature struct {
	FuncName string
	Params   []NamedType
	Results  []NamedType
}

type NamedType struct {
	Name  string
	Type  string
	Value string
}

type FunctionSignatures []FunctionSignature

func (fsigs FunctionSignatures) JSON() string {
	bz := amino.MustMarshalJSON(fsigs)
	return string(bz)
}
