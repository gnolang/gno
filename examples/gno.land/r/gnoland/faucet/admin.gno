package faucet

import (
	"errors"
	"std"
)

func AdminSetInPause(inPause bool) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}
	gInPause = inPause
	return ""
}

func AdminSetMessage(message string) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}
	gMessage = message
	return ""
}

func AdminSetTransferLimit(amount int64, denom string) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}
	gLimit = std.Coins{std.Coin{Denom: denom, Amount: amount}}
	return ""
}

func AdminSetAdminAddr(addr std.Address) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}
	gAdminAddr = addr
	return ""
}

func AdminAddController(addr std.Address) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}

	k := -1
	for _, v := range gControllers {
		if v == addr {
			return addr.String() + " exists, no need to add."
		}
	}
	for i, v := range gControllers {
		if v == "" {

			k = i
			break
		}
	}
	if k < 0 {
		return "can not add more controllers than allowed"
	}

	gControllers[k] = addr

	return ""
}

func AdminRemoveController(addr std.Address) string {
	if err := assertIsAdmin(); err != nil {
		return err.Error()
	}

	removed := false
	for i, v := range gControllers {
		if v == addr {

			gControllers[i] = ""
			removed = true
		}
	}

	if removed == false {
		return addr.String() + " is not on the controller list"
	}
	return ""
}

func assertIsAdmin() error {
	caller := std.GetOrigCaller()
	if caller != gAdminAddr {
		return errors.New("restricted for admin")
	}
	return nil
}
