package brainfuck

import (
	"bytes"
	"testing"
)

func TestBasicOperations(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		input    string
		expected string
	}{
		{
			name:     "Increment and Output",
			code:     "+++.",
			input:    "",
			expected: "\x03",
		},
		{
			name:     "Decrement and Output",
			code:     "+++--.",
			input:    "",
			expected: "\x01",
		},
		{
			name:     "Pointer Movement",
			code:     ">+++>++>+<<<.>.>.>.",
			input:    "",
			expected: "\x00\x03\x02\x01",
		},
		{
			name:     "Pointer at Zero Boundary",
			code:     "<<<+++.",
			input:    "",
			expected: "\x03",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := Run([]byte(tc.code), []byte(tc.input))
			if !bytes.Equal(output, []byte(tc.expected)) {
				t.Errorf("Expected %v, got %v", []byte(tc.expected), output)
			}
		})
	}
}

func TestLoops(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		input    string
		expected string
	}{
		{
			name:     "Simple Loop",
			code:     "+++[>+<-]>.",
			input:    "",
			expected: "\x03",
		},
		{
			name:     "Nested Loops",
			code:     "++[>++[>++<-]<-]>>.",
			input:    "",
			expected: "\x08",
		},
		{
			name:     "Skip Loop if Zero",
			code:     "[>+++<-]>.",
			input:    "",
			expected: "\x00",
		},
		{
			name:     "Unmatched Opening Bracket",
			code:     "[.",
			input:    "",
			expected: "",
		},
		{
			name:     "Unmatched Closing Bracket",
			code:     "].",
			input:    "",
			expected: "\x00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := Run([]byte(tc.code), []byte(tc.input))
			if !bytes.Equal(output, []byte(tc.expected)) {
				t.Errorf("Expected %v, got %v", []byte(tc.expected), output)
			}
		})
	}
}

func TestInput(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		input    string
		expected string
	}{
		{
			name:     "Read Single Input",
			code:     ",.",
			input:    "A",
			expected: "A",
		},
		{
			name:     "Read Multiple Inputs",
			code:     ",>,>,.",
			input:    "ABC",
			expected: "C",
		},
		{
			name:     "Input Modification",
			code:     ",+.",
			input:    "A",
			expected: "B",
		},
		{
			name:     "Input Exhaustion",
			code:     ",.,.,.,..",
			input:    "AB",
			expected: "AB\x00\x00\x00",
		},
		{
			name:     "Output Input Stream",
			code:     ",[.,]",
			input:    "hello, world!\n",
			expected: "hello, world!\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := Run([]byte(tc.code), []byte(tc.input))
			if !bytes.Equal(output, []byte(tc.expected)) {
				t.Errorf("Expected %v, got %v", []byte(tc.expected), output)
			}
		})
	}
}

func TestBsort(t *testing.T) {
	// [bsort.b -- bubble sort
	// (c) 2016 Daniel B. Cristofani
	// http://brainfuck.org/]
	// [This program sorts the bytes of its input by bubble sort.]
	code := `
		>>,[>>,]<<[
		[<<]>>>>[
		<<[>+<<+>-]
		>>[>+<<<<[->]>[<]>>-]
		<<<[[-]>>[>+<-]>>[<<<+>>>-]]
		>>[[<+>-]>>]<
		]<<[>>+<<-]<<
		]>>>>[.>>]`
	input := "ckduwcomaz"
	expected := "accdkmouwz"

	output := Run([]byte(code), []byte(input))
	if string(output) != expected {
		t.Errorf("Hello World test failed. Expected '%s', got '%s'", expected, string(output))
	}
}

func TestHelloWorld(t *testing.T) {
	// Classic "Hello World!" program from wikipedia
	code := `
+++++ +++++             initialize counter (cell #0) to 10
[                       use loop to set 70/100/30/10
    > +++++ ++              add  7 to cell #1
    > +++++ +++++           add 10 to cell #2
    > +++                   add  3 to cell #3
    > +                     add  1 to cell #4
<<<< -                  decrement counter (cell #0)
]
> ++ .                  print 'H'
> + .                   print 'e'
+++++ ++ .              print 'l'
.                       print 'l'
+++ .                   print 'o'
> ++ .                  print ' '
<< +++++ +++++ +++++ .  print 'W'
> .                     print 'o'
+++ .                   print 'r'
----- - .               print 'l'
----- --- .             print 'd'
> + .                   print '!'
> .                     print 'eol'`
	expected := "Hello World!\n"

	output := Run([]byte(code), []byte{})
	if string(output) != expected {
		t.Errorf("Hello World test failed. Expected '%s', got '%s'", expected, string(output))
	}
}

func TestEnsureMemory(t *testing.T) {
	tests := []struct {
		name        string
		initialMem  []byte
		pointer     int
		expectedLen int
	}{
		{
			name:        "Already Sufficient",
			initialMem:  []byte{0, 0, 0},
			pointer:     2,
			expectedLen: 3,
		},
		{
			name:        "Extend By One",
			initialMem:  []byte{0, 0, 0},
			pointer:     3,
			expectedLen: 4,
		},
		{
			name:        "Extend By Many",
			initialMem:  []byte{},
			pointer:     10,
			expectedLen: 11,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memory := tc.initialMem
			ensureMemory(&memory, tc.pointer)

			if len(memory) != tc.expectedLen {
				t.Errorf("Expected memory length %d, got %d", tc.expectedLen, len(memory))
			}

			// Check that we can access the pointer without panic
			_ = memory[tc.pointer]
		})
	}
}

// Test various complex programs to ensure the interpreter works correctly
func TestComplexPrograms(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		input    string
		expected string
	}{
		{
			name:     "Multiply Two Numbers",
			code:     ",>,<[>[->+>+<<]>>[-<<+>>]<<<-]>>.",
			input:    "\x03\x04", // 3 * 4 = 12
			expected: "\x0c",     // 12 in ASCII
		},
		{
			name:     "Cat Program",
			code:     ",[.,]",
			input:    "Hello!",
			expected: "Hello!",
		},
		{
			name:     "ASCII to Decimal",
			code:     ",>++++++[<-------->-]<.", // ASCII value minus 48
			input:    "5",                       // ASCII 53
			expected: "\x05",                    // 5 in decimal
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := Run([]byte(tc.code), []byte(tc.input))
			if !bytes.Equal(output, []byte(tc.expected)) {
				t.Errorf("Expected %v, got %v", []byte(tc.expected), output)
			}
		})
	}
}
