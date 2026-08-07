package gnolang

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"iter"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/ristretto/v2"
	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/txlog"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/colors"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/trace"
	"github.com/gnolang/gno/tm2/pkg/store/utils"
	stringz "github.com/gnolang/gno/tm2/pkg/strings"
	"github.com/pmezard/go-difflib/difflib"
)

// PackageGetter specifies how the store may retrieve packages which are not
// already in its cache. PackageGetter should return nil when the requested
// package does not exist. store should be used to run the machine, or otherwise
// call any methods which may call store.GetPackage; avoid using any "global"
// store as the one passed to the PackageGetter may be a fork of that (ie.
// the original is not meant to be written to). Loading dependencies may
// cause writes to happen to the store, such as MemPackages to iavlstore.
type PackageGetter func(pkgPath string, store Store) (*PackageNode, *PackageValue)

// NativeResolver is a function which can retrieve native bodies of native functions.
type NativeResolver func(pkgName string, name Name) func(m *Machine)

// StorageDiffs maps realm paths to their storage size difference (in bytes).
type StorageDiffs = map[string]int64

// Store is the central interface that specifies the communications between the
// GnoVM and the underlying data store; currently, generally the gno.land
// blockchain, or the file system.
type Store interface {
	// STABLE
	BeginTransaction(baseStore, iavlStore store.Store, gctx *store.GasContext, gasMeter store.GasMeter) TransactionStore
	GetPackageGetter() PackageGetter
	SetPackageGetter(PackageGetter)
	GetPackage(pkgPath string, isImport bool) *PackageValue
	SetCachePackage(*PackageValue)
	GetPackageRealm(pkgPath string) *Realm
	SetPackageRealm(*Realm)
	GetRealmByID(pid PkgID) *Realm // interrealm v2: cache-backed PkgID lookup
	GetObject(oid ObjectID) Object
	GetObjectSafe(oid ObjectID) Object
	SetObject(Object) int64 // returns size difference of the object
	GetStagingPackage() *PackageValue
	SetStagingPackage(pv *PackageValue)
	DelObject(Object) int64 // returns size difference of the object
	GetType(tid TypeID) Type
	GetTypeSafe(tid TypeID) Type
	SetCacheType(Type)
	SetType(Type)
	GetPackageNode(pkgPath string) *PackageNode
	GetBlockNode(Location) BlockNode
	GetBlockNodeSafe(Location) BlockNode
	SetBlockNode(BlockNode)
	RealmStorageDiffs() StorageDiffs // returns storage changes per realm within the message

	// UNSTABLE
	GetAllocator() *Allocator
	SetAllocator(alloc *Allocator)
	GetPreprocessAllocator() *Allocator
	SetPreprocessAllocator(alloc *Allocator)
	NumMemPackages() int64
	// Upon restart, all packages will be re-preprocessed; This
	// loads BlockNodes and Types onto the store for persistence
	// version 1.
	AddMemPackage(mpkg *std.MemPackage, mptype MemPackageType)
	DeleteMemPackage(path string)
	GetMemPackage(path string) *std.MemPackage
	GetMemPackageAll(path string) *std.MemPackage
	GetMemFile(path string, name string) *std.MemFile
	FindPathsByPrefix(prefix string) iter.Seq[string]
	// Yields each indexed package's PROD mempackage (test/filetest files
	// live under the #allbutprod sibling and are not included), in index
	// order. A package with no production .gno files has no prod blob and
	// is skipped.
	IterMemPackage() <-chan *std.MemPackage
	ClearObjectCache() // run before processing a message
	GarbageCollectObjectCache(gcCycle int64)
	SetNativeResolver(NativeResolver)                              // for native functions
	GetNative(pkgPath string, name Name) func(m *Machine)          // for native functions
	PopulateStdlibCache(paths []string)                            // populate stdlib byte cache at node start
	PopulateStdlibCacheFrom(paths []string, baseStore store.Store) // populate from a specific store
	SetLogStoreOps(dst io.Writer)
	LogFinalizeRealm(rlmpath string) // to mark finalization of realm boundaries
	Print()
}

// TransactionStore is a store where the operations modifying the underlying store's
// caches are temporarily held in a buffer, and then executed together after
// executing Write.
type TransactionStore interface {
	Store

	// Write commits the current buffered transaction data to the underlying store.
	// It also clears the current buffer of the transaction.
	Write()
}

// Gas consumption descriptors.
const (
	GasComputeMapKeyDesc = "ComputeMapKey"
	GasAminoDecodeDesc   = "AminoDecodePerByte"
	GasAminoEncodeDesc   = "AminoEncodePerByte"
)

// GasConfig defines amino compute gas costs for GnoVM stores.
// Storage I/O gas is handled separately by the cache.Store via GasContext.
type GasConfig struct {
	GasAminoEncode int64 // per byte cost for amino marshal
	GasAminoDecode int64 // per byte cost for amino unmarshal
}

// DefaultGasConfig returns a default gas config.
func DefaultGasConfig() GasConfig {
	return GasConfig{
		GasAminoEncode: amino.GasEncodePerByte,
		GasAminoDecode: amino.GasDecodePerByte,
	}
}

type defaultStore struct {
	// underlying stores used to keep data
	baseStore store.Store // for objects, types, nodes
	iavlStore store.Store // for escaped object hashes

	// Shared stdlib byte cache. Populated at node start, inherited
	// across transactions. Maps backend object key to raw bytes
	// (hash || amino). Each tx unmarshals its own copy — no aliasing.
	// Stdlib objects are immutable after genesis, so bytes never change.
	stdlibKeyBytes map[string][]byte

	// transaction-scoped
	// cacheNodes is an actual store - BlockNodes are not stored in the underlying
	// DB and must be re-initialized using PreprocessAllFilesAndSaveBlockNodes.
	cacheObjects map[ObjectID]Object
	cacheTypes   map[TypeID]Type
	cacheNodes   txlog.Map[Location, BlockNode]
	// cacheRealms is the per-tx *Realm pointer cache, parallel to
	// cacheObjects. Single source of truth for in-memory *Realm
	// pointers within a tx — populated by GetPackageRealm and
	// SetPackageRealm; consulted by GetRealmByID and fillPackage.
	// Ensures pv.Realm and any other in-tx caller observe the same
	// pointer (so in-memory Time/sumDiff mutations are visible
	// everywhere).
	cacheRealms map[PkgID]*Realm
	alloc       *Allocator // for accounting for cached items

	// preprocessAlloc, when non-nil, is the per-tx hard-cap allocator
	// installed by the keeper (AddPackage / Run) before RunMemPackage.
	// Sub-Machines spun up during Preprocess (evalStaticType, evalConst,
	// etc. at preprocess.go:3947, 4112, 4175, 4258) pick it up via
	// NewMachineWithOptions's nil-Alloc fallback. preAlloc.collect is
	// nil → Allocate hard-panics on maxBytes overflow rather than
	// attempting a GC retry (which would undercount because GC doesn't
	// visit m.Values, the operand stack). gasMeter is shared with the
	// outer tx Machine so CPU and alloc gas both bill against tx gas.
	// Inherited via BeginTransaction so nested forked stores see it.
	// Cleared by the keeper's defer at handler exit; not serialized.
	preprocessAlloc *Allocator

	// Partially restored package; occupies memory and tracked for GC,
	// this is more efficient than iterating over cacheObjects.
	stagingPackage *PackageValue

	// store configuration; cannot be modified in a transaction
	pkgGetter      PackageGetter  // non-realm packages
	nativeResolver NativeResolver // for injecting natives
	aminoCache     *ristretto.Cache[[]byte, Type]

	// transient
	opslog  io.Writer // for logging store operations.
	current []string  // for detecting import cycles.

	// gas
	gctx      *store.GasContext // for storage I/O gas (nil = no charging)
	gasMeter  store.GasMeter
	gasConfig GasConfig

	// realm storage changes on message level.
	realmStorageDiffs StorageDiffs // maps realm path to size diff
}

var globalAminoCache = sync.OnceValue[*ristretto.Cache[[]byte, Type]](func() *ristretto.Cache[[]byte, Type] {
	rc, err := ristretto.NewCache(&ristretto.Config[[]byte, Type]{
		NumCounters: 1_000_000,       // maximum number of keys in cache
		MaxCost:     128 * (1 << 20), // 128 MB
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	return rc
})

func NewStore(alloc *Allocator, baseStore, iavlStore store.Store) *defaultStore {
	ds := &defaultStore{
		baseStore: baseStore,
		iavlStore: iavlStore,
		alloc:     alloc,

		// cacheObjects is set; objects in the store will be copied over for any transaction.
		cacheObjects: make(map[ObjectID]Object),
		cacheTypes:   make(map[TypeID]Type),
		cacheNodes:   txlog.NewSyncGoMap[Location, BlockNode](),
		cacheRealms:  make(map[PkgID]*Realm),

		// stdlib byte cache
		stdlibKeyBytes: make(map[string][]byte),

		// reset at the message level
		realmStorageDiffs: make(StorageDiffs),

		// store configuration
		pkgGetter:      nil,
		nativeResolver: nil,
		gasConfig:      DefaultGasConfig(),
		aminoCache:     globalAminoCache(),
	}
	InitStoreCaches(ds)
	return ds
}

// If nil baseStore and iavlStore, the baseStores are re-used.
func (ds *defaultStore) BeginTransaction(baseStore, iavlStore store.Store, gctx *store.GasContext, gasMeter store.GasMeter) TransactionStore {
	if baseStore == nil {
		baseStore = ds.baseStore
	}
	if iavlStore == nil {
		iavlStore = ds.iavlStore
	}
	ds2 := &defaultStore{
		// underlying stores
		baseStore: baseStore,
		iavlStore: iavlStore,

		// transaction-scoped
		cacheObjects: make(map[ObjectID]Object),
		cacheTypes:   make(map[TypeID]Type),
		cacheNodes:   txlog.Wrap(ds.cacheNodes),
		cacheRealms:  make(map[PkgID]*Realm),
		alloc:        ds.alloc.Fork().Reset(),

		// Inherit the per-tx preprocess allocator so sub-Machines spun
		// up via NewMachine(pkg, store) inside preprocess (which fork
		// the store via BeginTransaction first) see the same hard-cap
		// allocator the keeper installed.
		preprocessAlloc: ds.preprocessAlloc,

		// stdlib byte cache (shared reference)
		stdlibKeyBytes: ds.stdlibKeyBytes,

		// store configuration
		pkgGetter:      ds.pkgGetter,
		nativeResolver: ds.nativeResolver,

		// gas
		gctx:       gctx,
		gasMeter:   gasMeter,
		gasConfig:  ds.gasConfig,
		aminoCache: ds.aminoCache,

		// transient
		current: nil,
		opslog:  nil,
		// reset at the message level
		realmStorageDiffs: make(StorageDiffs),
	}
	InitStoreCaches(ds2)

	return transactionStore{ds2}
}

type transactionStore struct {
	*defaultStore
}

func (t transactionStore) Write() {
	t.cacheNodes.(txlog.MapCommitter[Location, BlockNode]).Commit()
}

func (transactionStore) SetNativeResolver(ns NativeResolver) {
	panic("SetNativeResolver may not be called in a transaction store")
}

// CopyCachesFromStore allows to copy a store's internal object, type and
// BlockNode cache into the dst store.
// This is mostly useful for testing, where many stores have to be initialized.
func CopyFromCachedStore(destStore, cachedStore Store, cachedBase, cachedIavl store.Store) {
	ds, ss := destStore.(transactionStore), cachedStore.(*defaultStore)

	iter := cachedBase.Iterator(nil, nil, nil)
	for ; iter.Valid(); iter.Next() {
		ds.baseStore.Set(ds.gctx, iter.Key(), iter.Value())
	}
	iter = cachedIavl.Iterator(nil, nil, nil)
	for ; iter.Valid(); iter.Next() {
		ds.iavlStore.Set(ds.gctx, iter.Key(), iter.Value())
	}

	for k, v := range ss.cacheNodes.Iterate() {
		ds.cacheNodes.Set(k, v)
	}
}

func (ds *defaultStore) GetAllocator() *Allocator {
	return ds.alloc
}

func (ds *defaultStore) SetAllocator(alloc *Allocator) {
	ds.alloc = alloc
}

func (ds *defaultStore) GetPreprocessAllocator() *Allocator {
	return ds.preprocessAlloc
}

func (ds *defaultStore) SetPreprocessAllocator(alloc *Allocator) {
	ds.preprocessAlloc = alloc
}

// Used by cmd/gno (e.g. lint) to inject target package as MPTest.
func (ds *defaultStore) GetPackageGetter() (pg PackageGetter) {
	return ds.pkgGetter
}

func (ds *defaultStore) SetPackageGetter(pg PackageGetter) {
	ds.pkgGetter = pg
}

func (ds *defaultStore) GetStagingPackage() *PackageValue {
	return ds.stagingPackage
}

func (ds *defaultStore) SetStagingPackage(pv *PackageValue) {
	ds.stagingPackage = pv
}

// Gets package from cache, or loads it from baseStore, or gets it from package getter.
func (ds *defaultStore) GetPackage(pkgPath string, isImport bool) *PackageValue {
	defer func() {
		ds.stagingPackage = nil
	}()

	// helper to detect circular imports
	if isImport {
		if slices.Contains(ds.current, pkgPath) {
			panic(fmt.Sprintf("import cycle detected: %q (through %v)", pkgPath, ds.current))
		}
		ds.current = append(ds.current, pkgPath)
		defer func() {
			ds.current = ds.current[:len(ds.current)-1]
		}()
	}

	oid := ObjectIDFromPkgPath(pkgPath)
	// Synthetic packages are never persisted: check only the cache — a
	// backend read would be a guaranteed miss charged as a full I/O
	// read, and the pkgGetter cannot resolve them either.
	if IsSyntheticPath(pkgPath) {
		if oo, exists := ds.cacheObjects[oid]; exists {
			return oo.(*PackageValue)
		}
		return nil
	}
	// Get package from cache or baseStore
	oo := ds.GetObjectSafe(oid)
	if oo != nil {
		return oo.(*PackageValue)
	}
	// otherwise, fetch from pkgGetter.
	if ds.pkgGetter != nil {
		if pn, pv := ds.pkgGetter(pkgPath, ds); pv != nil {
			// e.g. tests/imports_tests loads example/gno.land/r/... realms.
			// if pv.IsRealm() {
			// 	panic("realm packages cannot be gotten from pkgGetter")
			// }
			ds.SetBlockNode(pn)
			// NOTE: not SetObject() here, we don't want to overwrite the value
			// from pkgGetter. Realm values obtained this way will get written
			// elsewhere later.
			ds.cacheObjects[oid] = pv
			// cache all types. usually preprocess() sets types,
			// but packages gotten from the pkgGetter may skip this step,
			// so fill in store.CacheTypes here.
			for _, tv := range pv.GetBlock(nil).Values {
				if tv.T == nil {
					// tv.T is nil here only when only predefined.
					// (for other types, .T == nil even after definition).
				} else if tv.T.Kind() == TypeKind {
					t := tv.GetType()
					ds.SetType(t)
				}
			}
			return pv
		}
	}
	// otherwise, package does not exist.
	return nil
}

// Used to set throwaway packages.
// NOTE: To check whether a mem package has been run, use GetMemPackage()
// instead of implementing HasCachePackage().
func (ds *defaultStore) SetCachePackage(pv *PackageValue) {
	oid := ObjectIDFromPkgPath(pv.PkgPath)
	if _, exists := ds.cacheObjects[oid]; exists {
		panic(fmt.Sprintf("package %s already exists in cache", pv.PkgPath))
	}
	ds.cacheObjects[oid] = pv
}

// Some atomic operation. Consults cacheRealms before reading from
// baseStore; populates the cache on read so that subsequent in-tx
// callers observe the same *Realm pointer.
func (ds *defaultStore) GetPackageRealm(pkgPath string) (rlm *Realm) {
	var size int
	if bm.Enabled {
		old := bm.StartStore(bm.StoreGetPackageRealm)
		defer func() { bm.StopStore(bm.StoreGetPackageRealm, old, size) }()
	}
	oid := ObjectIDFromPkgPath(pkgPath)
	if cached, ok := ds.cacheRealms[oid.PkgID]; ok {
		return cached
	}
	key := backendRealmKey(oid)
	bz := ds.baseStore.Get(ds.gctx, []byte(key))
	if bz == nil {
		return nil
	}
	gas := overflow.Mulp(ds.gasConfig.GasAminoDecode, store.Gas(len(bz)))
	ds.consumeGas(gas, GasAminoDecodeDesc)
	if trace.StoreGasEnabled {
		trace.Store("DECODE_REALM", gas, []byte(key), len(bz), "none")
	}
	amino.MustUnmarshal(bz, &rlm)
	size = len(bz)
	if debugAssert {
		if rlm.ID != oid.PkgID {
			panic(fmt.Sprintf("unexpected realm id: expected %v but got %v",
				oid.PkgID, rlm.ID))
		}
	}
	ds.cacheRealms[oid.PkgID] = rlm
	return rlm
}

// GetRealmByID looks up a Realm via the per-tx cache, falling back
// to a path-resolution + baseStore load on miss. Single source of
// truth for in-memory *Realm pointers within a tx. Used by
// PushFrameCall borrow rule #2 and by cross-realm finalize
// (touchForeignRealm).
func (ds *defaultStore) GetRealmByID(pid PkgID) *Realm {
	if rlm, ok := ds.cacheRealms[pid]; ok {
		return rlm
	}
	path := pkgPathFromPkgID(ds, pid)
	if path == "" {
		return nil
	}
	return ds.GetPackageRealm(path)
}

// An atomic operation to set the package realm info (id counter etc).
// Refreshes the cacheRealms entry so subsequent in-tx reads see the
// updated *Realm pointer.
func (ds *defaultStore) SetPackageRealm(rlm *Realm) {
	var size int
	if bm.Enabled {
		old := bm.StartStore(bm.StoreSetPackageRealm)
		defer func() { bm.StopStore(bm.StoreSetPackageRealm, old, size) }()
	}
	oid := ObjectIDFromPkgPath(rlm.Path)
	key := backendRealmKey(oid)
	bz := amino.MustMarshal(rlm)
	gas := overflow.Mulp(ds.gasConfig.GasAminoEncode, store.Gas(len(bz)))
	ds.consumeGas(gas, GasAminoEncodeDesc)
	if trace.StoreGasEnabled {
		trace.Store("ENCODE_REALM", gas, []byte(key), len(bz), "none")
	}
	ds.baseStore.Set(ds.gctx, []byte(key), bz)
	ds.cacheRealms[rlm.ID] = rlm
	size = len(bz)
}

// NOTE: it can be use to retrieve a package by ObjectID, but
// does not consult the packageGetter, only lookup the cache/store.
// NOTE: current implementation behavior requires
// all []TypedValue types and TypeValue{} types to be
// loaded (non-ref) types.
func (ds *defaultStore) GetObject(oid ObjectID) Object {
	oo := ds.GetObjectSafe(oid)
	if oo == nil {
		panic(fmt.Sprintf("unexpected object with id %s", oid.String()))
	}
	return oo
}

func (ds *defaultStore) GetObjectSafe(oid ObjectID) Object {
	// check cache.
	if oo, exists := ds.cacheObjects[oid]; exists {
		return oo
	}
	// check baseStore.
	if ds.baseStore != nil {
		if oo := ds.loadObjectSafe(oid); oo != nil {
			return oo
		}
	}
	return nil
}

// loads and caches an object.
// CONTRACT: object isn't already in the cache.
func (ds *defaultStore) loadObjectSafe(oid ObjectID) Object {
	var size int
	if bm.Enabled {
		old := bm.StartStore(bm.StoreGetObject)
		defer func() { bm.StopStore(bm.StoreGetObject, old, size) }()
	}
	key := backendObjectKey(oid)
	var hashbz []byte
	if oid.PkgID.IsStdlibPkg() {
		hashbz = ds.stdlibKeyBytes[key] // from byte cache — no I/O gas
	}
	if hashbz == nil {
		hashbz = ds.baseStore.Get(ds.gctx, []byte(key)) // from store — charges I/O gas
	}
	if hashbz != nil {
		size = len(hashbz)
		hash := hashbz[:HashSize]
		bz := hashbz[HashSize:]
		fromCache := oid.PkgID.IsStdlibPkg() && ds.stdlibKeyBytes[key] != nil
		var oo Object
		gas := overflow.Mulp(ds.gasConfig.GasAminoDecode, store.Gas(len(bz)))
		ds.consumeGas(gas, GasAminoDecodeDesc)
		if trace.StoreGasEnabled {
			trace.Store("DECODE_OBJ", gas, []byte(key), len(hashbz),
				fmt.Sprintf("cached=%v,meter=%v", fromCache, ds.gasMeter != nil))
		}
		amino.MustUnmarshal(bz, &oo)
		if debug {
			debug.Printf("loadObjectSafe by oid: %v, type of oo: %v\n", oid, reflect.TypeOf(oo))
		}

		// See copyValueWithRefs — child Objects become RefValue slots
		// in the serialized amino bytes, and internalRefSize accounts
		// for those slots.
		ss := oo.GetShallowSize()
		rs := internalRefSize(oo)
		// Allocate atomically: one Allocate call prevents GC from
		// intercepting between shallow-size and RefValue-size accounting.
		ds.alloc.Allocate(ss + rs)

		if debugAssert {
			if oo.GetObjectID() != oid {
				panic(fmt.Sprintf("unexpected object id: expected %v but got %v",
					oid, oo.GetObjectID()))
			}
		}
		oo.SetHash(ValueHash{NewHashlet(hash)})
		if debugAssert {
			// Verify stored hash matches actual content hash.
			if computed := HashBytes(bz); computed != NewHashlet(hash) {
				panic(fmt.Sprintf(
					"stored hash mismatch for %s: stored %X, computed %X",
					oid, hash, computed.Bytes()))
			}
		}

		if pv, ok := oo.(*PackageValue); ok {
			ds.SetStagingPackage(pv)
			ds.fillPackage(pv)
		}

		ds.cacheObjects[oid] = oo
		oo.GetObjectInfo().LastObjectSize = int64(size)
		_ = fillTypesOfValue(ds, oo)
		return oo
	}
	return nil
}

func (ds *defaultStore) fillPackage(pv *PackageValue) {
	pv.GetBlock(ds) // preload
	if pv.Realm == nil {
		if pv.IsRealm() {
			pv.Realm = ds.GetPackageRealm(pv.PkgPath)
		} else if isImmutableLibraryPath(pv.PkgPath) {
			// /p/ and stdlib packages carry a frozen, immutable realm so
			// borrow rule #2 lands m.Realm on it (not nil). It is not
			// persisted (IsRealm()==false), so recreate it
			// deterministically — the realm ID is derived from the pkgpath.
			// _test overlays are excluded (see isImmutableLibraryPath).
			pv.Realm = NewRealm(pv.PkgPath)
		}
	}
	// Re-derive denormalized PkgID cache (pv.PkgID is marked
	// json:"-" so amino skipped it on load).
	if pv.PkgID.IsZero() {
		pv.PkgID = PkgIDFromPkgPath(pv.PkgPath)
	}
	// pv.fBlocksMap is left empty: file blocks load lazily via
	// GetFileBlock when a function in that file is first called
	// (see FuncValue.GetParent). Loading a package therefore no longer
	// reads every file block up front — a call materializes only the
	// file blocks it touches.
	//
	// Preserve historical hydration and gas accounting when there is no
	// unused file to skip. A package with <= 1 file loads that one block on
	// first call anyway, so laziness buys nothing there — only a gas shift;
	// keep the eager path for them (this also keeps master's gas for the
	// rare import that reads only package-level vars).
	if len(pv.FNames) <= 1 {
		pv.deriveFBlocksMap(ds)
	}
}

// NOTE: unlike GetObject(), SetObject() is also used to persist updated
// package values.
func (ds *defaultStore) SetObject(oo Object) int64 {
	var size int
	if bm.Enabled {
		old := bm.StartStore(bm.StoreSetObject)
		defer func() { bm.StopStore(bm.StoreSetObject, old, size) }()
	}
	oid := oo.GetObjectID()
	// replace children/fields with Ref.
	o2 := copyValueWithRefs(oo)
	if debugAssert {
		// Invariant: every function-local declared type referenced by the
		// persist-copy must be known to the store — SetType'd at addpkg
		// (saveFuncLocalTypes) or loaded via GetType — so it must be in
		// cacheTypes. A miss means a RefType was minted that the store
		// cannot resolve on reload — the object would be persisted
		// permanently unreadable.
		ds.assertNoDanglingLocalTypeRef(o2)
	}
	// marshal to binary.
	bz := amino.MustMarshalAny(o2)
	gas := overflow.Mulp(ds.gasConfig.GasAminoEncode, store.Gas(len(bz)))
	ds.consumeGas(gas, GasAminoEncodeDesc)
	if trace.StoreGasEnabled {
		trace.Store("ENCODE_OBJ", gas, []byte(backendObjectKey(oid)), len(bz), "none")
	}
	// set hash.
	hash := HashBytes(bz) // XXX objectHash(bz)???
	if len(hash) != HashSize {
		panic("should not happen")
	}
	oo.SetHash(ValueHash{hash})
	// difference between object size and cached value
	diff := int64(len(hash)+len(bz)) - o2.(Object).GetObjectInfo().LastObjectSize
	// make store op log entry
	if ds.opslog != nil {
		obj := o2.(Object)
		if oo.GetIsNewReal() {
			obj.GetObjectInfo().LastObjectSize += diff
			fmt.Fprintf(ds.opslog, "c[%v](%d)=%s\n",
				obj.GetObjectID(),
				diff,
				prettyJSON(amino.MustMarshalJSON(obj)))
		} else {
			old := ds.loadForLog(oid)
			old.SetHash(ValueHash{})

			// need to do this marshal+unmarshal dance to ensure we get as close
			// as possible the output of amino that we'll get of MustMarshalJSON(old),
			// ie. empty slices should be null, not [].
			var pureNew Object
			amino.MustUnmarshalAny(amino.MustMarshalAny(obj), &pureNew)
			pureNew.GetObjectInfo().LastObjectSize = obj.GetObjectInfo().LastObjectSize

			s, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:       difflib.SplitLines(string(prettyJSON(amino.MustMarshalJSON(old)))),
				B:       difflib.SplitLines(string(prettyJSON(amino.MustMarshalJSON(pureNew)))),
				Context: 3,
			})
			if err != nil {
				panic(err)
			}
			if s == "" {
				fmt.Fprintf(ds.opslog, "u[%v]=(noop)\n", obj.GetObjectID())
			} else {
				s = "    " + strings.TrimSpace(strings.ReplaceAll(s, "\n", "\n    "))
				fmt.Fprintf(ds.opslog, "u[%v](%d)=\n%s\n", obj.GetObjectID(), diff, s)
			}
		}
	}
	// save bytes to backend.
	if ds.baseStore != nil {
		key := backendObjectKey(oid)
		hashbz := make([]byte, len(hash)+len(bz))
		copy(hashbz, hash.Bytes())
		copy(hashbz[HashSize:], bz)
		ds.baseStore.Set(ds.gctx, []byte(key), hashbz)
		size = len(hashbz)
		oo.GetObjectInfo().LastObjectSize = int64(size)
	}
	// save object to cache.
	if debug {
		if !oid.IsFinalized() {
			panic("object id must be finalized at SetObject")
		}
		if oo2, exists := ds.cacheObjects[oid]; exists {
			if oo != oo2 {
				panic(fmt.Sprintf(
					"duplicate object: set %s (oid: %s) but %s (oid %s) already exists",
					oo.String(), oid.String(), oo2.String(), oo2.GetObjectID().String()))
			}
		}
	}
	ds.cacheObjects[oid] = oo
	// if escaped, add hash to iavl.
	if oo.GetIsEscaped() && ds.iavlStore != nil {
		var key, value []byte
		key = []byte(oid.String())
		value = hash.Bytes()
		ds.iavlStore.Set(ds.gctx, key, value)
		if trace.StoreGasEnabled {
			trace.Store("IAVL_SET_ESCAPED", 0, key, len(value), "none")
		}
	}
	return diff
}

func (ds *defaultStore) loadForLog(oid ObjectID) Object {
	// Pass nil gctx: this read exists only to render the opslog diff,
	// which is a test-only observability feature (filetest harness).
	// Charging gas here would make tx gas depend on whether opslog is
	// enabled, breaking determinism if the feature is ever wired into
	// production diagnostics.
	key := backendObjectKey(oid)
	hashbz := ds.baseStore.Get(nil, []byte(key))
	if hashbz == nil {
		return nil
	}
	bz := hashbz[HashSize:]
	var oo Object
	amino.MustUnmarshal(bz, &oo)
	oo.GetObjectInfo().LastObjectSize = int64(len(hashbz))
	return oo
}

func (ds *defaultStore) DelObject(oo Object) int64 {
	if bm.Enabled {
		old := bm.StartStore(bm.StoreDeleteObject)
		defer func() { bm.StopStore(bm.StoreDeleteObject, old, 0) }()
	}
	// Storage I/O gas for delete is charged by cache.Store via GasContext.
	// No amino compute gas — delete doesn't marshal/unmarshal.
	oid := oo.GetObjectID()
	size := oo.GetObjectInfo().LastObjectSize
	// delete from cache. Lock-step evict cacheRealms when the object
	// being deleted is a PackageValue: keeps the
	// pv.Realm == cacheRealms[pid] invariant.
	delete(ds.cacheObjects, oid)
	if _, isPV := oo.(*PackageValue); isPV {
		delete(ds.cacheRealms, oid.PkgID)
	}
	// delete from backend.
	if ds.baseStore != nil {
		key := backendObjectKey(oid)
		ds.baseStore.Delete(ds.gctx, []byte(key))
	}
	// delete escaped hash from iavl.
	if oo.GetIsEscaped() && ds.iavlStore != nil {
		key := []byte(oid.String())
		ds.iavlStore.Delete(ds.gctx, key)
		if trace.StoreGasEnabled {
			trace.Store("IAVL_DEL_ESCAPED", 0, key, 0, "none")
		}
	}
	// make realm op log entry
	if ds.opslog != nil {
		fmt.Fprintf(ds.opslog, "d[%v](%d)\n", oo.GetObjectID(), -size)
	}
	return size
}

// NOTE: not used quite yet.
// NOTE: The implementation matches that of GetObject() in anticipation of what
// the persistent type system might work like.
func (ds *defaultStore) GetType(tid TypeID) Type {
	tt := ds.GetTypeSafe(tid)
	if tt == nil {
		panic(fmt.Sprintf("unexpected type with id %s", tid.String()))
	}
	return tt
}

func (ds *defaultStore) GetTypeSafe(tid TypeID) Type {
	// check cache.
	if tt, exists := ds.cacheTypes[tid]; exists {
		return tt
	}

	// check backend.
	if ds.baseStore != nil {
		key := backendTypeKey(tid)
		bz := ds.stdlibKeyBytes[key] // from byte cache — no I/O gas
		if bz == nil {
			bz = ds.baseStore.Get(ds.gctx, []byte(key)) // from store — charges I/O gas
		}
		if bz != nil {
			gas := overflow.Mulp(ds.gasConfig.GasAminoDecode, store.Gas(len(bz)))
			ds.consumeGas(gas, GasAminoDecodeDesc)
			if trace.StoreGasEnabled {
				trace.Store("DECODE_TYPE", gas, []byte(key), len(bz), "none")
			}
			cacheSum := sha256.Sum256(bz)
			var tt Type
			if val, ok := ds.aminoCache.Get(cacheSum[:]); ok {
				tt = copyTypeWithRefs(val)
			} else {
				amino.MustUnmarshal(bz, &tt)
				// len(bz) is not the proper cost of tt, but is good enough
				ds.aminoCache.Set(cacheSum[:], copyTypeWithRefs(tt), int64(len(bz)))
			}
			if debugAssert {
				if tt.TypeID() != tid {
					panic(fmt.Sprintf("unexpected type id: expected %v but got %v",
						tid, tt.TypeID()))
				}
			}
			// set in cache.
			ds.cacheTypes[tid] = tt
			// after setting in cache, fill tt.
			fillType(ds, tt)
			return tt
		}
	}
	return nil
}

func (ds *defaultStore) SetCacheType(tt Type) {
	tid := tt.TypeID()
	if tt2, exists := ds.cacheTypes[tid]; exists {
		if tt != tt2 {
			panic(fmt.Sprintf("cannot re-register %q with different type", tid))
		}
		// else, already set.
	} else {
		ds.cacheTypes[tid] = tt
	}
}

func (ds *defaultStore) SetType(tt Type) {
	var size int
	if bm.Enabled {
		old := bm.StartStore(bm.StoreSetType)
		defer func() { bm.StopStore(bm.StoreSetType, old, size) }()
	}
	tid := tt.TypeID()
	// Idempotent: if this TypeID is already known in-cache, do nothing.
	// The cache is populated either by a previous SetType (which also
	// persisted to backend) or by loading from backend on GetType — either
	// way the backend already has the canonical entry.
	if _, exists := ds.cacheTypes[tid]; exists {
		return
	}
	// save type to backend.
	if ds.baseStore != nil {
		key := backendTypeKey(tid)
		tcopy := copyTypeWithRefs(tt)
		bz := amino.MustMarshalAny(tcopy)
		cacheSum := sha256.Sum256(bz)
		ds.aminoCache.Set(cacheSum[:], tcopy, int64(len(bz)))
		gas := overflow.Mulp(ds.gasConfig.GasAminoEncode, store.Gas(len(bz)))
		ds.consumeGas(gas, GasAminoEncodeDesc)
		if trace.StoreGasEnabled {
			trace.Store("ENCODE_TYPE", gas, []byte(key), len(bz), "none")
		}
		ds.baseStore.Set(ds.gctx, []byte(key), bz)
		size = len(bz)
	}
	// save type to cache.
	ds.cacheTypes[tid] = tt
}

// assertNoDanglingLocalTypeRef (debugAssert only) walks a persist-copy
// produced by copyValueWithRefs and panics if it references a function-local
// declared type (RefType whose TypeID carries a location) that is not in
// cacheTypes. See the SetObject call site for the invariant.
func (ds *defaultStore) assertNoDanglingLocalTypeRef(val Value) {
	switch cv := val.(type) {
	case *ArrayValue:
		for i := range cv.List {
			ds.assertNoDanglingLocalTypeRefTV(&cv.List[i])
		}
	case *StructValue:
		for i := range cv.Fields {
			ds.assertNoDanglingLocalTypeRefTV(&cv.Fields[i])
		}
	case *MapValue:
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			ds.assertNoDanglingLocalTypeRefTV(&cur.Key)
			ds.assertNoDanglingLocalTypeRefTV(&cur.Value)
		}
	case *FuncValue:
		ds.assertNoDanglingLocalTypeRefType(cv.Type)
		for i := range cv.Captures {
			ds.assertNoDanglingLocalTypeRefTV(&cv.Captures[i])
		}
	case *BoundMethodValue:
		ds.assertNoDanglingLocalTypeRef(cv.Func)
		ds.assertNoDanglingLocalTypeRefTV(&cv.Receiver)
	case *Block:
		for i := range cv.Values {
			ds.assertNoDanglingLocalTypeRefTV(&cv.Values[i])
		}
		ds.assertNoDanglingLocalTypeRefTV(&cv.Blank)
	case *HeapItemValue:
		ds.assertNoDanglingLocalTypeRefTV(&cv.Value)
	case TypeValue:
		ds.assertNoDanglingLocalTypeRefType(cv.Type)
	default:
		// Scalars carry no type refs; PointerValue/SliceValue bases and
		// RefValue children are separate objects with their own SetObject.
	}
}

func (ds *defaultStore) assertNoDanglingLocalTypeRefTV(tv *TypedValue) {
	ds.assertNoDanglingLocalTypeRefType(tv.T)
	ds.assertNoDanglingLocalTypeRef(tv.V)
}

func (ds *defaultStore) assertNoDanglingLocalTypeRefType(t Type) {
	switch ct := t.(type) {
	case RefType:
		// RefTypes only wrap declared types ("path.Name" or "path[loc].Name",
		// see refOrCopyType), so a bracket identifies a function-local type.
		if strings.Contains(ct.ID.String(), "[") {
			if _, exists := ds.cacheTypes[ct.ID]; exists {
				return
			}
			// Not in this transaction's cache: the type must already be in
			// the backend, written at addpkg by saveFuncLocalTypes. Raw key
			// probe (not GetTypeSafe) so the debug-only assert has no amino
			// decode cost and no cache side effects.
			if ds.baseStore != nil {
				key := backendTypeKey(ct.ID)
				if ds.baseStore.Get(ds.gctx, []byte(key)) != nil {
					return
				}
			}
			panic(fmt.Sprintf(
				"dangling function-local type ref %s in persisted value", ct.ID))
		}
	case *DeclaredType:
		ds.assertNoDanglingLocalTypeRefType(ct.Base)
	case FieldType:
		ds.assertNoDanglingLocalTypeRefType(ct.Type)
	case *FuncType:
		for _, param := range ct.Params {
			ds.assertNoDanglingLocalTypeRefType(param)
		}
		for _, result := range ct.Results {
			ds.assertNoDanglingLocalTypeRefType(result)
		}
	case *SliceType, *ArrayType, *PointerType:
		ds.assertNoDanglingLocalTypeRefType(ct.Elem())
	case *MapType:
		ds.assertNoDanglingLocalTypeRefType(ct.Key)
		ds.assertNoDanglingLocalTypeRefType(ct.Value)
	case *tupleType:
		for _, et := range ct.Elts {
			ds.assertNoDanglingLocalTypeRefType(et)
		}
	case *InterfaceType:
		for _, method := range ct.Methods {
			ds.assertNoDanglingLocalTypeRefType(method)
		}
	case *StructType:
		for _, field := range ct.Fields {
			ds.assertNoDanglingLocalTypeRefType(field)
		}
	default:
		// nil, primitives, TypeType, PackageType, blockType, heapItemType.
	}
}

// Convenience
func (ds *defaultStore) GetPackageNode(pkgPath string) *PackageNode {
	return ds.GetBlockNode(PackageNodeLocation(pkgPath)).(*PackageNode)
}

func (ds *defaultStore) GetBlockNode(loc Location) BlockNode {
	bn := ds.GetBlockNodeSafe(loc)
	if bn == nil {
		panic(fmt.Sprintf("unexpected node with location %s", loc.String()))
	}
	return bn
}

func (ds *defaultStore) GetBlockNodeSafe(loc Location) BlockNode {
	var size int
	if bm.Enabled {
		old := bm.StartStore(bm.StoreGetBlockNode)
		defer func() { bm.StopStore(bm.StoreGetBlockNode, old, size) }()
	}
	// check cache.
	if bn, exists := ds.cacheNodes.Get(loc); exists {
		return bn
	}
	// check backend.
	if ds.baseStore != nil {
		key := backendNodeKey(loc)
		bz := ds.baseStore.Get(ds.gctx, []byte(key))
		if bz != nil {
			var bn BlockNode
			amino.MustUnmarshal(bz, &bn)
			size = len(bz)
			if debugAssert {
				if bn.GetLocation() != loc {
					panic(fmt.Sprintf("unexpected node location: expected %v but got %v",
						loc, bn.GetLocation()))
				}
			}
			ds.cacheNodes.Set(loc, bn)
			return bn
		}
	}
	return nil
}

func (ds *defaultStore) SetBlockNode(bn BlockNode) {
	loc := bn.GetLocation()
	if loc.IsZero() {
		panic("unexpected zero location in blocknode")
	}
	// save node to backend.
	// if ds.baseStore != nil {
	// TODO: implement copyValueWithRefs() for Nodes.
	// key := backendNodeKey(loc)
	// ds.backend.Set([]byte(key), bz)
	// }
	// save node to cache.
	ds.cacheNodes.Set(loc, bn)
	// XXX duplicate?
	// XXX
}

func (ds *defaultStore) NumMemPackages() int64 {
	ctrkey := []byte(backendPackageIndexCtrKey())
	ctrbz := ds.baseStore.Get(ds.gctx, ctrkey)
	if ctrbz == nil {
		return 0
	} else {
		ctr, err := strconv.Atoi(string(ctrbz))
		if err != nil {
			panic(err)
		}
		return int64(ctr)
	}
}

func (ds *defaultStore) incGetPackageIndexCounter() uint64 {
	ctrkey := []byte(backendPackageIndexCtrKey())
	ctrbz := ds.baseStore.Get(ds.gctx, ctrkey)
	if ctrbz == nil {
		nextbz := strconv.Itoa(1)
		ds.baseStore.Set(ds.gctx, ctrkey, []byte(nextbz))
		return 1
	} else {
		ctr, err := strconv.Atoi(string(ctrbz))
		if err != nil {
			panic(err)
		}
		nextbz := strconv.Itoa(ctr + 1)
		ds.baseStore.Set(ds.gctx, ctrkey, []byte(nextbz))
		return uint64(ctr) + 1
	}
}

// mptype is passed in as a redundant parameter as convenience to assert that
// mpkg.Type is what is expected.
// If MPAnyAll, mpkg may be either MPStdlibAll or MPProdAll, and likewise for
// MPAnyProd and MPAnyTest.
// MPFiletests are not allowed, as they are currently only read from disk (e.g.
// test/files). However, MP*All may include filetests files.
//
// MP*All packages are stored as two keys (a prod blob plus a #allbutprod
// sibling) and the writes are conditional, so this is NOT a full replace across
// both keys. Re-adding an MP*All package at an existing path (e.g. a private
// redeploy) MUST call DeleteMemPackage(path) first, or a stale sibling/prod blob
// can survive. The keeper's AddPackage does this; new MP*All re-add callers must
// too. (Already-filtered non-All types are stored under a single key and
// overwrite cleanly.)
func (ds *defaultStore) AddMemPackage(mpkg *std.MemPackage, mptype MemPackageType) {
	var size int
	if bm.Enabled {
		old := bm.StartStore(bm.StoreAddMemPackage)
		defer func() { bm.StopStore(bm.StoreAddMemPackage, old, size) }()
	}
	mpkgtype := mpkg.Type.(MemPackageType)
	if !mpkgtype.IsStorable() {
		panic(fmt.Sprintf("mempackage type is not storable: %v", mpkgtype))
	}
	mptype = mptype.Decide(mpkg.Path)
	if mpkgtype != mptype {
		panic(fmt.Sprintf("unexpected mempackage type: expected %v but got %v", mptype, mpkgtype))
	}
	err := ValidateMemPackageAny(mpkg)
	if err != nil {
		panic(fmt.Errorf("invalid mempackage: %w", err))
	}
	ctr := ds.incGetPackageIndexCounter()
	idxkey := []byte(backendPackageIndexKey(ctr))
	ds.baseStore.Set(ds.gctx, idxkey, []byte(mpkg.Path))
	pathkey := []byte(backendPackagePathKey(mpkg.Path))
	// MP*All packages are split for storage: production files under pathkey
	// (the import/preprocess hot path, so its read/decode gas is charged on
	// prod bytes only), and the remaining test/filetest files under a sibling
	// "#allbutprod" key read only by query paths (see GetMemPackageAll). Other
	// (already-filtered) types are stored whole under pathkey as before.
	if mpkgtype.IsAll() {
		prod, allButProd := splitProdAllButProd(mpkg)
		// prod is nil for a package with no production .gno files (e.g. an
		// xxx_test-only package): it has no prod blob, and readers treat a
		// missing prod blob as nil. splitProdAllButProd folds that package's
		// non-.gno files into the #allbutprod sibling so GetMemPackageAll can
		// reconstruct the full package losslessly.
		if prod != nil {
			size += ds.setMemPackageBlob(pathkey, prod)
		}
		if len(allButProd.Files) > 0 {
			size += ds.setMemPackageBlob([]byte(backendPackageAllButProdKey(mpkg.Path)), allButProd)
		}
	} else {
		size += ds.setMemPackageBlob(pathkey, mpkg)
	}
}

// DeleteMemPackage removes both the production blob (pkg:<path>) and the
// #allbutprod sibling for path. It is a no-op for keys that do not exist. Used
// before a private-package redeploy: AddMemPackage stores an MP*All package as
// two keys, and its conditional writes are not a full replace across both, so a
// stale sibling (or, for a now-prod-less package, a stale prod blob) could
// otherwise survive a re-add and be served by GetMemPackage/GetMemPackageAll.
func (ds *defaultStore) DeleteMemPackage(path string) {
	ds.iavlStore.Delete(ds.gctx, []byte(backendPackagePathKey(path)))
	ds.iavlStore.Delete(ds.gctx, []byte(backendPackageAllButProdKey(path)))
}

// setMemPackageBlob amino-marshals mpkg, charges encode gas, writes it under key
// in the iavl store, and returns the encoded byte length.
func (ds *defaultStore) setMemPackageBlob(key []byte, mpkg *std.MemPackage) int {
	bz := amino.MustMarshal(mpkg)
	gas := overflow.Mulp(ds.gasConfig.GasAminoEncode, store.Gas(len(bz)))
	ds.consumeGas(gas, GasAminoEncodeDesc)
	if trace.StoreGasEnabled {
		trace.Store("ENCODE_MEMPKG", gas, key, len(bz), "none")
	}
	ds.iavlStore.Set(ds.gctx, key, bz)
	if trace.StoreGasEnabled {
		trace.Store("IAVL_SET_MEMPKG", 0, key, len(bz), "none")
	}
	return len(bz)
}

// splitProdAllButProd partitions an MP*All mempackage into its production blob
// and the complement ("all but prod") sibling so that prod ∪ allButProd ==
// mpkg.Files exactly, with no overlap and no drops.
//
// The production blob holds the importable subset: non-.gno files plus non-test
// .gno files, typed MP*Prod. However, a package with no production .gno files
// (e.g. an xxx_test-only package) has no valid prod blob — an empty mempackage
// fails validation, and readers treat a missing prod blob as nil — so in that
// case prod is returned nil and ALL of its files (including non-.gno files such
// as gnomod.toml, LICENSE, README, *.md, *.toml) are folded into the complement
// instead, otherwise they would be silently dropped from storage.
//
// The complement keeps the package's Name/Path/Info and the original MP*All type
// as an inert sentinel: it is written and read only by the store and never
// enters the MemPackageType dispatch paths.
func splitProdAllButProd(mpkg *std.MemPackage) (prod, allButProd *std.MemPackage) {
	prod = MPFProd.FilterMemPackage(mpkg)
	allButProd = &std.MemPackage{
		Name: mpkg.Name,
		Path: mpkg.Path,
		Info: mpkg.Info,
		Type: mpkg.Type,
	}
	// If there are no production .gno files, the prod blob will not be written
	// (it would fail validation). Fold every non-prod-.gno file into the
	// complement so reconstruction stays lossless.
	prodSkipped := prod.IsEmpty()
	if prodSkipped {
		prod = nil
	}
	for _, mfile := range mpkg.Files {
		if IsTestFile(mfile.Name) || (prodSkipped && !strings.HasSuffix(mfile.Name, ".gno")) {
			allButProd.Files = append(allButProd.Files, mfile.Copy())
		}
	}
	allButProd.Sort()
	return prod, allButProd
}

// GetMemPackage retrieves the MemPackage at the given path.
// It returns nil if the package could not be found.
func (ds *defaultStore) GetMemPackage(path string) *std.MemPackage {
	return ds.getMemPackage(path, false)
}

func (ds *defaultStore) getMemPackage(path string, isRetry bool) *std.MemPackage {
	var size int
	if bm.Enabled {
		old := bm.StartStore(bm.StoreGetMemPackage)
		defer func() { bm.StopStore(bm.StoreGetMemPackage, old, size) }()
	}
	pathkey := []byte(backendPackagePathKey(path))
	bz := ds.iavlStore.Get(ds.gctx, pathkey)
	if trace.StoreGasEnabled {
		trace.Store("IAVL_GET_MEMPKG", 0, pathkey, len(bz), "none")
	}
	if bz == nil {
		// If this is the first try, attempt using GetPackage to retrieve the
		// package, first. GetPackage can leverage pkgGetter, which in most
		// implementations works by running Machine.RunMemPackage with save = true,
		// which would add the package to the store after running.
		// Some packages may never be persisted, thus why we only attempt this twice.
		if !isRetry && ds.pkgGetter != nil {
			if pv := ds.GetPackage(path, false); pv != nil {
				return ds.getMemPackage(path, true)
			}
		}
		return nil
	}
	gas := overflow.Mulp(ds.gasConfig.GasAminoDecode, store.Gas(len(bz)))
	ds.consumeGas(gas, GasAminoDecodeDesc)
	if trace.StoreGasEnabled {
		trace.Store("DECODE_MEMPKG", gas, pathkey, len(bz), "none")
	}

	var mpkg *std.MemPackage
	amino.MustUnmarshal(bz, &mpkg)
	size = len(bz)
	return mpkg
}

// getMemPackageAllButProd reads and decodes the "#allbutprod" sibling blob (the
// test/filetest files) for path, or nil if there is none. Charges decode gas.
func (ds *defaultStore) getMemPackageAllButProd(path string) *std.MemPackage {
	bz := ds.iavlStore.Get(ds.gctx, []byte(backendPackageAllButProdKey(path)))
	if bz == nil {
		return nil
	}
	gas := overflow.Mulp(ds.gasConfig.GasAminoDecode, store.Gas(len(bz)))
	ds.consumeGas(gas, GasAminoDecodeDesc)
	var mpkg *std.MemPackage
	amino.MustUnmarshal(bz, &mpkg)
	return mpkg
}

// GetMemPackageAll retrieves the complete MemPackage at path, including test and
// filetest files, by merging the production blob with its "#allbutprod" sibling.
// It returns nil if the package does not exist. The import/run hot path uses the
// prod-only GetMemPackage; GetMemPackageAll is for query/tooling paths that must
// see test files (e.g. vm/qfile).
func (ds *defaultStore) GetMemPackageAll(path string) *std.MemPackage {
	prod := ds.GetMemPackage(path)
	allButProd := ds.getMemPackageAllButProd(path)
	if prod == nil && allButProd == nil {
		return nil
	}
	base := prod
	if base == nil {
		base = allButProd
	}
	merged := &std.MemPackage{
		Name: base.Name,
		Path: base.Path,
		Info: base.Info,
		Type: MPAnyAll.Decide(path), // MPUserAll or MPStdlibAll.
	}
	if prod != nil {
		merged.Files = append(merged.Files, prod.Files...)
	}
	if allButProd != nil {
		merged.Files = append(merged.Files, allButProd.Files...)
	}
	merged.Sort()
	return merged
}

// GetMemFile retrieves the MemFile with the given name, contained in the
// MemPackage at the given path. It returns nil if the file or the package
// do not exist. It consults the full package (prod + test files) so that
// test/filetest files remain retrievable.
func (ds *defaultStore) GetMemFile(path string, name string) *std.MemFile {
	mpkg := ds.GetMemPackageAll(path)
	if mpkg == nil {
		return nil
	}
	memFile := mpkg.GetFile(name)
	return memFile
}

// FindPathsByPrefix retrieves all paths starting with the given prefix.
func (ds *defaultStore) FindPathsByPrefix(prefix string) iter.Seq[string] {
	// If prefix is empty range every package
	startKey := []byte(backendPackageGlobalPath("\x00"))
	endKey := []byte(backendPackageGlobalPath("\xFF"))
	if len(prefix) > 0 {
		startKey = []byte(backendPackageGlobalPath(prefix))
		// Create endkey by incrementing last byte of startkey
		endKey = slices.Clone(startKey)
		endKey[len(endKey)-1]++
	}

	return func(yield func(string) bool) {
		iter := ds.iavlStore.Iterator(ds.gctx, startKey, endKey)
		defer iter.Close()

		var last string
		var hasLast bool
		for ; iter.Valid(); iter.Next() {
			key := string(iter.Key())
			// A package's prod key (pkg:<path>) and its #allbutprod sibling map
			// to the same path. Strip the sibling suffix and de-dup so each
			// package is yielded exactly once — including an empty-prod package
			// (only a sibling key). Prod and sibling keys for a path are
			// adjacent in iavl order, so de-dup against the previous suffices.
			key = strings.TrimSuffix(key, "#allbutprod")
			// A prefix containing '#' (impossible in a valid package path,
			// but reachable from raw query input, e.g. vm/qpaths) can range
			// over a sibling key whose trimmed form no longer carries the
			// requested prefix; don't yield such a path. Compared in key
			// space (against startKey): stdlib paths decode without their
			// "_/" key marker, so path-space comparison would wrongly drop
			// legitimate stdlib matches.
			if len(prefix) > 0 && !strings.HasPrefix(key, string(startKey)) {
				continue
			}
			path := decodeBackendPackagePathKey(key)
			if hasLast && path == last {
				continue
			}
			last, hasLast = path, true
			if !yield(path) {
				return
			}
		}
	}
}

// IterMemPackage yields each indexed package's PROD mempackage in index
// order, skipping prod-less packages. See the Store interface doc.
func (ds *defaultStore) IterMemPackage() <-chan *std.MemPackage {
	ctrkey := []byte(backendPackageIndexCtrKey())
	ctrbz := ds.baseStore.Get(ds.gctx, ctrkey)
	if ctrbz == nil {
		return nil
	} else {
		ctr, err := strconv.Atoi(string(ctrbz))
		if err != nil {
			panic(err)
		}
		ch := make(chan *std.MemPackage)
		go func() {
			for i := uint64(1); i <= uint64(ctr); i++ {
				idxkey := []byte(backendPackageIndexKey(i))
				path := ds.baseStore.Get(ds.gctx, idxkey)
				if path == nil {
					panic(fmt.Sprintf(
						"missing package index %d", i))
				}
				mpkg := ds.GetMemPackage(string(path))
				if mpkg == nil {
					// Prod-less package (e.g. xxx_test-only): no prod
					// blob to yield. On-chain this is unreachable — the
					// vm keeper rejects prod-less packages at AddPackage
					// — so this skip is defensive, for non-chain stores.
					continue
				}
				ch <- mpkg
			}
			close(ch)
		}()
		return ch
	}
}

func (ds *defaultStore) consumeGas(gas int64, descriptor string) {
	// In the tests, the defaultStore may not set the gas meter.
	if ds.gasMeter != nil {
		ds.gasMeter.ConsumeGas(gas, descriptor)
	}
}

// PopulateStdlibCache scans the baseStore for all stdlib object keys
// and caches their raw bytes. Called once at node start (from
// VMKeeper.Initialize on restart, and after loadStdlibs at genesis).
// Each stdlib package's objects are found via a prefix iterator on
// "oid:<pkgid_hex>:".
func (ds *defaultStore) PopulateStdlibCache(paths []string) {
	ds.populateStdlibCache(paths, ds.baseStore)
}

func (ds *defaultStore) PopulateStdlibCacheFrom(paths []string, baseStore store.Store) {
	ds.populateStdlibCache(paths, baseStore)
}

func (ds *defaultStore) populateStdlibCache(paths []string, baseStore store.Store) {
	for _, path := range paths {
		// Cache object bytes (oid:<pkgid_hex>:*).
		pid := PkgIDFromPkgPath(path)
		prefix := "oid:" + hex.EncodeToString(pid.Hashlet[:]) + ":"
		start := []byte(prefix)
		endPrefix := prefix[:len(prefix)-1] + string(rune(prefix[len(prefix)-1]+1))
		end := []byte(endPrefix)
		// nil gctx: stdlib-cache population runs at node startup only,
		// never under a tx or query meter — must remain gas-free.
		iter := baseStore.Iterator(nil, start, end)
		for ; iter.Valid(); iter.Next() {
			key := string(iter.Key())
			val := iter.Value()
			bz := make([]byte, len(val))
			copy(bz, val)
			ds.stdlibKeyBytes[key] = bz
		}
		iter.Close()

		// Cache type bytes (tid:<path>.*).
		// Type keys use the full package path (e.g., "tid:strings.Builder").
		// Range: [tid:<path>. , tid:<path>/) captures all package-level types.
		tprefix := "tid:" + path + "."
		tstart := []byte(tprefix)
		tend := []byte("tid:" + path + "/")
		// nil gctx: see comment above.
		titer := baseStore.Iterator(nil, tstart, tend)
		for ; titer.Valid(); titer.Next() {
			key := string(titer.Key())
			val := titer.Value()
			bz := make([]byte, len(val))
			copy(bz, val)
			ds.stdlibKeyBytes[key] = bz
		}
		titer.Close()
	}
}

// It resturns storage changes per realm within message
func (ds *defaultStore) RealmStorageDiffs() StorageDiffs {
	return ds.realmStorageDiffs
}

// Unstable.
// This function is used to clear the object cache every transaction.
// It also sets a new allocator.
func (ds *defaultStore) ClearObjectCache() {
	ds.alloc.Reset()
	ds.cacheObjects = make(map[ObjectID]Object) // new cache.
	// Lock-step reset cacheRealms.
	ds.cacheRealms = make(map[PkgID]*Realm)
	ds.realmStorageDiffs = make(StorageDiffs)
	ds.opslog = nil // new ops log.
	ds.SetCachePackage(Uverse())
}

func (ds *defaultStore) GarbageCollectObjectCache(gcCycle int64) {
	for objId, obj := range ds.cacheObjects {
		// Skip .uverse packages.
		if pv, ok := obj.(*PackageValue); ok && pv.PkgPath == ".uverse" {
			continue
		}
		if obj.GetLastGCCycle() < gcCycle {
			delete(ds.cacheObjects, objId)
			// Lock-step evict cacheRealms when a PackageValue is
			// evicted. Falls back to PkgID derivation from PkgPath
			// if the PV's PkgID hasn't been stamped yet.
			if pv, isPV := obj.(*PackageValue); isPV {
				pid := objId.PkgID
				if pid.IsZero() {
					pid = PkgIDFromPkgPath(pv.PkgPath)
				}
				delete(ds.cacheRealms, pid)
			}
		}
	}
}

func (ds *defaultStore) SetNativeResolver(ns NativeResolver) {
	ds.nativeResolver = ns
}

func (ds *defaultStore) GetNative(pkgPath string, name Name) func(m *Machine) {
	if ds.nativeResolver != nil {
		return ds.nativeResolver(pkgPath, name)
	}
	return nil
}

// Set to nil to disable.
func (ds *defaultStore) SetLogStoreOps(buf io.Writer) {
	if enabled.Load() {
		ds.opslog = buf
	} else {
		ds.opslog = nil
	}
}

func (ds *defaultStore) LogFinalizeRealm(rlmpath string) {
	if ds.opslog != nil {
		fmt.Fprintf(ds.opslog, "finalizerealm[%q]\n", rlmpath)
	}
}

// for debugging
func (ds *defaultStore) Print() {
	fmt.Println(colors.Yellow("//----------------------------------------"))
	fmt.Println(colors.Green("defaultStore:baseStore..."))
	utils.Print(ds.baseStore)
	fmt.Println(colors.Yellow("//----------------------------------------"))
	fmt.Println(colors.Green("defaultStore:iavlStore..."))
	utils.Print(ds.iavlStore)
	fmt.Println(colors.Yellow("//----------------------------------------"))
	fmt.Println(colors.Green("defaultStore:cacheTypes..."))
	for tid, typ := range ds.cacheTypes {
		fmt.Printf("- %v: %v\n", tid,
			stringz.TrimN(fmt.Sprintf("%v", typ), 50))
	}
	fmt.Println(colors.Yellow("//----------------------------------------"))
	fmt.Println(colors.Green("defaultStore:cacheNodes..."))
	ds.cacheNodes.Iterate()(func(loc Location, bn BlockNode) bool {
		fmt.Printf("- %v: %v\n", loc,
			stringz.TrimN(fmt.Sprintf("%v", bn), 50))
		return true
	})
	fmt.Println(colors.Red("//----------------------------------------"))
}

// ----------------------------------------
// backend keys

func backendObjectKey(oid ObjectID) string {
	return "oid:" + oid.String()
}

// oid: associated package value object id.
func backendRealmKey(oid ObjectID) string {
	return "oid:" + oid.String() + "#realm"
}

func backendTypeKey(tid TypeID) string {
	return "tid:" + tid.String()
}

func backendNodeKey(loc Location) string {
	return "node:" + loc.String()
}

func backendPackageIndexCtrKey() string {
	return "pkgidx:counter"
}

func backendPackageIndexKey(index uint64) string {
	return fmt.Sprintf("pkgidx:%020d", index)
}

// We need to prefix stdlibs path with `_` to maitain them lexicographically
// ordered with domain path
func backendPackagePathKey(path string) string {
	if IsStdlib(path) {
		return backendPackageStdlibPath(path)
	}

	return backendPackageGlobalPath(path)
}

func backendPackageStdlibPath(path string) string { return "pkg:_/" + path }

func backendPackageGlobalPath(path string) string { return "pkg:" + path }

// backendPackageAllButProdKey returns the sibling key holding a package's
// test/filetest files (everything in an MP*All package but its production
// subset). It suffixes the package path key with "#allbutprod"; "#" cannot
// appear in a valid package path, so it never collides with a real package key.
func backendPackageAllButProdKey(path string) string {
	return backendPackagePathKey(path) + "#allbutprod"
}

func decodeBackendPackagePathKey(key string) string {
	path := strings.TrimPrefix(key, "pkg:")
	return strings.TrimPrefix(path, "_/")
}

// ----------------------------------------
// builtin types and packages

func InitStoreCaches(store Store) {
	uverse := UverseNode()
	for _, tv := range uverse.GetStaticBlock().Values {
		if tv.T != nil && tv.T.Kind() == TypeKind {
			uverseType := tv.GetType()
			switch uverseType.(type) {
			case PrimitiveType:
				store.SetCacheType(uverseType)
			case *DeclaredType:
				store.SetCacheType(uverseType)
			}
		}
	}
	store.SetCachePackage(Uverse())
}
