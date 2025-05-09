package evm

import (
	"encoding/hex"
	"net/url"
	"std"
	"strconv"

	// "std"
	"strings"

	"gno.land/p/demo/evm"
	"gno.land/p/demo/uint256"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
)

// global

const br = "\n\n"

var env = &evm.Env{
	Logs:        []byte{},
	Storage:     map[uint256.Uint]*uint256.Uint{},
	Accounts:    []*evm.Caller{},
	Contracts:   []*evm.SmartContract{},
	Transaction: []evm.Transaction{},
}

var gnoUserToAccountHelper = map[std.Address]*evm.Caller{}
var addressToAccountHelper = map[string]*evm.Caller{}

// helpers

func assertGetConnectedUserFromRealm() *evm.Caller {
	addr := std.PreviousRealm().Address()
	if addr == "" {
		panic("invalid realm: realm must be user realm")
	}
	caller, ok := gnoUserToAccountHelper[addr]
	if !ok {
		panic("invalid realm: user not connected")
	}
	return caller
}

// public functions

func Publish(code string) {
	crossing()

	// check for caller
	caller := assertGetConnectedUserFromRealm()

	// check for bytecode is ok
	x, err := hex.DecodeString(code)
	if err != nil {
		panic("invalid contract: " + err.Error())
	}

	// push the contract
	addr := env.Contracts.Push(&evm.SmartContract{
		Code:    x,
		Creator: caller,
	})
	env.Transaction.PushCreation(0, caller, uint(addr))
}

func Faucet(n int) {
	crossing()
	assertGetConnectedUserFromRealm().Gas += uint(n)
}

func Register(addrStr string) {
	crossing()

	// check for desired address length
	if !strings.HasPrefix(strings.ToLower(addrStr), "0x") {
		addrStr = "0x" + addrStr
	}

	// check for desired address number
	evmAddr, err := uint256.FromHex(addrStr)
	if err != nil {
		panic("invalid address: address cant be converted into uint256, " + err.Error())
	}
	evmAddr.And(evmAddr, uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffff")) // xor to have 20 byte addr

	// check if user exists already
	addrStr = evmAddr.Hex() // to make it the same for everyone
	_, ok := addressToAccountHelper[addrStr]
	if ok {
		panic("invalid address: this address is already in user")
	}

	// create the user
	gnoAddr := std.PreviousRealm().Address()
	caller := &evm.Caller{
		EVMAddress: evmAddr,
		GnoAddress: gnoAddr,
		Balance: uint256.Zero(),
		Gas: 0,
	}
	env.Accounts = append(env.Accounts, caller)
	gnoUserToAccountHelper[gnoAddr] = caller
	addressToAccountHelper[addrStr] = caller
}

func Call(path string, gasLimit uint) {
	crossing()

	// check the args & caller
	path, err := url.QueryUnescape(path)
	if err != nil {
		panic("call: cant unescape string, " + err.Error())
	}
	caller := assertGetConnectedUserFromRealm()
	callexpr := strings.Split(path, "/")
	if len(callexpr) < 3 {
		panic("call: invalid number of arguments, expected 'call/address:<int hex>/function(types)/arg1/arg2/...'")
	}

	// get contract index
	contractIndex := uint64(0)
	if contractIndex, err = strconv.ParseUint(callexpr[1], 16, 64); err != nil || contractIndex >= uint64(len(env.Contracts)) {
		panic("call: bad address (bad number, or invalid address)")
	}

	// call the contract with parametters
	list := []*uint256.Uint{}
	for i := 3; i < len(callexpr); i++ {
		n, err := strconv.Atoi(callexpr[i])
		if err != nil {
			panic("call: invalid parametter, " + err.Error())
		}
		list = append(list, uint256.NewUint(uint64(n)))
	}

	// make the call
	funcName := callexpr[2]
	res, err := caller.RawContractCall(env, gasLimit, funcName, hex.EncodeToString(env.Contracts[contractIndex].Code), list...)
	if err != nil {
		panic("call runtime: " + err.Error())
	}
	env.Transaction.PushExecution(0, caller, res, uint(contractIndex)) // TODO: get the execution price
}

func Render(path string) string {
	// create the title + TODO: description
	title := md.H1("Gnotherum") + br

	// create the login part
	loginPart := md.H2("Login") + br
	loginPart += "There are " + strconv.Itoa(len(env.Accounts)) + " callers." + br
	registerLnk := txlink.Call("Register", "addrStr", "0")
	loginPart += "Click " + md.Link("here", registerLnk) + " to register yourself as a caller."

	// create the action part
	actionPart := md.H2("Actions") + br
	publishLnk := txlink.Call("Publish")
	faucetLnk := txlink.Call("Faucet", "n", "100000")
	callLnk := txlink.Call("Call", "path", "call/0/foo(uint256)/42", "gasLimit", "100000")
	actionPart += md.BulletList([]string{
		"Click " + md.Link("here", publishLnk) + " to publish a new contract.",
		"Click " + md.Link("here", faucetLnk) + " to give gas to your account (faucet).",
		"Click " + md.Link("here", callLnk) + " to call a contract.",
	})

	// create the transaction part
	transactionsPart := md.H2("Transactions")
	transTable := mdtable.Table{}
	transTable.Headers = []string{"id", "name"}
	for k, v := range env.Transaction {
		transTable.Append([]string{strconv.Itoa(k), v.String()})
	}
	transactionsPart += transTable.String()

	// put everything together
	return title + md.Columns([]string{
		loginPart,
		actionPart,
		transactionsPart,
	})
}
