package evm

const maxlen = 30000

func Execute(code []byte) []byte {
	var (
		//		memory  = make([]byte, maxlen) // memory tape
		stack = NewByteStack() // byte stack
		//		pointer = 0                    // initial memory pointer
		buf []byte // output buffer
	)

	// Loop through each character in the code
	for i := 0; i < len(code); i++ {
		switch code[i] {
		case 0x01:
			// Add stack 1 et stack 2
			a1, err := stack.Pop()
			if err != nil {
				panic(err)
			}

			a2, err := stack.Pop()
			if err != nil {
				panic(err)
			}

			stack.Push(a1 + a2)
		}
	}
	return buf
}
