package cstypes

import (
	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	btypes "github.com/gnolang/gno/pkgs/bft/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/bft/consensus/types",
	"tm",
	amino.GetCallersDirname(),
).
	WithGoPkgName("cstypes").
	WithDependencies(
		abci.Package,
		btypes.Package,
	).
	WithTypes(

		// Round state types
		&RoundState{},
		HRS{},
		RoundStateSimple{},
		PeerRoundState{},

		// Misc
		HeightVoteSet{},

		// Event types
		EventNewRoundStep{},
		EventNewValidBlock{},
		EventNewRound{},
		EventCompleteProposal{},
		EventTimeoutPropose{},
		EventTimeoutWait{},
	))
