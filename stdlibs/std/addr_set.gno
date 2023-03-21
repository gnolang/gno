package std

import (
	"avl"
	"errors"
)

//----------------------------------------
// AddressSet

type AddressSet interface {
	Size() int
	AddAddress(Address) error
	HasAddress(Address) bool
}

//----------------------------------------
// AddressList implements AddressSet.
// AddressList is a slice of Address and can contain duplicated addresses
// to prevent address duplication, use AddressTree
type AddressList []Address

func NewAddressList() *AddressList {
	return &AddressList{}
}

func (alist *AddressList) Size() int {
	return len(*alist)
}

func (alist *AddressList) AddAddress(newAddr Address) error {
	// TODO optimize with binary algorithm
	*alist = append(*alist, newAddr)
	return nil
}

func (alist *AddressList) HasAddress(newAddr Address) bool {
	// TODO optimize with binary algorithm
	for _, addr := range *alist {
		if addr == newAddr {
			return true
		}
	}
	return false
}

//----------------------------------------
// AddressTree implements AddressSet.
// AddresTree stores Addresses in a MutableTree and can't store duplicates
// if you need address duplication, use AddressList
type AddressTree struct {
	tree *avl.MutTree
}

func NewAddressTree() *AddressTree {
	return &AddressTree{
		tree: avl.NewMutTree(),
	}
}

func (atree AddressTree) Size() int {
	return atree.tree.Size()
}

func (atree *AddressTree) AddAddress(newAddr Address) error {
	if atree.tree.Has(newAddr.String()) {
		return errors.New("address already exists")
	}
	atree.tree.Set(newAddr.String(), struct{}{})
	return nil
}

func (atree *AddressTree) HasAddress(newAddr Address) bool {
	return atree.tree.Has(newAddr.String())
}
