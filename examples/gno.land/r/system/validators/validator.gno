package validators

import (
	"std"

	"gno.land/p/demo/ufmt"
)

type Validator struct {
	name    string
	address std.Address
	pubkey  string
	power   int64
}

func newValidator(name string, address std.Address, pubkey string, power int64) *Validator {
	return &Validator{name, address, pubkey, power}
}

func (v Validator) JSON() string {
	format := `{ "name": "%s", "address": "%s", "pubkey": "%s", "power": %d }`
	return ufmt.Sprintf(format, v.name, v.address, v.pubkey, v.power)
}

func (v Validator) Render() string {
	format := "|%s|%s|%d|"
	return ufmt.Sprintf(format, v.name, v.address, v.power)
}
