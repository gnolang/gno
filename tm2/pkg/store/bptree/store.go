package bptree

import (
	goerrors "errors"
	"fmt"
	"math/bits"

	ics23 "github.com/cosmos/ics23/go"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bp "github.com/gnolang/gno/tm2/pkg/bptree"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	serrors "github.com/gnolang/gno/tm2/pkg/store/errors"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

const (
	defaultCacheSize = 10000

	ProofOpBptreeCommitment = "ics23:bptree"
)

// StoreConstructor implements store.CommitStoreConstructor.
func StoreConstructor(db dbm.DB, opts types.StoreOptions) types.CommitStore {
	tree := bp.NewMutableTreeWithDB(db, defaultCacheSize, bp.NewNopLogger())
	return UnsafeNewStore(tree, opts)
}

var (
	_ types.Store          = (*Store)(nil)
	_ types.CommitStore    = (*Store)(nil)
	_ types.Queryable      = (*Store)(nil)
	_ types.DepthEstimator = (*Store)(nil)
)

// Store implements types.Store and CommitStore backed by a B+ tree.
type Store struct {
	tree Tree
	mtree *bp.MutableTree // kept for operations that need the concrete type
	opts types.StoreOptions
}

func UnsafeNewStore(tree *bp.MutableTree, opts types.StoreOptions) *Store {
	return &Store{
		tree:  &mutableTreeAdapter{tree},
		mtree: tree,
		opts:  opts,
	}
}

// ExpectedDepth returns log₂(size) / log₂(B) as a deterministic estimate
// of B+ tree traversal depth.
func (st *Store) ExpectedDepth() int64 {
	size := st.tree.Size()
	if size <= 1 {
		return 1
	}
	// log₂(size) / log₂(32) = log₂(size) / 5
	depth := int64(bits.Len64(uint64(size))) / 5
	if depth < 1 {
		depth = 1
	}
	return depth
}

// GetImmutable returns a read-only store at a specific version.
func (st *Store) GetImmutable(version int64) (*Store, error) {
	if !st.VersionExists(version) {
		return nil, bp.ErrVersionDoesNotExist
	}
	iTree, err := st.tree.GetImmutableTree(version)
	if err != nil {
		return nil, err
	}
	// Wire up value resolver so ImmutableTree.Get returns actual values
	if st.mtree != nil {
		iTree.SetValueResolver(func(vk []byte) ([]byte, error) {
			return st.mtree.GetValueByKey(vk)
		})
	}
	opts := st.opts
	opts.Immutable = true
	return &Store{
		tree:  &immutableTreeAdapter{iTree},
		mtree: st.mtree,
		opts:  opts,
	}, nil
}

// --- Committer ---

func (st *Store) Commit() types.CommitID {
	hash, version, err := st.tree.SaveVersion()
	if err != nil {
		panic(err)
	}

	// Prune old versions per strategy
	previous := version - 1
	if st.opts.KeepRecent < previous {
		toRelease := previous - st.opts.KeepRecent
		if st.opts.KeepEvery == 0 || toRelease%st.opts.KeepEvery != 0 {
			err := st.tree.DeleteVersionsTo(toRelease)
			if errCause := errors.Cause(err); errCause != nil && !goerrors.Is(errCause, bp.ErrVersionDoesNotExist) {
				panic(err)
			}
		}
	}

	return types.CommitID{Version: version, Hash: hash}
}

func (st *Store) LastCommitID() types.CommitID {
	return types.CommitID{
		Version: st.tree.Version(),
		Hash:    st.tree.Hash(),
	}
}

func (st *Store) GetStoreOptions() types.StoreOptions { return st.opts }
func (st *Store) SetStoreOptions(opts types.StoreOptions) { st.opts = opts }

func (st *Store) LoadLatestVersion() error {
	// Load discovers versions and loads the latest
	latestV, err := st.mtree.Load()
	if err != nil {
		return err
	}
	if st.opts.Immutable {
		iTree, err := st.mtree.GetImmutable(latestV)
		if err != nil {
			return err
		}
		st.tree = &immutableTreeAdapter{iTree}
	} else {
		st.tree = &mutableTreeAdapter{st.mtree}
	}
	return nil
}

func (st *Store) LoadVersion(ver int64) error {
	if ver == 0 {
		return nil // version 0 is always "empty"
	}
	if st.opts.Immutable {
		if _, err := st.mtree.Load(); err != nil {
			return err
		}
		iTree, err := st.mtree.GetImmutable(ver)
		if err != nil {
			return err
		}
		st.tree = &immutableTreeAdapter{iTree}
		return nil
	}
	// Load() discovers versions and loads the latest.
	// Then LoadVersion loads the specific requested version.
	latestV, err := st.mtree.Load()
	if err != nil {
		return err
	}
	if latestV == ver {
		// Already loaded the right version
		st.tree = &mutableTreeAdapter{st.mtree}
		return nil
	}
	_, err = st.mtree.LoadVersion(ver)
	if err != nil {
		return err
	}
	st.tree = &mutableTreeAdapter{st.mtree}
	return nil
}

func (st *Store) VersionExists(version int64) bool {
	return st.tree.VersionExists(version)
}

// --- Store ---

func (st *Store) CacheWrap() types.Store { return cache.New(st) }
func (st *Store) Write()                 { panic("unexpected .Write() on bptree.Store") }

func (st *Store) Set(gctx *types.GasContext, key, value []byte) {
	types.AssertValidValue(value)
	_, err := st.tree.Set(key, value)
	if err != nil {
		panic(err)
	}
}

func (st *Store) Get(gctx *types.GasContext, key []byte) (value []byte) {
	v, err := st.tree.Get(key)
	if err != nil {
		panic(err)
	}
	return v
}

func (st *Store) Has(gctx *types.GasContext, key []byte) (exists bool) {
	has, err := st.tree.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

func (st *Store) Delete(gctx *types.GasContext, key []byte) {
	_, _, err := st.tree.Remove(key)
	if err != nil {
		panic(err)
	}
}

// --- Iterator ---

func (st *Store) Iterator(gctx *types.GasContext, start, end []byte) types.Iterator {
	return st.makeIterator(start, end, true)
}

func (st *Store) ReverseIterator(gctx *types.GasContext, start, end []byte) types.Iterator {
	return st.makeIterator(start, end, false)
}

func (st *Store) makeIterator(start, end []byte, ascending bool) types.Iterator {
	// For immutable stores, use the immutable tree's iterator but with
	// the mutable tree's ndb for value resolution.
	switch t := st.tree.(type) {
	case *immutableTreeAdapter:
		if st.mtree != nil {
			itr := bp.NewIteratorWithNDB(t.ImmutableTree, start, end, ascending, st.mtree)
			return &bptreeIterator{itr: itr, start: start, end: end}
		}
		itr, err := t.ImmutableTree.Iterator(start, end, ascending)
		if err != nil {
			panic(err)
		}
		return &bptreeIterator{itr: itr, start: start, end: end}
	default:
		itr, err := st.mtree.Iterator(start, end, ascending)
		if err != nil {
			panic(err)
		}
		return &bptreeIterator{itr: itr, start: start, end: end}
	}
}

// bptreeIterator wraps bp.Iterator to satisfy types.Iterator.
type bptreeIterator struct {
	itr        *bp.Iterator
	start, end []byte
}

func (it *bptreeIterator) Domain() (start, end []byte) { return it.start, it.end }
func (it *bptreeIterator) Valid() bool                  { return it.itr.Valid() }
func (it *bptreeIterator) Key() []byte                  { return it.itr.Key() }
func (it *bptreeIterator) Value() []byte                { return it.itr.Value() }
func (it *bptreeIterator) Next()                        { it.itr.Next() }
func (it *bptreeIterator) Close() error                  { return it.itr.Close() }
func (it *bptreeIterator) Error() error                  { return it.itr.Error() }

// --- Query ---

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

func (st *Store) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(req.Data) == 0 {
		res.Error = serrors.ErrTxDecode("Query cannot be zero length")
		return
	}

	tree := st.tree
	res.Height = getHeight(tree, req)

	switch req.Path {
	case "/key":
		key := req.Data
		res.Key = key

		if !st.VersionExists(res.Height) {
			res.Log = errors.Wrap(bp.ErrVersionDoesNotExist, "").Error()
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

		// Generate ICS23 proof for the specific version
		iTree, err := tree.GetImmutableTree(res.Height)
		if err != nil {
			panic(fmt.Sprintf("version exists but could not retrieve tree: %v", err))
		}
		// Wire value resolver for proof generation
		if st.mtree != nil {
			iTree.SetValueResolver(func(vk []byte) ([]byte, error) {
				return st.mtree.GetValueByKey(vk)
			})
		}

		var proof *ics23.CommitmentProof
		if value != nil {
			proof, err = iTree.GetMembershipProof(key)
		} else {
			proof, err = iTree.GetNonMembershipProof(key)
		}
		// Release the version-reader reservation acquired by GetImmutableTree.
		iTree.Close()

		if err != nil {
			res.Log = err.Error()
			break
		}
		res.Proof = &merkle.Proof{Ops: []merkle.ProofOp{
			NewBptreeCommitmentOp(key, proof).ProofOp(),
		}}

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
		res.Error = serrors.ErrUnknownRequest(fmt.Sprintf("Unexpected query path: %v", req.Path))
	}

	return
}

// --- ICS23 Proof Integration ---

// CommitmentOp wraps an ICS23 proof for the B+ tree.
type CommitmentOp struct {
	Key   []byte
	Proof *ics23.CommitmentProof
}

func NewBptreeCommitmentOp(key []byte, proof *ics23.CommitmentProof) CommitmentOp {
	return CommitmentOp{Key: key, Proof: proof}
}

func (op CommitmentOp) ProofOp() merkle.ProofOp {
	bz, err := op.Proof.Marshal()
	if err != nil {
		panic(err)
	}
	return merkle.ProofOp{
		Type: ProofOpBptreeCommitment,
		Key:  op.Key,
		Data: bz,
	}
}

func (op CommitmentOp) GetKey() []byte { return op.Key }

func (op CommitmentOp) Run(args [][]byte) ([][]byte, error) {
	root, err := op.Proof.Calculate()
	if err != nil {
		return nil, fmt.Errorf("could not calculate root: %w", err)
	}

	switch len(args) {
	case 0:
		// Verify absence
		if !ics23.VerifyNonMembership(bp.BptreeSpec, root, op.Proof, op.Key) {
			return nil, fmt.Errorf("proof did not verify absence of key: %s", string(op.Key))
		}
	case 1:
		// Verify existence
		if !ics23.VerifyMembership(bp.BptreeSpec, root, op.Proof, op.Key, args[0]) {
			return nil, fmt.Errorf("proof did not verify existence of key %s", op.Key)
		}
	default:
		return nil, fmt.Errorf("args must be length 0 or 1, got: %d", len(args))
	}

	return [][]byte{root}, nil
}

// Ensure CommitmentOp implements merkle.ProofOperator.
var _ merkle.ProofOperator = CommitmentOp{}

// BptreeCommitmentOpDecoder decodes a merkle.ProofOp into a CommitmentOp.
func BptreeCommitmentOpDecoder(pop merkle.ProofOp) (merkle.ProofOperator, error) {
	if pop.Type != ProofOpBptreeCommitment {
		return nil, fmt.Errorf("unexpected ProofOp.Type: %s", pop.Type)
	}
	proof := &ics23.CommitmentProof{}
	err := proof.Unmarshal(pop.Data)
	if err != nil {
		return nil, err
	}
	return CommitmentOp{
		Key:   pop.Key,
		Proof: proof,
	}, nil
}

// RegisterProofRuntime registers the B+ tree proof decoder with the given runtime.
// This should be called during app initialization alongside the existing IAVL
// and simple merkle decoders.
func RegisterProofRuntime(prt *merkle.ProofRuntime) {
	prt.RegisterOpDecoder(ProofOpBptreeCommitment, BptreeCommitmentOpDecoder)
}

