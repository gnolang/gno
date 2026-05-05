package std

type Realm interface {
	Address() Address
	PkgPath() string
	Previous() Realm
	Origin() Realm
	String() string
}

type Address string

func (a Address) String() string { return string(a) }
func (a Address) IsValid() bool  { return false }
