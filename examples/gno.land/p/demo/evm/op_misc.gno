package evm

import (
	"errors"
	"math"
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
		s.code.Increment()
		if s.code.Ill() {
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + outOfBounds.Error())
		}
		stackData, err := s.code.Read(n)
		if err != nil {
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + err.Error())
		}
		s.stack.Push(stackData)
		return nil
	}
}

func opSwap(n int) func(*State) error {
	return func(s *State) error {
		if s.stack.Len() <= n {
			return errors.New("can't SWAP" + strconv.Itoa(n) + ": stack too small")
		}
		temp := make([]*uint256.Uint, n+1)
		for i := 0; i <= n; i++ {
			val, err := s.stack.Pop()
			if err != nil {
				for j := i - 1; j >= 0; j-- {
					s.stack.Push(temp[j])
				}
				return errors.New("can't SWAP" + strconv.Itoa(n) + err.Error())
			}
			temp[i] = val
		}
		temp[0], temp[n] = temp[n], temp[0]
		for i := n; i >= 0; i-- {
			s.stack.Push(temp[i])
		}
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
	opDest := OpCode(s.code.At(inc))
	if opDest != JUMPDEST {
		return errors.New("can't JUMP: JUMP needs to land on a JUMPDEST, but here, it's a 0x" + strconv.FormatUint(uint64(opDest), 16) + ".")
	}
	s.code.rip = inc // + 1
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
	opDest := OpCode(s.code.At(inc))
	if opDest != JUMPDEST {
		return errors.New("can't JUMP: JUMP needs to land on a JUMPDEST, but here, it's a 0x" + strconv.FormatUint(uint64(opDest), 16) + ".")
	}
	s.code.rip = inc // + 1
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

// func opCodeCopy(s *State) error {
// 	dest, err := s.stack.Pop()
// 	if err != nil {
// 		return err
// 	}
// 	src, err := s.stack.Pop()
// 	if err != nil {
// 		return err
// 	}
// 	size, err := s.stack.Pop()
// 	if err != nil {
// 		return err
// 	}

// 	memLen := uint64(len(s.memory))
// 	farest := dest.Uint64() + size.Uint64()
// 	memoryCost := uint64(0)
// 	if farest > memLen {
// 		extend := farest - memLen
// 		s.memory = append(s.memory, make([]byte, extend)...)

// 		memorySizeWord := (extend + 31) / 32
// 		memoryCost = uint64(math.Pow(float64(memorySizeWord), 2))/512 + (3 * memorySizeWord)
// 	}

// 	codeLen := uint64(len(s.code.instructions))
// 	farest = src.Uint64() + size.Uint64()
// 	if farest > codeLen {
// 		copy(s.memory[dest.Uint64():], s.code.instructions[src.Uint64():])
// 		sizeCopied := codeLen - src.Uint64()
// 		copy(s.memory[dest.Uint64()+sizeCopied:], make([]byte, size.Uint64()-sizeCopied))
// 	} else {
// 		copy(s.memory[dest.Uint64():], s.code.instructions[src.Uint64():src.Uint64()+size.Uint64()])
// 	}

// 	minimumWordSize := (size.Uint64() + 31) / 32
// 	dynamicGas := 3*minimumWordSize + memoryCost
// 	s.gasConsumed += uint(dynamicGas)
// 	return nil
// }

// EXTCODESIZE can't be implemented, it only push 0 on stack
func opExtCodeSize(s *State) error {
	_, err := s.stack.Pop()
	if err != nil {
		return err
	}
	s.stack.Push(uint256.Zero())
	return nil
}

// EXTCODESIZE can't be implemented, it only copy 0's in memory
func opExtCodeCopy(s *State) error {
	_, err := s.stack.Pop() // address of contract code to copy
	if err != nil {
		return err
	}

	dest, err := s.stack.Pop()
	if err != nil {
		return err
	}
	_, err = s.stack.Pop() // offset of contract code to copy
	if err != nil {
		return err
	}
	size, err := s.stack.Pop()
	if err != nil {
		return err
	}

	memLen := uint64(len(s.memory))
	farest := dest.Uint64() + size.Uint64()
	memoryCost := uint64(0)
	if farest > memLen {
		extend := farest - memLen
		s.memory = append(s.memory, make([]byte, extend)...)

		memorySizeWord := (extend + 31) / 32
		memoryCost = uint64(math.Pow(float64(memorySizeWord), 2))/512 + (3 * memorySizeWord)
	}

	copy(s.memory[dest.Uint64():], make([]byte, size.Uint64()))

	minimumWordSize := (size.Uint64() + 31) / 32

	gasCost := uint(3*minimumWordSize + memoryCost)
	s.gasConsumed += gasCost
	return nil
}

func opCodeCopy(s *State) error {
	if s.stack.length < 3 {
		return tooFewStack
	}
	memOff, _ := s.stack.Pop()
	codeOff, _ := s.stack.Pop()
	size, _ := s.stack.Pop()

	var memCost uint64
	codeLen := uint256.NewUint(uint64(s.code.length))
	codeEnd := uint256.Zero().Add(codeOff, size)
	memLen := uint256.NewUint(uint64(len(s.memory)))
	memEnd := uint256.Zero().Add(memOff, size)
	var data []byte
	if codeOff.Gte(codeLen) { // if too far, the data is just zeroes
		data = make([]byte, size.Uint64())
	} else {
		codeEffectiveEnd := uint256.Zero().Min(codeEnd, codeLen)
		data = s.code.instructions[codeOff.Uint64():codeEffectiveEnd.Uint64()]
		if codeEnd.Gt(codeLen) {
			data = append(data, make([]byte, uint256.Zero().Sub(codeEnd, codeLen).Uint64())...)
		}
	}
	if memEnd.Gt(memLen) { // if too big, just extend the memory of memEnd - memLen
		memExp := uint256.Zero().Sub(memEnd, memLen).Uint64()
		s.memory = append(s.memory, make([]byte, memExp)...)
		memorySizeWord := (memExp + 31) / 32
		memCost = uint64(math.Pow(float64(memorySizeWord), 2))/512 + (3 * memorySizeWord)
	}
	copy(s.memory[memOff.Uint64():memEnd.Uint64()], data)

	// burn dynamic gas
	minimumWordSize := (size.Uint64() + 31) / 32
	s.gasConsumed += uint(3*minimumWordSize + memCost)
	return nil
}

func opGas(s *State) error {
	if s.gasBet < 0 {
		s.stack.Push(uint256.Zero())
	} else {
		s.stack.Push(uint256.NewUint(uint64(s.gasBet - int(s.gasConsumed))))
	}
	return nil
}

func opTimestamp(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(time.Now().Unix())))
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
	s.code.Stop()
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
	s.tmp.logs = []LogEntry{}
	s.tmp.storage = make(map[uint256.Uint]*uint256.Uint)
	return invalidPanic
}

func opCaller(s *State) error {
	s.stack.Push(s.caller.EVMAddress.Clone())
	return nil
}

func opCallvalue(s *State) error {
	s.stack.Push(s.callValue.Clone())
	return nil
}

func opGasprice(s *State) error {
	s.stack.Push(uint256.NewUint(gasPrice))
	return nil
}
