package gno

type TestStore struct {
	GetPackageFn func(pkgPath string) *PackageValue
}

func (ts TestStore) GetPackage(pkgPath string) *PackageValue {
	return ts.GetPackageFn(pkgPath)
}

func (ts TestStore) SetPackage(*PackageValue) {
	panic("should not happen")
}

func (ts TestStore) GetObject(oid ObjectID) Object {
	panic("should not happen")
}

func (ts TestStore) SetObject(oo Object) {
	panic("should not happen")
}

func (ts TestStore) GetType(tid TypeID) Type {
	panic("should not happen")
}

func (ts TestStore) SetType(tt Type) {
	panic("should not happen")
}
