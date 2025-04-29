package evm

func Execute(code []byte) ([]byte, error) {
	st := newState()
	err := st.executeByteCode(code)
	if err != nil {
		return []byte{}, err
	}
	return st.getResult(), nil
}
