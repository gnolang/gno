package evm

func Execute(code []byte) ([]byte, error) {
	st := newState()
	err := st.executeByteCodeWithGas(code, -1)
	if err != nil {
		return []byte{}, err
	}
	return st.getResult(), nil
}
