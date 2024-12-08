package main

import (
	"log"
	"os"
	"path/filepath"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

const maxAllocTx = 500 * 1000 * 1000

type BenchStore struct {
	mulStore store.MultiStore
	gnoStore gno.Store
}

func (bStore BenchStore) Write() {
	bStore.mulStore.MultiWrite()
}

func benchmarkDiskStore() BenchStore {
	storeDir, err := filepath.Abs("bench_store")
	if err != nil {
		log.Fatal("unable to get absolute path for storage directory.", err)
	}
	err = os.RemoveAll(storeDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("Directory does not exist.", err)
		} else {
			log.Fatal("can not clean up storage directory", err)
		}
	}

	db, err := dbm.NewDB("gnolang", dbm.GoLevelDBBackend, filepath.Join(storeDir, config.DefaultDBDir))
	if err != nil {
		log.Fatalf("error initializing database %q using path %q: %s\n", dbm.GoLevelDBBackend, storeDir, err)
	}

	baseKey := store.NewStoreKey("baseKey")
	iavlKey := store.NewStoreKey("iavlKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()
	msCache := ms.MultiCacheWrap()

	alloc := gno.NewAllocator(maxAllocTx)
	baseSDKStore := msCache.GetStore(baseKey)
	iavlSDKStore := msCache.GetStore(iavlKey)
	gStore := gno.NewStore(alloc, baseSDKStore, iavlSDKStore)

	return BenchStore{
		mulStore: msCache,
		gnoStore: gStore,
	}
}
