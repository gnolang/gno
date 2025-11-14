package gnolang

import (
	"github.com/gnolang/gno/gnovm/pkg/instrumentation"
	"github.com/gnolang/gno/gnovm/pkg/profiler"
)

type profilingState struct {
	profiler      *profiler.Profiler
	sink          *profiler.SinkAdapter
	options       profiler.Options
	sampleCounter int
}

func (m *Machine) profileSink() instrumentation.Sink {
	if m.profileState == nil {
		return nil
	}
	return m.profileState.sink
}

func (m *Machine) refreshInstrumentationSink() {
	combined := combineInstrumentationSinks(m.baseInstrumentation, m.profileSink())
	m.instrumentation = combined
	if m.Alloc != nil {
		m.Alloc.SetInstrumentationSink(combined)
	}
	if m.Store != nil {
		if alloc := m.Store.GetAllocator(); alloc != nil {
			alloc.SetInstrumentationSink(combined)
		}
	}
}

// StartProfiling enables profiling with the provided options.
func (m *Machine) StartProfiling(options profiler.Options) {
	if m.profileState != nil {
		m.StopProfiling()
	}
	if options.SampleRate <= 0 {
		options.SampleRate = 1000
	}
	p := profiler.NewProfiler(options.Type, options.SampleRate)
	p.StartProfiling(nil, options)
	sink := profiler.NewSinkAdapter(p, options)
	m.profileState = &profilingState{
		profiler: p,
		sink:     sink,
		options:  options,
	}
	m.refreshInstrumentationSink()
}

// StopProfiling stops the active profiler (if any) and returns the result.
func (m *Machine) StopProfiling() *profiler.Profile {
	if m.profileState == nil {
		return nil
	}
	profile := m.profileState.profiler.StopProfiling()
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
		Frames: frames,
		Cycles: m.Cycles,
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
		Func:   funcName,
		File:   file,
		Line:   line,
		Cycles: m.Cycles,
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
