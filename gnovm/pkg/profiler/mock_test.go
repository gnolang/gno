package profiler

type machineMock struct {
	Cycles  int64
	GasUsed int64
}

func (m *machineMock) GetCycles() int64       { return m.Cycles }
func (m *machineMock) GetGasUsed() int64      { return m.GasUsed }
func (m *machineMock) GetFrames() []FrameInfo { return nil }

type mockMachineInfo struct {
	frames  []FrameInfo
	cycles  int64
	gasUsed int64
}

func (m *mockMachineInfo) GetFrames() []FrameInfo { return m.frames }
func (m *mockMachineInfo) GetCycles() int64       { return m.cycles }
func (m *mockMachineInfo) GetGasUsed() int64      { return m.gasUsed }

type mockFrameInfo struct {
	isCall   bool
	funcName string
	fileName string
	pkgPath  string
	source   SourceInfo
}

func (f *mockFrameInfo) IsCall() bool          { return f.isCall }
func (f *mockFrameInfo) GetFuncName() string   { return f.funcName }
func (f *mockFrameInfo) GetFileName() string   { return f.fileName }
func (f *mockFrameInfo) GetPkgPath() string    { return f.pkgPath }
func (f *mockFrameInfo) GetSource() SourceInfo { return f.source }

type mockSourceInfo struct {
	line   int
	column int
}

func (s *mockSourceInfo) GetLine() int   { return s.line }
func (s *mockSourceInfo) GetColumn() int { return s.column }
