package simpledao

type (
	executeDelegate   func() error
	isExpiredDelegate func() bool
)

type mockExecutor struct {
	executeFn   executeDelegate
	isExpiredFn isExpiredDelegate
}

func (m *mockExecutor) Execute() error {
	if m.executeFn != nil {
		return m.executeFn()
	}

	return nil
}

func (m *mockExecutor) IsExpired() bool {
	if m.isExpiredFn != nil {
		return m.isExpiredFn()
	}

	return false
}
