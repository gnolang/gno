package testing

import (
	"encoding/hex"

	"gno.land/p/aib/ics23"
)

func NewExistenceProof(spec *ics23.ProofSpec, key []byte, value, leaf string, paths [][]string) *ics23.ExistenceProof {
	ep := &ics23.ExistenceProof{
		Key:   key,
		Value: hexDec(value),
		Leaf: &ics23.LeafOp{
			Hash:         spec.LeafSpec.Hash,
			PrehashKey:   spec.LeafSpec.PrehashKey,
			PrehashValue: spec.LeafSpec.PrehashValue,
			Length:       spec.LeafSpec.Length,
			Prefix:       hexDec(leaf),
		},
	}
	for i, _ := range paths {
		ep.Path = append(ep.Path, &ics23.InnerOp{
			Hash:   spec.InnerSpec.Hash,
			Prefix: hexDec(paths[i][0]),
			Suffix: hexDec(paths[i][1]),
		})
	}
	return ep
}

func hexDec(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
