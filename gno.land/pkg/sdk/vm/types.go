package vm

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

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

type AirdropInfo struct {
	Address crypto.Address `json:"address"`
	Amount  std.Coins      `json:"amount"`
	Claimed bool           `json:"claimed"`
}
