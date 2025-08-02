package profiler

// Machine is a mock type for testing
type Machine struct {
	Cycles int64
}

// GetCycles implements MachineInfo interface for testing
func (m *Machine) GetCycles() int64 {
	return m.Cycles
}

// GetFrames implements MachineInfo interface for testing
func (m *Machine) GetFrames() []FrameInfo {
	return nil
}
