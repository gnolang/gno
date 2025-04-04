package brainfuck

// Run takes code and input as byte slices and returns the output as a byte slice
func Run(code []byte, input []byte) []byte {
	memory  := []byte{}
	output  := []byte{}
	pointer := 0
	cursor  := 0
	i       := 0 // input index
	
	for cursor < len(code) {
		switch code[cursor] {
		case '>': // Move pointer right
			pointer++
			ensureMemory(&memory, pointer)
		case '<': // Move pointer left
			if pointer > 0 {
				pointer--
			} else { // we will not support wrapping, 
				pointer = 0
			}
		case '+': // Increment current cell
			ensureMemory(&memory, pointer)
			memory[pointer]++
		case '-': // Decrement current cell
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
				memory[pointer] = 0
			}
		case '[': // Jump forward to matching ] if current cell is 0
			ensureMemory(&memory, pointer)
			if memory[pointer] == 0 {
				brackets := 1
				for brackets > 0 {
					cursor++
					if cursor >= len(code) {
						return output
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
				brackets := 1
				for brackets > 0 {
					cursor--
					if cursor < 0 {
						return output
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
	}
	return output
}

func ensureMemory(memory *[]byte, pointer int) {
	for pointer >= len(*memory) {
		*memory = append(*memory, 0)
	}
}