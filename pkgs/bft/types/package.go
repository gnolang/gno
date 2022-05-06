package types

import (
	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/bitarray"
	"github.com/gnolang/gno/pkgs/crypto/merkle"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/bft/types",
	"tm",
	amino.GetCallersDirname(),
).
	WithDependencies(
		abci.Package,
		bitarray.Package,
		merkle.Package,
	).
	WithTypes(

		// Proposal
		Proposal{},

		// Block types
		Block{},
		Header{},
		Data{},
		//EvidenceData{},
		Commit{},
		BlockID{},
		CommitSig{},
		Vote{},
		//Tx{},
		//Txs{},
		Part{},
		PartSet{},
		PartSetHeader{},

		// Internal state types
		Validator{},
		ValidatorSet{},

		// Event types
		EventNewBlock{},
		EventNewBlockHeader{},
		EventTx{},
		EventVote{},
		EventString(""),
		EventValidatorSetUpdates{},

		// Evidence types
		DuplicateVoteEvidence{},
		MockGoodEvidence{},
		MockRandomGoodEvidence{},
		MockBadEvidence{},

		// Misc.
		TxResult{},
		MockAppState{},
		VoteSet{},
	))
