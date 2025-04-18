package evm

const maxlen = 30000

func Execute(code []byte) ([]byte, error) {
	var (
		memory = make([]byte, maxlen) // memory tape
		stack  = NewByteStack()       // byte stack
		rip    = 0                    // initial memory pointer
		buf    = []byte{}             // output buffer
	)

	// Loop through each character in the code
	for rip != -1 && rip < len(code) {
		opData, ok := opCodes[OpCode(code[rip])]
		if ok {
			return nil, opNotImplemented(&rip, code, stack, memory)
		}
		opData.operation(&rip, code, stack, memory)
	}
	return buf, nil
}
