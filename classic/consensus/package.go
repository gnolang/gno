package consensus

import (
	"github.com/tendermint/classic/types"
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/consensus",
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
