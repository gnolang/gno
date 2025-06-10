package evm

import (
	"errors"
	"math"

	"gno.land/p/demo/uint256"
)

func getVariables(stack *evmStack) (gas *uint256.Uint, address *uint256.Uint, moneyToGive *uint256.Uint, argsOffset *uint256.Uint, argsSize *uint256.Uint, retOffset *uint256.Uint, retSize *uint256.Uint, err error) {
	if stack.length < 7 {
		err = tooFewStack
		return
	}
	gas, _ = stack.Pop()
	address, _ = stack.Pop()
	moneyToGive, _ = stack.Pop()
	argsOffset, _ = stack.Pop()
	argsSize, _ = stack.Pop()
	retOffset, _ = stack.Pop()
	retSize, _ = stack.Pop()
	return
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
	// get the values pushed in stack
	gas, address, moneyToGive, argsOffset, argsSize, retOffset, retSize, err := getVariables(&s.stack)
	if err != nil {
		return err
	}

	// check for enough money
	if s.tmp.accounts[s.callerAddress].Balance.Cmp(moneyToGive) == -1 {
		s.stack.Push(uint256.Zero())
		return nil
	}
	if uint(gas.Uint64()) > s.tmp.accounts[s.callerAddress].Gas {
		s.stack.Push(uint256.Zero())
		return notEnoughGas
	}
	memoryCost := uint64(0)
	callData := getCallData(argsOffset, s, argsSize)

	// create a tmp state
	tmpState := s.clone()
	tmpState.setCallData(callData)
	tmpState.setGasLimit(uint(gas.Uint64()))
	tmpState.setCallerAddress(address.Uint64())

	// check for precompiled
	var res []byte
	if address.Uint64() <= nbPrecompiledContract {
		f, ok := precompiled[address.Uint64()]
		if ok {
			err = f(tmpState)
		}
	} else {
		// get the contrat code
		contract := tmpState.tmp.accounts[address.Uint64()]
		if contract == nil || contract.AccountType != ContractAccount {
			s.stack.Push(uint256.Zero())
			return nil
		}
		contractCode := contract.Code

		// execute the bytecode in tmpState
		tmpState.setCode(contractCode)
		err = tmpState.executeByteCode()
	}
	s.gasConsumed += tmpState.gasConsumed

	// check for error, and push 1 or 0 if success
	if err != nil {
		s.stack.Push(uint256.Zero())
		return nil
	}

	s.stack.Push(uint256.One())
	res = tmpState.getResult()

	// extend memory as needed, and copy the result (res) in parent process
	// also change storages, gas consumed and stuff like that
	extendMemory(s, retOffset, retSize, &memoryCost)
	copy(s.memory[retOffset.Uint64():retOffset.Uint64()+retSize.Uint64()], res)

	s.gasConsumed += uint(memoryCost)
	s.lastCallState = tmpState
	s.tmp.storage = tmpState.tmp.storage
	s.tmp.logs = tmpState.tmp.logs
	s.tmp.accounts = tmpState.tmp.accounts
	s.tmp.warmStorage = tmpState.tmp.warmStorage
	if s.tmp.accounts[s.callerAddress] != nil && s.tmp.accounts[s.callerAddress].Balance != nil {
		s.tmp.accounts[s.callerAddress].Balance.Sub(s.tmp.accounts[s.callerAddress].Balance, moneyToGive)
	}
	if s.tmp.accounts[address.Uint64()] != nil && s.tmp.accounts[address.Uint64()].Balance != nil {
		s.tmp.accounts[address.Uint64()].Balance.Add(s.tmp.accounts[address.Uint64()].Balance, moneyToGive)
	}
	return nil
}

func opCallCode(s *State) error {
	// get the values pushed in stack
	gas, address, moneyToGive, argsOffset, argsSize, retOffset, retSize, err := getVariables(&s.stack)
	if err != nil {
		return err
	}

	// check for enough money
	if s.tmp.accounts[s.callerAddress].Balance.Cmp(moneyToGive) == -1 {
		s.stack.Push(uint256.Zero())
		return nil
	}
	if uint(gas.Uint64()) > s.tmp.accounts[s.callerAddress].Gas {
		s.stack.Push(uint256.Zero())
		return notEnoughGas
	}
	memoryCost := uint64(0)
	callData := getCallData(argsOffset, s, argsSize)

	// create a tmp state
	tmpState := s.clone()
	tmpState.setCallData(callData)
	tmpState.setGasLimit(uint(gas.Uint64()))
	tmpState.transientStorage = s.transientStorage

	// check for precompiled
	var res []byte
	if address.Uint64() <= nbPrecompiledContract {
		f, ok := precompiled[address.Uint64()]
		if ok {
			err = f(tmpState)
		}
	} else {
		// get the contrat code
		contract := tmpState.tmp.accounts[address.Uint64()]
		if contract == nil || contract.AccountType != ContractAccount {
			s.stack.Push(uint256.Zero())
			return nil
		}
		contractCode := contract.Code

		// execute the bytecode in tmpState
		tmpState.setCode(contractCode)
		err = tmpState.executeByteCode()
	}
	s.gasConsumed += tmpState.gasConsumed

	// check for error, and push 1 or 0 if success
	if err != nil {
		s.stack.Push(uint256.Zero())
		return nil
	}

	s.stack.Push(uint256.One())
	res = tmpState.getResult()

	// extend memory as needed, and copy the result (res) in parent process
	// also change storages, gas consumed and stuff like that
	extendMemory(s, retOffset, retSize, &memoryCost)
	copy(s.memory[retOffset.Uint64():retOffset.Uint64()+retSize.Uint64()], res)

	s.gasConsumed += uint(memoryCost)
	s.lastCallState = tmpState
	s.tmp.storage = tmpState.tmp.storage
	s.tmp.logs = tmpState.tmp.logs
	s.tmp.accounts = tmpState.tmp.accounts
	s.tmp.warmStorage = tmpState.tmp.warmStorage
	if s.tmp.accounts[s.callerAddress] != nil && s.tmp.accounts[s.callerAddress].Balance != nil {
		s.tmp.accounts[s.callerAddress].Balance.Sub(s.tmp.accounts[s.callerAddress].Balance, moneyToGive)
	}
	if s.tmp.accounts[address.Uint64()] != nil && s.tmp.accounts[address.Uint64()].Balance != nil {
		s.tmp.accounts[address.Uint64()].Balance.Add(s.tmp.accounts[address.Uint64()].Balance, moneyToGive)
	}
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

func opReturnDataCopy(s *State) error {

	destOffset, err := s.stack.Pop()
	if err != nil {
		return err
	}
	offset, err := s.stack.Pop()
	if err != nil {
		return err
	}
	size, err := s.stack.Pop()
	if err != nil {
		return err
	}

	if s.lastCallState == nil {
	}

	if offset.Add(offset, size).Uint64() > uint64(len(s.lastCallState.result)) {
		return errors.New("return data copy overflow")
	}

	memoryCost := uint64(0)
	extendMemory(s, destOffset, size, &memoryCost)

	res := s.lastCallState.getResult()
	copy(s.memory[destOffset.Uint64():destOffset.Uint64()+size.Uint64()], res[offset.Uint64():])

	minimumWordSize := (size.Uint64() + 31) / 32
	s.gasConsumed += uint(memoryCost + 3*minimumWordSize)

	return nil
}
