package gnolang

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm"
	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/txlog"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/colors"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/utils"
	stringz "github.com/gnolang/gno/tm2/pkg/strings"
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
	SetPackageGetter(PackageGetter)
	GetPackage(pkgPath string, isImport bool) *PackageValue
	SetCachePackage(*PackageValue)
	GetPackageRealm(pkgPath string) *Realm
	SetPackageRealm(*Realm)
	GetObject(oid ObjectID) Object
	GetObjectSafe(oid ObjectID) Object
	SetObject(Object)
	DelObject(Object)
	GetType(tid TypeID) Type
	GetTypeSafe(tid TypeID) Type
	SetCacheType(Type)
	SetType(Type)
	GetBlockNode(Location) BlockNode // to get a PackageNode, use PackageNodeLocation().
	GetBlockNodeSafe(Location) BlockNode
	SetBlockNode(BlockNode)

	// UNSTABLE
	Go2GnoType(rt reflect.Type) Type
	GetAllocator() *Allocator
	NumMemPackages() int64
	// Upon restart, all packages will be re-preprocessed; This
	// loads BlockNodes and Types onto the store for persistence
	// version 1.
	AddMemPackage(memPkg *gnovm.MemPackage)
	GetMemPackage(path string) *gnovm.MemPackage
	GetMemFile(path string, name string) *gnovm.MemFile
	IterMemPackage() <-chan *gnovm.MemPackage
	ClearObjectCache()                                    // run before processing a message
	SetNativeResolver(NativeResolver)                     // for "new" natives XXX
	GetNative(pkgPath string, name Name) func(m *Machine) // for "new" natives XXX
	SetLogStoreOps(enabled bool)
	SprintStoreOps() string
	LogSwitchRealm(rlmpath string) // to mark change of realm boundaries
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
		GasGetType:         52,   // per byte cost
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
	cacheObjects map[ObjectID]Object            // this is a real cache, reset with every transaction.
	cacheTypes   txlog.Map[TypeID, Type]        // this re-uses the parent store's.
	cacheNodes   txlog.Map[Location, BlockNode] // until BlockNode persistence is implemented, this is an actual store.
	alloc        *Allocator                     // for accounting for cached items

	// store configuration; cannot be modified in a transaction
	pkgGetter        PackageGetter         // non-realm packages
	cacheNativeTypes map[reflect.Type]Type // reflect doc: reflect.Type are comparable
	nativeResolver   NativeResolver        // for injecting natives

	// transient
	opslog  []StoreOp // for debugging and testing.
	current []string  // for detecting import cycles.

	// gas
	gasMeter  store.GasMeter
	gasConfig GasConfig
}

func NewStore(alloc *Allocator, baseStore, iavlStore store.Store) *defaultStore {
	ds := &defaultStore{
		baseStore: baseStore,
		iavlStore: iavlStore,
		alloc:     alloc,

		// cacheObjects is set; objects in the store will be copied over for any transaction.
		cacheObjects: make(map[ObjectID]Object),
		cacheTypes:   txlog.GoMap[TypeID, Type](map[TypeID]Type{}),
		cacheNodes:   txlog.GoMap[Location, BlockNode](map[Location]BlockNode{}),

		// store configuration
		pkgGetter:        nil,
		cacheNativeTypes: make(map[reflect.Type]Type),
		nativeResolver:   nil,
		gasConfig:        DefaultGasConfig(),
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
		cacheTypes:   txlog.Wrap(ds.cacheTypes),
		cacheNodes:   txlog.Wrap(ds.cacheNodes),
		alloc:        ds.alloc.Fork().Reset(),

		// store configuration
		pkgGetter:        ds.pkgGetter,
		cacheNativeTypes: ds.cacheNativeTypes,
		nativeResolver:   ds.nativeResolver,

		// gas meter
		gasMeter:  gasMeter,
		gasConfig: ds.gasConfig,

		// transient
		current: nil,
		opslog:  nil,
	}
	ds2.SetCachePackage(Uverse())

	return transactionStore{ds2}
}

type transactionStore struct{ *defaultStore }

func (t transactionStore) Write() {
	t.cacheTypes.(txlog.MapCommitter[TypeID, Type]).Commit()
	t.cacheNodes.(txlog.MapCommitter[Location, BlockNode]).Commit()
}

func (transactionStore) SetPackageGetter(pg PackageGetter) {
	panic("SetPackageGetter may not be called in a transaction store")
}

// XXX: we should block Go2GnoType, because it uses a global cache map;
// but it's called during preprocess and thus breaks some testing code.
// let's wait until we remove Go2Gno entirely.
// https://github.com/gnolang/gno/issues/1361
// func (transactionStore) Go2GnoType(reflect.Type) Type {
// 	panic("Go2GnoType may not be called in a transaction store")
// }

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

	ss.cacheTypes.Iterate()(func(k TypeID, v Type) bool {
		ds.cacheTypes.Set(k, v)
		return true
	})
	ss.cacheNodes.Iterate()(func(k Location, v BlockNode) bool {
		ds.cacheNodes.Set(k, v)
		return true
	})
}

func (ds *defaultStore) GetAllocator() *Allocator {
	return ds.alloc
}

func (ds *defaultStore) SetPackageGetter(pg PackageGetter) {
	ds.pkgGetter = pg
}

// Gets package from cache, or loads it from baseStore, or gets it from package getter.
func (ds *defaultStore) GetPackage(pkgPath string, isImport bool) *PackageValue {
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
	// first, check cache.
	oid := ObjectIDFromPkgPath(pkgPath)
	if oo, exists := ds.cacheObjects[oid]; exists {
		pv := oo.(*PackageValue)
		return pv
	}
	// else, load package.
	if ds.baseStore != nil {
		if oo := ds.loadObjectSafe(oid); oo != nil {
			pv := oo.(*PackageValue)
			_ = pv.GetBlock(ds) // preload
			// get package associated realm if nil.
			if pv.IsRealm() && pv.Realm == nil {
				rlm := ds.GetPackageRealm(pkgPath)
				pv.Realm = rlm
			}
			// Rederive pv.fBlocksMap.
			pv.deriveFBlocksMap(ds)
			return pv
		}
	}
	// otherwise, fetch from pkgGetter.
	if ds.pkgGetter != nil {
		if pn, pv := ds.pkgGetter(pkgPath, ds); pv != nil {
			// e.g. tests/imports_tests loads example/gno.land/r/... realms.
			// if pv.IsRealm() {
			// 	panic("realm packages cannot be gotten from pkgGetter")
			// }
			ds.SetBlockNode(pn)
			// NOTE: not SetObject() here,
			// we don't want to overwrite
			// the value from pkgGetter.
			// Realm values obtained this way
			// will get written elsewhere
			// later.
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
					ds.SetCacheType(t)
				}
			}
			return pv
		}
	}
	// otherwise, package does not exist.
	return nil
}

// Used to set throwaway packages.
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
	gas := overflow.Mul64p(ds.gasConfig.GasGetPackageRealm, store.Gas(len(bz)))
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
	gas := overflow.Mul64p(ds.gasConfig.GasSetPackageRealm, store.Gas(len(bz)))
	ds.consumeGas(gas, GasSetPackageRealmDesc)
	ds.baseStore.Set([]byte(key), bz)
	size = len(bz)
}

// NOTE: does not consult the packageGetter, so instead
// call GetPackage() for packages.
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
	// check cache.
	if oo, exists := ds.cacheObjects[oid]; exists {
		return oo
	}
	// check baseStore.
	if ds.baseStore != nil {
		if oo := ds.loadObjectSafe(oid); oo != nil {
			if debug {
				if _, ok := oo.(*PackageValue); ok {
					panic("packages must be fetched with GetPackage()")
				}
			}
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
		ds.alloc.AllocateAmino(int64(len(bz)))
		gas := overflow.Mul64p(ds.gasConfig.GasGetObject, store.Gas(len(bz)))
		ds.consumeGas(gas, GasGetObjectDesc)
		amino.MustUnmarshal(bz, &oo)
		if debug {
			if oo.GetObjectID() != oid {
				panic(fmt.Sprintf("unexpected object id: expected %v but got %v",
					oid, oo.GetObjectID()))
			}
		}
		oo.SetHash(ValueHash{NewHashlet(hash)})
		ds.cacheObjects[oid] = oo
		_ = fillTypesOfValue(ds, oo)
		return oo
	}
	return nil
}

// NOTE: unlike GetObject(), SetObject() is also used to persist updated
// package values.
func (ds *defaultStore) SetObject(oo Object) {
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
	gas := overflow.Mul64p(ds.gasConfig.GasSetObject, store.Gas(len(bz)))
	ds.consumeGas(gas, GasSetObjectDesc)
	// set hash.
	hash := HashBytes(bz) // XXX objectHash(bz)???
	if len(hash) != HashSize {
		panic("should not happen")
	}
	oo.SetHash(ValueHash{hash})
	// save bytes to backend.
	if ds.baseStore != nil {
		key := backendObjectKey(oid)
		hashbz := make([]byte, len(hash)+len(bz))
		copy(hashbz, hash.Bytes())
		copy(hashbz[HashSize:], bz)
		ds.baseStore.Set([]byte(key), hashbz)
		size = len(hashbz)
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
	// make store op log entry
	if ds.opslog != nil {
		var op StoreOpType
		if oo.GetIsNewReal() {
			op = StoreOpNew
		} else {
			op = StoreOpMod
		}
		ds.opslog = append(ds.opslog,
			StoreOp{Type: op, Object: o2.(Object)})
	}
	// if escaped, add hash to iavl.
	if oo.GetIsEscaped() && ds.iavlStore != nil {
		var key, value []byte
		key = []byte(oid.String())
		value = hash.Bytes()
		ds.iavlStore.Set(key, value)
	}
}

func (ds *defaultStore) DelObject(oo Object) {
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
	// delete from cache.
	delete(ds.cacheObjects, oid)
	// delete from backend.
	if ds.baseStore != nil {
		key := backendObjectKey(oid)
		ds.baseStore.Delete([]byte(key))
	}
	// make realm op log entry
	if ds.opslog != nil {
		ds.opslog = append(ds.opslog,
			StoreOp{Type: StoreOpDel, Object: oo})
	}
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
	if tt, exists := ds.cacheTypes.Get(tid); exists {
		return tt
	}
	// check backend.
	if ds.baseStore != nil {
		key := backendTypeKey(tid)
		bz := ds.baseStore.Get([]byte(key))
		if bz != nil {
			gas := overflow.Mul64p(ds.gasConfig.GasGetType, store.Gas(len(bz)))
			ds.consumeGas(gas, GasGetTypeDesc)
			var tt Type
			amino.MustUnmarshal(bz, &tt)
			if debug {
				if tt.TypeID() != tid {
					panic(fmt.Sprintf("unexpected type id: expected %v but got %v",
						tid, tt.TypeID()))
				}
			}
			// set in cache.
			ds.cacheTypes.Set(tid, tt)
			// after setting in cache, fill tt.
			fillType(ds, tt)
			return tt
		}
	}
	return nil
}

func (ds *defaultStore) SetCacheType(tt Type) {
	tid := tt.TypeID()
	if tt2, exists := ds.cacheTypes.Get(tid); exists {
		if tt != tt2 {
			panic(fmt.Sprintf("cannot re-register %q with different type", tid))
		} else {
			// already set.
		}
	} else {
		ds.cacheTypes.Set(tid, tt)
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
	if tt2, exists := ds.cacheTypes.Get(tid); exists {
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
		gas := overflow.Mul64p(ds.gasConfig.GasSetType, store.Gas(len(bz)))
		ds.consumeGas(gas, GasSetTypeDesc)
		ds.baseStore.Set([]byte(key), bz)
		size = len(bz)
	}
	// save type to cache.
	ds.cacheTypes.Set(tid, tt)
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
	if ds.baseStore != nil {
		// TODO: implement copyValueWithRefs() for Nodes.
		// key := backendNodeKey(loc)
		// ds.backend.Set([]byte(key), bz)
	}
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

func (ds *defaultStore) AddMemPackage(memPkg *gnovm.MemPackage) {
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
	memPkg.Validate() // NOTE: duplicate validation.
	ctr := ds.incGetPackageIndexCounter()
	idxkey := []byte(backendPackageIndexKey(ctr))
	bz := amino.MustMarshal(memPkg)
	gas := overflow.Mul64p(ds.gasConfig.GasAddMemPackage, store.Gas(len(bz)))
	ds.consumeGas(gas, GasAddMemPackageDesc)
	ds.baseStore.Set(idxkey, []byte(memPkg.Path))
	pathkey := []byte(backendPackagePathKey(memPkg.Path))
	ds.iavlStore.Set(pathkey, bz)
	size = len(bz)
}

// GetMemPackage retrieves the MemPackage at the given path.
// It returns nil if the package could not be found.
func (ds *defaultStore) GetMemPackage(path string) *gnovm.MemPackage {
	return ds.getMemPackage(path, false)
}

func (ds *defaultStore) getMemPackage(path string, isRetry bool) *gnovm.MemPackage {
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
	gas := overflow.Mul64p(ds.gasConfig.GasGetMemPackage, store.Gas(len(bz)))
	ds.consumeGas(gas, GasGetMemPackageDesc)

	var memPkg *gnovm.MemPackage
	amino.MustUnmarshal(bz, &memPkg)
	size = len(bz)
	return memPkg
}

// GetMemFile retrieves the MemFile with the given name, contained in the
// MemPackage at the given path. It returns nil if the file or the package
// do not exist.
func (ds *defaultStore) GetMemFile(path string, name string) *gnovm.MemFile {
	memPkg := ds.GetMemPackage(path)
	if memPkg == nil {
		return nil
	}
	memFile := memPkg.GetFile(name)
	return memFile
}

func (ds *defaultStore) IterMemPackage() <-chan *gnovm.MemPackage {
	ctrkey := []byte(backendPackageIndexCtrKey())
	ctrbz := ds.baseStore.Get(ctrkey)
	if ctrbz == nil {
		return nil
	} else {
		ctr, err := strconv.Atoi(string(ctrbz))
		if err != nil {
			panic(err)
		}
		ch := make(chan *gnovm.MemPackage, 0)
		go func() {
			for i := uint64(1); i <= uint64(ctr); i++ {
				idxkey := []byte(backendPackageIndexKey(i))
				path := ds.baseStore.Get(idxkey)
				if path == nil {
					panic(fmt.Sprintf(
						"missing package index %d", i))
				}
				memPkg := ds.GetMemPackage(string(path))
				ch <- memPkg
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

// Unstable.
// This function is used to clear the object cache every transaction.
// It also sets a new allocator.
func (ds *defaultStore) ClearObjectCache() {
	ds.alloc.Reset()
	ds.cacheObjects = make(map[ObjectID]Object) // new cache.
	ds.opslog = nil                             // new ops log.
	ds.SetCachePackage(Uverse())
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

// ----------------------------------------
// StoreOp

type StoreOpType uint8

const (
	StoreOpNew StoreOpType = iota
	StoreOpMod
	StoreOpDel
	StoreOpSwitchRealm
)

type StoreOp struct {
	Type    StoreOpType
	Object  Object // ref'd objects
	RlmPath string // for StoreOpSwitchRealm
}

// used by the tests/file_test system to check
// veracity of realm operations.
func (sop StoreOp) String() string {
	switch sop.Type {
	case StoreOpNew:
		return fmt.Sprintf("c[%v]=%s",
			sop.Object.GetObjectID(),
			prettyJSON(amino.MustMarshalJSON(sop.Object)))
	case StoreOpMod:
		return fmt.Sprintf("u[%v]=%s",
			sop.Object.GetObjectID(),
			prettyJSON(amino.MustMarshalJSON(sop.Object)))
	case StoreOpDel:
		return fmt.Sprintf("d[%v]",
			sop.Object.GetObjectID())
	case StoreOpSwitchRealm:
		return fmt.Sprintf("switchrealm[%q]",
			sop.RlmPath)
	default:
		panic("should not happen")
	}
}

func (ds *defaultStore) SetLogStoreOps(enabled bool) {
	if enabled {
		ds.ResetStoreOps()
	} else {
		ds.opslog = nil
	}
}

// resets .realmops.
func (ds *defaultStore) ResetStoreOps() {
	ds.opslog = make([]StoreOp, 0, 1024)
}

// for test/file_test.go, to test realm changes.
func (ds *defaultStore) SprintStoreOps() string {
	ss := make([]string, 0, len(ds.opslog))
	for _, sop := range ds.opslog {
		ss = append(ss, sop.String())
	}
	return strings.Join(ss, "\n")
}

func (ds *defaultStore) LogSwitchRealm(rlmpath string) {
	ds.opslog = append(ds.opslog,
		StoreOp{Type: StoreOpSwitchRealm, RlmPath: rlmpath})
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
	ds.cacheTypes.Iterate()(func(tid TypeID, typ Type) bool {
		fmt.Printf("- %v: %v\n", tid,
			stringz.TrimN(fmt.Sprintf("%v", typ), 50))
		return true
	})
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
	return fmt.Sprintf("pkgidx:counter")
}

func backendPackageIndexKey(index uint64) string {
	return fmt.Sprintf("pkgidx:%020d", index)
}

func backendPackagePathKey(path string) string {
	return "pkg:" + path
}

// ----------------------------------------
// builtin types and packages

func InitStoreCaches(store Store) {
	types := []Type{
		BoolType, UntypedBoolType,
		StringType, UntypedStringType,
		IntType, Int8Type, Int16Type, Int32Type, Int64Type, UntypedRuneType,
		UintType, Uint8Type, Uint16Type, Uint32Type, Uint64Type,
		BigintType, UntypedBigintType,
		gTypeType,
		gPackageType,
		blockType{},
		Float32Type, Float64Type,
		gErrorType, // from uverse.go
	}
	for _, tt := range types {
		store.SetCacheType(tt)
	}
	store.SetCachePackage(Uverse())
}
