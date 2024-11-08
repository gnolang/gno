package airdrop

import (
	"std"

	"gno.land/p/demo/ufmt"
)

type AirdropData interface {
	Bytes() []byte
	Address() std.Address
	Amount() uint64
}

// Data is complient with AirdropData interface
type Data struct {
	data struct {
		Address std.Address `json:"address"`
		Amount  uint64      `json:"amount"`
	}
}

func NewData(addr std.Address, amount uint64) Data {
	d := Data{}
	d.data.Address = addr
	d.data.Amount = amount
	return d
}

func (d Data) Bytes() []byte {
	out := ufmt.Sprintf(`{"address":"%s","amount":"%d"}`, d.data.Address, d.data.Amount)
	return []byte(out)
}

func (d Data) Address() std.Address {
	return d.data.Address
}

func (d Data) Amount() uint64 {
	return d.data.Amount
}
