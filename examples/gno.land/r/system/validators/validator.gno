package validators

import (
	"std"

	"gno.land/p/demo/ufmt"
)

type Validator struct {
	name    string
	address std.Address // owner address
	pubkey  string      // node pubkey(id)
	power   int64       // voting power
}

func newValidator(name string, address std.Address, pubkey string, power int64) *Validator {
	return &Validator{name, address, pubkey, power}
}

func (v Validator) Render() string {
	format := "|%s|%s|%s|%d|"
	return ufmt.Sprintf(format, v.name, v.address, v.pubkey, v.power)
}
