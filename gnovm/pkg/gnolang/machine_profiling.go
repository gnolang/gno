package gnolang

import (
	"github.com/gnolang/gno/gnovm/pkg/profiler"
)

// Adapter types to implement profiler interfaces

// machineAdapter adapts Machine to profiler.MachineInfo
type machineAdapter struct {
	m *Machine
}

func (ma *machineAdapter) GetFrames() []profiler.FrameInfo {
	frames := make([]profiler.FrameInfo, len(ma.m.Frames))
	for i := range ma.m.Frames {
		frames[i] = &frameAdapter{f: &ma.m.Frames[i]}
	}
	return frames
}

func (ma *machineAdapter) GetCycles() int64 {
	return ma.m.Cycles
}

// frameAdapter adapts Frame to profiler.FrameInfo
type frameAdapter struct {
	f *Frame
}

func (fa *frameAdapter) IsCall() bool {
	return fa.f.IsCall()
}

func (fa *frameAdapter) GetFuncName() string {
	if fa.f.Func == nil {
		return ""
	}
	return string(fa.f.Func.Name)
}

func (fa *frameAdapter) GetFileName() string {
	if fa.f.Func == nil {
		return ""
	}
	return fa.f.Func.FileName
}

func (fa *frameAdapter) GetPkgPath() string {
	if fa.f.Func == nil {
		return ""
	}
	return fa.f.Func.PkgPath
}

func (fa *frameAdapter) GetSource() profiler.SourceInfo {
	return &sourceAdapter{n: fa.f.Source}
}

// sourceAdapter adapts Node to profiler.SourceInfo
type sourceAdapter struct {
	n Node
}

func (sa *sourceAdapter) GetLine() int {
	if sa.n == nil {
		return 0
	}
	return sa.n.GetLine()
}

func (sa *sourceAdapter) GetColumn() int {
	if sa.n == nil {
		return 0
	}
	return sa.n.GetColumn()
}

// storeAdapter adapts Store to profiler.Store
type storeAdapter struct {
	s Store
}

func (sa *storeAdapter) GetMemFile(pkgPath, name string) *profiler.MemFile {
	memFile := sa.s.GetMemFile(pkgPath, name)
	if memFile == nil {
		return nil
	}
	return &profiler.MemFile{
		Name: memFile.Name,
		Body: memFile.Body,
	}
}

// Helper functions to create adapters

func adaptMachine(m *Machine) profiler.MachineInfo {
	return &machineAdapter{m: m}
}

func adaptStore(s Store) profiler.Store {
	if s == nil {
		return nil
	}
	return &storeAdapter{s: s}
}

// StartProfiling starts profiling the VM
func (m *Machine) StartProfiling(options profiler.Options) {
	if m.profiler == nil {
		m.profiler = profiler.NewProfiler()
	}
	m.profiler.StartProfiling(adaptMachine(m), options)
}

// StopProfiling stops profiling and returns the profile
func (m *Machine) StopProfiling() *profiler.Profile {
	if m.profiler == nil {
		return nil
	}
	return m.profiler.StopProfiling()
}

// RecordProfileSample records a profiling sample if profiling is enabled
func (m *Machine) RecordProfileSample() {
	if m.profiler != nil && m.profiler.IsEnabled() {
		m.profiler.RecordSample(adaptMachine(m))
	}
}

// GetProfile returns the current profile without stopping profiling
func (m *Machine) GetProfile() *profiler.Profile {
	if m.profiler == nil {
		return nil
	}
	return m.profiler.GetProfile()
}

// EnableLineProfiling enables line-level profiling
func (m *Machine) EnableLineProfiling() {
	if m.profiler != nil {
		m.profiler.EnableLineProfiling()
	}
}

// DisableLineProfiling disables line-level profiling
func (m *Machine) DisableLineProfiling() {
	if m.profiler != nil {
		m.profiler.DisableLineProfiling()
	}
}

// IsProfilingEnabled returns true if profiling is currently enabled
func (m *Machine) IsProfilingEnabled() bool {
	return m.profiler != nil && m.profiler.IsEnabled()
}

// RecordAllocation records a memory allocation if profiling is enabled
func (m *Machine) RecordAllocation(size int64, count int64, allocType string) {
	if m.profiler != nil && m.profiler.IsEnabled() {
		m.profiler.RecordAlloc(adaptMachine(m), size, count, allocType)
	}
}

// SetProfiler sets the profiler for the machine (used for testing)
func (m *Machine) SetProfiler(p *profiler.Profiler) {
	m.profiler = p
}

// getCurrentLocation extracts the current execution location from the machine state
func (m *Machine) getCurrentLocation() *profiler.ProfileLocation {
	if m == nil {
		return nil
	}

	// Get file name from current frame
	var fileName string
	if len(m.Frames) > 0 {
		frame := &m.Frames[len(m.Frames)-1]
		if frame.Func != nil && frame.Func.FileName != "" {
			fileName = frame.Func.FileName
		}
	}

	// Try to get location from current statement
	if len(m.Stmts) > 0 {
		stmt := m.PeekStmt(1)
		if stmt != nil {
			loc := &profiler.ProfileLocation{
				Line:   stmt.GetLine(),
				Column: stmt.GetColumn(),
				File:   fileName,
			}

			// Get file information from current frame
			if len(m.Frames) > 0 {
				frame := &m.Frames[len(m.Frames)-1]
				if frame.Func != nil && frame.Func.FileName != "" {
					loc.File = frame.Func.FileName
				}
			}

			// Get function information from current frame
			if len(m.Frames) > 0 {
				frame := &m.Frames[len(m.Frames)-1]
				if frame.Func != nil {
					loc.Function = string(frame.Func.Name)
					if frame.Func.PkgPath != "" {
						loc.Function = frame.Func.PkgPath + "." + loc.Function
					}
				}
			}

			return loc
		}
	}

	// Try to get location from current expression
	if len(m.Exprs) > 0 {
		expr := m.PeekExpr(1)
		if expr != nil {
			loc := &profiler.ProfileLocation{
				Line:   expr.GetLine(),
				Column: expr.GetColumn(),
			}

			// Get file information from current frame
			if len(m.Frames) > 0 {
				frame := &m.Frames[len(m.Frames)-1]
				if frame.Func != nil && frame.Func.FileName != "" {
					loc.File = frame.Func.FileName
				}
			}

			// Get function information from current frame
			if len(m.Frames) > 0 {
				frame := &m.Frames[len(m.Frames)-1]
				if frame.Func != nil {
					loc.Function = string(frame.Func.Name)
					if frame.Func.PkgPath != "" {
						loc.Function = frame.Func.PkgPath + "." + loc.Function
					}
				}
			}

			return loc
		}
	}

	// Fall back to frame source if available
	if len(m.Frames) > 0 {
		frame := &m.Frames[len(m.Frames)-1]
		if frame.Source != nil {
			loc := &profiler.ProfileLocation{
				Line:   frame.Source.GetLine(),
				Column: frame.Source.GetColumn(),
			}

			// Get file information from current frame
			if len(m.Frames) > 0 {
				frame := &m.Frames[len(m.Frames)-1]
				if frame.Func != nil && frame.Func.FileName != "" {
					loc.File = frame.Func.FileName
				}
			}

			// Get function information
			if frame.Func != nil {
				loc.Function = string(frame.Func.Name)
				if frame.Func.PkgPath != "" {
					loc.Function = frame.Func.PkgPath + "." + loc.Function
				}
			}

			return loc
		}
	}

	return nil
}
