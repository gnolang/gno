package profiler

// MachineInfo provides access to VM machine state for profiling
type MachineInfo interface {
	GetFrames() []FrameInfo
	GetCycles() int64
	GetGasUsed() int64 // Total gas used so far
}

// FrameInfo provides access to stack frame information
type FrameInfo interface {
	IsCall() bool
	GetFuncName() string
	GetFileName() string
	GetPkgPath() string
	GetSource() SourceInfo
}

// SourceInfo provides source location information
type SourceInfo interface {
	GetLine() int
	GetColumn() int
}

// Store provides access to source files
type Store interface {
	GetMemFile(pkgPath, name string) *MemFile
}

// MemFile represents an in-memory file
type MemFile struct {
	Name string
	Body string
}
