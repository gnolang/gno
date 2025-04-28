package evm

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"gno.land/p/demo/uint256"
)

func TestExecuteEvm(t *testing.T) {
	testCases := []struct {
		name              string
		code              []byte
		expected_result   []byte
		expected_memory   []byte
		expected_stack    []*uint256.Uint
		expected_error    error
		expected_gas_used uint
		ignore_mem        bool
	}{
		{
			name:              "simple addition",
			code:              []byte{0x60, 1, 0x60, 1, 0x01, 0x00}, // PUSH1 1, PUSH1 1, ADD
			expected_stack:    []*uint256.Uint{uint256.NewUint(2)},
			expected_gas_used: 9,
		},
		{
			name:              "error can't add on empty stack",
			code:              []byte{0x01}, // ADD
			expected_error:    errors.New("can't do operation: empty stack"),
			expected_gas_used: 3,
		},
		{
			name:              "error push1 out of bounds",
			code:              []byte{0x60}, // PUSH1
			expected_error:    errors.New("can't PUSH1: out of bounds"),
			expected_gas_used: 3,
		},
		{
			name:              "exchange 1st and 2nd stack items",
			code:              []byte{byte(PUSH0), byte(PUSH1), 0x42, byte(SWAP1), byte(STOP)}, // PUSH0, PUSH1 0x42, SWAP1
			expected_stack:    []*uint256.Uint{uint256.NewUint(66), uint256.Zero()},
			expected_gas_used: 9,
		},
		{
			name:              "push2",
			code:              []byte{byte(PUSH2), 0x42, 0x42, byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stack:    []*uint256.Uint{uint256.NewUint(16962)},
			expected_gas_used: 3,
		},
		{
			name:              "error push32 out of bounds",
			code:              []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // PUSH1
			expected_error:    errors.New("can't PUSH32: out of bounds"),
			expected_gas_used: 3,
		},
		{
			name:              "simple addition and mul",
			code:              []byte{byte(PUSH1), 10, byte(PUSH1), 11, byte(ADD), byte(PUSH1), 2, byte(MUL), byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stack:    []*uint256.Uint{uint256.NewUint(42)},
			expected_error:    nil,
			expected_gas_used: 17,
		},
		{
			name:              "big mul",
			code:              []byte{byte(PUSH31), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(PUSH1), 0xff, byte(MUL), byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stack:    []*uint256.Uint{uint256.Zero().SetBytes([]byte{0xfE, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01})},
			expected_gas_used: 11,
		},
		{
			name:              "mul random big",
			code:              []byte{byte(PUSH15), 0xdf, 0x01, 0xe8, 0x5f, 0x91, 0x2e, 0x37, 0xa3, 0x6b, 0x6c, 0x46, 0xde, 0xc5, 0x2f, 0x66, byte(PUSH17), 0x9c, 0x8b, 0xe1, 0xdc, 0x47, 0x25, 0x68, 0xe4, 0xee, 0x20, 0x3b, 0x83, 0xff, 0xff, 0xff, 0xff, 0xf0, byte(MUL), byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stack:    []*uint256.Uint{uint256.Zero().SetBytes([]byte{0x88, 0x5f, 0x04, 0x63, 0xf9, 0x35, 0xcc, 0xd3, 0x78, 0x6d, 0x30, 0x27, 0x7a, 0x17, 0x1a, 0x5f, 0x86, 0x81, 0x3e, 0x91, 0xf3, 0xe8, 0x8d, 0x96, 0x22, 0x3b, 0xd3, 0x92, 0x13, 0xad, 0x09, 0xa0})},
			expected_gas_used: 11,
		},
		{
			name:              "simple sub",
			code:              []byte{byte(PUSH1), 5, byte(PUSH1), 10, byte(SUB), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(5)},
			expected_gas_used: 9,
		},
		{
			name:              "simple sub to neg",
			code:              []byte{byte(PUSH1), 10, byte(PUSH1), 5, byte(SUB), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.Zero().Xor(uint256.Zero().SetAllOne(), uint256.NewUint(4))},
			expected_gas_used: 9,
		},
		{
			name:              "big sub",
			code:              []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(SUB), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.Zero()},
			expected_gas_used: 9,
		},
		{
			name:              "simple div",
			code:              []byte{byte(PUSH1), 2, byte(PUSH1), 10, byte(DIV), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(5)},
			expected_gas_used: 11,
		},
		{
			name:              "big div",
			code:              []byte{byte(PUSH16), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(DIV), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.MustFromHex("0x100000000000000000000000000000001")},
			expected_gas_used: 11,
		},
		{
			name:              "simple sdiv",
			code:              []byte{byte(PUSH1), 2, byte(PUSH1), 10, byte(SDIV), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(5)},
			expected_gas_used: 11,
		},
		{
			name:              "neg sdiv",
			code:              []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfb, byte(SDIV), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(2)},
			expected_gas_used: 11,
		},
		{
			name:              "simple mod",
			code:              []byte{byte(PUSH1), 3, byte(PUSH1), 10, byte(MOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(1)},
			expected_gas_used: 11,
		},
		{
			name:              "simple mod 2",
			code:              []byte{byte(PUSH1), 5, byte(PUSH1), 17, byte(MOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(2)},
			expected_gas_used: 11,
		},
		{
			name:              "big mod",
			code:              []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(MOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(1)},
			expected_gas_used: 11,
		},
		{
			name:              "simple smod",
			code:              []byte{byte(PUSH1), 3, byte(PUSH1), 10, byte(SMOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(1)},
			expected_gas_used: 11,
		},
		{
			name:              "pos smod neg",
			code:              []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfd, byte(PUSH1), 13, byte(SMOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(1)},
			expected_gas_used: 11,
		},
		{
			name:              "neg smod pos",
			code:              []byte{byte(PUSH1), 3, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xf3, byte(SMOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")},
			expected_gas_used: 11,
		},
		{
			name:              "neg smod neg",
			code:              []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfd, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xf3, byte(SMOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")},
			expected_gas_used: 11,
		},
		{
			name:              "big smod",
			code:              []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(SMOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")},
			expected_gas_used: 11,
		},
		{
			name:              "simple addmod",
			code:              []byte{byte(PUSH1), 8, byte(PUSH1), 10, byte(PUSH1), 10, byte(ADDMOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(4)},
			expected_gas_used: 17,
		},
		{
			name:              "big addmod",
			code:              []byte{byte(PUSH1), 2, byte(PUSH1), 2, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(ADDMOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.One()},
			expected_gas_used: 17,
		},
		{
			name:              "simple mulmod",
			code:              []byte{byte(PUSH1), 8, byte(PUSH1), 10, byte(PUSH1), 10, byte(MULMOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(4)},
			expected_gas_used: 17,
		},
		{
			name:              "big mulmod",
			code:              []byte{byte(PUSH1), 12, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(MULMOD), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(9)},
			expected_gas_used: 17,
		},
		{
			name:              "simple exp",
			code:              []byte{byte(PUSH1), 8, byte(PUSH1), 10, byte(EXP), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(0x5f5e100)},
			expected_gas_used: 66,
		},
		{
			name:              "big exp",
			code:              []byte{byte(PUSH1), 0xff, byte(PUSH1), 2, byte(EXP), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.MustFromHex("0x8000000000000000000000000000000000000000000000000000000000000000")},
			expected_gas_used: 66,
		},
		{
			name:              "big exp overflow",
			code:              []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(PUSH1), 12, byte(EXP), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.Zero()},
			expected_gas_used: 1616,
		},
		{
			name:              "simple signextend",
			code:              []byte{byte(PUSH1), 8, byte(PUSH1), 10, byte(SIGNEXTEND), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(8)},
			expected_gas_used: 11,
		},
		{
			name:              "big signextend",
			code:              []byte{byte(PUSH1), 0xff, byte(PUSH1), 0, byte(SIGNEXTEND), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")},
			expected_gas_used: 11,
		},
		{
			name:              "jump jump !",
			code:              []byte{byte(PUSH1), 1, byte(PUSH1), 8, byte(JUMP), byte(PUSH1), 1, byte(ADD), byte(JUMPDEST), byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(1)},
			expected_gas_used: 14,
		},
		{
			name:              "jumpi yes",
			code:              []byte{byte(PUSH1), 1, byte(PUSH1), 10, byte(JUMPI), byte(PUSH1), 1, byte(PUSH1), 1, byte(ADD), byte(JUMPDEST), byte(STOP)},
			expected_stack:    []*uint256.Uint{},
			expected_gas_used: 16,
		},
		{
			name: "jumpi no",
			code: []byte{
				byte(PUSH1), 0,
				byte(PUSH1), 10,
				byte(JUMPI),
				byte(PUSH1), 1,
				byte(PUSH1), 1,
				byte(ADD),
				byte(JUMPDEST),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{uint256.NewUint(2)},
			expected_gas_used: 26,
		},
		{
			name: "check zero and jump",
			code: []byte{
				byte(PUSH1), 4,
				byte(PUSH1), 2,
				byte(PUSH1), 2,
				byte(ADD),
				byte(SUB),
				byte(ISZERO),
				byte(PUSH1), 14,
				byte(JUMPI),
				byte(PUSH1), 42,
				byte(JUMPDEST),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{},
			expected_gas_used: 31,
		},
		{
			name: "check zero and dont jump",
			code: []byte{
				byte(PUSH1), 4,
				byte(PUSH1), 2,
				byte(PUSH1), 3,
				byte(ADD),
				byte(SUB),
				byte(ISZERO),
				byte(PUSH1), 14,
				byte(JUMPI),
				byte(PUSH1), 42,
				byte(JUMPDEST),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{uint256.NewUint(42)},
			expected_gas_used: 35,
		},
		{
			name: "exchange 1st and 5th stack items",
			code: []byte{
				byte(PUSH1), 1,
				byte(PUSH1), 2,
				byte(PUSH1), 3,
				byte(PUSH1), 4,
				byte(PUSH1), 5,
				byte(PUSH1), 6,
				byte(SWAP5),
				byte(STOP),
			},
			expected_stack: []*uint256.Uint{
				uint256.NewUint(6),
				uint256.NewUint(2),
				uint256.NewUint(3),
				uint256.NewUint(4),
				uint256.NewUint(5),
				uint256.NewUint(1),
			},
			expected_gas_used: 21,
		},
		{
			name: "read raw mem at offset 12",
			code: []byte{
				byte(PUSH1), 12,
				byte(MLOAD),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{uint256.Zero()},
			ignore_mem:        true,
			expected_gas_used: 6,
		},
		{
			name: "read raw mem at offset 0",
			code: []byte{
				byte(PUSH1), 12,
				byte(MLOAD),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{uint256.Zero()},
			ignore_mem:        true,
			expected_gas_used: 6,
		},
		{
			name: "write mem",
			code: []byte{
				byte(PUSH1), 42, // data
				byte(PUSH1), 0, // offset
				byte(MSTORE),
				byte(PUSH1), 0,
				byte(MLOAD),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{uint256.NewUint(42)},
			ignore_mem:        true,
			expected_gas_used: 15,
		},
		{
			name: "write mem offsetted",
			code: []byte{
				byte(PUSH1), 42, // data
				byte(PUSH1), 12, // offset
				byte(MSTORE),
				byte(PUSH1), 12,
				byte(MLOAD),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{uint256.NewUint(42)},
			ignore_mem:        true,
			expected_gas_used: 15,
		},
		{
			name: "mem example https://www.evm.codes/",
			code: []byte{
				byte(PUSH1), 0,
				byte(MLOAD),

				byte(PUSH1), 0xff, // data
				byte(PUSH1), 0, // offset
				byte(MSTORE),
				byte(PUSH1), 0,
				byte(MLOAD),

				byte(PUSH1), 0xff, // data
				byte(PUSH1), 1, // offset
				byte(MSTORE),
				byte(PUSH1), 0,
				byte(MLOAD),

				byte(STOP),
			},
			expected_stack: []*uint256.Uint{
				uint256.Zero(),
				uint256.NewUint(0xff),
				uint256.Zero(),
			},
			ignore_mem:        true,
			expected_gas_used: 36,
		},
		{
			name: "mem save byte",
			code: []byte{
				byte(PUSH4), 0xaa, 0xbb, 0xcc, 0xdd, // data
				byte(PUSH1), 0, // offset
				byte(MSTORE8),
				byte(STOP),
			},
			expected_memory:   []byte{0xdd},
			expected_gas_used: 9,
		},
		{
			name: "simple dup5",
			code: []byte{
				byte(PUSH1), 1,
				byte(PUSH1), 2,
				byte(PUSH1), 3,
				byte(PUSH1), 4,
				byte(PUSH1), 5,
				byte(PUSH1), 6,
				byte(PUSH1), 7,
				byte(DUP5),
				byte(STOP),
			},
			expected_stack: []*uint256.Uint{
				uint256.NewUint(1),
				uint256.NewUint(2),
				uint256.NewUint(3),
				uint256.NewUint(4),
				uint256.NewUint(5),
				uint256.NewUint(6),
				uint256.NewUint(7),
				uint256.NewUint(3),
			},
			expected_gas_used: 24,
		},
		{
			name: "bitwise AND operation",
			code: []byte{
				byte(PUSH2), 0x00, 0xFF,
				byte(NOT),
				byte(PUSH2), 0x0F, 0x0F,
				byte(AND),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{uint256.NewUint(3840)}, // 0x0F00
			expected_gas_used: 12,
		},
		{
			name: "bitwise OR operation",
			code: []byte{
				byte(PUSH2), 0x00, 0xFF,
				byte(NOT),
				byte(PUSH2), 0x0F, 0x0F,
				byte(OR),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{uint256.Zero().SetBytes([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0f})},
			expected_gas_used: 12,
		},
		{
			name: "bitwise XOR operation",
			code: []byte{
				byte(PUSH2), 0x00, 0xFF,
				byte(NOT),
				byte(PUSH1), 0x0F,
				byte(XOR),
				byte(STOP),
			},
			expected_stack:    []*uint256.Uint{uint256.Zero().SetBytes([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0f})},
			expected_gas_used: 12,
		},
		{
			name: "just do some keccak256 with 0xffffffff",
			code: []byte{
				byte(PUSH1), 0xff, // store 0xffffffff
				byte(PUSH0),
				byte(MSTORE8),
				byte(PUSH1), 0xff,
				byte(PUSH1), 1,
				byte(MSTORE8),
				byte(PUSH1), 0xff,
				byte(PUSH1), 2,
				byte(MSTORE8),
				byte(PUSH1), 0xff,
				byte(PUSH1), 3,
				byte(MSTORE8),

				byte(PUSH1), 4, // size
				byte(PUSH0), // offset
				byte(KECCAK256),

				byte(STOP),
			},
			expected_stack: []*uint256.Uint{
				uint256.Zero().SetBytes([]byte{0x29, 0x04, 0x5A, 0x59, 0x20, 0x07, 0xD0, 0xC2, 0x46, 0xEF, 0x02, 0xC2, 0x22, 0x35, 0x70, 0xDA, 0x95, 0x22, 0xD0, 0xCF, 0x0F, 0x73, 0x28, 0x2C, 0x79, 0xA1, 0xBC, 0x8F, 0x0B, 0xB2, 0xC2, 0x38}),
			},
			expected_memory:   []byte{0xff, 0xff, 0xff, 0xff},
			expected_gas_used: 72,
		},
		{
			name: "checkin da mem size",
			code: []byte{
				byte(PUSH1), 0xff, // store 0xffffff
				byte(PUSH0),
				byte(MSTORE8),
				byte(PUSH1), 0xff,
				byte(PUSH1), 1,
				byte(MSTORE8),
				byte(PUSH1), 0xff,
				byte(PUSH1), 2,
				byte(MSTORE8),

				byte(MSIZE),

				byte(STOP),
			},
			expected_stack: []*uint256.Uint{
				uint256.NewUint(3),
			},
			expected_memory:   []byte{0xff, 0xff, 0xff},
			expected_gas_used: 29,
		},
		{
			name: "checkin da pc",
			code: []byte{
				byte(PUSH1), 0xff, // store 0xffffff
				byte(PUSH0),
				byte(MSTORE8),
				byte(PUSH1), 0xff,
				byte(PUSH1), 1,
				byte(MSTORE8),
				byte(PUSH1), 0xff,
				byte(PUSH1), 2,
				byte(MSTORE8),

				byte(PC),

				byte(STOP),
			},
			expected_stack: []*uint256.Uint{
				uint256.NewUint(14),
			},
			ignore_mem:        true,
			expected_gas_used: 29,
		},
		{
			name:              "simple push 2 byte",
			code:              []byte{byte(PUSH2), 0xff, 0x00, byte(STOP)},
			expected_stack:    []*uint256.Uint{uint256.NewUint(0xff00)},
			expected_gas_used: 3,
		},
		{
			name: "byte instruction from https://www.evm.codes/",
			code: []byte{
				byte(PUSH1), 0xff,
				byte(PUSH1), 31,
				byte(BYTE),

				byte(PUSH2), 0xff, 0x00,
				byte(PUSH1), 30,
				byte(BYTE),

				byte(STOP),
			},
			expected_stack: []*uint256.Uint{
				uint256.NewUint(0xff),
				uint256.NewUint(0xff),
			},
			expected_gas_used: 18,
		},
		{
			name: "set 0 in memory and inshala memory is 32 bytes",
			code: []byte{
				byte(PUSH0),
				byte(PUSH0),
				byte(MSTORE),
				byte(STOP),
			},
			expected_memory:   make([]byte, 32),
			expected_gas_used: 9,
		},
		{
			name: "check for 2 32 byte long stuff",
			code: []byte{
				byte(PUSH0),
				byte(PUSH0),
				byte(MSTORE),

				byte(PUSH32), 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
				byte(PUSH1), 32,
				byte(MSTORE),
				byte(STOP),
			},
			expected_memory:   append(make([]byte, 32), []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f}...),
			expected_gas_used: 18,
		},
		{
			name: "memcopy ahh https://www.evm.codes/ example",
			code: []byte{
				byte(PUSH0),
				byte(PUSH0),
				byte(MSTORE),

				byte(PUSH32), 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
				byte(PUSH1), 32,
				byte(MSTORE),

				byte(PUSH1), 32,
				byte(PUSH1), 32,
				byte(PUSH0),
				byte(MCOPY),
				byte(STOP),
			},
			expected_memory:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f},
			expected_gas_used: 30,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errStack, errResult, errMem := false, false, false

			st := newState()
			err := st.executeByteCode(tc.code)
			stack := st.getStackData()
			result := st.getResult()
			mem := st.memory

			if (err != tc.expected_error && (err == nil || tc.expected_error == nil)) || (err != nil && tc.expected_error != nil && err.Error() != tc.expected_error.Error()) {
				if err == nil {
					t.Errorf("I expected this error:\n -> '%s'\nbut I got no", tc.expected_error)
				} else if tc.expected_error == nil {
					t.Errorf("I expected no error but I got that one:\n -> '%s'", err)
				} else {
					t.Errorf("I expected this error:\n -> '%s'\nbut I got that one:\n -> '%s'", tc.expected_error, err)
				}
			}

			if st.gasConsumed != tc.expected_gas_used {
				t.Errorf("I expected this gas usage:\n -> %d\nbut I got:\n -> %d", tc.expected_gas_used, st.gasConsumed)
			}

			if len(stack) != len(tc.expected_stack) {
				errStack = true
			} else {
				for i := 0; i < len(stack); i++ {
					if !stack[i].Eq(tc.expected_stack[i]) {
						errStack = true
						break
					}
				}
			}

			if len(result) != len(tc.expected_result) {
				errResult = true
			} else {
				for i := 0; i < len(result); i++ {
					if result[i] != tc.expected_result[i] {
						errResult = true
						break
					}
				}
			}

			if len(mem) != len(tc.expected_memory) {
				errMem = true
			} else {
				for i := 0; i < len(result); i++ {
					if result[i] != tc.expected_memory[i] {
						errMem = true
						break
					}
				}
			}
			if tc.ignore_mem {
				fmt.Println("--- WARN: Memory is ignored for this test")
				errMem = false
			}

			if errStack {
				t.Errorf("I expected this stack at exit:")
				for i := 0; i < len(stack) || i < len(tc.expected_stack); i++ {
					expected := "empty"
					if i < len(tc.expected_stack) {
						expected = "'" + tc.expected_stack[i].Hex() + "'"
					}
					t.Errorf("-> %s", expected)
				}
				t.Error("but I got that one:")
				for i := 0; i < len(stack) || i < len(tc.expected_stack); i++ {
					actual := "empty"
					if i < len(stack) {
						actual = "'" + stack[i].Hex() + "'"
					}
					t.Errorf("-> %s", actual)
				}
			}
			if errResult {
				t.Errorf("I expected this result:")
				for i := 0; i < len(result) || i < len(tc.expected_result); i++ {
					expected := "empty"
					if i < len(tc.expected_result) {
						expected = "'" + strconv.FormatUint(uint64(tc.expected_result[i]), 16) + "'"
					}
					t.Errorf("-> %s", expected)
				}
				t.Error("but I got that one:")
				for i := 0; i < len(result) || i < len(tc.expected_result); i++ {
					actual := "empty"
					if i < len(result) {
						actual = "'" + strconv.FormatUint(uint64(result[i]), 16) + "'"
					}
					t.Errorf("-> %s", actual)
				}
			}
			if errMem {
				t.Errorf("I expected this memory:")
				for i := 0; i < len(mem) || i < len(tc.expected_memory); i++ {
					expected := "empty"
					if i < len(tc.expected_memory) {
						expected = "'" + strconv.FormatUint(uint64(tc.expected_memory[i]), 16) + "'"
					}
					t.Errorf("-> %s", expected)
				}
				t.Error("but I got that one:")
				for i := 0; i < len(mem) || i < len(tc.expected_memory); i++ {
					actual := "<empty>"
					if i < len(mem) {
						actual = "'" + strconv.FormatUint(uint64(mem[i]), 16) + "'"
					}
					t.Errorf("-> %s", actual)
				}
			}
		})
	}
}
