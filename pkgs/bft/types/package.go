package types

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/abci/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/bft/types",
	"tm",
	amino.GetCallersDirname(),
).
	WithDependencies(
		abci.Package,
	).
	WithTypes(

		// Block types
		Block{},
		Header{},
		Data{},
		//EvidenceData{},
		Commit{},
		BlockID{},
		CommitSig{},
		PartSetHeader{},
		Vote{},
		//Tx{},
		//Txs{},

		// Internal state types
		Validator{},

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
	))
