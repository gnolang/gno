package evm

import (
	"encoding/hex"
	"std"
	"testing"

	"gno.land/p/onbloc/uint256"
)

func TestCallSha256(t *testing.T) {
	// create me and the environnement
	env := NewEnv()
	me, err := env.Accounts.PushCaller(env, std.Address("g18dgpzqxnhmrsemnadngg403f8clf5l82jalxza")) // me
	if err != nil {
		t.Error(err)
	}

	// create the contract
	contractAddress := uint64(0x10) // lets begin at 16 (this is for my safety)
	env.Accounts[contractAddress] = &Account{
		AccountType: ContractAccount,
		EvmAddress:  contractAddress,
		GnoAddress:  "",
		Balance:     uint256.Zero(),
		Code: []byte{
			// First place the parameters in memory
			byte(PUSH0),
			byte(PUSH0),
			byte(MSTORE),

			// Do the call
			byte(PUSH1), 32, // retSize
			byte(PUSH1), 100, // retOffset
			byte(PUSH0),    // argsSize
			byte(PUSH0),    // argsOffset
			byte(PUSH0),    // value
			byte(PUSH1), 2, // address
			byte(PUSH4), 0xFF, 0xFF, 0xFF, 0xFF, // gas
			byte(CALL),

			// Return the resultttt
			byte(POP),
			byte(PUSH1), 32, // size
			byte(PUSH1), 100, // offset
			byte(RETURN),
		},
	}

	// make the call
	env.Accounts[me].Gas = 0xffffffff_ffffffff
	res, _, err := env.Accounts[me].executeCallHex(env, -1, "", env.Accounts[contractAddress].Code)
	if err != nil {
		t.Error(err)
	}

	// check the stringz
	hexString := hex.EncodeToString(res)
	expectedString := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hexString != expectedString {
		t.Error("expected e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855, got " + hexString)
	}
}
