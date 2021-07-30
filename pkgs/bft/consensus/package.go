package consensus

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/bft/consensus",
	"tm",
	amino.GetCallersDirname(),
).
	WithDependencies(
		types.Package,
	).
	WithTypes(

		// Consensus message types
		&NewRoundStepMessage{},
		&NewValidBlockMessage{},
		&ProposalMessage{},
		&ProposalPOLMessage{},
		&BlockPartMessage{},
		&VoteMessage{},
		&HasVoteMessage{},
		&VoteSetMaj23Message{},
		&VoteSetBitsMessage{},

		// WAL message types
		newRoundStepInfo{},
		msgInfo{},
		timeoutInfo{},
	))
