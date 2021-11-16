package store

import (
	"fmt"

	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/strings"

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

// TODO move to another file.
func Print(store Store) {
	fmt.Println("//----------------------------------------")
	fmt.Println("// store:", store)
	itr := store.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key, value := itr.Key(), itr.Value()
		var keystr, valuestr string
		if strings.IsASCIIText(string(key)) {
			keystr = string(key)
		} else {
			keystr = fmt.Sprintf("0x%X", key)
		}
		if strings.IsASCIIText(string(value)) {
			valuestr = string(value)
		} else {
			valuestr = fmt.Sprintf("0x%X", value)
		}
		fmt.Printf("%s: %s\n", keystr, valuestr)
	}
}
