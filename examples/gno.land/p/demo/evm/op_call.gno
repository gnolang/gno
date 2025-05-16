package evm

import (
	"errors"
	"math"

	"gno.land/p/demo/uint256"
)

func getVariables(stack *evmStack) (*uint256.Uint, *uint256.Uint, *uint256.Uint, *uint256.Uint, *uint256.Uint, *uint256.Uint, *uint256.Uint, error) {
	gas, err := stack.Pop()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	address, err := stack.Pop()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	moneyToGive, err := stack.Pop()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	argsOffset, err := stack.Pop()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	argsSize, err := stack.Pop()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	retOffset, err := stack.Pop()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	retSize, err := stack.Pop()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	return gas, address, moneyToGive, argsOffset, argsSize, retOffset, retSize, nil
}

func getCallData(argsOffset *uint256.Uint, s *State, argsSize *uint256.Uint) []byte {
	callData := make([]byte, 0)
	if int(argsOffset.Uint64()) >= len(s.memory) {
		callData = make([]byte, int(argsSize.Uint64()))
	} else if int(argsOffset.Uint64()+argsSize.Uint64()) > len(s.memory) {
		callData = make([]byte, int(argsSize.Uint64()))
		copy(callData, s.memory[argsOffset.Uint64():])
	} else {
		callData = s.memory[argsOffset.Uint64() : argsOffset.Uint64()+argsSize.Uint64()]
	}
	return callData
}

func opCall(s *State) error {
	gas, address, moneyToGive, argsOffset, argsSize, retOffset, retSize, err := getVariables(&s.stack)
	if err != nil {
		return err
	}

	if s.tmp.accounts[s.callerAddress].Balance.Cmp(moneyToGive) == -1 {
		s.stack.Push(uint256.Zero())
		return errors.New("not enough money to send " + moneyToGive.String() + " wei to call adress " + address.String())
	}
	if uint(gas.Uint64()) > s.tmp.accounts[s.callerAddress].Gas {
		return notEnoughGas
	}

	memoryCost := uint64(0)
	callData := getCallData(argsOffset, s, argsSize)

	tmpState := s.clone()
	tmpState.setCallData(callData)
	tmpState.setGasLimit(uint(gas.Uint64()))

	contract := tmpState.tmp.accounts[address.Uint64()]
	if contract == nil || contract.AccountType != ContractAccount {
		s.stack.Push(uint256.Zero())
		return nil
	}
	contractCode := contract.Code

	tmpState.setCode(contractCode)
	err = tmpState.executeByteCode()
	if err != nil {
		s.stack.Push(uint256.Zero())
		return nil
	}
	s.stack.Push(uint256.One())

	extendMemory(s, retOffset, retSize, &memoryCost)

	res := tmpState.getResult()
	copy(s.memory[retOffset.Uint64():retOffset.Uint64()+retSize.Uint64()], res)

	s.gasConsumed += uint(memoryCost) + tmpState.gasConsumed

	s.lastCallState = tmpState
	s.tmp.storage = tmpState.tmp.storage
	s.tmp.logs = tmpState.tmp.logs
	s.tmp.accounts = tmpState.tmp.accounts
	s.tmp.warmStorage = tmpState.tmp.warmStorage
	return nil
}

func extendMemory(s *State, retOffset *uint256.Uint, retSize *uint256.Uint, memoryCost *uint64) {
	memLen := uint64(len(s.memory))
	farest := retOffset.Uint64() + retSize.Uint64()
	if farest > memLen {
		extend := farest - memLen
		s.memory = append(s.memory, make([]byte, extend)...)

		memorySizeWord := (extend + 31) / 32
		*memoryCost = uint64(math.Pow(float64(memorySizeWord), 2))/512 + (3 * memorySizeWord)
	}
}

func opReturnDataSize(s *State) error {
	if s.lastCallState != nil {
		s.stack.Push(uint256.NewUint(uint64(len(s.lastCallState.result))))
	} else {
		s.stack.Push(uint256.Zero())
	}
	return nil
}
