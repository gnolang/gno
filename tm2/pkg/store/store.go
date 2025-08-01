package store

import (
	dbm "github.com/gnolang/gno/tm2/pkg/db"

	"github.com/gnolang/gno/tm2/pkg/store/rootmulti"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func NewCommitMultiStore(db dbm.DB) types.CommitMultiStore {
	return rootmulti.NewMultiStore(db)
}
