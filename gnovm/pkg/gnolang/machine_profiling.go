package gnolang

import (
	"unsafe"

	"github.com/gnolang/gno/gnovm/pkg/instrumentation"
	"github.com/gnolang/gno/gnovm/pkg/profiler"
)

type profilingState struct {
	profiler      *profiler.Profiler
	sink          instrumentation.Sink
	options       profiler.Options
	sampleCounter int
}

func (m *Machine) refreshInstrumentationSink() {
	var profileSink instrumentation.Sink
	if m.profileState != nil {
		profileSink = m.profileState.sink
	}
	combined := combineInstrumentationSinks(m.baseInstrumentation, profileSink)
	m.instrumentation = combined
	var allocSink instrumentation.Sink
	if combined != nil {
		allocSink = &allocationStackInjector{
			machine: m,
			sink:    combined,
		}
	}
	if m.Alloc != nil {
		m.Alloc.SetInstrumentationSink(allocSink)
	}
	if m.Store != nil {
		if alloc := m.Store.GetAllocator(); alloc != nil {
			alloc.SetInstrumentationSink(allocSink)
		}
	}
}

// StartProfiling enables profiling with the provided options.
func (m *Machine) StartProfiling(options profiler.Options) {
	p := profiler.NewProfiler(options.Type, options.SampleRate)
	p.StartProfiling(nil, options)
	m.StartProfilingWithSink(profiler.NewSinkAdapter(p, options), options)
	if m.profileState != nil {
		m.profileState.profiler = p
	}
}

// StartProfilingWithSink enables profiling using the provided instrumentation sink.
func (m *Machine) StartProfilingWithSink(s instrumentation.Sink, options profiler.Options) {
	if s == nil {
		return
	}
	if options.SampleRate <= 0 {
		options.SampleRate = 1000
	}
	m.profileState = &profilingState{
		sink:    s,
		options: options,
	}
	m.refreshInstrumentationSink()
}

// StopProfiling stops the active profiler (if any) and returns the result.
func (m *Machine) StopProfiling() *profiler.Profile {
	if m.profileState == nil {
		return nil
	}
	var profile *profiler.Profile
	if m.profileState.profiler != nil {
		profile = m.profileState.profiler.StopProfiling()
	}
	m.profileState = nil
	m.refreshInstrumentationSink()
	return profile
}

func combineInstrumentationSinks(sinks ...instrumentation.Sink) instrumentation.Sink {
	filtered := make([]instrumentation.Sink, 0, len(sinks))
	for _, sink := range sinks {
		if sink != nil {
			filtered = append(filtered, sink)
		}
	}
	switch len(filtered) {
	case 0:
		return nil
	case 1:
		return filtered[0]
	default:
		return &instrumentationFanout{sinks: filtered}
	}
}

type instrumentationFanout struct {
	sinks []instrumentation.Sink
}

// allocationStackInjector wraps an instrumentation sink so allocation events
// carry call stacks even though the allocator only emits size/object data.
// The profiler needs real stack frames to attribute allocations, so on each
// event we capture the current machine stack if the event omitted one.
type allocationStackInjector struct {
	machine *Machine
	sink    instrumentation.Sink
}

var (
	_ instrumentation.Sink         = (*allocationStackInjector)(nil)
	_ instrumentation.Capabilities = (*allocationStackInjector)(nil)
)

func (f *instrumentationFanout) OnSample(ctx *instrumentation.SampleContext) {
	for _, sink := range f.sinks {
		sink.OnSample(ctx)
	}
}

func (f *instrumentationFanout) OnAllocation(ev *instrumentation.AllocationEvent) {
	for _, sink := range f.sinks {
		sink.OnAllocation(ev)
	}
}

func (f *instrumentationFanout) OnLineSample(sample *instrumentation.LineSample) {
	for _, sink := range f.sinks {
		sink.OnLineSample(sample)
	}
}

func (f *instrumentationFanout) WantsSamples() bool {
	for _, sink := range f.sinks {
		if wantsSamples(sink) {
			return true
		}
	}
	return false
}

func (f *instrumentationFanout) WantsAllocations() bool {
	for _, sink := range f.sinks {
		if wantsAllocations(sink) {
			return true
		}
	}
	return false
}

func (f *instrumentationFanout) WantsLineSamples() bool {
	for _, sink := range f.sinks {
		if wantsLineSamples(sink) {
			return true
		}
	}
	return false
}

func wantsSamples(s instrumentation.Sink) bool {
	if caps, ok := s.(instrumentation.Capabilities); ok {
		return caps.WantsSamples()
	}
	return true
}

func wantsAllocations(s instrumentation.Sink) bool {
	if caps, ok := s.(instrumentation.Capabilities); ok {
		return caps.WantsAllocations()
	}
	return true
}

func wantsLineSamples(s instrumentation.Sink) bool {
	if caps, ok := s.(instrumentation.Capabilities); ok {
		return caps.WantsLineSamples()
	}
	return true
}

func (a *allocationStackInjector) OnSample(ctx *instrumentation.SampleContext) {
	if a.sink != nil {
		a.sink.OnSample(ctx)
	}
}

func (a *allocationStackInjector) OnAllocation(ev *instrumentation.AllocationEvent) {
	if a.sink == nil {
		return
	}
	if ev != nil && len(ev.Stack) == 0 && a.machine != nil {
		if frames := a.machine.captureFrameSnapshots(); len(frames) > 0 {
			// Copy the frames so downstream sinks can retain the stack.
			stack := make([]instrumentation.FrameSnapshot, len(frames))
			copy(stack, frames)
			ev.Stack = stack
		}
	}
	if ev != nil && ev.MachineID == 0 && a.machine != nil {
		ev.MachineID = uintptr(unsafe.Pointer(a.machine))
	}
	a.sink.OnAllocation(ev)
}

func (a *allocationStackInjector) OnLineSample(sample *instrumentation.LineSample) {
	if a.sink != nil {
		a.sink.OnLineSample(sample)
	}
}

func (a *allocationStackInjector) WantsSamples() bool {
	return wantsSamples(a.sink)
}

func (a *allocationStackInjector) WantsAllocations() bool {
	return wantsAllocations(a.sink)
}

func (a *allocationStackInjector) WantsLineSamples() bool {
	return wantsLineSamples(a.sink)
}

func (m *Machine) instrumentationCapabilities() instrumentation.Capabilities {
	if caps, ok := m.instrumentation.(instrumentation.Capabilities); ok {
		return caps
	}
	return nil
}

func (m *Machine) maybeEmitSample() {
	if m.profileState == nil || m.instrumentation == nil {
		return
	}
	if caps := m.instrumentationCapabilities(); caps != nil && !caps.WantsSamples() {
		return
	}
	rate := m.profileState.options.SampleRate
	if rate <= 0 {
		rate = 1
	}
	m.profileState.sampleCounter++
	if m.profileState.sampleCounter%rate != 0 {
		return
	}
	ctx := m.captureSampleContext()
	if ctx == nil {
		return
	}
	m.instrumentation.OnSample(ctx)
}

func (m *Machine) captureSampleContext() *instrumentation.SampleContext {
	frames := m.captureFrameSnapshots()
	if len(frames) == 0 {
		return nil
	}
	ctx := &instrumentation.SampleContext{
		Frames:    frames,
		Cycles:    m.Cycles,
		MachineID: uintptr(unsafe.Pointer(m)),
	}
	if m.GasMeter != nil {
		ctx.GasUsed = m.GasMeter.GasConsumed()
	}
	return ctx
}

func (m *Machine) captureFrameSnapshots() []instrumentation.FrameSnapshot {
	if len(m.Frames) == 0 {
		return nil
	}
	snaps := make([]instrumentation.FrameSnapshot, 0, len(m.Frames))
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if fr.Func == nil {
			continue
		}
		snap := instrumentation.FrameSnapshot{
			FuncName: string(fr.Func.Name),
			File:     fr.Func.FileName,
			PkgPath:  fr.Func.PkgPath,
			IsCall:   fr.IsCall(),
		}
		if fr.Source != nil {
			snap.Line = fr.Source.GetLine()
			snap.Column = fr.Source.GetColumn()
		}
		snaps = append(snaps, snap)
	}
	return snaps
}

func (m *Machine) recordLineSampleIfNeeded() {
	if m.instrumentation == nil {
		return
	}
	if caps := m.instrumentationCapabilities(); caps != nil && !caps.WantsLineSamples() {
		return
	}
	sample := m.captureLineSample()
	if sample == nil {
		return
	}
	m.instrumentation.OnLineSample(sample)
}

func (m *Machine) captureLineSample() *instrumentation.LineSample {
	if len(m.Frames) == 0 {
		return nil
	}
	line := m.currentLineNumber()
	if line <= 0 {
		return nil
	}
	frame := &m.Frames[len(m.Frames)-1]
	if frame.Func == nil {
		return nil
	}
	funcName := string(frame.Func.Name)
	if frame.Func.PkgPath != "" {
		funcName = frame.Func.PkgPath + "." + funcName
	}
	file := frame.Func.FileName
	if file == "" {
		return nil
	}
	return &instrumentation.LineSample{
		Func:      funcName,
		File:      file,
		Line:      line,
		Cycles:    m.Cycles,
		MachineID: uintptr(unsafe.Pointer(m)),
	}
}

func (m *Machine) currentLineNumber() int {
	if len(m.Exprs) > 0 {
		if expr := m.PeekExpr(1); expr != nil {
			return expr.GetLine()
		}
	}
	if len(m.Stmts) > 0 {
		stmt := m.PeekStmt(1)
		if stmt == nil {
			return 0
		}
		if bs, ok := stmt.(*bodyStmt); ok {
			if idx := bs.NextBodyIndex - 1; 0 <= idx && idx < len(bs.Body) {
				stmt = bs.Body[idx]
			}
		}
		return stmt.GetLine()
	}
	return 0
}
