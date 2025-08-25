package std

type Realm interface {
	Address() Address
	PkgPath() string
	Coins() Gnocoins
	Send(coins Gnocoins, to Address) error
	Previous() Realm
	Origin() Realm
	String() string
}

type Address string

func (a Address) String() string { return string(a) }
func (a Address) IsValid() bool  { return false }

type Gnocoins []Gnocoin

type Gnocoin struct {
	Denom  string
	Amount int64
}
