package evm

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"

	// "std"
	"strings"

	"gno.land/p/demo/evm"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/uint256"
)

var env = &evm.Env{
	Logs:        []byte{},
	Storage:     map[uint256.Uint]*uint256.Uint{},
	Accounts:    []*evm.Caller{},
	Contracts:   []*evm.SmartContract{},
	Transaction: []evm.Transaction{},
	GasLimit: uint256.NewUint(1000),
}

func Publish(code string) {
	crossing()
	x, err := hex.DecodeString(code)
	if err != nil {
		panic(err)
	}
	// TODO: env.Contracts.AddNew(...)
	env.Contracts = append(env.Contracts, &evm.SmartContract{
		Code: x,
		Creator: &evm.Caller{},
	})
}

func Call(funcName string, contractIndex int, list ...*uint256.Uint) string{
	crossing()
	me := &evm.Caller{
		Gas: 10000000,
	}

	res, err := me.RawContractCall(env, 10000000, funcName, hex.EncodeToString(env.Contracts[contractIndex].Code), list...)
	if err != nil {
		return "error: " + err.Error()
	}
	return ufmt.Sprintf("%d\n", binary.BigEndian.Uint64(res[len(res)-8:len(res)]))
}

func Render(path string) string {
	callexpr := strings.Split(path, "/")
	if len(callexpr) < 4 {
		return "not enough parametters"
	}

	contractIndex := 0
	if contractIndex, err := strconv.Atoi(callexpr[1]); err != nil || contractIndex >= len(env.Contracts) {
		return "bad number (or too big)"
	}

	list := []*uint256.Uint{}
	for i := 3; i < len(callexpr); i++ {
		n, err := strconv.Atoi(callexpr[i])
		if err != nil {
			return "error: " + err.Error()
		}
		list = append(list, uint256.NewUint(uint64(n)))
	}
	return cross(Call)(callexpr[2], contractIndex, list...)
}
