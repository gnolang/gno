package gno

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/pkgs/amino"
	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/iavl"
)

const iavlCacheSize = 1024 * 1024 // TODO increase and parameterize.

type PackageGetter func(pkgPath string) *PackageValue

type Store interface {
	SetPackageGetter(PackageGetter)
	GetPackage(pkgPath string) *PackageValue
	SetPackage(*PackageValue)
	GetObject(oid ObjectID) Object
	SetObject(Object)
	DelObject(Object)
	GetType(tid TypeID) Type
	SetType(Type)
	GetBlockNode(Location) BlockNode
	SetBlockNode(BlockNode)
	SetLogStoreOps(enabled bool)
	SprintStoreOps() string
	ClearCache()
	Print()
}

// Used to keep track of in-mem objects during tx.
type defaultStore struct {
	pkgGetter    PackageGetter // non-realm packages
	cacheObjects map[ObjectID]Object
	cacheTypes   map[TypeID]Type
	cacheNodes   map[Location]BlockNode
	backend      dbm.DB
	iavldb       dbm.DB
	iavltree     *iavl.MutableTree

	opslog  []StoreOp           // for debugging.
	current map[string]struct{} // for detecting import cycles.
}

func NewStore(backend dbm.DB) *defaultStore {
	var iavldb dbm.DB
	var iavltree *iavl.MutableTree
	if backend != nil {
		iavldb = dbm.NewPrefixDB(backend, []byte("iavl/"))
		iavltree = iavl.NewMutableTree(iavldb, iavlCacheSize)
	}
	ds := &defaultStore{
		pkgGetter:    nil,
		cacheObjects: make(map[ObjectID]Object),
		cacheTypes:   make(map[TypeID]Type),
		cacheNodes:   make(map[Location]BlockNode),
		backend:      backend,
		iavldb:       iavldb,
		iavltree:     iavltree,
		current:      make(map[string]struct{}),
	}
	SetBuiltinTypes(ds)
	return ds
}

func (ds *defaultStore) SetPackageGetter(pg PackageGetter) {
	ds.pkgGetter = pg
}

func (ds *defaultStore) GetPackage(pkgPath string) *PackageValue {
	oid := ObjectIDFromPkgPath(pkgPath)
	// first, try to load (package) object as usual.
	pvo := ds.GetObjectSafe(oid)
	if pv, ok := pvo.(*PackageValue); ok {
		return pv
	}
	// otherwise, fetch from pkgGetter.
	if ds.pkgGetter != nil {
		if _, exists := ds.current[pkgPath]; exists {
			panic(fmt.Sprintf("import cycle detected: %q", pkgPath))
		}
		ds.current[pkgPath] = struct{}{}
		defer delete(ds.current, pkgPath)
		if pv := ds.pkgGetter(pkgPath); pv != nil {
			if pv.IsRealm() {
				panic("realm packages cannot be gotten from pkgGetter")
			}
			ds.cacheObjects[oid] = pv
			return pv
		}
	}
	panic(fmt.Sprintf("unexpected package with path %q", pkgPath))
}

// Setting an already cached package (eg modifying it) fails unless realm package.
func (ds *defaultStore) SetPackage(pv *PackageValue) {
	oid := pv.ObjectInfo.ID
	if oid.IsZero() {
		// .SetRealm() should have set object id.
		panic("should not happen")
	}
	ds.SetObject(pv)
}

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
	// check backend.
	if ds.backend != nil {
		key := backendObjectKey(oid)
		bz := ds.backend.Get([]byte(key))
		if bz != nil {
			var oo Object
			amino.MustUnmarshal(bz, &oo)
			if debug {
				if oo.GetObjectID() != oid {
					panic(fmt.Sprintf("unexpected object id: expected %v but got %v",
						oid, oo.GetObjectID()))
				}
			}
			_ = fillTypes(ds, oo)
			ds.cacheObjects[oid] = oo
			return oo
		}
	}
	return nil
}

func (ds *defaultStore) SetObject(oo Object) {
	oid := oo.GetObjectID()
	// replace children/fields with Ref.
	o2 := copyWithRefs(nil, oo)
	// marshal to binary.
	bz := amino.MustMarshalAny(o2)
	// set hash.
	hash := HashBytes(bz) // XXX objectHash(bz)???
	oo.SetHash(ValueHash{hash})
	// save bytes to backend.
	if ds.backend != nil {
		key := backendObjectKey(oid)
		ds.backend.Set([]byte(key), bz)
	}
	// save object to cache.
	if debug {
		if oid.IsZero() {
			panic("object id cannot be zero")
		}
		if oo2, exists := ds.cacheObjects[oid]; exists {
			if oo != oo2 {
				panic("duplicate object")
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
			StoreOp{op, o2.(Object)})
	}
	// if escaped, add hash to iavl.
	if oo.GetIsEscaped() {
		var key, value []byte
		key = []byte(oid.String())
		value = hash.Bytes()
		ds.iavltree.Set(key, value)
	}
}

func (ds *defaultStore) DelObject(oo Object) {
	oid := oo.GetObjectID()
	// delete from cache.
	delete(ds.cacheObjects, oid)
	// delete from backend.
	if ds.backend != nil {
		key := backendObjectKey(oid)
		ds.backend.Delete([]byte(key))
	}
	// make realm op log entry
	if ds.opslog != nil {
		ds.opslog = append(ds.opslog,
			StoreOp{StoreOpDel, oo})
	}
}

// NOTE: not used quite yet.
// NOTE: The implementation matches that of GetObject() in anticipation of what
// the persistent type system might work like.
func (ds *defaultStore) GetType(tid TypeID) Type {
	// check cache.
	if tt, exists := ds.cacheTypes[tid]; exists {
		return tt
	}
	// check backend.
	if ds.backend != nil {
		key := backendTypeKey(tid)
		bz := ds.backend.Get([]byte(key))
		if bz != nil {
			var tt Type
			amino.MustUnmarshal(bz, &tt)
			if debug {
				if tt.TypeID() != tid {
					panic(fmt.Sprintf("unexpected type id: expected %v but got %v",
						tid, tt.TypeID()))
				}
			}
			ds.cacheTypes[tid] = tt
			return tt
		}
	}
	panic(fmt.Sprintf("unexpected type with id %s", tid.String()))
}

// NOTE: not used quite yet.
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
	if ds.backend != nil {
		// TODO: implement copyWithRefs() for Types.
		// key := backendTypeKey(tid)
		// ds.backend.Set([]byte(key), bz)
	}
	// save type to cache.
	ds.cacheTypes[tid] = tt
	// mark type as saved.
	if !tt.GetIsSaved() {
		tt.SetIsSaved()
	}
}

func (ds *defaultStore) GetBlockNode(loc Location) BlockNode {
	// check cache.
	if bn, exists := ds.cacheNodes[loc]; exists {
		return bn
	}
	// check backend.
	if ds.backend != nil {
		key := backendNodeKey(loc)
		bz := ds.backend.Get([]byte(key))
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
	panic(fmt.Sprintf("unexpected node with location %s", loc.String()))
}

func (ds *defaultStore) SetBlockNode(bn BlockNode) {
	loc := bn.GetLocation()
	if loc.IsZero() {
		panic("unexpected zero location in blocknode")
	}
	// save node to backend.
	if ds.backend != nil {
		// TODO: implement copyWithRefs() for Nodes.
		// key := backendNodeKey(loc)
		// ds.backend.Set([]byte(key), bz)
	}
	// save node to cache.
	ds.cacheNodes[loc] = bn
}

func (ds *defaultStore) Flush() {
	// XXX
}

//----------------------------------------
// StoreOp

type StoreOpType uint8

const (
	StoreOpNew StoreOpType = iota
	StoreOpMod
	StoreOpDel
)

type StoreOp struct {
	Type   StoreOpType
	Object Object // ref'd objects
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

func (ds *defaultStore) ClearCache() {
	ds.cacheObjects = make(map[ObjectID]Object)
	// NOTE: types/nodes are not yet persisted, so keep cache.
	// ds.cacheTypes = make(map[TypeID]Type)
	// ds.cacheNodes = make(map[Location]BlockNode)
}

// for debugging
func (ds *defaultStore) Print() {
	ds.backend.Print()
}

//----------------------------------------
// backend keys

func backendPackageKey(pkgPath string) string {
	return "pkg:" + pkgPath
}

func backendObjectKey(oid ObjectID) string {
	return "oid:" + oid.String()
}

func backendTypeKey(tid TypeID) string {
	return "tid:" + tid.String()
}

func backendNodeKey(loc Location) string {
	return "node:" + loc.String()
}

//----------------------------------------
// builtin types

func SetBuiltinTypes(store Store) {
	types := []Type{
		BoolType,
		StringType,
		IntType, Int8Type, Int16Type, Int32Type, Int64Type,
		UintType, Uint8Type, Uint16Type, Uint32Type, Uint64Type,
		BigintType,
		gTypeType,
		gPackageType,
		blockType{},
		Float32Type, Float64Type,
	}
	for _, tt := range types {
		store.SetType(tt)
	}
}
