package rootmulti

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/errors"

	"github.com/gnolang/gno/tm2/pkg/store/cachemulti"
	serrors "github.com/gnolang/gno/tm2/pkg/store/errors"
	"github.com/gnolang/gno/tm2/pkg/store/immut"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

const (
	latestVersionKey = "s/latest"
	commitInfoKeyFmt = "s/%d" // s/<version>
)

// multiStore is composed of many CommitStores. Name contrasts with
// cacheMultiStore which is for cache-wrapping other MultiStores. It implements
// the CommitMultiStore interface.
type multiStore struct {
	db           dbm.DB
	lastCommitID types.CommitID
	storeOpts    types.StoreOptions
	storesParams map[types.StoreKey]storeParams
	stores       map[types.StoreKey]types.CommitStore
	keysByName   map[string]types.StoreKey
}

var (
	_ types.CommitMultiStore = (*multiStore)(nil)
	_ types.Queryable        = (*multiStore)(nil)
)

func NewMultiStore(db dbm.DB) *multiStore {
	return &multiStore{
		db:           db,
		storesParams: make(map[types.StoreKey]storeParams),
		stores:       make(map[types.StoreKey]types.CommitStore),
		keysByName:   make(map[string]types.StoreKey),
	}
}

// Implements CommitMultiStore
func (ms *multiStore) GetStoreOptions() types.StoreOptions {
	return ms.storeOpts
}

// Implements CommitMultiStore
func (ms *multiStore) SetStoreOptions(opts types.StoreOptions) {
	ms.storeOpts = opts
	for _, store := range ms.stores {
		store.SetStoreOptions(opts)
	}
}

// Implements CommitMultiStore.
func (ms *multiStore) MountStoreWithDB(key types.StoreKey, cons types.CommitStoreConstructor, db dbm.DB) {
	if key == nil {
		panic("MountIAVLStore() key cannot be nil")
	}
	if _, ok := ms.storesParams[key]; ok {
		panic(fmt.Sprintf("Store duplicate store key %v", key))
	}
	if _, ok := ms.keysByName[key.Name()]; ok {
		panic(fmt.Sprintf("Store duplicate store key name %v", key))
	}
	ms.storesParams[key] = storeParams{
		key:         key,
		constructor: cons,
		db:          db,
	}
	ms.keysByName[key.Name()] = key
}

// Implements CommitMultiStore.
func (ms *multiStore) GetCommitStore(key types.StoreKey) types.CommitStore {
	return ms.stores[key]
}

// Implements CommitMultiStore.
func (ms *multiStore) LoadLatestVersion() error {
	ver := getLatestVersion(ms.db)
	return ms.LoadVersion(ver)
}

// Implements CommitMultiStore.
func (ms *multiStore) LoadVersion(ver int64) error {
	if ver == 0 {
		// Special logic for version 0 where there is no need to get commit
		// information.
		newStores := make(map[types.StoreKey]types.CommitStore)
		for key, storeParams := range ms.storesParams {
			store, err := ms.constructStore(storeParams)
			if err != nil {
				return errors.New("failed to load Store: %v", err)
			}
			store.SetStoreOptions(ms.storeOpts)
			err = store.LoadVersion(ver)
			if err != nil {
				return errors.New("failed to load Store version %d: %v", ver, err)
			}
			// NOTE(tb): tm2/iavl used to return empty hash for empty tree, but this
			// is no longer the case for cosmos/iavl, since this change:
			// https://github.com/cosmos/iavl/pull/304
			// For that reason, the following check is commented as no longer
			// relevant.
			// if !store.LastCommitID().IsZero() {
			// return errors.New("failed to load Store: non-empty CommitID for zero state")
			// }
			newStores[key] = store
		}
		ms.stores = newStores
		ms.lastCommitID = types.CommitID{}
		return nil
	}

	// Load store commit infos @ version ver.
	cInfo, err := getCommitInfo(ms.db, ver)
	if err != nil {
		return err
	}

	// Convert StoreInfos slice to map.
	infos := make(map[types.StoreKey]storeInfo)
	for _, storeInfo := range cInfo.StoreInfos {
		infos[ms.nameToKey(storeInfo.Name)] = storeInfo
	}

	// Load each Store and check CommitID for each.
	newStores := make(map[types.StoreKey]types.CommitStore)
	for key, storeParams := range ms.storesParams {
		var id types.CommitID
		if info, ok := infos[key]; ok {
			id = info.Core.CommitID
		}
		store, err := ms.constructStore(storeParams)
		if err != nil {
			return fmt.Errorf("failed to load Store: %w", err)
		}
		store.SetStoreOptions(ms.storeOpts)
		err = store.LoadVersion(ver)
		if err != nil {
			return errors.New("failed to load Store version %d: %v", ver, err)
		}
		if !store.LastCommitID().Equals(id) {
			return errors.New("failed to load Store: wrong commit id: %v vs %v",
				store.LastCommitID(),
				id)
		}
		newStores[key] = store
	}

	ms.lastCommitID = cInfo.CommitID()
	ms.stores = newStores

	return nil
}

// ----------------------------------------
// +CommitStore

// Implements Committer/CommitStore.
func (ms *multiStore) LastCommitID() types.CommitID {
	return ms.lastCommitID
}

// Implements Committer/CommitStore.
func (ms *multiStore) Commit() types.CommitID {
	// Commit stores.
	version := ms.lastCommitID.Version + 1
	commitInfo := commitStores(version, ms.stores)

	// Need to update atomically.
	batch := ms.db.NewBatch()
	defer batch.Close()
	setCommitInfo(batch, version, commitInfo)
	setLatestVersion(batch, version)
	batch.Write()

	// Prepare for next version.
	commitID := types.CommitID{
		Version: version,
		Hash:    commitInfo.Hash(),
	}
	ms.lastCommitID = commitID
	return commitID
}

// ----------------------------------------
// +MultiStore

// Implements MultiStore.
func (ms *multiStore) MultiCacheWrap() types.MultiStore {
	stores := make(map[types.StoreKey]types.Store)
	for k, v := range ms.stores {
		stores[k] = v
	}

	return cachemulti.New(stores, ms.keysByName)
}

// Implements MultiStore.
func (ms *multiStore) MultiWrite() {
	panic("unexpected .MultiWrite() on rootmulti.Store. Commit()?")
}

// Implements CommitMultiStore.
func (ms *multiStore) MultiImmutableCacheWrapWithVersion(version int64) (types.MultiStore, error) {
	ims := &multiStore{
		db:           dbm.NewImmutableDB(ms.db),
		storeOpts:    ms.storeOpts,
		storesParams: ms.storesParams,
		keysByName:   ms.keysByName,
	}
	ims.storeOpts.Immutable = true
	err := ims.LoadVersion(version)
	if err != nil {
		return nil, err
	}
	stores := make(map[types.StoreKey]types.Store, len(ims.stores))
	for storeKey, store := range ims.stores {
		stores[storeKey] = immut.New(store)
	}
	return cachemulti.New(stores, ims.keysByName), nil
}

// Implements MultiStore.
// If the store does not exist, panics.
func (ms *multiStore) GetStore(key types.StoreKey) types.Store {
	store := ms.stores[key]
	if store == nil {
		panic("Could not load store " + key.String())
	}
	return store
}

// Implements MultiStore

// getStoreByName will first convert the original name to
// a special key, before looking up the CommitStore.
// This is not exposed to the extensions (which will need the
// StoreKey), but is useful in main, and particularly app.Query,
// in order to convert human strings into CommitStores.
func (ms *multiStore) getStoreByName(name string) types.Store {
	key := ms.keysByName[name]
	if key == nil {
		return nil
	}
	return ms.stores[key]
}

// ---------------------- Query ------------------

// Query calls substore.Query with the same `req` where `req.Path` is
// modified to remove the substore prefix.
// Ie. `req.Path` here is `/<substore>/<path>`, and trimmed to `/<path>` for the substore.
// TODO: add proof for `multistore -> substore`.
func (ms *multiStore) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	// Query just routes this to a substore.
	path := req.Path
	storeName, subpath, err := parsePath(path)
	if err != nil {
		res.Error = err
		return
	}

	store := ms.getStoreByName(storeName)
	if store == nil {
		msg := fmt.Sprintf("no such store: %s", storeName)
		res.Error = serrors.ErrUnknownRequest(msg)
		return
	}

	queryable, ok := store.(types.Queryable)
	if !ok {
		msg := fmt.Sprintf("store %s doesn't support queries", storeName)
		res.Error = serrors.ErrUnknownRequest(msg)
		return
	}

	// trim the path and make the query
	req.Path = subpath
	res = queryable.Query(req)

	if !req.Prove {
		return res
	}
	if res.Proof == nil || len(res.Proof.Ops) == 0 {
		res.Error = serrors.ErrInternal("proof is unexpectedly empty; ensure height has not been pruned")
		return
	}

	commitInfo, errMsg := getCommitInfo(ms.db, res.Height)
	if errMsg != nil {
		res.Error = serrors.ErrInternal(errMsg.Error())
		return
	}

	proofOp, errMsg := types.ProofOpFromMap(commitInfo.toMap(), storeName)
	if errMsg != nil {
		res.Error = serrors.ErrInternal(errMsg.Error())
		return
	}

	// Append proof op.
	res.Proof.Ops = append(res.Proof.Ops, proofOp)

	// TODO: handle in another TM v0.26 update PR
	// res.Proof = buildMultiStoreProof(res.Proof, storeName, commitInfo.StoreInfos)
	return res
}

// parsePath expects a format like /<storeName>[/<subpath>]
// Must start with /, subpath may be empty
// Returns error if it doesn't start with /
func parsePath(path string) (storeName string, subpath string, err serrors.Error) {
	if !strings.HasPrefix(path, "/") {
		err = serrors.ErrUnknownRequest(fmt.Sprintf("invalid path: %s", path))
		return
	}

	paths := strings.SplitN(path[1:], "/", 2)
	storeName = paths[0]

	if len(paths) == 2 {
		subpath = "/" + paths[1]
	}

	return
}

// ----------------------------------------

func (ms *multiStore) constructStore(params storeParams) (store types.CommitStore, err error) {
	var db dbm.DB
	if params.db != nil {
		db = dbm.NewPrefixDB(params.db, []byte("s/_/"))
	} else {
		db = dbm.NewPrefixDB(ms.db, []byte("s/k:"+params.key.Name()+"/"))
	}
	opts := ms.storeOpts

	// XXX: use these:
	// return iavl.LoadStore(db, id, ms.pruningOpts, ms.lazyLoading)
	// return commitDBStoreAdapter{dbadapter.Store{db}}, nil
	store = params.constructor(db, opts)
	return store, nil
}

func (ms *multiStore) nameToKey(name string) types.StoreKey {
	for key := range ms.storesParams {
		if key.Name() == name {
			return key
		}
	}
	panic("Unknown name " + name)
}

// ----------------------------------------
// storeParams

type storeParams struct {
	key         types.StoreKey
	constructor types.CommitStoreConstructor
	db          dbm.DB
}

// ----------------------------------------
// commitInfo

// NOTE: Keep commitInfo a simple immutable struct.
type commitInfo struct {
	// Version
	Version int64

	// Store info for
	StoreInfos []storeInfo
}

func (ci commitInfo) toMap() map[string][]byte {
	m := make(map[string][]byte, len(ci.StoreInfos))
	for _, storeInfo := range ci.StoreInfos {
		m[storeInfo.Name] = storeInfo.GetHash()
	}
	return m
}

// Hash returns the simple merkle root hash of the stores sorted by name.
func (ci commitInfo) Hash() []byte {
	// TODO: cache to ci.hash []byte
	return merkle.SimpleHashFromMap(ci.toMap())
}

func (ci commitInfo) CommitID() types.CommitID {
	return types.CommitID{
		Version: ci.Version,
		Hash:    ci.Hash(),
	}
}

// ----------------------------------------
// storeInfo

// storeInfo contains the name and core reference for an
// underlying store.  It is the leaf of the Stores top
// level simple merkle tree.
type storeInfo struct {
	Name string
	Core storeCore
}

type storeCore struct {
	// StoreType StoreType
	CommitID types.CommitID
	// ... maybe add more state
}

func (si storeInfo) GetHash() []byte {
	// NOTE(tb): ics23 compatibility: return the commit hash and not the hash
	// of the commit hash.
	// See similar change in SDK https://github.com/cosmos/cosmos-sdk/pull/6323
	// Problem: this causes app hash mismatch when upgrading from an existing store.
	return si.Core.CommitID.Hash
}

// ----------------------------------------
// Misc.

func getLatestVersion(db dbm.DB) int64 {
	var latest int64
	latestBytes, err := db.Get([]byte(latestVersionKey))
	if err != nil {
		panic(err)
	}
	if latestBytes == nil {
		return 0
	}

	if err := amino.UnmarshalSized(latestBytes, &latest); err != nil {
		panic(err)
	}

	return latest
}

// Set the latest version.
func setLatestVersion(batch dbm.Batch, version int64) {
	latestBytes, _ := amino.MarshalSized(version)
	batch.Set([]byte(latestVersionKey), latestBytes)
}

// Commits each store and returns a new commitInfo.
func commitStores(version int64, storeMap map[types.StoreKey]types.CommitStore) commitInfo {
	storeInfos := make([]storeInfo, 0, len(storeMap))

	for key, store := range storeMap {
		// Commit
		commitID := store.Commit()
		/* Print all items.
		itr := store.Iterator(nil, nil)
		for ; itr.Valid(); itr.Next() {
			k, v := itr.Key(), itr.Value()
			fmt.Println("STORE ENTRY",
			colors.ColoredBytes(k, colors.Green, colors.Blue),
			colors.ColoredBytes(v, colors.Cyan, colors.Blue))
		}
		itr.Close()
		*/
		// Record CommitID
		si := storeInfo{}
		si.Name = key.Name()
		si.Core.CommitID = commitID
		// si.Core.StoreType = store.GetStoreType()
		storeInfos = append(storeInfos, si)
	}

	ci := commitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}
	return ci
}

// Gets commitInfo from disk.
func getCommitInfo(db dbm.DB, ver int64) (commitInfo, error) {
	// Get from DB.
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, ver)
	cInfoBytes, err := db.Get([]byte(cInfoKey))
	if err != nil {
		return commitInfo{}, fmt.Errorf("failed to get Store: %w", err)
	}
	if cInfoBytes == nil {
		return commitInfo{}, fmt.Errorf("failed to get Store: no data")
	}

	var cInfo commitInfo

	if err := amino.UnmarshalSized(cInfoBytes, &cInfo); err != nil {
		return commitInfo{}, fmt.Errorf("failed to get Store: %w", err)
	}

	return cInfo, nil
}

// Set a commitInfo for given version.
func setCommitInfo(batch dbm.Batch, version int64, cInfo commitInfo) {
	cInfoBytes := amino.MustMarshalSized(cInfo)
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, version)
	batch.Set([]byte(cInfoKey), cInfoBytes)
}
