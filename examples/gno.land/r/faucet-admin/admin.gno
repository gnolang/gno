package admin

import (
	"errors"
	"std"
)


// getter
func AdminAddress() std.Address {
	return adminAddress
}

func ControllerAddress() std.Address {
	return controlAddress
}

func InPause() bool {
	return setPause
}

func GetMessage() string {
	return faucetMsg
}


var reqAddr = make(map[std.Address]uint)

// TODO
// Rules need to be more clear / more detail / more reasonable
func PerTransferSend(addr std.Address) std.Coins {
	// if found exists > it's their (n)th faucet request > give 1 gnot
	if _, found := reqAddr[addr]; found {
		return std.Coins{std.Coin{"ugnot", 1000000}}
	} 

	// if not found > it's their 1st faucet request > give 10 gnot
	reqAddr[addr] = 1
	return std.Coins{std.Coin{"ugnot", 10000000}} 
}


// setter
// only admin is allowed to execute this function
func SetAdminInPause(inPause bool) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	setPause = inPause
	return nil
}

func SetAdminMessage(message string) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	faucetMsg = message
	return nil
}

func SetAdminPerTransferSend(send std.Coins) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	faucetMount = send
	return nil
}

func SetAdminAdminAddr(addr std.Address) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	adminAddress = addr
	return nil
}

func SetAdminControllerAddr(addr std.Address) error {
	if err := assertIsAdmin(); err != nil {
		return err
	}
	controlAddress = addr
	return nil
}

func assertIsAdmin() error {
	caller := std.GetOrigCaller()
	if caller != adminAddress {
		return errors.New("restricted for admin")
	}
	return nil
}