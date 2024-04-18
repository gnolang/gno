package std

import (
	"errors"

	"gno.land/p/demo/avl"
)

type AddressSet interface {
	Size() int
	AddAddress(Address) error
	HasAddress(Address) bool
}

type AddressList avl.Tree

func NewAddressList() *AddressList {
	return avl.Newtree()
}

func (alist AddressList) Size() int {
	return alist.Size()
}

func (alist *AddressList) AddAddress(newAddr Address) error {
	if alist.Set(newAddr, newAddr) {
		return errors.New("address already exists")
	}
	return nil
}

func (alist AddressList) HasAddress(newAddr Address) bool {
	return alist.Has(newAddr)
}
