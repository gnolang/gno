package evm

import (
	"errors"
	"strconv"
	"time"

	"gno.land/p/demo/evm/keccak256"
	"gno.land/p/demo/uint256"
)

func opPush0(s *State) error {
	s.stack.Push(uint256.Zero())
	return nil
}

func opPush(n int) func(*State) error {
	return func(s *State) error {
		s.code.increment()
		if s.code.ill() {
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + outOfBounds.Error())
		}
		stackData, err := s.code.read(n)
		if err != nil {
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + err.Error())
		}
		s.stack.Push(stackData.Clone())
		return nil
	}
}

func opSwap(n int) func(*State) error {
	return func(s *State) error {
		sLen := s.stack.Len()
		if sLen <= n {
			return errors.New("can't SWAP" + strconv.Itoa(n) + ": stack too small")
		}
		sLen -= 1
		s.stack.data[sLen], s.stack.data[sLen-n] = s.stack.data[sLen-n], s.stack.data[sLen]
		return nil
	}
}

func opDup(n int) func(*State) error {
	return func(s *State) error {
		if s.stack.Len() < n {
			return errors.New("can't DUP" + strconv.Itoa(n) + ": stack too small")
		}
		val, err := s.stack.Peek(n - 1)
		if err != nil {
			return errors.New("can't DUP" + strconv.Itoa(n) + ": " + err.Error())
		}

		s.stack.Push(val.Clone())
		return nil
	}
}

func opPop(s *State) error {
	_, err := s.stack.Pop()
	return err
}

func opLog(nTopics int) func(*State) error {
	return func(s *State) error {
		if s.context == staticContext {
			return errors.New("can't log from staticCall")
		}

		topics := make([][]byte, nTopics)
		for i := 0; i < nTopics; i++ {
			topicVal, err := s.stack.Pop()
			if err != nil {
				return err
			}
			arr := topicVal.Bytes32()
			topics[i] = make([]byte, 32)
			copy(topics[i], arr[:])
		}
		offsetVal, err := s.stack.Pop()
		if err != nil {
			return err
		}
		sizeVal, err := s.stack.Pop()
		if err != nil {
			return err
		}
		offset := offsetVal.Uint64()
		size := sizeVal.Uint64()
		if offset > uint64(len(s.memory)) {
			return errors.New("memory out of bounds")
		}
		if offset+size > uint64(len(s.memory)) {
			size = uint64(len(s.memory)) - offset
		}
		data := make([]byte, size)
		copy(data, s.memory[offset:offset+size])
		s.tmp.logs = append(s.tmp.logs, LogEntry{
			Topics: topics,
			Data:   data,
		})
		return nil
	}
}

func opJump(s *State) error {
	dest, err := s.stack.Pop()
	if err != nil {
		return err
	}
	inc := int(dest.Uint64())
	opDest := OpCode(s.code.at(inc))
	if opDest != JUMPDEST {
		opData, ok := opCodes[opDest]
		opCodeStr := "0x" + strconv.FormatUint(uint64(opDest), 16)
		if ok {
			opCodeStr = opData.name
		}
		return errors.New("can't JUMP: JUMP needs to land on a JUMPDEST, but at " + dest.Hex() + " there is a " + opCodeStr)
	}
	s.code.rip = inc
	return nil
}

func opNop(s *State) error {
	return nil
}

func opJumpi(s *State) error {
	dest, err := s.stack.Pop()
	if err != nil {
		return err
	}
	doJump, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if doJump.IsZero() {
		return nil
	}
	inc := int(dest.Uint64())
	opDest := OpCode(s.code.at(inc))
	if opDest != JUMPDEST {
		opData, ok := opCodes[opDest]
		opCodeStr := "0x" + strconv.FormatUint(uint64(opDest), 16)
		if ok {
			opCodeStr = opData.name
		}
		return errors.New("can't JUMPI: JUMPI needs to land on a JUMPDEST, but at " + dest.Hex() + " there is a " + opCodeStr)
	}
	s.code.rip = inc
	return nil
}

func opIszero(s *State) error {
	num, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if num.IsZero() {
		s.stack.Push(uint256.NewUint(1))
	} else {
		s.stack.Push(uint256.Zero())
	}
	return nil
}

func opKeccak256(s *State) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	size, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	if off64 > uint64(len(s.memory)) {
		s.stack.Push(uint256.Zero())
		return nil
	}

	minWordSize := uint256.Zero().Div(uint256.Zero().Add(size, uint256.NewUint(31)), uint256.NewUint(32))
	s.gasConsumed += uint(6 * minWordSize.Uint64()) // TODO: runtime gas with mem extension (idk how to expand)

	slice := s.memory[off64 : off64+size.Uint64()]
	resBytes := keccak256.Keccak256(slice)
	resUint := uint256.Zero().SetBytes(resBytes)
	s.stack.Push(resUint)
	return nil
}

func opPc(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(s.code.rip)))
	return nil
}

func opCodeSize(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(s.code.length)))
	return nil
}

// EXTCODESIZE can't be implemented, it only push 0 on stack
func opExtCodeSize(s *State) error {
	_, err := s.stack.Pop()
	if err != nil {
		return err
	}
	s.stack.Push(uint256.Zero())
	return nil
}

func opGas(s *State) error {
	if s.gasLimit < 0 {
		s.stack.Push(uint256.Zero())
	} else {
		s.stack.Push(uint256.NewUint(uint64(s.gasLimit - int(s.gasConsumed))))
	}
	return nil
}

func opTimestamp(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(time.Now().Unix())))
	return nil
}

func opPrevRandao(s *State) error {
	s.stack.Push(s.env.BlockDifficulty.Clone())
	return nil
}

func opGaslimit(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(s.gasLimit)))
	return nil
}

func opReturn(s *State) error {
	if s.stack.Len() < 2 {
		return tooFewStack
	}
	off, _ := s.stack.Pop()
	size, _ := s.stack.Pop()
	memLen256 := uint256.NewUint(uint64(len(s.memory)))
	if !size.IsZero() {
		// extend memory if needed and consume gas
		end := uint256.Zero().Add(off, size)
		if end.Gt(memLen256) {
			more := end.Sub(end, memLen256).Uint64()
			s.memory = append(s.memory, make([]byte, more)...)
			s.gasConsumed += uint(more)
		}
		s.result = append(s.result, s.memory[off.Uint64():end.Uint64()]...)
	}
	s.code.stop()
	return nil
}

func opRevert(s *State) error {
	_ = opInvalid(s)
	err := opReturn(s)
	if err != nil {
		return err
	}
	return revertError
}

func opInvalid(s *State) error {
	return invalidPanic
}

func opCaller(s *State) error {
	s.stack.Push(uint256.NewUint(s.callerAddress))
	return nil
}

func opCallvalue(s *State) error {
	s.stack.Push(s.callValue.Clone())
	return nil
}

func opGasprice(s *State) error {
	s.stack.Push(uint256.NewUint(GasPrice))
	return nil
}

// TODO: tests
func opChainId(s *State) error {
	if s.env != nil && s.env.ChainID != nil {
		s.stack.Push(s.env.ChainID.Clone())
	} else {
		s.stack.Push(uint256.Zero())
	}
	return nil
}

// TODO: tests
func opSelfBalance(s *State) error {
	s.stack.Push(s.tmp.accounts[s.callerAddress].Balance.Clone())
	return nil
}

// TODO: tests
func opNumber(s *State) error {
	blockNumber := uint64(time.Now().Unix())
	s.stack.Push(uint256.NewUint(blockNumber))
	return nil
}

func opBaseFee(s *State) error {
	prevBaseFee := uint64(202700000000) // wei
	includedGas := uint64(30000000)
	targetGas := uint64(15000000)
	if env := s.env; env != nil {
		if env.PrevBaseFee != 0 {
			prevBaseFee = env.PrevBaseFee
		}
		if env.IncludedGas != 0 {
			includedGas = env.IncludedGas
		}
		if env.TargetGas != 0 {
			targetGas = env.TargetGas
		}
	}
	delta := float64(int64(includedGas)-int64(targetGas)) / float64(targetGas)
	adjustment := delta * float64(prevBaseFee) / 8               // adjust base fee (12.5% limit)
	feeResult := uint64(float64(prevBaseFee) + adjustment + 0.5) // rounded
	s.stack.Push(uint256.NewUint(feeResult))
	return nil
}

func opBlobBaseFee(s *State) error {
	if s.env != nil && s.env.BlobBaseFee != nil {
		s.stack.Push(s.env.BlobBaseFee.Clone())
	} else {
		s.stack.Push(uint256.Zero())
	}
	return nil
}

func opBlobHash(s *State) error {
	index, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if s.env == nil || s.env.BlobHashes == nil {
		s.stack.Push(uint256.Zero())
		return nil
	}
	i := index.Uint64()
	if i >= uint64(len(s.env.BlobHashes)) {
		s.stack.Push(uint256.Zero())
		return nil
	}
	blobHash := s.env.BlobHashes[i]
	s.stack.Push(blobHash.Clone())
	return nil
}

func opCoinbase(s *State) error {
	if s.env != nil && s.env.Coinbase != nil {
		s.stack.Push(s.env.Coinbase.Clone()) // miner address -> coinbase
	} else {
		s.stack.Push(uint256.Zero())
	}
	return nil
}

func opBlockHash(s *State) error {
	blockNumber, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if s.env == nil || s.env.BlockNumber == nil {
		s.stack.Push(uint256.Zero())
		return nil
	}
	currentBlock := s.env.BlockNumber.Uint64()
	blockNb := blockNumber.Uint64()
	// valid block numbers range -> [currentBlock-256, currentBlock-1]
	if blockNb >= currentBlock || blockNb < currentBlock-256 {
		s.stack.Push(uint256.Zero())
		return nil
	}
	hash, ok := s.env.BlockHashes[blockNb]
	if !ok {
		s.stack.Push(uint256.Zero())
		return nil
	}
	s.stack.Push(hash.Clone())
	return nil
}

func opExtCodeHash(s *State) error {
	addr256, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if s.env == nil {
		s.stack.Push(uint256.Zero())
		return nil
	}
	addr := addr256.Uint64()
	account, ok := s.env.Accounts[addr]
	if !ok {
		s.stack.Push(uint256.Zero())
		return nil
	}
	if len(account.Code) == 0 {
		emptyHash := keccak256.Keccak256(nil) // if the account has no code -> return keccak256("")
		emptyHash256 := uint256.Zero().SetBytes(emptyHash)
		s.stack.Push(emptyHash256)
		return nil
	}
	codeHash := keccak256.Keccak256(account.Code)
	codeHash256 := uint256.Zero().SetBytes(codeHash)
	s.stack.Push(codeHash256)
	return nil
}

func opAddress(s *State) error {
	s.stack.Push(uint256.NewUint(s.callerAddress))
	return nil
}

func opBalance(s *State) error {
	addr, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if s.tmp.accounts[addr.Uint64()] == nil {
		s.stack.Push(uint256.Zero())
		return nil
	}
	s.stack.Push(s.tmp.accounts[addr.Uint64()].Balance)

	warm := false
	for _, e := range s.tmp.warmAccountAddress {
		if e.Eq(addr) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmAccountAddress = append(s.tmp.warmAccountAddress, *addr)
		s.gasConsumed += 2500
	}

	return nil
}

func opOrigin(s *State) error {
	s.stack.Push(uint256.NewUint(s.originAddress))
	return nil
}
