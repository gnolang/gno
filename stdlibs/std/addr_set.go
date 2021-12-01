package std

import "errors"

//----------------------------------------
// AddressSet

type AddressSet interface {
	Size() int
	AddAddress(Address) error
	HasAddress(Address) bool
}

//----------------------------------------
// AddressList implements AddressSet.
// TODO implement AddressTree with avl.

type AddressList []Address

func NewAddressList() *AddressList {
	return &AddressList{}
}

func (alist *AddressList) Size() int {
	return len(*alist)
}

func (alist *AddressList) AddAddress(newAddr Address) error {
	// TODO optimize with binary algorithm
	for _, addr := range *alist {
		if addr == newAddr {
			return errors.New("address already exists")
		}
	}
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
