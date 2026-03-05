package gnolang

import (
	"crypto/sha256"
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

// Store is the central interface that specifies the communications between the
// GnoVM and the underlying data store; currently, generally the gno.land
// blockchain, or the file system.
type Store interface {
	// STABLE
	BeginTransaction(baseStore, iavlStore store.Store, gasMeter store.GasMeter) TransactionStore
	GetPackageGetter() PackageGetter
	SetPackageGetter(PackageGetter)
	GetPackage(pkgPath string, isImport bool) *PackageValue
	SetCachePackage(*PackageValue)
	GetPackageRealm(pkgPath string) *Realm
	SetPackageRealm(*Realm)
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
	RealmStorageDiffs() map[string]int64 // returns storage changes per realm within the message

	// UNSTABLE
	GetAllocator() *Allocator
	SetAllocator(alloc *Allocator)
	NumMemPackages() int64
	// Upon restart, all packages will be re-preprocessed; This
	// loads BlockNodes and Types onto the store for persistence
	// version 1.
	AddMemPackage(mpkg *std.MemPackage, mptype MemPackageType)
	GetMemPackage(path string) *std.MemPackage
	GetMemFile(path string, name string) *std.MemFile
	FindPathsByPrefix(prefix string) iter.Seq[string]
	IterMemPackage() <-chan *std.MemPackage
	ClearObjectCache() // run before processing a message
	GarbageCollectObjectCache(gcCycle int64)
	SetNativeResolver(NativeResolver)                     // for native functions
	GetNative(pkgPath string, name Name) func(m *Machine) // for native functions
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
	GasGetObjectDesc       = "GetObjectPerByte"
	GasSetObjectDesc       = "SetObjectPerByte"
	GasGetTypeDesc         = "GetTypePerByte"
	GasSetTypeDesc         = "SetTypePerByte"
	GasGetPackageRealmDesc = "GetPackageRealmPerByte"
	GasSetPackageRealmDesc = "SetPackageRealmPerByte"
	GasAddMemPackageDesc   = "AddMemPackagePerByte"
	GasGetMemPackageDesc   = "GetMemPackagePerByte"
	GasDeleteObjectDesc    = "DeleteObjectFlat"
)

// GasConfig defines gas cost for each operation on KVStores
type GasConfig struct {
	GasGetObject       int64
	GasSetObject       int64
	GasGetType         int64
	GasSetType         int64
	GasGetPackageRealm int64
	GasSetPackageRealm int64
	GasAddMemPackage   int64
	GasGetMemPackage   int64
	GasDeleteObject    int64
}

// DefaultGasConfig returns a default gas config for KVStores.
func DefaultGasConfig() GasConfig {
	return GasConfig{
		GasGetObject:       16,   // per byte cost
		GasSetObject:       16,   // per byte cost
		GasGetType:         5,    // per byte cost
		GasSetType:         52,   // per byte cost
		GasGetPackageRealm: 524,  // per byte cost
		GasSetPackageRealm: 524,  // per byte cost
		GasAddMemPackage:   8,    // per byte cost
		GasGetMemPackage:   8,    // per byte cost
		GasDeleteObject:    3715, // flat cost
	}
}

type defaultStore struct {
	// underlying stores used to keep data
	baseStore store.Store // for objects, types, nodes
	iavlStore store.Store // for escaped object hashes

	// transaction-scoped
	// cacheNodes is an actual store - BlockNodes are not stored in the underlying
	// DB and must be re-initialized using PreprocessAllFilesAndSaveBlockNodes.
	cacheObjects map[ObjectID]Object
	cacheTypes   map[TypeID]Type
	cacheNodes   txlog.Map[Location, BlockNode]
	alloc        *Allocator // for accounting for cached items

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
	gasMeter  store.GasMeter
	gasConfig GasConfig

	// realm storage changes on message level.
	realmStorageDiffs map[string]int64 // maps realm path to size diff
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
		cacheNodes:   txlog.GoMap[Location, BlockNode](map[Location]BlockNode{}),

		// reset at the message level
		realmStorageDiffs: make(map[string]int64),

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
func (ds *defaultStore) BeginTransaction(baseStore, iavlStore store.Store, gasMeter store.GasMeter) TransactionStore {
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
		alloc:        ds.alloc.Fork().Reset(),

		// store configuration
		pkgGetter:      ds.pkgGetter,
		nativeResolver: ds.nativeResolver,

		// gas meter
		gasMeter:   gasMeter,
		gasConfig:  ds.gasConfig,
		aminoCache: ds.aminoCache,

		// transient
		current: nil,
		opslog:  nil,
		// reset at the message level
		realmStorageDiffs: make(map[string]int64),
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

	iter := cachedBase.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		ds.baseStore.Set(iter.Key(), iter.Value())
	}
	iter = cachedIavl.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		ds.iavlStore.Set(iter.Key(), iter.Value())
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

// Some atomic operation.
func (ds *defaultStore) GetPackageRealm(pkgPath string) (rlm *Realm) {
	var size int
	if bm.StorageEnabled {
		bm.StartStore(bm.StoreGetPackageRealm)
		defer func() {
			bm.StopStore(size)
		}()
	}
	oid := ObjectIDFromPkgPath(pkgPath)
	key := backendRealmKey(oid)
	bz := ds.baseStore.Get([]byte(key))
	if bz == nil {
		return nil
	}
	gas := overflow.Mulp(ds.gasConfig.GasGetPackageRealm, store.Gas(len(bz)))
	ds.consumeGas(gas, GasGetPackageRealmDesc)
	amino.MustUnmarshal(bz, &rlm)
	size = len(bz)
	if debug {
		if rlm.ID != oid.PkgID {
			panic(fmt.Sprintf("unexpected realm id: expected %v but got %v",
				oid.PkgID, rlm.ID))
		}
	}
	return rlm
}

// An atomic operation to set the package realm info (id counter etc).
func (ds *defaultStore) SetPackageRealm(rlm *Realm) {
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}

	var size int
	if bm.StorageEnabled {
		bm.StartStore(bm.StoreSetPackageRealm)
		defer func() {
			bm.StopStore(size)
		}()
	}
	oid := ObjectIDFromPkgPath(rlm.Path)
	key := backendRealmKey(oid)
	bz := amino.MustMarshal(rlm)
	gas := overflow.Mulp(ds.gasConfig.GasSetPackageRealm, store.Gas(len(bz)))
	ds.consumeGas(gas, GasSetPackageRealmDesc)
	ds.baseStore.Set([]byte(key), bz)
	size = len(bz)
}

// NOTE: it can be use to retrieve a package by ObjectID, but
// does not consult the packageGetter, only lookup the cache/store.
// NOTE: current implementation behavior requires
// all []TypedValue types and TypeValue{} types to be
// loaded (non-ref) types.
func (ds *defaultStore) GetObject(oid ObjectID) Object {
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}
	oo := ds.GetObjectSafe(oid)
	if oo == nil {
		panic(fmt.Sprintf("unexpected object with id %s", oid.String()))
	}
	return oo
}

func (ds *defaultStore) GetObjectSafe(oid ObjectID) Object {
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}
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
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}

	var size int

	if bm.StorageEnabled {
		bm.StartStore(bm.StoreGetObject)
		defer func() {
			bm.StopStore(size)
		}()
	}
	key := backendObjectKey(oid)
	hashbz := ds.baseStore.Get([]byte(key))
	if hashbz != nil {
		size = len(hashbz)
		hash := hashbz[:HashSize]
		bz := hashbz[HashSize:]
		var oo Object
		gas := overflow.Mulp(ds.gasConfig.GasGetObject, store.Gas(len(bz)))
		ds.consumeGas(gas, GasGetObjectDesc)
		amino.MustUnmarshal(bz, &oo)
		if debug {
			debug.Printf("loadObjectSafe by oid: %v, type of oo: %v\n", oid, reflect.TypeOf(oo))
		}

		ds.alloc.Allocate(oo.GetShallowSize())
		// Alloc values other than shallow value,
		// RefValue, e.g. keep sync with copyValueWithRefs().
		AllocExpanded(ds.alloc, oo)

		if debug {
			if oo.GetObjectID() != oid {
				panic(fmt.Sprintf("unexpected object id: expected %v but got %v",
					oid, oo.GetObjectID()))
			}
		}
		oo.SetHash(ValueHash{NewHashlet(hash)})

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
	if pv.IsRealm() && pv.Realm == nil {
		rlm := ds.GetPackageRealm(pv.PkgPath)
		pv.Realm = rlm
	}
	// Rederive pv.fBlocksMap.
	pv.deriveFBlocksMap(ds)
}

func AllocExpanded(alloc *Allocator, val Value) {
	var size int64
	defer func() {
		alloc.Allocate(size)
	}()

	switch v := val.(type) {
	case *PackageValue:
		if _, ok := v.Block.(RefValue); ok {
			size += allocRefValue // .Block ref
		}

		// include RefValue size
		for _, fb := range v.FBlocks {
			if _, ok := fb.(RefValue); !ok {
				continue
			}
			size += allocRefValue
		}
	case *Block:
		for _, v := range v.Values {
			if _, ok := v.V.(RefValue); ok {
				size += allocRefValue
			}
		}

		if _, ok := v.Parent.(RefValue); ok {
			size += allocRefValue
		}
	case *ArrayValue:
		if v.Data == nil {
			for _, tv := range v.List {
				if _, ok := tv.V.(RefValue); ok {
					size += allocRefValue
				}
			}
		}
	case *StructValue:
		for _, tv := range v.Fields {
			if _, ok := tv.V.(RefValue); ok {
				size += allocRefValue
			}
		}
	case *MapValue:
		for cur := v.List.Head; cur != nil; cur = cur.Next {
			if _, ok := cur.Key.V.(RefValue); ok {
				size += allocRefValue
			}

			if _, ok := cur.Value.V.(RefValue); ok {
				size += allocRefValue
			}
		}
	case *BoundMethodValue:
		if _, ok := v.Receiver.V.(RefValue); ok {
			size += allocRefValue
		}
	case *HeapItemValue:
		if _, ok := v.Value.V.(RefValue); ok {
			size += allocRefValue
		}
	case RefValue:
		// do nothing
	case *PointerValue:
		if _, ok := v.Base.(RefValue); ok {
			size += allocRefValue
		}
	case *SliceValue:
		if _, ok := v.Base.(RefValue); ok {
			size += allocRefValue
		}
	case *FuncValue:
		for _, tv := range v.Captures {
			if _, ok := tv.V.(RefValue); ok {
				size += allocRefValue
			}
		}

		if _, ok := v.Parent.(RefValue); ok {
			size += allocRefValue
		}
	case StringValue:
		// do nothing
	case BigintValue:
		// do nothing
	case BigdecValue:
		// do nothing
	case DataByteValue:
		// do nothing
	case TypeValue:
		// do nothing
	}
}

// NOTE: unlike GetObject(), SetObject() is also used to persist updated
// package values.
func (ds *defaultStore) SetObject(oo Object) int64 {
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}
	var size int
	if bm.StorageEnabled {
		bm.StartStore(bm.StoreSetObject)
		defer func() {
			bm.StopStore(size)
		}()
	}
	oid := oo.GetObjectID()
	// replace children/fields with Ref.
	o2 := copyValueWithRefs(oo)
	// marshal to binary.
	bz := amino.MustMarshalAny(o2)
	gas := overflow.Mulp(ds.gasConfig.GasSetObject, store.Gas(len(bz)))
	ds.consumeGas(gas, GasSetObjectDesc)
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
		ds.baseStore.Set([]byte(key), hashbz)
		size = len(hashbz)
		oo.GetObjectInfo().LastObjectSize = int64(size)
	}
	// save object to cache.
	if debug {
		if oid.IsZero() {
			panic("object id cannot be zero")
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
		ds.iavlStore.Set(key, value)
	}
	return diff
}

func (ds *defaultStore) loadForLog(oid ObjectID) Object {
	key := backendObjectKey(oid)
	hashbz := ds.baseStore.Get([]byte(key))
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
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}
	if bm.StorageEnabled {
		bm.StartStore(bm.StoreDeleteObject)
		defer func() {
			// delete is a signle operation, not a func of size of bytes
			bm.StopStore(0)
		}()
	}
	ds.consumeGas(ds.gasConfig.GasDeleteObject, GasDeleteObjectDesc)
	oid := oo.GetObjectID()
	size := oo.GetObjectInfo().LastObjectSize
	// delete from cache.
	delete(ds.cacheObjects, oid)
	// delete from backend.
	if ds.baseStore != nil {
		key := backendObjectKey(oid)
		ds.baseStore.Delete([]byte(key))
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
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}

	// check cache.
	if tt, exists := ds.cacheTypes[tid]; exists {
		return tt
	}

	// check backend.
	if ds.baseStore != nil {
		key := backendTypeKey(tid)
		bz := ds.baseStore.Get([]byte(key))
		if bz != nil {
			gas := overflow.Mulp(ds.gasConfig.GasGetType, store.Gas(len(bz)))
			ds.consumeGas(gas, GasGetTypeDesc)
			cacheSum := sha256.Sum256(bz)
			var tt Type
			if val, ok := ds.aminoCache.Get(cacheSum[:]); ok {
				tt = copyTypeWithRefs(val)
			} else {
				amino.MustUnmarshal(bz, &tt)
				// len(bz) is not the proper cost of tt, but is good enough
				ds.aminoCache.Set(cacheSum[:], copyTypeWithRefs(tt), int64(len(bz)))
			}
			if debug {
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
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}
	var size int

	if bm.StorageEnabled {
		bm.StartStore(bm.StoreSetType)
		defer func() {
			bm.StopStore(size)
		}()
	}
	tid := tt.TypeID()
	// return if tid already known.
	if tt2, exists := ds.cacheTypes[tid]; exists {
		if tt != tt2 {
			// this can happen for a variety of reasons.
			// TODO classify them and optimize.
			return
		}
	}
	// save type to backend.
	if ds.baseStore != nil {
		key := backendTypeKey(tid)
		tcopy := copyTypeWithRefs(tt)
		bz := amino.MustMarshalAny(tcopy)
		cacheSum := sha256.Sum256(bz)
		ds.aminoCache.Set(cacheSum[:], tcopy, int64(len(bz)))
		gas := overflow.Mulp(ds.gasConfig.GasSetType, store.Gas(len(bz)))
		ds.consumeGas(gas, GasSetTypeDesc)
		ds.baseStore.Set([]byte(key), bz)
		size = len(bz)
	}
	// save type to cache.
	ds.cacheTypes[tid] = tt
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
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}

	var size int

	if bm.StorageEnabled {
		bm.StartStore(bm.StoreGetBlockNode)
		defer func() {
			bm.StopStore(size)
		}()
	}
	// check cache.
	if bn, exists := ds.cacheNodes.Get(loc); exists {
		return bn
	}
	// check backend.
	if ds.baseStore != nil {
		key := backendNodeKey(loc)
		bz := ds.baseStore.Get([]byte(key))
		if bz != nil {
			var bn BlockNode
			amino.MustUnmarshal(bz, &bn)
			size = len(bz)
			if debug {
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
	ctrbz := ds.baseStore.Get(ctrkey)
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
	ctrbz := ds.baseStore.Get(ctrkey)
	if ctrbz == nil {
		nextbz := strconv.Itoa(1)
		ds.baseStore.Set(ctrkey, []byte(nextbz))
		return 1
	} else {
		ctr, err := strconv.Atoi(string(ctrbz))
		if err != nil {
			panic(err)
		}
		nextbz := strconv.Itoa(ctr + 1)
		ds.baseStore.Set(ctrkey, []byte(nextbz))
		return uint64(ctr) + 1
	}
}

// mptype is passed in as a redundant parameter as convenience to assert that
// mpkg.Type is what is expected.
// If MPAnyAll, mpkg may be either MPStdlibAll or MPProdAll, and likewise for
// MPAnyProd and MPAnyTest.
// MPFiletests are not allowed, as they are currently only read from disk (e.g.
// test/files). However, MP*All may include filetests files.
func (ds *defaultStore) AddMemPackage(mpkg *std.MemPackage, mptype MemPackageType) {
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}
	var size int

	if bm.StorageEnabled {
		bm.StartStore(bm.StoreAddMemPackage)
		defer func() {
			bm.StopStore(size)
		}()
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
	bz := amino.MustMarshal(mpkg)
	gas := overflow.Mulp(ds.gasConfig.GasAddMemPackage, store.Gas(len(bz)))
	ds.consumeGas(gas, GasAddMemPackageDesc)
	ds.baseStore.Set(idxkey, []byte(mpkg.Path))
	pathkey := []byte(backendPackagePathKey(mpkg.Path))
	ds.iavlStore.Set(pathkey, bz)
	size = len(bz)
}

// GetMemPackage retrieves the MemPackage at the given path.
// It returns nil if the package could not be found.
func (ds *defaultStore) GetMemPackage(path string) *std.MemPackage {
	return ds.getMemPackage(path, false)
}

func (ds *defaultStore) getMemPackage(path string, isRetry bool) *std.MemPackage {
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}

	var size int

	if bm.StorageEnabled {
		bm.StartStore(bm.StoreGetMemPackage)
		defer func() {
			bm.StopStore(size)
		}()
	}
	pathkey := []byte(backendPackagePathKey(path))
	bz := ds.iavlStore.Get(pathkey)
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
	gas := overflow.Mulp(ds.gasConfig.GasGetMemPackage, store.Gas(len(bz)))
	ds.consumeGas(gas, GasGetMemPackageDesc)

	var mpkg *std.MemPackage
	amino.MustUnmarshal(bz, &mpkg)
	size = len(bz)
	return mpkg
}

// GetMemFile retrieves the MemFile with the given name, contained in the
// MemPackage at the given path. It returns nil if the file or the package
// do not exist.
func (ds *defaultStore) GetMemFile(path string, name string) *std.MemFile {
	mpkg := ds.GetMemPackage(path)
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
		iter := ds.iavlStore.Iterator(startKey, endKey)
		defer iter.Close()

		for ; iter.Valid(); iter.Next() {
			path := decodeBackendPackagePathKey(string(iter.Key()))
			if !yield(path) {
				return
			}
		}
	}
}

func (ds *defaultStore) IterMemPackage() <-chan *std.MemPackage {
	ctrkey := []byte(backendPackageIndexCtrKey())
	ctrbz := ds.baseStore.Get(ctrkey)
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
				path := ds.baseStore.Get(idxkey)
				if path == nil {
					panic(fmt.Sprintf(
						"missing package index %d", i))
				}
				mpkg := ds.GetMemPackage(string(path))
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

// It resturns storage changes per realm within message
func (ds *defaultStore) RealmStorageDiffs() map[string]int64 {
	return ds.realmStorageDiffs
}

// Unstable.
// This function is used to clear the object cache every transaction.
// It also sets a new allocator.
func (ds *defaultStore) ClearObjectCache() {
	ds.alloc.Reset()
	ds.cacheObjects = make(map[ObjectID]Object) // new cache.
	ds.realmStorageDiffs = make(map[string]int64)
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
	if enabled {
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
