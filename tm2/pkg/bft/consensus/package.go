package consensus

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bft/consensus",
	"tm",
	amino.GetCallersDirname(),
).
	WithDependencies(
		cstypes.Package,
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
