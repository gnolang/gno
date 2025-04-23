package evm

import (
	"errors"
	"strconv"
	"testing"

	"gno.land/p/demo/evm/uint256"
)

func TestExecuteEvm(t *testing.T) {
	testCases := []struct {
		name            string
		code            []byte
		expected_result []byte
		expected_stack  []*uint256.Uint
		expected_error  error
	}{
		{
			name:           "simple addition",
			code:           []byte{0x60, 1, 0x60, 1, 0x01, 0x00}, // PUSH1 1, PUSH1 1, ADD
			expected_stack: []*uint256.Uint{uint256.NewFrom64s(2)},
		},
		{
			name:           "error can't add on empty stack",
			code:           []byte{0x01}, // ADD
			expected_error: errors.New("can't ADD: empty stack"),
		},
		{
			name:           "error push1 out of bounds",
			code:           []byte{0x60}, // PUSH1
			expected_error: errors.New("can't PUSH1: out of bounds"),
		},
		{
			name:           "exchange 1st and 2nd stack items",
			code:           []byte{byte(PUSH0), byte(PUSH1), 0x42, byte(SWAP1), byte(STOP)}, // PUSH0, PUSH1 0x42, SWAP1
			expected_stack: []*uint256.Uint{uint256.NewFrom64s(66), uint256.New()},
		},
		{
			name:           "push2",
			code:           []byte{byte(PUSH2), 0x42, 0x42, byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stack: []*uint256.Uint{uint256.NewFrom64s(16962)},
		},
		{
			name:           "error push32 out of bounds",
			code:           []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // PUSH1
			expected_error: errors.New("can't PUSH32: out of bounds"),
		},
		{
			name:           "simple addition and mul",
			code:           []byte{byte(PUSH1), 10, byte(PUSH1), 11, byte(ADD), byte(PUSH1), 2, byte(MUL), byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stack: []*uint256.Uint{uint256.NewFrom64s(42)},
			expected_error: nil,
		},
		{
			name:           "big mul",
			code:           []byte{byte(PUSH31), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(PUSH1), 0xff, byte(MUL), byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stack: []*uint256.Uint{uint256.NewFromBytes([]byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01})},
		},
		{
			name:           "mul random big",
			code:           []byte{byte(PUSH15), 0xdf, 0x01, 0xe8, 0x5f, 0x91, 0x2e, 0x37, 0xa3, 0x6b, 0x6c, 0x46, 0xde, 0xc5, 0x2f, 0x66, byte(PUSH17), 0x9c, 0x8b, 0xe1, 0xdc, 0x47, 0x25, 0x68, 0xe4, 0xee, 0x20, 0x3b, 0x83, 0xff, 0xff, 0xff, 0xff, 0xf0, byte(MUL), byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stack: []*uint256.Uint{uint256.NewFromBytes([]byte{0x88, 0x5f, 0x04, 0x63, 0xf9, 0x35, 0xcc, 0xd3, 0x78, 0x6d, 0x30, 0x27, 0x7a, 0x17, 0x1a, 0x5f, 0x86, 0x81, 0x3e, 0x91, 0xf3, 0xe8, 0x8d, 0x96, 0x22, 0x3b, 0xd3, 0x92, 0x13, 0xad, 0x09, 0xa0})},
		},
		{
			name:           "simple sub",
			code:           []byte{byte(PUSH1), 5, byte(PUSH1), 10, byte(SUB), byte(STOP)},
			expected_stack: []*uint256.Uint{uint256.NewFrom64s(5)},
		},
		{
			name:           "big sub",
			code:           []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(SUB), byte(STOP)},
			expected_stack: []*uint256.Uint{uint256.New()},
		},
		{
			name:           "simple div",
			code:           []byte{byte(PUSH1), 2, byte(PUSH1), 10, byte(DIV), byte(STOP)},
			expected_stack: []*uint256.Uint{uint256.NewFrom64s(5)},
		},
		{
			name:           "big div",
			code:           []byte{byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(PUSH32), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(SUB), byte(STOP)},
			expected_stack: []*uint256.Uint{uint256.New()},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errStack, errResult := false, false

			st := newState()
			err := st.executeByteCode(tc.code)
			stack := st.getStackData()
			result := st.getResult()

			if (err != tc.expected_error && (err == nil || tc.expected_error == nil)) || (err != nil && tc.expected_error != nil && err.Error() != tc.expected_error.Error()) {
				if err == nil {
					t.Errorf("I expected this error:\n -> '%s'\nbut I got no", tc.expected_error)
				} else if tc.expected_error == nil {
					t.Errorf("I expected no error but I got that one:\n -> '%s'", err)
				} else {
					t.Errorf("I expected this error:\n -> '%s'\nbut I got that one:\n -> '%s'", tc.expected_error, err)
				}
			}

			if len(stack) != len(tc.expected_stack) {
				errStack = true
			} else {
				for i := 0; i < len(stack); i++ {
					if !stack[i].Equals(tc.expected_stack[i]) {
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

			if errStack {
				t.Errorf("I expected this stack at exit:")
				for i := 0; i < len(stack) || i < len(tc.expected_stack); i++ {
					expected := "empty"
					if i < len(tc.expected_stack) {
						expected = "'" + tc.expected_stack[i].HexString() + "'"
					}
					t.Errorf("-> %s", expected)
				}
				t.Error("but I got that one:")
				for i := 0; i < len(stack) || i < len(tc.expected_stack); i++ {
					actual := "empty"
					if i < len(stack) {
						actual = "'" + stack[i].HexString() + "'"
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
		})
	}
}
