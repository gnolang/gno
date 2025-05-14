package evm

func Execute(code []byte) ([]byte, error) {
	st := newEmptyState()
	st.setCode(code)
	err := st.executeByteCode()
	if err != nil {
		return []byte{}, err
	}
	return st.getResult(), nil
}
