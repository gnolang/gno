package profiler

import "testing"

type mockMachineInfo struct {
	frames []FrameInfo
	cycles int64
	gas    int64
	id     uintptr
}

func (m *mockMachineInfo) GetFrames() []FrameInfo { return m.frames }
func (m *mockMachineInfo) GetCycles() int64       { return m.cycles }
func (m *mockMachineInfo) GetGasUsed() int64      { return m.gas }
func (m *mockMachineInfo) Identity() uintptr      { return m.id }

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

	if got := p.totalCycles; got != 110 {
		t.Fatalf("expected total cycles 110, got %d", got)
	}
	if got := p.totalGas; got != 25 {
		t.Fatalf("expected total gas 25, got %d", got)
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

	// Next callback has frames and should record the delta (80-30)
	machine.frames = []FrameInfo{mockFrame{name: "foo", file: "foo.gno", line: 1}}
	machine.cycles = 80
	p.RecordSample(machine)

	if got := p.totalCycles; got != 50 {
		t.Fatalf("expected total cycles 50, got %d", got)
	}
}

func TestProfilerRecordLineSampleHandlesCounterReset(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.EnableLineProfiling()
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	file := "foo.gno"
	line := 10

	p.RecordLineSample("foo", file, line, 100, 0)
	p.RecordLineSample("foo", file, line, 5, 0) // counter reset -> delta should be 5

	stats := p.lineSamples[file][line]
	if stats == nil {
		t.Fatalf("expected line stats for %s:%d", file, line)
	}
	if got := stats.cycles; got != 105 {
		t.Fatalf("expected cumulative cycles 105, got %d", got)
	}

	// Skip a test file to ensure baseline still updates
	p.RecordLineSample("foo", "foo_test.gno", 10, 50, 0)

	// Next valid line sample should use delta from the skipped baseline (60-50)
	p.RecordLineSample("foo", file, line, 60, 0)
	if got := p.lineSamples[file][line].cycles; got != 115 {
		t.Fatalf("expected cumulative cycles 115, got %d", got)
	}
}

func TestLineSamplesPopulateFunctionStats(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.EnableLineProfiling()
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	// Sampling via call stacks should be ignored while line profiling is enabled.
	machine := &mockMachineInfo{
		frames: []FrameInfo{mockFrame{
			name:    "foo",
			file:    "foo.gno",
			pkgPath: "pkg",
			line:    1,
		}},
		cycles: 10,
	}
	p.RecordSample(machine)
	if len(p.functionLines) != 0 {
		t.Fatalf("expected no function line stats from stack sampling when line profiling is enabled")
	}

	funcName := "pkg.foo"
	file := "foo.gno"
	canonicalFile := canonicalFilePath(file, funcName)

	p.RecordLineSample(funcName, file, 5, 100, 0)
	p.RecordLineSample(funcName, file, 5, 130, 0)

	info := p.functionLines[funcName]
	if info == nil {
		t.Fatalf("expected function line stats for %s", funcName)
	}
	stat := info.fileSamples[canonicalFile][5]
	if stat == nil {
		t.Fatalf("expected stats for %s:%d", canonicalFile, 5)
	}
	if stat.cycles != 130 {
		t.Fatalf("expected 130 cycles, got %d", stat.cycles)
	}
	if stat.count != 2 {
		t.Fatalf("expected 2 samples, got %d", stat.count)
	}
	if info.totalCycles != 130 {
		t.Fatalf("expected total cycles 130, got %d", info.totalCycles)
	}
}

func TestLineProfilingTracksGas(t *testing.T) {
	p := NewProfiler(ProfileGas, 1)
	p.EnableLineProfiling()
	p.StartProfiling(nil, Options{Type: ProfileGas, SampleRate: 1})

	funcName := "pkg.foo"
	file := "pkg/foo.gno"
	line := 5

	// Seed line-level stats so gas can be merged into the existing entry.
	p.RecordLineSample(funcName, file, line, 10, 1)

	machine := &mockMachineInfo{
		frames: []FrameInfo{mockFrame{
			name:    "foo",
			file:    file,
			pkgPath: "pkg",
			line:    line,
		}},
		cycles: 20,
		gas:    15,
		id:     1,
	}
	p.RecordSample(machine)

	profile := p.StopProfiling()
	if profile == nil {
		t.Fatalf("expected profile to be returned")
	}

	fn := profile.FunctionLines[funcName]
	if fn == nil {
		t.Fatalf("expected function line data for %s", funcName)
	}
	canonicalFile := canonicalFilePath(file, funcName)
	stat := fn.fileSamples[canonicalFile][line]
	if stat == nil {
		t.Fatalf("expected line stats for %s:%d", canonicalFile, line)
	}
	if stat.gas != 15 {
		t.Fatalf("expected gas 15, got %d", stat.gas)
	}
	if fn.totalGas != 15 {
		t.Fatalf("expected total gas 15, got %d", fn.totalGas)
	}
}

func TestProfilerSeparatesBaselinesPerMachine(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	m1 := &mockMachineInfo{
		id: 1,
		frames: []FrameInfo{mockFrame{
			name:    "foo",
			file:    "foo.gno",
			line:    1,
			pkgPath: "pkg",
		}},
		cycles: 10,
	}
	m2 := &mockMachineInfo{
		id: 2,
		frames: []FrameInfo{mockFrame{
			name:    "bar",
			file:    "bar.gno",
			line:    1,
			pkgPath: "pkg",
		}},
		cycles: 5,
	}

	p.RecordSample(m1) // +10
	p.RecordSample(m2) // +5
	m1.cycles = 20
	p.RecordSample(m1) // +10 from m1 baseline, should not include m2

	if got := p.totalCycles; got != 25 {
		t.Fatalf("expected isolated baselines, total cycles 25, got %d", got)
	}
}

func TestProfilerResetsBaselineWhenIdentityReused(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	id := uintptr(1)
	m1 := &mockMachineInfo{
		id: id,
		frames: []FrameInfo{mockFrame{
			name:    "foo",
			file:    "foo.gno",
			line:    1,
			pkgPath: "pkg",
		}},
		cycles: 100,
		gas:    10,
	}
	p.RecordSample(m1)
	if got := p.totalCycles; got != 100 {
		t.Fatalf("expected cycles 100 after first sample, got %d", got)
	}

	// Reuse identity with a fresh machine whose counters restarted.
	m2 := &mockMachineInfo{
		id:     id,
		frames: m1.frames,
		cycles: 100,
		gas:    10,
	}
	p.RecordSample(m2)
	if got := p.totalCycles; got != 200 {
		t.Fatalf("expected cycles 200 after baseline reset, got %d", got)
	}
	if got := p.totalGas; got != 20 {
		t.Fatalf("expected gas 20 after baseline reset, got %d", got)
	}
}

func TestProfilerIgnoresRestartWhileRunning(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	m := &mockMachineInfo{
		id:     1,
		frames: []FrameInfo{mockFrame{name: "foo", file: "foo.gno", line: 1}},
		cycles: 10,
	}
	p.RecordSample(m)

	// Second start should be ignored, preserving accumulated state.
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	m.cycles = 20
	p.RecordSample(m)

	if got := p.totalCycles; got != 20 {
		t.Fatalf("expected totals to continue after redundant start, got %d", got)
	}
}

func TestRecordAllocPopulatesCallTreeAndLines(t *testing.T) {
	p := NewProfiler(ProfileMemory, 1)
	p.StartProfiling(nil, Options{Type: ProfileMemory, SampleRate: 1})

	m := &mockMachineInfo{
		id:     1,
		frames: []FrameInfo{mockFrame{name: "allocFn", file: "pkg/foo.gno", pkgPath: "pkg", line: 7}},
	}
	p.RecordAlloc(m, 64, 2, "struct")
	profile := p.StopProfiling()

	if profile.CallTree == nil || len(profile.CallTree.Children) == 0 || profile.CallTree.Children[0].AllocBytes == 0 {
		t.Fatalf("expected call tree to accumulate allocation bytes")
	}
	if stats := profile.LineStats["pkg/foo.gno"][7]; stats == nil || stats.allocBytes == 0 || stats.allocations == 0 {
		t.Fatalf("expected line stats to include allocation data")
	}
}

func TestProfilerFiltersTestingPackage(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	machine := &mockMachineInfo{
		frames: []FrameInfo{
			mockFrame{name: "RunTest", file: "run.gno", pkgPath: "testing", line: 1},
			mockFrame{name: "Target", file: "demo.gno", pkgPath: "gno.land/p/demo", line: 2},
		},
		cycles: 100,
	}

	p.RecordSample(machine)

	if _, ok := p.funcStats["testing.RunTest"]; ok {
		t.Fatalf("expected testing package functions to be filtered out")
	}
	if _, ok := p.funcStats["gno.land/p/demo.Target"]; !ok {
		t.Fatalf("expected user function to remain in stats")
	}

	p.EnableLineProfiling()
	p.RecordLineSample("testing.RunTest", "run.gno", 1, 10, 0)
	p.RecordLineSample("gno.land/p/demo.Target", "demo.gno", 2, 20, 0)

	if _, ok := p.functionLines["testing.RunTest"]; ok {
		t.Fatalf("expected no line data for testing package")
	}
	if _, ok := p.functionLines["gno.land/p/demo.Target"]; !ok {
		t.Fatalf("expected line data for user function")
	}
}

func TestProfilerRecordAllocUsesCallStack(t *testing.T) {
	p := NewProfiler(ProfileMemory, 1)
	p.StartProfiling(nil, Options{Type: ProfileMemory, SampleRate: 1})

	frame := mockFrame{
		name:    "Leaf",
		file:    "leaf.gno",
		pkgPath: "gno.land/p/demo",
		line:    10,
	}
	machine := &mockMachineInfo{
		frames: []FrameInfo{frame},
	}

	const allocBytes = 512
	const allocObjects = 3
	p.RecordAlloc(machine, allocBytes, allocObjects, "test")

	profile := p.StopProfiling()
	if profile == nil {
		t.Fatalf("expected profile to be returned")
	}

	var stat *FunctionStat
	for _, fn := range profile.Functions {
		if fn.Name == "gno.land/p/demo.Leaf" {
			stat = fn
			break
		}
	}
	if stat == nil {
		t.Fatalf("expected allocation stats for gno.land/p/demo.Leaf, got %+v", profile.Functions)
	}
	if stat.AllocBytes != allocBytes {
		t.Fatalf("expected %d alloc bytes, got %d", allocBytes, stat.AllocBytes)
	}
	if stat.AllocObjects != allocObjects {
		t.Fatalf("expected %d alloc objects, got %d", allocObjects, stat.AllocObjects)
	}
	if stat.CallCount != 1 {
		t.Fatalf("expected call count 1, got %d", stat.CallCount)
	}
}

func TestProfilerAnonymousFunctionNameFallback(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	machine := &mockMachineInfo{
		frames: []FrameInfo{
			mockFrame{
				name:    "",
				file:    "anon.gno",
				pkgPath: "gno.land/p/demo",
				line:    1,
			},
		},
		cycles: 10,
	}

	p.RecordSample(machine)
	profile := p.StopProfiling()
	if profile == nil {
		t.Fatalf("expected profile data")
	}
	if len(profile.Functions) != 1 {
		t.Fatalf("expected single entry, got %d", len(profile.Functions))
	}
	if got := profile.Functions[0].Name; got != "gno.land/p/demo.<anonymous>" {
		t.Fatalf("expected anonymous fallback name, got %q", got)
	}
}
