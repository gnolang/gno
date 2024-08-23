package gnolang

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/colors"
	"github.com/gnolang/gno/tm2/pkg/std"
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

// inject natives into a new or loaded package (value and node)
type PackageInjector func(store Store, pn *PackageNode)

// NativeStore is a function which can retrieve native bodies of native functions.
type NativeStore func(pkgName string, name Name) func(m *Machine)

// Store is the central interface that specifies the communications between the
// GnoVM and the underlying data store; currently, generally the Gno.land
// blockchain, or the file system.
type Store interface {
	// STABLE
	BeginTransaction(baseStore, iavlStore store.Store) TransactionStore
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
	GetBlockNode(Location) BlockNode
	GetBlockNodeSafe(Location) BlockNode
	SetBlockNode(BlockNode)
	// UNSTABLE
	SetStrictGo2GnoMapping(bool)
	Go2GnoType(rt reflect.Type) Type
	GetAllocator() *Allocator
	NumMemPackages() int64
	// Upon restart, all packages will be re-preprocessed; This
	// loads BlockNodes and Types onto the store for persistence
	// version 1.
	AddMemPackage(memPkg *std.MemPackage)
	GetMemPackage(path string) *std.MemPackage
	GetMemFile(path string, name string) *std.MemFile
	IterMemPackage() <-chan *std.MemPackage
	ClearObjectCache()                                    // run before processing a message
	SetPackageInjector(PackageInjector)                   // for natives
	SetNativeStore(NativeStore)                           // for "new" natives XXX
	GetNative(pkgPath string, name Name) func(m *Machine) // for "new" natives XXX
	SetLogStoreOps(enabled bool)
	SprintStoreOps() string
	LogSwitchRealm(rlmpath string) // to mark change of realm boundaries
	ClearCache()
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

type defaultStore struct {
	// underlying stores used to keep data
	baseStore store.Store // for objects, types, nodes
	iavlStore store.Store // for escaped object hashes

	// transaction-scoped
	cacheObjects map[ObjectID]Object          // this is a real cache, reset with every transaction.
	cacheTypes   hashMap[TypeID, Type]        // this re-uses the parent store's.
	cacheNodes   hashMap[Location, BlockNode] // until BlockNode persistence is implemented, this is an actual store.
	alloc        *Allocator                   // for accounting for cached items

	// store configuration; cannot be modified in a transaction
	pkgGetter        PackageGetter         // non-realm packages
	cacheNativeTypes map[reflect.Type]Type // reflect doc: reflect.Type are comparable
	pkgInjector      PackageInjector       // for injecting natives
	nativeStore      NativeStore           // for injecting natives
	go2gnoStrict     bool                  // if true, native->gno type conversion must be registered.

	// transient
	current []string  // for detecting import cycles.
	opslog  []StoreOp // for debugging and testing.
}

type hashMap[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
	Delete(K)
	Iterate() func(yield func(K, V) bool)
}

type txLogMap[K comparable, V any] struct {
	source hashMap[K, V]
	dirty  map[K]deletable[V]
}

func newTxLog[K comparable, V any](source hashMap[K, V]) *txLogMap[K, V] {
	return &txLogMap[K, V]{
		source: source,
		dirty:  make(map[K]deletable[V]),
	}
}

// write commits the data in dirty to the map in source.
func (b *txLogMap[K, V]) write() {
	for k, v := range b.dirty {
		if v.deleted {
			b.source.Delete(k)
		} else {
			b.source.Set(k, v.v)
		}
	}
	b.dirty = make(map[K]deletable[V])
}

func (b txLogMap[K, V]) Get(k K) (V, bool) {
	if bufValue, ok := b.dirty[k]; ok {
		if bufValue.deleted {
			var zeroV V
			return zeroV, false
		}
		return bufValue.v, true
	}

	return b.source.Get(k)
}

func (b txLogMap[K, V]) Set(k K, v V) {
	b.dirty[k] = deletable[V]{v: v}
}

func (b txLogMap[K, V]) Delete(k K) {
	b.dirty[k] = deletable[V]{deleted: true}
}

func (b txLogMap[K, V]) Iterate() func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {
		b.source.Iterate()(func(k K, v V) bool {
			if dirty, ok := b.dirty[k]; ok {
				if dirty.deleted {
					return true
				}
				return yield(k, dirty.v)
			}

			// not in dirty
			return yield(k, v)
		})
		// yield for new values
		for k, v := range b.dirty {
			if v.deleted {
				continue
			}
			_, ok := b.source.Get(k)
			if ok {
				continue
			}
			if !yield(k, v.v) {
				break
			}
		}
	}
}

type mapWrapper[K comparable, V any] map[K]V

func (m mapWrapper[K, V]) Get(k K) (V, bool) {
	v, ok := m[k]
	return v, ok
}

func (m mapWrapper[K, V]) Set(k K, v V) {
	m[k] = v
}

func (m mapWrapper[K, V]) Delete(k K) {
	delete(m, k)
}

func (m mapWrapper[K, V]) Iterate() func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
}

type deletable[V any] struct {
	v       V
	deleted bool
}

func NewStore(alloc *Allocator, baseStore, iavlStore store.Store) *defaultStore {
	ds := &defaultStore{
		baseStore: baseStore,
		iavlStore: iavlStore,
		alloc:     alloc,

		// cacheObjects is set; objects in the store will be copied over for any transaction.
		cacheObjects: make(map[ObjectID]Object),
		cacheTypes:   mapWrapper[TypeID, Type](map[TypeID]Type{}),
		cacheNodes:   mapWrapper[Location, BlockNode](map[Location]BlockNode{}),

		// store configuration
		pkgGetter:        nil,
		cacheNativeTypes: make(map[reflect.Type]Type),
		pkgInjector:      nil,
		nativeStore:      nil,
		go2gnoStrict:     true,
	}
	InitStoreCaches(ds)
	return ds
}

// If nil baseStore and iavlStore, the baseStores are re-used.
func (ds *defaultStore) BeginTransaction(baseStore, iavlStore store.Store) TransactionStore {
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
		cacheTypes:   newTxLog(ds.cacheTypes),
		cacheNodes:   newTxLog(ds.cacheNodes),
		alloc:        ds.alloc.Fork().Reset(),

		// store configuration
		pkgGetter:        ds.pkgGetter,
		cacheNativeTypes: ds.cacheNativeTypes,
		pkgInjector:      ds.pkgInjector,
		nativeStore:      ds.nativeStore,
		go2gnoStrict:     ds.go2gnoStrict,

		// transient
		current: nil,
		opslog:  nil,
	}
	ds2.SetCachePackage(Uverse())

	return transactionStore{ds2}
}

type transactionStore struct{ *defaultStore }

func (t transactionStore) Write() {
	t.cacheTypes.(*txLogMap[TypeID, Type]).write()
	t.cacheNodes.(*txLogMap[Location, BlockNode]).write()
}

func (transactionStore) SetPackageGetter(pg PackageGetter) {
	panic("SetPackageGetter may not be called in a transaction store")
}

func (transactionStore) ClearCache() {
	panic("ClearCache may not be called in a transaction store")
}

// XXX: we should block Go2GnoType, because it uses a global cache map;
// but it's called during preprocess and thus breaks some testing code.
// let's wait until we remove Go2Gno entirely.
// func (transactionStore) Go2GnoType(reflect.Type) Type {
// 	panic("Go2GnoType may not be called in a transaction store")
// }

func (transactionStore) SetPackageInjector(inj PackageInjector) {
	panic("SetPackageInjector may not be called in a transaction store")
}

func (transactionStore) SetNativeStore(ns NativeStore) {
	panic("SetNativeStore may not be called in a transaction store")
}

func (transactionStore) SetStrictGo2GnoMapping(strict bool) {
	panic("SetStrictGo2GnoMapping may not be called in a transaction store")
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
			// get package node.
			pl := PackageNodeLocation(pkgPath)
			pn, ok := ds.GetBlockNodeSafe(pl).(*PackageNode)
			if !ok {
				// Do not inject packages from packageGetter
				// that don't have corresponding *PackageNodes.
			} else {
				// Inject natives after load.
				if ds.pkgInjector != nil {
					if pn.HasAttribute(ATTR_INJECTED) {
						// e.g. in checktx or simulate or query.
						pn.PrepareNewValues(pv)
					} else {
						// pv.GetBlock(ds) // preload pv.Block
						ds.pkgInjector(ds, pn)
						pn.SetAttribute(ATTR_INJECTED, true)
						pn.PrepareNewValues(pv)
					}
				}
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
			// inject natives after init.
			if ds.pkgInjector != nil {
				if pn.HasAttribute(ATTR_INJECTED) {
					// not sure why this would happen.
					panic("should not happen")
					// pn.PrepareNewValues(pv)
				} else {
					ds.pkgInjector(ds, pn)
					pn.SetAttribute(ATTR_INJECTED, true)
					pn.PrepareNewValues(pv)
				}
			}
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
	oid := ObjectIDFromPkgPath(pkgPath)
	key := backendRealmKey(oid)
	bz := ds.baseStore.Get([]byte(key))
	if bz == nil {
		return nil
	}
	amino.MustUnmarshal(bz, &rlm)
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
	oid := ObjectIDFromPkgPath(rlm.Path)
	key := backendRealmKey(oid)
	bz := amino.MustMarshal(rlm)
	ds.baseStore.Set([]byte(key), bz)
}

// NOTE: does not consult the packageGetter, so instead
// call GetPackage() for packages.
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
	key := backendObjectKey(oid)
	hashbz := ds.baseStore.Get([]byte(key))
	if hashbz != nil {
		hash := hashbz[:HashSize]
		bz := hashbz[HashSize:]
		var oo Object
		ds.alloc.AllocateAmino(int64(len(bz)))
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
	oid := oo.GetObjectID()
	// replace children/fields with Ref.
	o2 := copyValueWithRefs(oo)
	// marshal to binary.
	bz := amino.MustMarshalAny(o2)
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
		ds.Print()
		panic(fmt.Sprintf("unexpected type with id %s", tid.String()))
	}
	return tt
}

func (ds *defaultStore) GetTypeSafe(tid TypeID) Type {
	// check cache.
	if tt, exists := ds.cacheTypes.Get(tid); exists {
		return tt
	}
	// check backend.
	if ds.baseStore != nil {
		key := backendTypeKey(tid)
		bz := ds.baseStore.Get([]byte(key))
		if bz != nil {
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
			// NOTE: not sure why this would happen.
			panic("should not happen")
		} else {
			// already set.
		}
	} else {
		ds.cacheTypes.Set(tid, tt)
	}
}

func (ds *defaultStore) SetType(tt Type) {
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
		ds.baseStore.Set([]byte(key), bz)
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

func (ds *defaultStore) AddMemPackage(memPkg *std.MemPackage) {
	memPkg.Validate() // NOTE: duplicate validation.
	ctr := ds.incGetPackageIndexCounter()
	idxkey := []byte(backendPackageIndexKey(ctr))
	bz := amino.MustMarshal(memPkg)
	ds.baseStore.Set(idxkey, []byte(memPkg.Path))
	pathkey := []byte(backendPackagePathKey(memPkg.Path))
	ds.iavlStore.Set(pathkey, bz)
}

// GetMemPackage retrieves the MemPackage at the given path.
// It returns nil if the package could not be found.
func (ds *defaultStore) GetMemPackage(path string) *std.MemPackage {
	return ds.getMemPackage(path, false)
}

func (ds *defaultStore) getMemPackage(path string, isRetry bool) *std.MemPackage {
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

	var memPkg *std.MemPackage
	amino.MustUnmarshal(bz, &memPkg)
	return memPkg
}

// GetMemFile retrieves the MemFile with the given name, contained in the
// MemPackage at the given path. It returns nil if the file or the package
// do not exist.
func (ds *defaultStore) GetMemFile(path string, name string) *std.MemFile {
	memPkg := ds.GetMemPackage(path)
	if memPkg == nil {
		return nil
	}
	memFile := memPkg.GetFile(name)
	return memFile
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
		ch := make(chan *std.MemPackage, 0)
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

// Unstable.
// This function is used to clear the object cache every transaction.
// It also sets a new allocator.
func (ds *defaultStore) ClearObjectCache() {
	ds.alloc.Reset()
	ds.cacheObjects = make(map[ObjectID]Object) // new cache.
	ds.opslog = nil                             // new ops log.
	ds.SetCachePackage(Uverse())
}

func (ds *defaultStore) SetPackageInjector(inj PackageInjector) {
	ds.pkgInjector = inj
}

func (ds *defaultStore) SetNativeStore(ns NativeStore) {
	ds.nativeStore = ns
}

func (ds *defaultStore) GetNative(pkgPath string, name Name) func(m *Machine) {
	if ds.nativeStore != nil {
		return ds.nativeStore(pkgPath, name)
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

func (ds *defaultStore) ClearCache() {
	ds.cacheObjects = make(map[ObjectID]Object)
	ds.cacheTypes = mapWrapper[TypeID, Type](map[TypeID]Type{})
	ds.cacheNodes = mapWrapper[Location, BlockNode](map[Location]BlockNode{})
	ds.cacheNativeTypes = make(map[reflect.Type]Type)
	// restore builtin types to cache.
	InitStoreCaches(ds)
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
	return fmt.Sprintf("pkg:" + path)
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
