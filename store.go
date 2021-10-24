package gno

type Store interface {
	GetPackage(pkgPath string) *PackageValue
	SetPackage(*PackageValue)
	GetObject(oid ObjectID) Object
	SetObject(Object)
	GetType(tid TypeID) Type
	SetType(Type)
}

// Used to keep track of in-mem objects during tx.
type CacheStore struct {
	CachePkgs    map[string]*PackageValue
	CacheObjects map[ObjectID]Object
	CacheTypes   map[TypeID]Type
	Store        Store
}

func NewCacheStore(store Store) CacheStore {
	return CacheStore{
		CachePkgs:    make(map[string]*PackageValue),
		CacheObjects: make(map[ObjectID]Object),
		CacheTypes:   make(map[TypeID]Type),
		Store:        store,
	}
}

func (cs CacheStore) GetPackage(pkgPath string) *PackageValue {
	if pv, exists := cs.CachePkgs[pkgPath]; exists {
		return pv
	}
	if cs.Store != nil {
		pv := cs.Store.GetPackage(pkgPath)
		cs.CachePkgs[pkgPath] = pv
		return pv
	} else {
		return nil
	}
}

func (cs CacheStore) SetPackage(pv *PackageValue) {
	pkgPath := pv.PkgPath
	if debug {
		if pv2, ex := cs.CachePkgs[pkgPath]; ex {
			if ex && pv != pv2 {
				panic("duplicate package value")
			}
		}
	}
	cs.CachePkgs[pkgPath] = pv
}

func (cs CacheStore) GetObject(oid ObjectID) Object {
	if oo, exists := cs.CacheObjects[oid]; exists {
		return oo
	}
	if cs.Store != nil {
		oo := cs.Store.GetObject(oid)
		cs.CacheObjects[oid] = oo
		return oo
	} else {
		return nil
	}
}

func (cs CacheStore) SetObject(oo Object) {
	oid := oo.GetObjectID()
	if debug {
		if oid.IsZero() {
			panic("object id cannot be zero")
		}
		if oo2, ex := cs.CacheObjects[oid]; ex {
			if ex && oo != oo2 {
				panic("duplicate object")
			}
		}
	}
	cs.CacheObjects[oid] = oo
}

func (cs CacheStore) GetType(tid TypeID) Type {
	if tt, exists := cs.CacheTypes[tid]; exists {
		return tt
	}
	if cs.Store != nil {
		tt := cs.Store.GetType(tid)
		cs.CacheTypes[tid] = tt
		return tt
	} else {
		return nil
	}
}

func (cs CacheStore) SetType(tt Type) {
	tid := tt.TypeID()
	if debug {
		if tt2, ex := cs.CacheTypes[tid]; ex {
			if ex && tt != tt2 {
				panic("duplicate type")
			}
		}
	}
	cs.CacheTypes[tid] = tt
}

func (cs CacheStore) Flush() {
	// XXX
}
