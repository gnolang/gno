package evm

import (
	"net/url"
	"std"
	"strconv"

	"strings"

	"gno.land/p/demo/evm"
	"gno.land/p/demo/uint256"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
)

// global

const br = "\n\n"

var (
	env = evm.NewEnv()

	gnoUserToAccountHelper = map[std.Address]*evm.Account{}
	addressToAccountHelper = map[string]*evm.Account{}
)

// helpers

func assertGetConnectedUserFromRealm() *evm.Account {
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

func splitCallData(callData string) (string, []*uint256.Uint) {
	callData, err := url.QueryUnescape(callData)
	if err != nil {
		panic("call: cant unescape string, " + err.Error())
	}
	callexpr := strings.Split(callData, " ")

	list := []*uint256.Uint{}
	for i := 1; i < len(callexpr); i++ {
		n, err := uint256.FromHex(callexpr[i])
		if err != nil {
			panic("call: invalid parametter, " + err.Error())
		}
		list = append(list, n)
	}
	return callexpr[0], list
}

// public functions
func ExecuteByteCode(_ realm, codeStr, callData, valueToSend string) {
	// check for caller
	caller := assertGetConnectedUserFromRealm()

	// parse args
	callDataFunc, args := splitCallData(callData)

	// execute raw byte code
	res, err := caller.Call(env, uint256.MustFromDecimal(valueToSend), codeStr, callDataFunc, args...)
	if err != nil {
		panic("execute bytecode runtime error: " + err.Error())
	}
	env.Transaction.PushExecution(0, caller, res, 0, true) // TODO: get the execution price
}

func CreateContract(_ realm, initCode, valueToSend string) {
	// check for caller
	caller := assertGetConnectedUserFromRealm()

	// execute raw byte code
	_, err := caller.CreateContract(env, initCode, uint256.MustFromDecimal(valueToSend))
	if err != nil {
		panic("execute bytecode runtime error: " + err.Error())
	}
	env.Transaction.PushContractCreation(0, caller, 0) // TODO: get the execution price
}

func Faucet(_ realm, weiToGive uint, gasToGive uint) {
	usr := assertGetConnectedUserFromRealm()
	usr.Balance.Add(usr.Balance, uint256.NewUint(uint64(weiToGive)))
	usr.Gas += gasToGive
}

func BuyGas(_ realm, nbGasToBuy uint64) {
	usr := assertGetConnectedUserFromRealm()
	price := nbGasToBuy * evm.GasPrice
	if usr.Balance.Lt(uint256.NewUint(uint64(price))) {
		panic("not enough wei to buy " + strconv.FormatUint(nbGasToBuy, 10) +
			" gas: " + usr.Balance.String() + " wei on account but it cost " + strconv.FormatUint(price, 10) + " wei")
	}
	usr.Balance.Sub(usr.Balance, uint256.NewUint(price))
	usr.Gas += uint(nbGasToBuy)
}

func Register(_ realm) {
	// create the user
	gnoAddr := std.PreviousRealm().Address()
	addr, err := env.Accounts.PushCaller(env, gnoAddr)

	if err != nil {
		panic("invalid contract: " + err.Error())
	}
	gnoUserToAccountHelper[gnoAddr] = env.Accounts[addr]
	env.Transaction.PushCallerCreation(0, env.Accounts[addr], addr)
}

func CallContract(_ realm, contractAddress, callData string, valueToSend string) {
	caller := assertGetConnectedUserFromRealm()

	// get contract index
	contractIndex := uint256.Zero()
	contractIndex, err := uint256.FromHex(contractAddress)
	if err != nil || env.Accounts[contractIndex.Uint64()] == nil {
		panic("call contract: bad address (bad number, or invalid address) or " + err.Error())
	}

	// parse args
	callDataFunc, args := splitCallData(callData)

	// execute contract code
	res, err := caller.CallContract(env, contractIndex.Uint64(), uint256.MustFromDecimal(valueToSend), callDataFunc, args...)
	if err != nil {
		panic("call contract runtime error: " + err.Error())
	}
	env.Transaction.PushExecution(0, caller, res, contractIndex.Uint64(), false) // TODO: get the execution price
}

func Render(path string) string {
	// create the title + TODO: description
	title := md.H1("Gnotherum") + br

	// create the login part
	loginPart := md.H2("Login") + br
	loginPart += "There are " + strconv.Itoa(len(env.Accounts)) + " accounts." + br
	registerLnk := txlink.Call("Register", "addrStr", "0")
	loginPart += "Click " + md.Link("here", registerLnk) + " to create an account."

	// create the action part
	actionPart := md.H2("Actions") + br
	executeByteCodeLnk := txlink.Call("ExecuteByteCode", "codeStr", "00", "callData", "foo(uint256,uint256) 0x42 0x1", "valueToSend", "10")
	faucetLnk := txlink.Call("Faucet", "weiToGive", "100000")
	buyGasLnk := txlink.Call("BuyGas", "nbGasToBuy", "100000")
	callLnk := txlink.Call("CallContract", "contractAddress", "12", "callData", "foo(uint256,uint256) 0x42 0x1", "valueToSend", "10")
	createLnk := txlink.Call("CreateContract", "initCode", "0xCACABEBE", "valueToSend", "10")
	actionPart += md.BulletList([]string{
		"Click " + md.Link("here", executeByteCodeLnk) + " to execute raw EVM bytecode.",
		"Click " + md.Link("here", faucetLnk) + " to give wei to your account (faucet).",
		"Click " + md.Link("here", buyGasLnk) + " to buy gas from wei.",
		"Click " + md.Link("here", callLnk) + " to call a contract.",
		"Click " + md.Link("here", createLnk) + " to create a new contract from it's init code.",
	})

	// list transactions
	transactionsPart := md.H2("Transactions")
	transTable := mdtable.Table{}
	transTable.Headers = []string{"id", "name"}
	for k, v := range env.Transaction {
		transTable.Append([]string{strconv.Itoa(k), v.String()})
	}
	transactionsPart += transTable.String()

	// list accounts
	accountPart := md.H2("Accounts")
	accountTable := mdtable.Table{}
	accountTable.Headers = []string{"id", "type", "gas", "wei"}
	for k, v := range env.Accounts {
		accountTable.Append([]string{"0x" + strconv.FormatUint(uint64(k), 16), v.AccountType.String(), strconv.Itoa(int(v.Gas)) + " gas", v.Balance.String() + " wei"})
	}
	accountPart += accountTable.String()

	// put everything together
	return title + md.Columns([]string{
		loginPart,
		actionPart,
	}) + md.Columns([]string{
		transactionsPart,
		accountPart,
	})
}
