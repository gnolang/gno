package faucet

import (
	"chain"
	"chain/runtime"
	"errors"
)

func AdminSetInPause(cur realm, inPause bool) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}
	gInPause = inPause
	return ""
}

func AdminSetMessage(cur realm, message string) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}
	gMessage = message
	return ""
}

func AdminSetTransferLimit(cur realm, amount int64) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}
	gLimit = chain.NewCoin("ugnot", amount)
	return ""
}

func AdminSetAdminAddr(cur realm, addr address) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}
	gAdminAddr = addr
	return ""
}

func AdminAddController(cur realm, addr address) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}

	size := gControllers.Size()

	if size >= gControllersMaxSize {
		return "can not add more controllers than allowed"
	}

	if gControllers.Has(addr.String()) {
		return addr.String() + " exists, no need to add."
	}

	gControllers.Set(addr.String(), addr)

	return ""
}

func AdminRemoveController(cur realm, addr address) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}

	if !gControllers.Has(addr.String()) {
		return addr.String() + " is not on the controller list"
	}

	_, ok := gControllers.Remove(addr.String())

	// it not should happen.
	// we will check anyway to prevent issues in the underline implementation.

	if !ok {
		return addr.String() + " is not on the controller list"
	}

	return ""
}

func assertIsAdmin() error {
	caller := runtime.PreviousRealm().Address()
	if caller != gAdminAddr {
		return errors.New("restricted for admin")
	}
	return nil
}
