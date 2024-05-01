package gnolang

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// return nil if package doesn't exist.
type PackageGetter func(pkgPath string) (*PackageNode, *PackageValue)

// inject natives into a new or loaded package (value and node)
type PackageInjector func(store Store, pn *PackageNode)

// NativeStore is a function which can retrieve native bodies of native functions.
type NativeStore func(pkgName string, name Name) func(m *Machine)

type Store interface {
	// STABLE
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
	ClearObjectCache()                                    // for each delivertx.
	Fork() Store                                          // for checktx, simulate, and queries.
	SwapStores(baseStore, iavlStore store.Store)          // for gas wrappers.
	SetPackageInjector(PackageInjector)                   // for natives
	SetNativeStore(NativeStore)                           // for "new" natives XXX
	GetNative(pkgPath string, name Name) func(m *Machine) // for "new" natives XXX
	SetLogStoreOps(enabled bool)
	SprintStoreOps() string
	LogSwitchRealm(rlmpath string) // to mark change of realm boundaries
	ClearCache()
	Print()
}

// Used to keep track of in-mem objects during tx.
type defaultStore struct {
	alloc            *Allocator    // for accounting for cached items
	pkgGetter        PackageGetter // non-realm packages
	cacheObjects     map[ObjectID]Object
	cacheTypes       map[TypeID]Type
	cacheNodes       map[Location]BlockNode
	cacheNativeTypes map[reflect.Type]Type // go spec: reflect.Type are comparable
	baseStore        store.Store           // for objects, types, nodes
	iavlStore        store.Store           // for escaped object hashes
	pkgInjector      PackageInjector       // for injecting natives
	nativeStore      NativeStore           // for injecting natives
	go2gnoStrict     bool                  // if true, native->gno type conversion must be registered.

	// transient
	opslog  []StoreOp           // for debugging and testing.
	current map[string]struct{} // for detecting import cycles.
}

func NewStore(alloc *Allocator, baseStore, iavlStore store.Store) *defaultStore {
	ds := &defaultStore{
		alloc:            alloc,
		pkgGetter:        nil,
		cacheObjects:     make(map[ObjectID]Object),
		cacheTypes:       make(map[TypeID]Type),
		cacheNodes:       make(map[Location]BlockNode),
		cacheNativeTypes: make(map[reflect.Type]Type),
		baseStore:        baseStore,
		iavlStore:        iavlStore,
		go2gnoStrict:     true,
		current:          make(map[string]struct{}),
	}
	InitStoreCaches(ds)
	return ds
}

func (ds *defaultStore) GetAllocator() *Allocator {
	return ds.alloc
}

func (ds *defaultStore) SetPackageGetter(pg PackageGetter) {
	ds.pkgGetter = pg
}

// Gets package from cache, or loads it from baseStore, or gets it from package getter.
func (ds *defaultStore) GetPackage(pkgPath string, isImport bool) *PackageValue {
	// detect circular imports
	if isImport {
		if _, exists := ds.current[pkgPath]; exists {
			panic(fmt.Sprintf("import cycle detected: %q", pkgPath))
		}
		ds.current[pkgPath] = struct{}{}
		defer delete(ds.current, pkgPath)
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
		if pn, pv := ds.pkgGetter(pkgPath); pv != nil {
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
				if tv.T.Kind() == TypeKind {
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
	o2 := copyValueWithRefs(nil, oo)
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
	if tt, exists := ds.cacheTypes[tid]; exists {
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
			// NOTE: not sure why this would happen.
			panic("should not happen")
		} else {
			// already set.
		}
	} else {
		ds.cacheTypes[tid] = tt
	}
}

func (ds *defaultStore) SetType(tt Type) {
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
		ds.baseStore.Set([]byte(key), bz)
	}
	// save type to cache.
	ds.cacheTypes[tid] = tt
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
	if bn, exists := ds.cacheNodes[loc]; exists {
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
			ds.cacheNodes[loc] = bn
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
	ds.cacheNodes[loc] = bn
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

func (ds *defaultStore) GetMemPackage(path string) *std.MemPackage {
	pathkey := []byte(backendPackagePathKey(path))
	bz := ds.iavlStore.Get(pathkey)
	if bz == nil {
		panic(fmt.Sprintf(
			"missing package at path %s", string(pathkey)))
	}
	var memPkg *std.MemPackage
	amino.MustUnmarshal(bz, &memPkg)
	return memPkg
}

func (ds *defaultStore) GetMemFile(path string, name string) *std.MemFile {
	memPkg := ds.GetMemPackage(path)
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
	if len(ds.current) > 0 {
		ds.current = make(map[string]struct{})
	}
	ds.SetCachePackage(Uverse())
}

// Unstable.
// This function is used to handle queries and checktx transactions.
func (ds *defaultStore) Fork() Store {
	ds2 := &defaultStore{
		alloc:            ds.alloc.Fork().Reset(),
		pkgGetter:        ds.pkgGetter,
		cacheObjects:     make(map[ObjectID]Object), // new cache.
		cacheTypes:       ds.cacheTypes,
		cacheNodes:       ds.cacheNodes,
		cacheNativeTypes: ds.cacheNativeTypes,
		baseStore:        ds.baseStore,
		iavlStore:        ds.iavlStore,
		pkgInjector:      ds.pkgInjector,
		nativeStore:      ds.nativeStore,
		go2gnoStrict:     ds.go2gnoStrict,
		opslog:           nil, // new ops log.
		current:          make(map[string]struct{}),
	}
	ds2.SetCachePackage(Uverse())
	return ds2
}

// TODO: consider a better/faster/simpler way of achieving the overall same goal?
func (ds *defaultStore) SwapStores(baseStore, iavlStore store.Store) {
	ds.baseStore = baseStore
	ds.iavlStore = iavlStore
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

func (ds *defaultStore) Flush() {
	// XXX
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
	ds.cacheTypes = make(map[TypeID]Type)
	ds.cacheNodes = make(map[Location]BlockNode)
	ds.cacheNativeTypes = make(map[reflect.Type]Type)
	// restore builtin types to cache.
	InitStoreCaches(ds)
}

// for debugging
func (ds *defaultStore) Print() {
	fmt.Println("//----------------------------------------")
	fmt.Println("defaultStore:baseStore...")
	store.Print(ds.baseStore)
	fmt.Println("//----------------------------------------")
	fmt.Println("defaultStore:iavlStore...")
	store.Print(ds.iavlStore)
	fmt.Println("//----------------------------------------")
	fmt.Println("defaultStore:cacheTypes...")
	for tid, typ := range ds.cacheTypes {
		fmt.Printf("- %v: %v\n", tid, typ)
	}
	fmt.Println("//----------------------------------------")
	fmt.Println("defaultStore:cacheNodes...")
	for loc, bn := range ds.cacheNodes {
		fmt.Printf("- %v: %v\n", loc, bn)
	}
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
