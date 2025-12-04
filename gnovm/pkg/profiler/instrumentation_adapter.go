package profiler

import (
	"github.com/gnolang/gno/gnovm/pkg/instrumentation"
)

// SinkAdapter bridges instrumentation events to the profiler.
type SinkAdapter struct {
	profiler *Profiler
	opts     Options
}

// NewSinkAdapter builds a sink adapter for the given profiler and options.
func NewSinkAdapter(p *Profiler, opts Options) *SinkAdapter {
	return &SinkAdapter{
		profiler: p,
		opts:     opts,
	}
}

// WantsSamples reports whether CPU/gas samples are required.
func (sa *SinkAdapter) WantsSamples() bool {
	switch sa.opts.Type {
	case ProfileCPU, ProfileGas:
		return true
	default:
		return false
	}
}

// WantsAllocations reports whether memory allocation events should be emitted.
func (sa *SinkAdapter) WantsAllocations() bool {
	return sa.opts.Type == ProfileMemory
}

// WantsLineSamples reports whether line-level samples are desired.
func (sa *SinkAdapter) WantsLineSamples() bool {
	return sa.profiler.IsLineProfilingEnabled()
}

// OnSample handles CPU/Gas samples emitted by the VM.
func (sa *SinkAdapter) OnSample(ctx *instrumentation.SampleContext) {
	if sa.profiler == nil || ctx == nil {
		return
	}
	sa.profiler.RecordSample(adapterMachineInfo{
		frames: ctx.Frames,
		cycles: ctx.Cycles,
		gas:    ctx.GasUsed,
		id:     ctx.MachineID,
	})
}

// OnAllocation handles memory allocation events.
func (sa *SinkAdapter) OnAllocation(ev *instrumentation.AllocationEvent) {
	if sa.profiler == nil || ev == nil {
		return
	}
	sa.profiler.RecordAlloc(adapterMachineInfo{
		frames: ev.Stack,
		id:     ev.MachineID,
	}, ev.Bytes, ev.Objects, ev.Kind)
}

// OnLineSample forwards line-level samples to the profiler.
func (sa *SinkAdapter) OnLineSample(sample *instrumentation.LineSample) {
	if sa.profiler == nil || sample == nil {
		return
	}
	sa.profiler.RecordLineSample(sample.Func, sample.File, sample.Line, sample.Cycles, sample.MachineID)
}

// adapterMachineInfo implements MachineInfo using instrumentation snapshots.
type adapterMachineInfo struct {
	frames []instrumentation.FrameSnapshot
	cycles int64
	gas    int64
	id     uintptr
}

func (a adapterMachineInfo) GetFrames() []FrameInfo {
	frames := make([]FrameInfo, 0, len(a.frames))
	for i := range a.frames {
		frames = append(frames, frameSnapshotAdapter{snap: a.frames[i]})
	}
	return frames
}

func (a adapterMachineInfo) GetCycles() int64  { return a.cycles }
func (a adapterMachineInfo) GetGasUsed() int64 { return a.gas }

// Identity returns a stable identifier for the originating machine, if known.
func (a adapterMachineInfo) Identity() uintptr {
	return a.id
}

type frameSnapshotAdapter struct {
	snap instrumentation.FrameSnapshot
}

func (f frameSnapshotAdapter) IsCall() bool {
	if !f.snap.IsCall && f.snap.Line == 0 {
		// Default to true when explicit info missing to preserve previous behavior.
		return true
	}
	return f.snap.IsCall
}

func (f frameSnapshotAdapter) GetFuncName() string { return f.snap.FuncName }
func (f frameSnapshotAdapter) GetFileName() string { return f.snap.File }
func (f frameSnapshotAdapter) GetPkgPath() string  { return f.snap.PkgPath }

func (f frameSnapshotAdapter) GetSource() SourceInfo {
	return sourceSnapshotAdapter{line: f.snap.Line, column: f.snap.Column}
}

type sourceSnapshotAdapter struct {
	line   int
	column int
}

func (s sourceSnapshotAdapter) GetLine() int {
	return s.line
}

func (s sourceSnapshotAdapter) GetColumn() int {
	return s.column
}
