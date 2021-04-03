package types

import (
	"github.com/tendermint/classic/abci/types"
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/types",
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
