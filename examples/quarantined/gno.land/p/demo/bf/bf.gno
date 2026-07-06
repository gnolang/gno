package bf

import (
	"strings"
)

const maxlen = 30000

func Execute(code string) string {
	var (
		memory  = make([]byte, maxlen) // memory tape
		pointer = 0                    // initial memory pointer
		buf     strings.Builder
	)

	// Loop through each character in the code
	for i := 0; i < len(code); i++ {
		switch code[i] {
		case '>':
			// Increment memory pointer
			pointer++
			if pointer >= maxlen {
				pointer = 0
			}
		case '<':
			// Decrement memory pointer
			pointer--
			if pointer < 0 {
				pointer = maxlen - 1
			}
		case '+':
			// Increment the byte at the memory pointer
			memory[pointer]++
		case '-':
			// Decrement the byte at the memory pointer
			memory[pointer]--
		case '.':
			// Output the byte at the memory pointer
			buf.WriteByte(memory[pointer])
		case ',':
			// Input a byte and store it in the memory
			panic("unsupported")
			// fmt.Scan(&memory[pointer])
		case '[':
			// Jump forward past the matching ']' if the byte at the memory pointer is zero
			if memory[pointer] == 0 {
				braceCount := 1
				for braceCount > 0 {
					i++
					if code[i] == '[' {
						braceCount++
					} else if code[i] == ']' {
						braceCount--
					}
				}
			}
		case ']':
			// Jump backward to the matching '[' if the byte at the memory pointer is nonzero
			if memory[pointer] != 0 {
				braceCount := 1
				for braceCount > 0 {
					i--
					if code[i] == ']' {
						braceCount++
					} else if code[i] == '[' {
						braceCount--
					}
				}
				i-- // Move back one more to compensate for the upcoming increment in the loop
			}
		}
	}
	return buf.String()
}
