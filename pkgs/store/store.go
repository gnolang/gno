package store

import (
	dbm "github.com/gnolang/gno/pkgs/db"

	"github.com/gnolang/gno/pkgs/store/rootmulti"
	"github.com/gnolang/gno/pkgs/store/types"
)

func NewCommitMultiStore(db dbm.DB) types.CommitMultiStore {
	return rootmulti.NewMultiStore(db)
}

func NewPruningOptionsFromString(strategy string) (opt PruningOptions) {
	switch strategy {
	case "nothing":
		opt = PruneNothing
	case "everything":
		opt = PruneEverything
	case "syncable":
		opt = PruneSyncable
	default:
		opt = PruneSyncable
	}
	return
}
