package types

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/abci/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/bft/consensus/types",
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
