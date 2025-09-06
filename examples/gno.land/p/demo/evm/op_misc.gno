package evm

import (
	"errors"
	"strconv"
	"time"

	"gno.land/p/demo/evm/keccak256"
	"gno.land/p/onbloc/uint256"
)

func opPush0(s *state) error {
	s.stack.push(uint256.Zero())
	return nil
}

func opPush(n int) func(*state) error {
	return func(s *state) error {
		s.code.increment()
		if s.code.ill() {
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + outOfBounds.Error())
		}
		stackData, err := s.code.read(n)
		if err != nil {
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + err.Error())
		}
		s.stack.push(stackData.Clone())
		return nil
	}
}

func opSwap(n int) func(*state) error {
	return func(s *state) error {
		sLen := s.stack.len()
		if sLen <= n {
			return errors.New("can't SWAP" + strconv.Itoa(n) + ": stack too small")
		}
		sLen -= 1
		s.stack.data[sLen], s.stack.data[sLen-n] = s.stack.data[sLen-n], s.stack.data[sLen]
		return nil
	}
}

func opDup(n int) func(*state) error {
	return func(s *state) error {
		if s.stack.len() < n {
			return errors.New("can't DUP" + strconv.Itoa(n) + ": stack too small")
		}
		val, err := s.stack.peek(n - 1)
		if err != nil {
			return errors.New("can't DUP" + strconv.Itoa(n) + ": " + err.Error())
		}

		s.stack.push(val.Clone())
		return nil
	}
}

func opPop(s *state) error {
	_, err := s.stack.pop()
	return err
}

func opLog(nTopics int) func(*state) error {
	return func(s *state) error {
		if s.context == staticContext {
			return errors.New("can't log from staticCall")
		}

		topics := make([][]byte, nTopics)
		for i := 0; i < nTopics; i++ {
			topicVal, err := s.stack.pop()
			if err != nil {
				return err
			}
			arr := topicVal.Bytes32()
			topics[i] = make([]byte, 32)
			copy(topics[i], arr[:])
		}
		offsetVal, err := s.stack.pop()
		if err != nil {
			return err
		}
		sizeVal, err := s.stack.pop()
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
		s.tmp.logs = append(s.tmp.logs, logEntry{
			topics: topics,
			data:   data,
		})
		return nil
	}
}

func opJump(s *state) error {
	dest, err := s.stack.pop()
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

func opNop(s *state) error {
	return nil
}

func opJumpi(s *state) error {
	dest, err := s.stack.pop()
	if err != nil {
		return err
	}
	doJump, err := s.stack.pop()
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

func opIszero(s *state) error {
	num, err := s.stack.pop()
	if err != nil {
		return err
	}
	if num.IsZero() {
		s.stack.push(uint256.NewUint(1))
	} else {
		s.stack.push(uint256.Zero())
	}
	return nil
}

func opKeccak256(s *state) error {
	off, err := s.stack.pop()
	if err != nil {
		return err
	}
	size, err := s.stack.pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	if off64 > uint64(len(s.memory)) {
		s.stack.push(uint256.Zero())
		return nil
	}

	minWordSize := uint256.Zero().Div(uint256.Zero().Add(size, uint256.NewUint(31)), uint256.NewUint(32))
	s.gasConsumed += uint(6 * minWordSize.Uint64()) // TODO: runtime gas with mem extension (idk how to expand)

	slice := s.memory[off64 : off64+size.Uint64()]
	resBytes := keccak256.Keccak256(slice)
	resUint := uint256.Zero().SetBytes(resBytes)
	s.stack.push(resUint)
	return nil
}

func opPc(s *state) error {
	s.stack.push(uint256.NewUint(uint64(s.code.rip)))
	return nil
}

func opCodeSize(s *state) error {
	s.stack.push(uint256.NewUint(uint64(s.code.length)))
	return nil
}

// EXTCODESIZE can't be implemented, it only push 0 on stack
func opExtCodeSize(s *state) error {
	_, err := s.stack.pop()
	if err != nil {
		return err
	}
	s.stack.push(uint256.Zero())
	return nil
}

func opGas(s *state) error {
	if s.gasLimit < 0 {
		s.stack.push(uint256.Zero())
	} else {
		s.stack.push(uint256.NewUint(uint64(s.gasLimit - int(s.gasConsumed))))
	}
	return nil
}

func opTimestamp(s *state) error {
	s.stack.push(uint256.NewUint(uint64(time.Now().Unix())))
	return nil
}

func opPrevRandao(s *state) error {
	s.stack.push(s.env.BlockDifficulty.Clone())
	return nil
}

func opGaslimit(s *state) error {
	s.stack.push(uint256.NewUint(uint64(s.gasLimit)))
	return nil
}

func opReturn(s *state) error {
	if s.stack.len() < 2 {
		return tooFewStack
	}
	off, _ := s.stack.pop()
	size, _ := s.stack.pop()
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

func opRevert(s *state) error {
	_ = opInvalid(s)
	err := opReturn(s)
	if err != nil {
		return err
	}
	return revertError
}

func opInvalid(s *state) error {
	return invalidPanic
}

func opCaller(s *state) error {
	s.stack.push(uint256.NewUint(s.callerAddress))
	return nil
}

func opCallvalue(s *state) error {
	s.stack.push(s.callValue.Clone())
	return nil
}

func opGasprice(s *state) error {
	s.stack.push(uint256.NewUint(GasPrice))
	return nil
}

// TODO: tests
func opChainId(s *state) error {
	if s.env != nil && s.env.ChainID != nil {
		s.stack.push(s.env.ChainID.Clone())
	} else {
		s.stack.push(uint256.Zero())
	}
	return nil
}

// TODO: tests
func opSelfBalance(s *state) error {
	s.stack.push(s.tmp.accounts[s.callerAddress].Balance.Clone())
	return nil
}

// TODO: tests
func opNumber(s *state) error {
	blockNumber := uint64(time.Now().Unix())
	s.stack.push(uint256.NewUint(blockNumber))
	return nil
}

func opBaseFee(s *state) error {
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
	s.stack.push(uint256.NewUint(feeResult))
	return nil
}

func opBlobBaseFee(s *state) error {
	if s.env != nil && s.env.BlobBaseFee != nil {
		s.stack.push(s.env.BlobBaseFee.Clone())
	} else {
		s.stack.push(uint256.Zero())
	}
	return nil
}

func opBlobHash(s *state) error {
	index, err := s.stack.pop()
	if err != nil {
		return err
	}
	if s.env == nil || s.env.BlobHashes == nil {
		s.stack.push(uint256.Zero())
		return nil
	}
	i := index.Uint64()
	if i >= uint64(len(s.env.BlobHashes)) {
		s.stack.push(uint256.Zero())
		return nil
	}
	blobHash := s.env.BlobHashes[i]
	s.stack.push(blobHash.Clone())
	return nil
}

func opCoinbase(s *state) error {
	if s.env != nil && s.env.Coinbase != nil {
		s.stack.push(s.env.Coinbase.Clone()) // miner address -> coinbase
	} else {
		s.stack.push(uint256.Zero())
	}
	return nil
}

func opBlockHash(s *state) error {
	blockNumber, err := s.stack.pop()
	if err != nil {
		return err
	}
	if s.env == nil || s.env.BlockNumber == nil {
		s.stack.push(uint256.Zero())
		return nil
	}
	currentBlock := s.env.BlockNumber.Uint64()
	blockNb := blockNumber.Uint64()
	// valid block numbers range -> [currentBlock-256, currentBlock-1]
	if blockNb >= currentBlock || blockNb < currentBlock-256 {
		s.stack.push(uint256.Zero())
		return nil
	}
	hash, ok := s.env.BlockHashes[blockNb]
	if !ok {
		s.stack.push(uint256.Zero())
		return nil
	}
	s.stack.push(hash.Clone())
	return nil
}

func opExtCodeHash(s *state) error {
	addr256, err := s.stack.pop()
	if err != nil {
		return err
	}
	if s.env == nil {
		s.stack.push(uint256.Zero())
		return nil
	}
	addr := addr256.Uint64()
	account, ok := s.env.Accounts[addr]
	if !ok {
		s.stack.push(uint256.Zero())
		return nil
	}
	if len(account.Code) == 0 {
		emptyHash := keccak256.Keccak256(nil) // if the account has no code -> return keccak256("")
		emptyHash256 := uint256.Zero().SetBytes(emptyHash)
		s.stack.push(emptyHash256)
		return nil
	}
	codeHash := keccak256.Keccak256(account.Code)
	codeHash256 := uint256.Zero().SetBytes(codeHash)
	s.stack.push(codeHash256)
	return nil
}

func opAddress(s *state) error {
	s.stack.push(uint256.NewUint(s.callerAddress))
	return nil
}

func opBalance(s *state) error {
	addr, err := s.stack.pop()
	if err != nil {
		return err
	}
	if s.tmp.accounts[addr.Uint64()] == nil {
		s.stack.push(uint256.Zero())
		return nil
	}
	s.stack.push(s.tmp.accounts[addr.Uint64()].Balance)

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

func opOrigin(s *state) error {
	s.stack.push(uint256.NewUint(s.originAddress))
	return nil
}
