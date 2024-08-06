package gnolang

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// KeyValueStore is a subset of tm2's store.Store.
type KeyValueStore interface {
	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key or value.
	Set(key, value []byte)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)
}

// StoreOptions are the options which may be passed when creating a new store.
type StoreOptions struct {
	// PackageInjector is deprecated.
	// It is an old method to inject native code into existing packages. New code
	// should use NativeResolver instead.
	//
	// TODO(morgan): remove with https://github.com/gnolang/gno/pull/1464
	PackageInjector func(pn *PackageNode)

	// NativeResolver is called to resolve the given combination of a pkgPath
	// and function name to a native function.
	NativeResolver func(pkgPath string, name Name) func(m *Machine)

	// Allocator is the store's allocator, which can limit how much the VM allocates.
	// May be nil.
	Allocator *Allocator

	// Go2GnoDefined allows mappings on defined types in the store.
	// By default, Go2GnoType only works on unnamed types. If this is enabled,
	// mappings will also be created and registered for named types.
	//
	// TODO(morgan): remove with https://github.com/gnolang/gno/issues/1361
	Go2GnoDefined bool

	// CachePackage is a package that should be directly added into the store's
	// cache. Can be set for throwaway packages.
	CachePackage *PackageValue
}

// GetNative uses StoreOptions' NativeResolver to resolve a symbol to a native function.
// Implements FullStore.
func (s *StoreOptions) GetNative(pkgPath string, name Name) func(m *Machine) {
	return s.NativeResolver(pkgPath, name)
}

// GetAllocator returns the StoreOptions' allocator.
// Implements FullStore.
func (s *StoreOptions) GetAllocator() *Allocator {
	return s.Allocator
}

// PackageStore keeps GnoVM's packages and realms.
type PackageStore interface {
	GetPackage(pkgPath string) *PackageValue
	SetPackageRealm(*Realm)

	// getPackageRealm exists, but is currently unexported as it is only used
	// internally in the store. You can likely use PackageValue.Realm.
	// SetPackage does not exist - use SetObject instead.
}

// NewPackageStore returns a new PackageStore.
// packageInjector is optional; all stores are required.
// realmStore is used to store the data for [Realm] values, typically a base store.
func NewPackageStore(
	os ObjectStore,
	bns BlockNodeStore,
	realmStore KeyValueStore,
	packageInjector func(pn *PackageNode),
) PackageStore {
	return &packageStore{
		os:              os,
		bns:             bns,
		realmStore:      realmStore,
		packageInjector: packageInjector,
	}
}

type packageStore struct {
	os              ObjectStore
	bns             BlockNodeStore
	realmStore      KeyValueStore
	packageInjector func(pn *PackageNode)
}

var _ PackageStore = (*packageStore)(nil)

func (ps *packageStore) GetPackage(pkgPath string) *PackageValue {
	oid := ObjectIDFromPkgPath(pkgPath)
	oo := ps.os.GetObject(oid)
	if oo == nil {
		// *PackageValue does not exist.
		return nil
	}
	pv := oo.(*PackageValue)
	if pv.fBlocksMap != nil {
		// *PackageValue is already loaded.
		return pv
	}
	// fBlocksMap == nil; the PackageValue needs to be initialized.

	// Resolve pv.Block to a *Block if it is a RefValue.
	_ = pv.GetBlock(ps.os)

	// Get associated realm.
	if pv.IsRealm() {
		rlm := ps.getPackageRealm(pkgPath)
		pv.Realm = rlm
	}

	pl := PackageNodeLocation(pkgPath)
	pn := ps.bns.GetBlockNode(pl).(*PackageNode)

	// Inject natives if applicable, and PrepareNewValues so we make sure
	// PackageNode is up-to-date. Finally, re-derive FBlocksMap.
	if ps.packageInjector != nil {
		ps.packageInjector(pn)
		pn.PrepareNewValues(pv)
	}
	pv.deriveFBlocksMap(ps.os)

	return pv
}

func (ps *packageStore) getPackageRealm(pkgPath string) (rlm *Realm) {
	oid := ObjectIDFromPkgPath(pkgPath)
	key := backendRealmKey(oid)
	bz := ps.realmStore.Get([]byte(key))
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

func (ps *packageStore) SetPackageRealm(rlm *Realm) {
	oid := ObjectIDFromPkgPath(rlm.Path)
	key := backendRealmKey(oid)
	bz := amino.MustMarshal(rlm)
	ps.realmStore.Set([]byte(key), bz)
}

// MemPackageStore is a store which keeps track of all of all the underlying
// packages' source code.
type MemPackageStore interface {
	NumMemPackages() int64
	AddMemPackage(memPkg *std.MemPackage)
	GetMemPackage(path string) *std.MemPackage
	// Can be used with Rangefunc: https://go.dev/wiki/RangefuncExperiment
	IterMemPackage() func(yield func(*std.MemPackage) bool)
}

// NewMemPackageStore creates a new [MemPackageStore], storing data in the given
// key/value stores.
//
// controlStore keeps track of the number of current mempackages and keeps an
// ordered list of all MemPackages in the order they were added.
// packageStore contains the actual package data.
func NewMemPackageStore(controlStore, packageStore KeyValueStore) MemPackageStore {
	return &memPackageStore{controlStore, packageStore}
}

type memPackageStore struct {
	controlStore KeyValueStore
	packageStore KeyValueStore
}

var _ MemPackageStore = (*memPackageStore)(nil)

func (mps *memPackageStore) NumMemPackages() int64 {
	ctrkey := []byte(backendPackageIndexCtrKey())
	ctrbz := mps.controlStore.Get(ctrkey)
	if ctrbz == nil {
		return 0
	} else {
		ctr, err := strconv.ParseInt(string(ctrbz), 10, 64)
		if err != nil {
			panic(err)
		}
		return ctr
	}
}

func (mps *memPackageStore) incrNumMemPackages() int64 {
	num := mps.NumMemPackages()
	num++

	bz := strconv.FormatInt(num, 10)
	ctrkey := []byte(backendPackageIndexCtrKey())
	mps.controlStore.Set(ctrkey, []byte(bz))
	return num
}

func (mps *memPackageStore) AddMemPackage(memPkg *std.MemPackage) {
	memPkg.Validate() // NOTE: duplicate validation.
	ctr := mps.incrNumMemPackages()
	idxkey := []byte(backendPackageIndexKey(ctr))
	bz := amino.MustMarshal(memPkg)
	mps.controlStore.Set(idxkey, []byte(memPkg.Path))
	pathkey := []byte(backendPackagePathKey(memPkg.Path))
	mps.packageStore.Set(pathkey, bz)
}

func (mps *memPackageStore) GetMemPackage(path string) *std.MemPackage {
	pathkey := []byte(backendPackagePathKey(path))
	bz := mps.packageStore.Get(pathkey)
	if bz == nil {
		return nil
	}

	var memPkg *std.MemPackage
	amino.MustUnmarshal(bz, &memPkg)
	return memPkg
}

func (mps *memPackageStore) IterMemPackage() func(yield func(*std.MemPackage) bool) {
	num := mps.NumMemPackages()
	if num == 0 {
		return func(_ func(*std.MemPackage) bool) {}
	}

	return func(yield func(*std.MemPackage) bool) {
		for i := int64(1); i <= num; i++ {
			idxkey := []byte(backendPackageIndexKey(int64(i)))
			path := mps.controlStore.Get(idxkey)
			if path == nil {
				panic(fmt.Sprintf(
					"missing package index %d", i))
			}

			memPkg := mps.GetMemPackage(string(path))
			if !yield(memPkg) {
				return
			}
		}
	}
}

// ObjectTypeStore is a combination of the Object and Type store, sometimes
// required by some functions.
type ObjectTypeStore interface {
	ObjectStore
	TypeStore
}

type objectTypeStore struct {
	ObjectStore
	TypeStore
}

// ObjectStore is a store which manages all [Object] values.
type ObjectStore interface {
	// NOTE: does not initialize *PackageValues, so instead call GetPackage()
	// for packages.
	// NOTE: current implementation behavior requires
	// all []TypedValue types and TypeValue{} types to be
	// loaded (non-ref) types.
	GetObject(oid ObjectID) Object
	// NOTE: unlike GetObject(), SetObject() is also used to persist updated
	// package values.
	SetObject(Object)
	DelObject(Object)
}

var _ ObjectStore = (*objectStore)(nil)

// NewObjectStore creates a new [ObjectStore].
// kvStore is used to keep the objects themselves; escapedStore instead just
// contains the hashes of escaped objects.
//
// To correctly recover objects from the database, a typeStore is required.
// An allocator may be provided.
func NewObjectStore(
	kvStore KeyValueStore,
	escapedStore KeyValueStore,
	ts TypeStore,
	alloc *Allocator,
) ObjectStore {
	return &objectStore{
		kvStore:      kvStore,
		escapedStore: escapedStore,
		ts:           ts,
		alloc:        alloc,
	}
}

type objectStoreCache struct {
	parent ObjectStore
	m      map[ObjectID]Object
}

var _ ObjectStore = (*objectStoreCache)(nil)

func (os *objectStoreCache) GetObject(oid ObjectID) Object {
	if obj, ok := os.m[oid]; ok {
		return obj
	}
	os.parent.GetObject(oid)
}

type objectStore struct {
	kvStore      KeyValueStore
	escapedStore KeyValueStore
	ts           TypeStore
	alloc        *Allocator
}

var _ ObjectStore = (*objectStore)(nil)

func (os *objectStore) GetObject(oid ObjectID) Object {
	key := backendObjectKey(oid)
	hashbz := os.kvStore.Get([]byte(key))
	if hashbz == nil {
		panic(fmt.Sprintf("unexpected object with id %s", oid.String()))
	}
	hash := hashbz[:HashSize]
	bz := hashbz[HashSize:]

	var oo Object
	os.alloc.AllocateAmino(int64(len(bz)))
	amino.MustUnmarshal(bz, &oo)
	if debug {
		if oo.GetObjectID() != oid {
			panic(fmt.Sprintf("unexpected object id: expected %v but got %v",
				oid, oo.GetObjectID()))
		}
	}
	oo.SetHash(ValueHash{NewHashlet(hash)})
	_ = fillTypesOfValue(objectTypeStore{
		ObjectStore: os,
		TypeStore:   os.ts,
	}, oo) // XXX: os should be a cached type - find a way to have fillTypesOfValue cleanly.
	if debug {
		if _, ok := oo.(*PackageValue); ok {
			panic("packages must be fetched with GetPackage()")
		}
	}
	return oo
}

func (ds *objectStore) SetObject(oo Object) {
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
	if ds.kvStore != nil {
		key := backendObjectKey(oid)
		hashbz := make([]byte, len(hash)+len(bz))
		copy(hashbz, hash.Bytes())
		copy(hashbz[HashSize:], bz)
		ds.kvStore.Set([]byte(key), hashbz)
	}
	// save object to cache.
	if debug {
		if oid.IsZero() {
			panic("object id cannot be zero")
		}
	}
	// XXX: opslog
	// if escaped, add hash to iavl.
	if oo.GetIsEscaped() {
		var key, value []byte
		key = []byte(oid.String())
		value = hash.Bytes()
		ds.escapedStore.Set(key, value)
	}
}

func (os *objectStore) DelObject(oo Object) {
	oid := oo.GetObjectID()
	// delete from backend.
	key := backendObjectKey(oid)
	os.kvStore.Delete([]byte(key))
	// XXX OPSLOG
}

type TypeStore interface {
	GetType(tid TypeID) Type
	HasType(tid TypeID) bool
	SetType(Type)
}

// NewTypeStore creates a new [TypeStore], storing the types in the given
// kvStore.
func NewTypeStore(kvStore KeyValueStore) TypeStore {
	return &typeStore{kvStore}
}

type typeStore struct {
	kvStore KeyValueStore
}

func (ts *typeStore) GetType(tid TypeID) Type {
	tt := ts.getTypeSafe(tid)
	if tt == nil {
		panic(fmt.Sprintf("unexpected type with id %s", tid.String()))
	}
	return tt
}

func (ts *typeStore) HasType(tid TypeID) bool {
	return ts.getTypeSafe(tid) != nil
}

func (ts *typeStore) getTypeSafe(tid TypeID) Type {
	// check backend.
	key := backendTypeKey(tid)
	bz := ts.kvStore.Get([]byte(key))
	if bz == nil {
		return nil
	}

	var tt Type
	amino.MustUnmarshal(bz, &tt)
	if debug {
		if tt.TypeID() != tid {
			panic(fmt.Sprintf("unexpected type id: expected %v but got %v",
				tid, tt.TypeID()))
		}
	}
	fillType(ts, tt)

	return tt
}

func (ts *typeStore) SetType(tt Type) {
	tid := tt.TypeID()
	// save type to backend.
	key := backendTypeKey(tid)
	tcopy := copyTypeWithRefs(tt)
	bz := amino.MustMarshalAny(tcopy)
	ts.kvStore.Set([]byte(key), bz)
}

type BlockNodeStore interface {
	GetBlockNode(Location) BlockNode
	SetBlockNode(BlockNode)
}

func NewBlockNodeStore() BlockNodeStore {
	return &blockNodeStore{
		m: make(map[Location]BlockNode),
	}
}

type blockNodeStore struct {
	// XXX: this implementation should be changed
	m map[Location]BlockNode
}

func (bns *blockNodeStore) GetBlockNode(l Location) BlockNode {
	return bns.m[l]
}

func (bns *blockNodeStore) SetBlockNode(bn BlockNode) {
	loc := bn.GetLocation()
	if loc.IsZero() {
		panic("unexpected zero location in blocknode")
	}
	bns.m[loc] = bn
}

type FullStore interface {
	PackageStore
	MemPackageStore
	ObjectStore
	TypeStore
	BlockNodeStore

	GetNative(pkgPath string, name Name) func(m *Machine)
	GetAllocator() *Allocator
	Go2GnoType(rt reflect.Type) Type
}

// DebuggingStore is an interface with the debugging features of a store.
// A Store does not always implement this; to call a debugging function,
// a store should be type asserted to a DebuggingStore first.
type DebuggingStore interface {
	SetLogStoreOps(enabled bool)
	SprintStoreOps() string
	LogSwitchRealm(rlmpath string) // to mark change of realm boundaries
	ClearCache()
	Print()
}

type LocalStore interface {
	FullStore
	Begin(baseStore, iavlStore KeyValueStore) TransactionStore2
}

type TransactionStore2 interface {
	FullStore
	Write()
}
