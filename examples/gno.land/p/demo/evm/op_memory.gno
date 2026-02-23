package evm

import (
	"errors"
	"math"

	"gno.land/p/onbloc/uint256"
)

func getByteCopyFromMemory(memory []byte, offset *uint256.Uint, size *uint256.Uint) []byte {
	callData := make([]byte, 0)
	if int(offset.Uint64()) >= len(memory) {
		callData = make([]byte, int(size.Uint64()))
	} else if int(offset.Uint64()+size.Uint64()) > len(memory) {
		callData = make([]byte, int(size.Uint64()))
		copy(callData, memory[offset.Uint64():])
	} else {
		callData = memory[offset.Uint64() : offset.Uint64()+size.Uint64()]
	}
	return callData
}

// Retrieve data stored in memory, at offset.
// > IN: `[offset]`
// > OUT: `[result]`
func opMload(s *state) error {
	off, err := s.stack.pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	if off64+32 > uint64(len(s.memory)) {
		s.stack.push(uint256.Zero())
		return nil
	}

	// TODO: add read-memory gas consuption
	read := s.memory[off64 : off64+32]
	if len(read) < 32 {
		read = append(read, make([]byte, 32-len(read))...)
	}
	s.stack.push(uint256.Zero().SetBytes(read))
	return nil
}

// Write data in memory at offset.
// > IN: `[offset, data]`
func opMstore(s *state) error {
	off, err := s.stack.pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	val, err := s.stack.pop()
	if err != nil {
		return err
	}
	bytes := val.Bytes32()
	memLen := uint64(len(s.memory))

	// if memory is empty, add the bytes, expension is 32 bytes
	if memLen == 0 {
		s.memory = append(make([]byte, off64), bytes[:]...)
		s.gasConsumed += 32
		return nil
	}
	// if offset + 32 is greater than memory len, expend/consume the diff
	if off64+32 >= memLen {
		add := off64 + 32 - memLen
		s.memory = append(s.memory, make([]byte, add)...)
		s.gasConsumed += uint(add)
	}

	// copy in the right place
	copy(s.memory[off64:off64+32], bytes[:])
	return nil
}

// Write first byte of data in memory at offset.
// > IN: `[offset, data]`
func opMstore8(s *state) error {
	off, err := s.stack.pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	val, err := s.stack.pop()
	if err != nil {
		return err
	}

	// get the byte
	b := byte(val.Uint64() & 0xff)
	memLen := uint64(len(s.memory))

	// if no mem and offset is greater than zero, just do some zeroes + add the byte
	if memLen == 0 && off64 != 0 {
		s.memory = append(make([]byte, off64), b)
		s.gasConsumed += uint(off64) + 1
		return nil
	}
	if off64+1 > memLen {
		add := off64 + 1 - memLen
		s.memory = append(s.memory, make([]byte, add)...)
		s.gasConsumed += uint(add)
	}
	s.memory[off64] = b
	return nil
}

// Get the memory size.
// > OUT: `[size]`
func opMsize(s *state) error {
	s.stack.push(uint256.NewUint(uint64(len(s.memory))))
	return nil
}

// opMcopy copies a segment of memory from a source offset to a destination offset for a specified size.
// It expands the memory if needed and accounts for gas consumption.
func opMcopy(s *state) error {
	dst, err := s.stack.pop()
	if err != nil {
		return err
	}
	src, err := s.stack.pop()
	if err != nil {
		return err
	}
	size, err := s.stack.pop()
	if err != nil {
		return err
	}

	memlen256 := uint256.NewUint(uint64(len(s.memory)))
	farest := uint256.Zero().Max(uint256.Zero().Add(src, size), uint256.Zero().Add(dst, size))
	if farest.Gt(memlen256) {
		extend := uint256.Zero().Sub(farest, memlen256).Uint64()
		s.memory = append(s.memory, make([]byte, extend)...)
		s.gasConsumed += uint(extend)
	}
	copy(s.memory[dst.Uint64():], s.memory[src.Uint64():src.Uint64()+size.Uint64()])
	return nil
}

// opTload loads a value from the transient storage at the given key and pushes it onto the stack.
// If the key does not exist, it pushes zero.
func opTload(s *state) error {
	if s.context == staticContext {
		return errors.New("can't modify transient memory from staticCall")
	}

	key, err := s.stack.pop()
	if err != nil {
		return err
	}

	read, ok := s.transientStorage[*key]
	if !ok {
		s.stack.push(uint256.Zero())
	}
	s.stack.push(read.Clone())
	return nil
}

// opTstore stores a value into the transient storage at the given key.
func opTstore(s *state) error {
	key, err := s.stack.pop()
	if err != nil {
		return err
	}
	value, err := s.stack.pop()
	if err != nil {
		return err
	}
	s.transientStorage[*key] = value
	return nil
}

// opSload loads a value from persistent storage at the given key and pushes it onto the stack.
// It tracks whether the storage slot is warm and charges gas accordingly.
// If the key does not exist, it pushes zero.
func opSload(s *state) error {
	key, err := s.stack.pop()
	if err != nil {
		return err
	}

	warm := false
	for _, e := range s.tmp.warmStorage {
		if e.Eq(key) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmStorage = append(s.tmp.warmStorage, *key)
		s.gasConsumed += 2000
	}

	read, ok := s.tmp.accounts[s.callerAddress].Storage[key.Uint64()]
	if !ok {
		s.stack.push(uint256.Zero())
		return nil
	}
	s.stack.push(read.Clone())
	return nil
}

// opSstore stores a value into persistent storage at the given key.
// It tracks whether the storage slot is warm and charges gas accordingly.
func opSstore(s *state) error {
	if s.context == staticContext {
		return errors.New("can't modify memory from staticCall")
	}
	key, err := s.stack.pop()
	if err != nil {
		return err
	}
	value, err := s.stack.pop()
	if err != nil {
		return err
	}

	warm := false
	for _, e := range s.tmp.warmStorage {
		if e.Eq(key) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmStorage = append(s.tmp.warmStorage, *key)
		s.gasConsumed += 2000
	}
	s.tmp.accounts[s.callerAddress].Storage[key.Uint64()] = value
	return nil
}

// opCalldataload loads 32 bytes from the call data at the given offset and pushes it onto the stack.
// If the requested range exceeds the call data size, it pads with zeros.
func opCalldataload(s *state) error {
	off, err := s.stack.pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	read := []byte{}
	if off64+32 > uint64(len(s.callData)) {
		read = s.callData[off64:]
	} else {
		read = s.callData[off64 : off64+32]
	}
	if len(read) < 32 {
		read = append(read, make([]byte, 32-len(read))...)
	}
	s.stack.push(uint256.Zero().SetBytes(read))
	return nil
}

// opCalldatasize pushes the size of the call data onto the stack.
func opCalldatasize(s *state) error {
	s.stack.push(uint256.NewUint(uint64(len(s.callData))))
	return nil
}

// opCalldatacopy copies a portion of call data to memory starting at the given destination offset.
// It expands memory as needed and charges gas accordingly.
func opCalldatacopy(s *state) error {
	dest, err := s.stack.pop()
	if err != nil {
		return err
	}
	src, err := s.stack.pop()
	if err != nil {
		return err
	}
	size, err := s.stack.pop()
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

	callDataLen := uint64(len(s.callData))
	farest = src.Uint64() + size.Uint64()
	if farest > callDataLen {
		copy(s.memory[dest.Uint64():], s.callData[src.Uint64():])
		sizeCopied := callDataLen - src.Uint64()
		copy(s.memory[dest.Uint64()+sizeCopied:], make([]byte, size.Uint64()-sizeCopied))
	} else {
		copy(s.memory[dest.Uint64():], s.callData[src.Uint64():src.Uint64()+size.Uint64()])
	}

	minimumWordSize := (size.Uint64() + 31) / 32
	dynamicGas := 3*minimumWordSize + memoryCost
	s.gasConsumed += uint(dynamicGas)
	return nil
}

// opExtCodeCopy copies contract code from an external address into memory.
// This implementation only copies zeros (no actual code) and accounts for gas usage.
// Memory is expanded as needed.
// btw: EXTCODESIZE can't be implemented, it only copy 0's in memory
func opExtCodeCopy(s *state) error {
	_, err := s.stack.pop() // address of contract code to copy
	if err != nil {
		return err
	}

	dest, err := s.stack.pop()
	if err != nil {
		return err
	}
	_, err = s.stack.pop() // offset of contract code to copy
	if err != nil {
		return err
	}
	size, err := s.stack.pop()
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

// opCodeCopy copies bytecode from the current contract code into memory at a specified offset and size.
// It expands memory as necessary and burns gas based on memory expansion and size copied.
func opCodeCopy(s *state) error {
	if s.stack.length < 3 {
		return tooFewStack
	}
	memOff, _ := s.stack.pop()
	codeOff, _ := s.stack.pop()
	size, _ := s.stack.pop()

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
