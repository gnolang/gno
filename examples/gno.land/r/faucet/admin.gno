package faucet

import (
	"errors"
	"std"
)

func AdminSetInPause(inPause bool) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	gInPause = inPause
	return nil
}

func AdminSetMessage(message string) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	gMessage = message
	return nil
}

func AdminSetPerTransferSend(send std.Coins) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	gPerTransferSend = send
	return nil
}

func AdminSetAdminAddr(addr std.Address) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	gAdminAddr = addr
	return nil
}

func AdminSetControllerAddr(addr std.Address) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	gControllerAddr = addr
	return nil
}

func assertIsAdmin() error {
	caller := std.GetOrigCaller()
	if caller != gAdminAddr {
		return errors.New("restricted for admin")
	}
	return nil
}
