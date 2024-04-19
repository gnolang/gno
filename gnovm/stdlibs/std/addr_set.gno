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

type AddressList struct {
	list *avl.Tree
}

func NewAddressList() *AddressList {
	return &AddressList{list: avl.NewTree()}
}

func (alist AddressList) Size() int {
	return alist.list.Size()
}

func (alist *AddressList) AddAddress(newAddr Address) error {
	if alist.list.Set(newAddr.String(), newAddr) {
		return errors.New("address already exists")
	}
	return nil
}

func (alist AddressList) HasAddress(newAddr Address) bool {
	return alist.list.Has(newAddr.String())
}
