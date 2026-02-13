package evm

import (
	"errors"
	"math"

	"gno.land/p/onbloc/uint256"
)

func getCallsVariables(stack *evmStack) (gas *uint256.Uint, addr *uint256.Uint, moneyToGive *uint256.Uint, argsOffset *uint256.Uint, argsSize *uint256.Uint, retOffset *uint256.Uint, retSize *uint256.Uint, err error) {
	if stack.length < 7 {
		err = tooFewStack
		return
	}
	gas, _ = stack.pop()
	addr, _ = stack.pop()
	moneyToGive, _ = stack.pop()
	argsOffset, _ = stack.pop()
	argsSize, _ = stack.pop()
	retOffset, _ = stack.pop()
	retSize, _ = stack.pop()
	return
}

func getDelegateCallVariables(stack *evmStack) (gas *uint256.Uint, addr *uint256.Uint, argsOffset *uint256.Uint, argsSize *uint256.Uint, retOffset *uint256.Uint, retSize *uint256.Uint, err error) {
	if stack.length < 6 {
		err = tooFewStack
		return
	}
	gas, _ = stack.pop()
	addr, _ = stack.pop()
	argsOffset, _ = stack.pop()
	argsSize, _ = stack.pop()
	retOffset, _ = stack.pop()
	retSize, _ = stack.pop()
	return
}

func opCall(s *state) error {
	// get the values pushed in stack
	gas, addr, moneyToGive, argsOffset, argsSize, retOffset, retSize, err := getCallsVariables(&s.stack)
	if err != nil {
		return err
	}

	if s.context == staticContext && !moneyToGive.IsZero() {
		return errors.New("can't send money from staticCall (call with non 0 value)")
	}

	// check for enough money
	if s.tmp.accounts[s.callerAddress].Balance.Cmp(moneyToGive) == -1 {
		s.stack.push(uint256.Zero())
		return nil
	}
	if s.gasLimit >= 0 && int(gas.Uint64()) > s.gasLimit-int(s.gasConsumed) {
		return notEnoughGas
	}
	memoryCost := uint64(0)
	callData := getByteCopyFromMemory(s.memory, argsOffset, argsSize)

	// create a tmp state
	tmpState := s.clone()
	tmpState.setCallData(callData)
	tmpState.setCallValue(moneyToGive)
	if s.gasLimit >= 0 {
		tmpState.setGasLimit(int(gas.Uint64()))
	}

	tmpState.setCallerAddress(addr.Uint64())

	// check for precompiled
	var res []byte
	if addr.Uint64() <= nbPrecompiledContract {
		f, ok := precompiled[precompiledAddress(addr.Uint64())]
		if ok {
			err = f(tmpState)
		}
	} else {
		// get the contrat code
		contract := tmpState.tmp.accounts[addr.Uint64()]
		if contract == nil || contract.AccountType != ContractAccount {
			s.stack.push(uint256.Zero())
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
		s.stack.push(uint256.Zero())
		return nil
	}

	s.stack.push(uint256.One())
	res = tmpState.getResult()

	// extend memory as needed, and copy the result (res) in parent process
	// also change storages, gas consumed and stuff like that
	extendMemory(s, retOffset, retSize, &memoryCost)
	copy(s.memory[retOffset.Uint64():retOffset.Uint64()+retSize.Uint64()], res)

	s.gasConsumed += uint(memoryCost)

	s.applyStateAfterReturn(tmpState)

	warm := false
	for _, e := range s.tmp.warmAccountAddress {
		if e.Eq(addr) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmAccountAddress = append(s.tmp.warmAccountAddress, *addr)
		s.gasConsumed += 2600
	}

	if s.tmp.accounts[s.callerAddress] != nil && s.tmp.accounts[s.callerAddress].Balance != nil {
		s.tmp.accounts[s.callerAddress].Balance.Sub(s.tmp.accounts[s.callerAddress].Balance, moneyToGive)
	}
	if s.tmp.accounts[addr.Uint64()] != nil && s.tmp.accounts[addr.Uint64()].Balance != nil {
		s.tmp.accounts[addr.Uint64()].Balance.Add(s.tmp.accounts[addr.Uint64()].Balance, moneyToGive)
	}
	return nil
}

// DEPRECATED
func opCallCode(s *state) error {
	// get the values pushed in stack
	gas, addr, moneyToGive, argsOffset, argsSize, retOffset, retSize, err := getCallsVariables(&s.stack)
	if err != nil {
		return err
	}

	// check for enough money
	if s.tmp.accounts[s.callerAddress].Balance.Cmp(moneyToGive) == -1 {
		s.stack.push(uint256.Zero())
		return nil
	}
	if s.gasLimit >= 0 && int(gas.Uint64()) > s.gasLimit-int(s.gasConsumed) {
		return notEnoughGas
	}
	memoryCost := uint64(0)
	callData := getByteCopyFromMemory(s.memory, argsOffset, argsSize)

	// create a tmp state
	tmpState := s.clone()
	tmpState.setCallData(callData)
	if s.gasLimit >= 0 {
		tmpState.setGasLimit(int(gas.Uint64()))
	}
	tmpState.setCallValue(moneyToGive)
	tmpState.transientStorage = s.transientStorage

	// check for precompiled
	var res []byte
	if addr.Uint64() <= nbPrecompiledContract {
		f, ok := precompiled[precompiledAddress(addr.Uint64())]
		if ok {
			err = f(tmpState)
		}
	} else {
		// get the contrat code
		contract := tmpState.tmp.accounts[addr.Uint64()]
		if contract == nil || contract.AccountType != ContractAccount {
			s.stack.push(uint256.Zero())
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
		s.stack.push(uint256.Zero())
		return nil
	}

	s.stack.push(uint256.One())
	res = tmpState.getResult()

	// extend memory as needed, and copy the result (res) in parent process
	// also change storages, gas consumed and stuff like that
	extendMemory(s, retOffset, retSize, &memoryCost)
	copy(s.memory[retOffset.Uint64():retOffset.Uint64()+retSize.Uint64()], res)

	s.gasConsumed += uint(memoryCost)

	s.applyStateAfterReturn(tmpState)

	warm := false
	for _, e := range s.tmp.warmAccountAddress {
		if e.Eq(addr) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmAccountAddress = append(s.tmp.warmAccountAddress, *addr)
		s.gasConsumed += 2600
	}

	return nil
}

func opDelegateCall(s *state) error {
	// get the values pushed in stack
	gas, addr, argsOffset, argsSize, retOffset, retSize, err := getDelegateCallVariables(&s.stack)
	if err != nil {
		return err
	}

	if s.gasLimit >= 0 && int(gas.Uint64()) > s.gasLimit-int(s.gasConsumed) {
		return notEnoughGas
	}
	memoryCost := uint64(0)
	callData := getByteCopyFromMemory(s.memory, argsOffset, argsSize)

	// create a tmp state
	tmpState := s.clone()
	tmpState.setCallData(callData)
	if s.gasLimit >= 0 {
		tmpState.setGasLimit(int(gas.Uint64()))
	}

	tmpState.setCallerAddress(s.originAddress)
	tmpState.transientStorage = s.transientStorage

	// check for precompiled
	var res []byte
	if addr.Uint64() <= nbPrecompiledContract {
		f, ok := precompiled[precompiledAddress(addr.Uint64())]
		if ok {
			err = f(tmpState)
		}
	} else {
		// get the contrat code
		contract := tmpState.tmp.accounts[addr.Uint64()]
		if contract == nil || contract.AccountType != ContractAccount {
			s.stack.push(uint256.Zero())
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
		s.stack.push(uint256.Zero())
		return nil
	}

	s.stack.push(uint256.One())
	res = tmpState.getResult()

	// extend memory as needed, and copy the result (res) in parent process
	// also change storages, gas consumed and stuff like that
	extendMemory(s, retOffset, retSize, &memoryCost)
	copy(s.memory[retOffset.Uint64():retOffset.Uint64()+retSize.Uint64()], res)

	s.gasConsumed += uint(memoryCost)

	s.applyStateAfterReturn(tmpState)

	warm := false
	for _, e := range s.tmp.warmAccountAddress {
		if e.Eq(addr) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmAccountAddress = append(s.tmp.warmAccountAddress, *addr)
		s.gasConsumed += 2600
	}

	return nil
}

func opStaticCall(s *state) error {
	// get the values pushed in stack
	gas, addr, argsOffset, argsSize, retOffset, retSize, err := getDelegateCallVariables(&s.stack)
	if err != nil {
		return err
	}

	if s.gasLimit >= 0 && int(gas.Uint64()) > s.gasLimit-int(s.gasConsumed) {
		return notEnoughGas
	}
	memoryCost := uint64(0)
	callData := getByteCopyFromMemory(s.memory, argsOffset, argsSize)

	// create a tmp state
	tmpState := s.clone()
	tmpState.context = staticContext
	tmpState.setCallData(callData)
	tmpState.setCallValue(uint256.Zero())
	if s.gasLimit >= 0 {
		tmpState.setGasLimit(int(gas.Uint64()))
	}
	tmpState.transientStorage = s.transientStorage

	// check for precompiled
	var res []byte
	if addr.Uint64() <= nbPrecompiledContract {
		f, ok := precompiled[precompiledAddress(addr.Uint64())]
		if ok {
			err = f(tmpState)
		}
	} else {
		// get the contrat code
		contract := tmpState.tmp.accounts[addr.Uint64()]
		if contract == nil || contract.AccountType != ContractAccount {
			s.stack.push(uint256.Zero())
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
		s.stack.push(uint256.Zero())
		return nil
	}

	s.stack.push(uint256.One())
	res = tmpState.getResult()

	// extend memory as needed, and copy the result (res) in parent process
	// also change storages, gas consumed and stuff like that
	extendMemory(s, retOffset, retSize, &memoryCost)
	copy(s.memory[retOffset.Uint64():retOffset.Uint64()+retSize.Uint64()], res)

	warm := false
	for _, e := range s.tmp.warmAccountAddress {
		if e.Eq(addr) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmAccountAddress = append(s.tmp.warmAccountAddress, *addr)
		s.gasConsumed += 2600
	}

	s.gasConsumed += uint(memoryCost)
	return nil
}

func extendMemory(s *state, retOffset *uint256.Uint, retSize *uint256.Uint, memoryCost *uint64) {
	memLen := uint64(len(s.memory))
	farest := retOffset.Uint64() + retSize.Uint64()
	if farest > memLen {
		extend := farest - memLen
		s.memory = append(s.memory, make([]byte, extend)...)

		memorySizeWord := (extend + 31) / 32
		*memoryCost = uint64(math.Pow(float64(memorySizeWord), 2))/512 + (3 * memorySizeWord)
	}
}

func opReturnDataSize(s *state) error {
	if s.lastCallState != nil {
		s.stack.push(uint256.NewUint(uint64(len(s.lastCallState.result))))
	} else {
		s.stack.push(uint256.Zero())
	}
	return nil
}

func opReturnDataCopy(s *state) error {

	destOffset, err := s.stack.pop()
	if err != nil {
		return err
	}
	offset, err := s.stack.pop()
	if err != nil {
		return err
	}
	size, err := s.stack.pop()
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
