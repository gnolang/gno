// Package brainfuck implements a minimalist Brainfuck language interpreter.
package brainfuck

// Run executes a byte slice encoded Brainfuck program with the provided input and
// returns the output. If no input is desired, use an empty string "". Memory is
// dynamically allocated and grows rightward as needed. The pointer cannot go below
// zero (does not wrap from the beginning to end). When input is exhausted, the ','
// command writes 0 to the current cell. Program exits early if unmatched brackets
// are encountered.
func Run(code, input []byte) []byte {
	memory := []byte{} // dynamic memory tape
	output := []byte{} // collected output bytes
	pointer := 0       // memory cell pointer
	cursor := 0        // code instruction pointer
	i := 0             // input index

	for cursor < len(code) {
		switch code[cursor] {
		case '>': // Move pointer right
			pointer++
			if pointer <= 0 { // Guard against potential overflow
				pointer = 0
			}
			ensureMemory(&memory, pointer)
		case '<': // Move pointer left
			pointer--
			if pointer <= 0 { // Guard against below zero
				pointer = 0
			}
		case '+': // Increment current cell (wraps: 255+1=0)
			ensureMemory(&memory, pointer)
			memory[pointer]++
		case '-': // Decrement current cell (wraps: 0-1=255)
			ensureMemory(&memory, pointer)
			memory[pointer]--
		case '.': // Output current cell
			ensureMemory(&memory, pointer)
			output = append(output, memory[pointer])
		case ',': // Read one byte from input
			ensureMemory(&memory, pointer)
			if i < len(input) {
				memory[pointer] = input[i]
				i++
			} else {
				memory[pointer] = 0 // Write 0 when input is exhausted
			}
		case '[': // Jump forward to matching ] if current cell is 0
			ensureMemory(&memory, pointer)
			if memory[pointer] == 0 {
				brackets := 1 // Track nesting level
				for brackets > 0 {
					cursor++
					if cursor >= len(code) {
						return output // Exit if unmatched bracket
					}
					if code[cursor] == '[' {
						brackets++
					} else if code[cursor] == ']' {
						brackets--
					}
				}
			}
		case ']': // Jump back to matching [ if current cell is not 0
			ensureMemory(&memory, pointer)
			if memory[pointer] != 0 {
				brackets := 1 // Track nesting level
				for brackets > 0 {
					cursor--
					if cursor < 0 {
						return output // Exit if unmatched bracket
					}
					if code[cursor] == ']' {
						brackets++
					} else if code[cursor] == '[' {
						brackets--
					}
				}
			}
		}
		cursor++
		if cursor < 0 { // Guard cursor against potential overflow
			cursor = 0
		}
		if i < 0 { // Guard input index against potential overflow
			i = 0
		}
	}
	return output
}

// ensureMemory grows the memory slice if needed to accommodate the specified
// pointer position. New cells are initialized to zero.
func ensureMemory(memory *[]byte, pointer int) {
	for pointer >= len(*memory) {
		*memory = append(*memory, 0)
	}
}
