package std

type Realm struct {
	Addr    Address
	PkgPath string
}

func (r Realm) IsUser() bool {
	return r.PkgPath == ""
}
