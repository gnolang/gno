package rootmulti

import (
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// XXX: This should be managed by the rootMultiStore which may want to register
// more proof ops?
func DefaultProofRuntime() (prt *merkle.ProofRuntime) {
	prt = merkle.NewProofRuntime()
	prt.RegisterOpDecoder(types.ProofOpIAVLCommitment, types.CommitmentOpDecoder)
	prt.RegisterOpDecoder(types.ProofOpSimpleMerkleCommitment, types.CommitmentOpDecoder)
	prt.RegisterOpDecoder(storebptree.ProofOpBptreeCommitment, storebptree.BptreeCommitmentOpDecoder)
	return
}
