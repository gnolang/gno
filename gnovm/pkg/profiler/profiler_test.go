package profiler

import "testing"

type mockMachineInfo struct {
	frames []FrameInfo
	cycles int64
	gas    int64
}

func (m *mockMachineInfo) GetFrames() []FrameInfo { return m.frames }
func (m *mockMachineInfo) GetCycles() int64       { return m.cycles }
func (m *mockMachineInfo) GetGasUsed() int64      { return m.gas }

type mockFrame struct {
	name    string
	file    string
	pkgPath string
	line    int
}

func (f mockFrame) IsCall() bool          { return true }
func (f mockFrame) GetFuncName() string   { return f.name }
func (f mockFrame) GetFileName() string   { return f.file }
func (f mockFrame) GetPkgPath() string    { return f.pkgPath }
func (f mockFrame) GetSource() SourceInfo { return mockSource{line: f.line} }

type mockSource struct {
	line int
}

func (s mockSource) GetLine() int   { return s.line }
func (s mockSource) GetColumn() int { return 0 }

func TestProfilerRecordSampleHandlesCounterReset(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	frame := mockFrame{name: "foo", file: "foo.gno", line: 1}
	machine := &mockMachineInfo{
		frames: []FrameInfo{frame},
		cycles: 100,
		gas:    20,
	}

	p.RecordSample(machine)

	// Simulate counter reset (e.g. profiling another machine)
	machine.cycles = 10
	machine.gas = 5
	p.RecordSample(machine)

	if len(p.stackSamples) != 2 {
		t.Fatalf("expected 2 stack samples, got %d", len(p.stackSamples))
	}
	if got := p.stackSamples[1].cycles; got != 10 {
		t.Fatalf("expected reset delta cycles 10, got %d", got)
	}
	if got := p.stackSamples[1].gasUsed; got != 5 {
		t.Fatalf("expected reset delta gas 5, got %d", got)
	}
}

func TestProfilerRecordSampleUpdatesBaselineWithoutFrames(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	machine := &mockMachineInfo{
		cycles: 30,
	}

	// First callback has no frames; should not produce a sample but must update baseline
	p.RecordSample(machine)
	if len(p.stackSamples) != 0 {
		t.Fatalf("expected no samples when stack empty, got %d", len(p.stackSamples))
	}

	// Next callback has frames and should record the delta (80-30)
	machine.frames = []FrameInfo{mockFrame{name: "foo", file: "foo.gno", line: 1}}
	machine.cycles = 80
	p.RecordSample(machine)

	if len(p.stackSamples) != 1 {
		t.Fatalf("expected 1 stack sample, got %d", len(p.stackSamples))
	}
	if got := p.stackSamples[0].cycles; got != 50 {
		t.Fatalf("expected delta cycles 50, got %d", got)
	}
}

func TestProfilerRecordLineSampleHandlesCounterReset(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.EnableLineProfiling()
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	file := "foo.gno"
	line := 10

	p.RecordLineSample("foo", file, line, 100)
	p.RecordLineSample("foo", file, line, 5) // counter reset -> delta should be 5

	stats := p.lineSamples[file][line]
	if stats == nil {
		t.Fatalf("expected line stats for %s:%d", file, line)
	}
	if got := stats.cycles; got != 105 {
		t.Fatalf("expected cumulative cycles 105, got %d", got)
	}
	if len(p.profile.Samples) != 2 {
		t.Fatalf("expected 2 line samples, got %d", len(p.profile.Samples))
	}
	if got := p.profile.Samples[1].Value[1]; got != 5 {
		t.Fatalf("expected delta cycles 5 in sample, got %d", got)
	}

	// Skip a test file to ensure baseline still updates
	p.RecordLineSample("foo", "foo_test.gno", 10, 50)
	if len(p.profile.Samples) != 2 {
		t.Fatalf("expected skip to avoid new sample, got %d entries", len(p.profile.Samples))
	}

	// Next valid line sample should use delta from the skipped baseline (60-50)
	p.RecordLineSample("foo", file, line, 60)
	if got := p.lineSamples[file][line].cycles; got != 115 {
		t.Fatalf("expected cumulative cycles 115, got %d", got)
	}
	if len(p.profile.Samples) != 3 {
		t.Fatalf("expected third line sample, got %d", len(p.profile.Samples))
	}
	if got := p.profile.Samples[2].Value[1]; got != 10 {
		t.Fatalf("expected delta cycles 10 after skip, got %d", got)
	}
}
