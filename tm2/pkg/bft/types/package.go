package types

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bft/types",
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
		// EvidenceData{},
		Commit{},
		BlockID{},
		CommitSig{},
		Vote{},
		// Tx{},
		// Txs{},
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
