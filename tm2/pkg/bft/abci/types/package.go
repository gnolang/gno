package abci

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bft/abci/types",
	"abci",
	amino.GetCallersDirname(),
).
	WithGoPkgName("abci").
	WithDependencies(
		merkle.Package,
	).
	WithTypes(

		/*
			pkg.Type{ // example of overriding type name.
				Type:             reflect.TypeOf(EmbeddedSt5{}),
				Name:             "EmbeddedSt5NameOverride",
				PointerPreferred: false,
			},
		*/

		// request types
		RequestBase{},
		RequestEcho{},
		RequestFlush{},
		RequestInfo{},
		RequestSetOption{},
		RequestInitChain{},
		RequestQuery{},
		RequestBeginBlock{},
		RequestCheckTx{},
		RequestDeliverTx{},
		RequestEndBlock{},
		RequestCommit{},

		// response types
		ResponseBase{},
		ResponseException{},
		ResponseEcho{},
		ResponseFlush{},
		ResponseInfo{},
		ResponseSetOption{},
		ResponseInitChain{},
		ResponseQuery{},
		ResponseBeginBlock{},
		ResponseCheckTx{},
		ResponseDeliverTx{},
		ResponseEndBlock{},
		ResponseCommit{},

		// error types
		StringError(""),

		// misc types
		ConsensusParams{},
		BlockParams{},
		ValidatorParams{},
		ValidatorUpdate{},
		LastCommitInfo{},
		VoteInfo{},
		// Validator{},
		// Violation{},

		// events
		EventString(""),
		StorageDepositEvent{},

		// mocks
		MockHeader{},

		// Params (abci/types/params.go)
	))
