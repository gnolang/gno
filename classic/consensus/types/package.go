package types

import (
	"github.com/tendermint/classic/abci/types"
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/consensus/types",
	"tm",
	amino.GetCallersDirname(),
).
	WithDependencies(
		abci.Package,
	).
	WithTypes(

		// Round state types
		&RoundState{},
		HRS{},
		RoundStateSimple{},
		PeerRoundState{},

		// Event types
		EventNewRoundStep{},
		EventNewValidBlock{},
		EventNewRound{},
		EventCompleteProposal{},
		EventTimeoutPropose{},
		EventTimeoutWait{},
	))
