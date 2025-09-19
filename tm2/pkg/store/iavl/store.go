package iavl

import (
	goerrors "errors"
	"fmt"
	"sync"

	ics23 "github.com/cosmos/ics23/go"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/iavl"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	serrors "github.com/gnolang/gno/tm2/pkg/store/errors"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

const (
	defaultIAVLCacheSize = 10000
)

// Implements store.CommitStoreConstructor.
func StoreConstructor(db dbm.DB, opts types.StoreOptions) types.CommitStore {
	tree := iavl.NewMutableTree(db, defaultIAVLCacheSize, true, iavl.NewNopLogger())
	store := UnsafeNewStore(tree, opts)
	return store
}

// ----------------------------------------

var (
	_ types.Store       = (*Store)(nil)
	_ types.CommitStore = (*Store)(nil)
	_ types.Queryable   = (*Store)(nil)
)

// Store Implements types.Store and CommitStore.
type Store struct {
	tree Tree
	opts types.StoreOptions
}

func UnsafeNewStore(tree *iavl.MutableTree, opts types.StoreOptions) *Store {
	st := &Store{
		tree: tree,
		opts: opts,
	}
	return st
}

// GetImmutable returns a reference to a new store backed by an immutable IAVL
// tree at a specific version (height) without any pruning options. This should
// be used for querying and iteration only. If the version does not exist or has
// been pruned, an error will be returned. Any mutable operations executed will
// result in a panic.
func (st *Store) GetImmutable(version int64) (*Store, error) {
	if !st.VersionExists(version) {
		return nil, iavl.ErrVersionDoesNotExist
	}

	iTree, err := st.tree.GetImmutable(version)
	if err != nil {
		return nil, err
	}

	opts := st.opts
	opts.Immutable = true

	return &Store{
		tree: &immutableTree{iTree},
		opts: opts,
	}, nil
}

// Implements Committer.
func (st *Store) Commit() types.CommitID {
	// Save a new version.
	hash, version, err := st.tree.SaveVersion()
	if err != nil {
		// TODO: Do we want to extend Commit to allow returning errors?
		panic(err)
	}

	// Release an old version of history, if not a sync waypoint.
	previous := version - 1
	if st.opts.KeepRecent < previous {
		toRelease := previous - st.opts.KeepRecent
		if st.opts.KeepEvery == 0 || toRelease%st.opts.KeepEvery != 0 {
			err := st.tree.DeleteVersionsTo(toRelease)
			if errCause := errors.Cause(err); errCause != nil && !goerrors.Is(errCause, iavl.ErrVersionDoesNotExist) {
				panic(err)
			}
		}
	}

	return types.CommitID{
		Version: version,
		Hash:    hash,
	}
}

// Implements Committer.
func (st *Store) LastCommitID() types.CommitID {
	return types.CommitID{
		Version: st.tree.Version(),
		Hash:    st.tree.Hash(),
	}
}

// Implements Committer.
func (st *Store) GetStoreOptions() types.StoreOptions {
	return st.opts
}

// Implements Committer.
func (st *Store) SetStoreOptions(opts2 types.StoreOptions) {
	st.opts = opts2
}

// Implements Committer.
func (st *Store) LoadLatestVersion() error {
	version, err := st.tree.GetLatestVersion()
	if err != nil {
		return err
	}
	return st.LoadVersion(version)
}

// Implements Committer.
func (st *Store) LoadVersion(ver int64) error {
	if st.opts.Immutable {
		immutTree, err := st.tree.(*iavl.MutableTree).GetImmutable(ver)
		if err != nil {
			return err
		}
		st.tree = &immutableTree{immutTree}
		return nil
	}
	_, err := st.tree.(*iavl.MutableTree).LoadVersion(ver)
	return err
}

// VersionExists returns whether or not a given version is stored.
func (st *Store) VersionExists(version int64) bool {
	return st.tree.VersionExists(version)
}

// Implements Store.
func (st *Store) CacheWrap() types.Store {
	return cache.New(st)
}

// Implements Store.
func (st *Store) Write() {
	panic("unexpected .Write() on iavl.Store. Hash()?")
}

// Implements types.Store.
func (st *Store) Set(key, value []byte) {
	types.AssertValidValue(value)
	_, err := st.tree.Set(key, value)
	if err != nil {
		panic(err)
	}
}

// Implements types.Store.
func (st *Store) Get(key []byte) (value []byte) {
	v, err := st.tree.Get(key)
	if err != nil {
		panic(err)
	}
	return v
}

// Implements types.Store.
func (st *Store) Has(key []byte) (exists bool) {
	has, err := st.tree.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

// Implements types.Store.
func (st *Store) Delete(key []byte) {
	_, _, err := st.tree.Remove(key)
	if err != nil {
		panic(err)
	}
}

// Implements types.Store.
func (st *Store) Iterator(start, end []byte) types.Iterator {
	var iTree *iavl.ImmutableTree

	switch tree := st.tree.(type) {
	case *immutableTree:
		iTree = tree.ImmutableTree
	case *iavl.MutableTree:
		iTree = tree.ImmutableTree
	}

	return newIAVLIterator(iTree, start, end, true)
}

// Implements types.Store.
func (st *Store) ReverseIterator(start, end []byte) types.Iterator {
	var iTree *iavl.ImmutableTree

	switch tree := st.tree.(type) {
	case *immutableTree:
		iTree = tree.ImmutableTree
	case *iavl.MutableTree:
		iTree = tree.ImmutableTree
	}

	return newIAVLIterator(iTree, start, end, false)
}

// Handle gatest the latest height, if height is 0
func getHeight(tree Tree, req abci.RequestQuery) int64 {
	height := req.Height
	if height == 0 {
		latest := tree.Version()
		if tree.VersionExists(latest - 1) {
			height = latest - 1
		} else {
			height = latest
		}
	}
	return height
}

// Query implements ABCI interface, allows queries
//
// by default we will return from (latest height -1),
// as we will have merkle proofs immediately (header height = data height + 1)
// If latest-1 is not present, use latest (which must be present)
// if you care to have the latest data to see a tx results, you must
// explicitly set the height you want to see
func (st *Store) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(req.Data) == 0 {
		msg := "Query cannot be zero length"
		res.Error = serrors.ErrTxDecode(msg)
		return
	}

	tree := st.tree

	// store the height we chose in the response, with 0 being changed to the
	// latest height
	res.Height = getHeight(tree, req)

	switch req.Path {
	case "/key": // get by key
		key := req.Data // data holds the key bytes

		res.Key = key
		if !st.VersionExists(res.Height) {
			res.Log = errors.Wrap(iavl.ErrVersionDoesNotExist, "").Error()
			break
		}

		value, err := tree.GetVersioned(key, res.Height)
		if err != nil {
			res.Log = err.Error()
			break
		}
		res.Value = value

		if !req.Prove {
			break
		}

		// Continue to prove existence/absence of value
		// Must convert store.Tree to iavl.MutableTree with given version
		iTree, err := tree.GetImmutable(res.Height)
		if err != nil {
			// sanity check: If value for given version was retrieved, immutable tree must also be retrievable
			panic(fmt.Sprintf("version exists in store but could not retrieve corresponding versioned tree in store, %v", err))
		}
		mtree := &iavl.MutableTree{
			ImmutableTree: iTree,
		}

		// Generate ics23 proof
		var proof *ics23.CommitmentProof
		if value != nil {
			// Get existence proof
			proof, err = mtree.GetMembershipProof(key)
		} else {
			// Get non-existence proof
			proof, err = mtree.GetNonMembershipProof(key)
		}
		if err != nil {
			res.Log = err.Error()
			break
		}
		// Encode and append proof
		res.Proof = &merkle.Proof{Ops: []merkle.ProofOp{types.NewIavlCommitmentOp(key, proof).ProofOp()}}

	case "/subspace":
		var KVs []types.KVPair

		subspace := req.Data
		res.Key = subspace

		iterator := types.PrefixIterator(st, subspace)
		for ; iterator.Valid(); iterator.Next() {
			KVs = append(KVs, types.KVPair{Key: iterator.Key(), Value: iterator.Value()})
		}

		iterator.Close()
		res.Value = amino.MustMarshalSized(KVs)

	default:
		msg := fmt.Sprintf("Unexpected Query path: %v", req.Path)
		res.Error = serrors.ErrUnknownRequest(msg)
		return
	}

	return
}

// ----------------------------------------

// Implements types.Iterator.
type iavlIterator struct {
	// Underlying store
	tree *iavl.ImmutableTree

	// Domain
	start, end []byte

	// Iteration order
	ascending bool

	// Channel to push iteration values.
	iterCh chan std.KVPair

	// Close this to release goroutine.
	quitCh chan struct{}

	// Close this to signal that state is initialized.
	initCh chan struct{}

	// ----------------------------------------
	// What follows are mutable state.
	mtx sync.Mutex

	invalid bool   // True once, true forever
	key     []byte // The current key
	value   []byte // The current value
}

var _ types.Iterator = (*iavlIterator)(nil)

// newIAVLIterator will create a new iavlIterator.
// CONTRACT: Caller must release the iavlIterator, as each one creates a new
// goroutine.
func newIAVLIterator(tree *iavl.ImmutableTree, start, end []byte, ascending bool) *iavlIterator {
	iter := &iavlIterator{
		tree:      tree,
		start:     types.Cp(start),
		end:       types.Cp(end),
		ascending: ascending,
		iterCh:    make(chan std.KVPair), // Set capacity > 0?
		quitCh:    make(chan struct{}),
		initCh:    make(chan struct{}),
	}
	go iter.iterateRoutine()
	go iter.initRoutine()
	return iter
}

// Run this to funnel items from the tree to iterCh.
func (iter *iavlIterator) iterateRoutine() {
	iter.tree.IterateRange(
		iter.start, iter.end, iter.ascending,
		func(key, value []byte) bool {
			select {
			case <-iter.quitCh:
				return true // done with iteration.
			case iter.iterCh <- std.KVPair{Key: key, Value: value}:
				return false // yay.
			}
		},
	)
	close(iter.iterCh) // done.
}

// Run this to fetch the first item.
func (iter *iavlIterator) initRoutine() {
	iter.receiveNext()
	close(iter.initCh)
}

// Implements types.Iterator.
func (iter *iavlIterator) Domain() (start, end []byte) {
	return iter.start, iter.end
}

// Implements types.Iterator.
func (iter *iavlIterator) Valid() bool {
	iter.waitInit()
	iter.mtx.Lock()

	validity := !iter.invalid
	iter.mtx.Unlock()
	return validity
}

// Implements types.Iterator.
func (iter *iavlIterator) Next() {
	iter.waitInit()
	iter.mtx.Lock()
	iter.assertIsValid(true)

	iter.receiveNext()
	iter.mtx.Unlock()
}

// Implements types.Iterator.
func (iter *iavlIterator) Key() []byte {
	iter.waitInit()
	iter.mtx.Lock()
	iter.assertIsValid(true)

	key := iter.key
	iter.mtx.Unlock()
	return key
}

// Implements types.Iterator.
func (iter *iavlIterator) Value() []byte {
	iter.waitInit()
	iter.mtx.Lock()
	iter.assertIsValid(true)

	val := iter.value
	iter.mtx.Unlock()
	return val
}

// Implements types.Iterator.
func (iter *iavlIterator) Close() error {
	close(iter.quitCh)
	return nil
}

// Implements types.Iterator.
func (iter *iavlIterator) Error() error {
	return nil
}

// ----------------------------------------

func (iter *iavlIterator) setNext(key, value []byte) {
	iter.assertIsValid(false)

	iter.key = key
	iter.value = value
}

func (iter *iavlIterator) setInvalid() {
	iter.assertIsValid(false)

	iter.invalid = true
}

func (iter *iavlIterator) waitInit() {
	<-iter.initCh
}

func (iter *iavlIterator) receiveNext() {
	kvPair, ok := <-iter.iterCh
	if ok {
		iter.setNext(kvPair.Key, kvPair.Value)
	} else {
		iter.setInvalid()
	}
}

// assertIsValid panics if the iterator is invalid. If unlockMutex is true,
// it also unlocks the mutex before panicking, to prevent deadlocks in code that
// recovers from panics
func (iter *iavlIterator) assertIsValid(unlockMutex bool) {
	if iter.invalid {
		if unlockMutex {
			iter.mtx.Unlock()
		}
		panic("invalid iterator")
	}
}
