package gno

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/pkgs/amino"
	dbm "github.com/gnolang/gno/pkgs/db"
)

type Store interface {
	GetPackage(pkgPath string) *PackageValue
	SetPackage(*PackageValue)
	GetObject(oid ObjectID) Object
	SetObject(Object)
	DelObject(Object)
	GetType(tid TypeID) Type
	SetType(Type)
	SetLogStoreOps(enabled bool)
}

// Used to keep track of in-mem objects during tx.
type defaultStore struct {
	builtinPkgs  map[string]*PackageValue // TODO merge with cachePkgs map[string]StorePackageItem
	cachePkgs    map[string]*PackageValue
	cacheObjects map[ObjectID]Object
	cacheTypes   map[TypeID]Type
	backend      dbm.DB

	opslog []StoreOp // for debugging.
}

func NewStore(backend dbm.DB) *defaultStore {
	return &defaultStore{
		builtinPkgs:  make(map[string]*PackageValue),
		cachePkgs:    make(map[string]*PackageValue),
		cacheObjects: make(map[ObjectID]Object),
		cacheTypes:   make(map[TypeID]Type),
		backend:      backend,
	}
}

func (ds *defaultStore) GetPackage(pkgPath string) *PackageValue {
	if pv, exists := ds.builtinPkgs[pkgPath]; exists {
		return pv
	}
	if pv, exists := ds.cachePkgs[pkgPath]; exists {
		return pv
	}
	if ds.backend != nil {
		key := backendPackageKey(pkgPath)
		bz := ds.backend.Get([]byte(key))
		if bz == nil {
			return nil
		} else {
			var pv = new(PackageValue)
			amino.MustUnmarshal(bz, pv)
			ds.cachePkgs[pkgPath] = pv
			return pv
		}
	} else {
		return nil
	}
}

func (ds *defaultStore) SetPackage(pv *PackageValue) {
	pkgPath := pv.PkgPath
	if debug {
		if _, exists := ds.builtinPkgs[pkgPath]; exists {
			panic("builtin packages should not be modified")
		}
		if pv2, exists := ds.cachePkgs[pkgPath]; exists {
			if pv != pv2 {
				panic("duplicate package value")
			}
		}
	}
	ds.cachePkgs[pkgPath] = pv
	if ds.backend != nil {
		key := backendPackageKey(pkgPath)
		bz := amino.MustMarshal(pv)
		ds.backend.Set([]byte(key), bz)
	}
}

func (ds *defaultStore) SetBuiltinPackage(pv *PackageValue) {
	pkgPath := pv.PkgPath
	if pv2, exists := ds.builtinPkgs[pkgPath]; exists {
		if pv != pv2 {
			panic("duplicate (builtin) package value")
		}
	}
	if _, exists := ds.cachePkgs[pkgPath]; exists {
		panic("duplicate package value -- already cached")
	}
	ds.builtinPkgs[pkgPath] = pv
}

func (ds *defaultStore) GetObject(oid ObjectID) Object {
	// check cache.
	if oo, exists := ds.cacheObjects[oid]; exists {
		return oo
	}
	// check backend.
	if ds.backend != nil {
		key := backendObjectKey(oid)
		bz := ds.backend.Get([]byte(key))
		if bz == nil {
			return nil
		}
		var oo Object
		amino.MustUnmarshal(bz, &oo)
		if debug {
			if oo.GetObjectID() != oid {
				panic(fmt.Sprintf("unexpected object id: expected %v but got %v",
					oid, oo.GetObjectID()))
			}
		}
		ds.cacheObjects[oid] = oo
		return oo
	}
	return nil
}

func (ds *defaultStore) SetObject(oo Object) {
	// replace children/fields with Ref.
	o2 := copyWithRefs(nil, oo)
	// marshal to binary.
	bz := amino.MustMarshal(o2)
	// set hash.
	hash := HashBytes(bz) // XXX objectHash(bz)???
	oo.SetHash(ValueHash{hash})
	// save bytes to backend.
	if ds.backend != nil {
		key := backendObjectKey(oo.GetObjectID())
		ds.backend.Set([]byte(key), bz)
	}
	// save object to cache.
	oid := oo.GetObjectID()
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
		ds.cacheTypes[tid] = tt
		return tt
	}
	return nil
}

// NOTE: not used quite yet.
func (ds *defaultStore) SetType(tt Type) {
	tid := tt.TypeID()
	if debug {
		if tt2, exists := ds.cacheTypes[tid]; exists {
			if tt != tt2 {
				panic("duplicate type")
			}
		}
	}
	// save type to backend.
	if ds.backend != nil {
		// TODO: implement copyWithRefs() for Types.
		// TODO: for now,
		// key := backendTypeKey(tid)
		// ds.backend.Set([]byte(key), bz)
	}
	// save type to cache.
	ds.cacheTypes[tid] = tt
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
