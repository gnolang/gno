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

func adaptMachine(m *Machine) profiler.MachineInfo {
	return &machineAdapter{m: m}
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
