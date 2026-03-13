package ledger

type mockLedgerDevice struct {
	exchange func([]byte) ([]byte, error)
	calls    [][]byte
}

func (m *mockLedgerDevice) Exchange(command []byte) ([]byte, error) {
	if m.exchange == nil {
		return nil, nil
	}
	cmdCopy := make([]byte, len(command))
	copy(cmdCopy, command)
	m.calls = append(m.calls, cmdCopy)
	return m.exchange(command)
}

func (m *mockLedgerDevice) Close() error {
	return nil
}
