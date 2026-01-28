package benchops

import (
	"strconv"
	"strings"
	"time"
)

// ---- Op measurement methods

// BeginOp starts timing for an opcode with its source location context.
func (p *Profiler) BeginOp(op Op, ctx OpContext) {
	entry := &opStackEntry{op: op, ctx: ctx}
	if p.timingEnabled {
		entry.startTime = time.Now()
	}
	p.currentOp = entry
}

// EndOp stops timing for the current opcode and records the measurement.
func (p *Profiler) EndOp() {
	entry := p.currentOp
	if entry == nil {
		panic("benchops: EndOp: no matching BeginOp")
	}
	p.currentOp = nil

	gas := GetOpGas(entry.op)

	var dur time.Duration
	if p.timingEnabled {
		dur = entry.elapsed + time.Since(entry.startTime)
	}
	p.opStats[entry.op].Record(gas, dur)

	// Record location stats if context was set
	if entry.ctx.File != "" && entry.ctx.Line > 0 {
		p.recordLocation(entry.op, entry.ctx, dur)
	}

	// Record stack sample if stack tracking is enabled
	if p.stackEnabled && len(p.callStack) > 0 {
		p.recordStackSample(gas, dur)
	}
}

// recordLocation records gas and optionally timing for a specific source location.
// Pass zero duration when timing is disabled.
func (p *Profiler) recordLocation(op Op, ctx OpContext, dur time.Duration) {
	key := ctx.File + ":" + strconv.Itoa(ctx.Line)
	stat := p.locationStats[key]
	if stat == nil {
		stat = &LocationStat{
			File:     ctx.File,
			Line:     ctx.Line,
			FuncName: ctx.FuncName,
			PkgPath:  ctx.PkgPath,
		}
		p.locationStats[key] = stat
	}
	stat.Count++
	stat.TotalNs += dur.Nanoseconds()
	stat.Gas += GetOpGas(op)
}

// ---- Store measurement methods

// BeginStore starts timing for a store operation.
func (p *Profiler) BeginStore(op StoreOp) {
	// Pause current opcode timing on first store call
	if len(p.storeStack) == 0 && p.currentOp != nil {
		if p.timingEnabled {
			p.currentOp.elapsed += time.Since(p.currentOp.startTime)
		}
		p.opStack = append(p.opStack, *p.currentOp)
		p.currentOp = nil
	}

	entry := storeStackEntry{op: op}
	if p.timingEnabled {
		entry.startTime = time.Now()
	}
	p.storeStack = append(p.storeStack, entry)
}

// EndStore stops timing for the current store operation and records the measurement.
func (p *Profiler) EndStore(size int) {
	if len(p.storeStack) == 0 {
		panic("benchops: EndStore: no matching BeginStore")
	}

	// Pop and record the store entry
	idx := len(p.storeStack) - 1
	entry := p.storeStack[idx]
	p.storeStack = p.storeStack[:idx]

	var dur time.Duration
	if p.timingEnabled {
		dur = time.Since(entry.startTime)
	}
	p.storeStats[entry.op].Record(size, dur)

	// Resume opcode timing when store stack empties
	if len(p.storeStack) == 0 && len(p.opStack) > 0 {
		idx := len(p.opStack) - 1
		restored := p.opStack[idx]
		p.opStack = p.opStack[:idx]

		if p.timingEnabled {
			restored.startTime = time.Now()
		}
		p.currentOp = &restored
	}
}

// ---- Native measurement methods

// TraceNative traces a native function call.
// Records timing by default (disable with WithoutTiming).
// Usage: defer profiler.TraceNative(benchops.NativeXxx)()
func (p *Profiler) TraceNative(op NativeOp) func() {
	entry := &nativeEntry{op: op}
	if p.timingEnabled {
		entry.startTime = time.Now()
	}
	p.currentNative = entry
	return p.endNative
}

// endNative stops timing for the current native function and records the measurement.
func (p *Profiler) endNative() {
	entry := p.currentNative
	if entry == nil {
		panic("benchops: endNative: no matching TraceNative")
	}
	p.currentNative = nil

	var dur time.Duration
	if p.timingEnabled {
		dur = time.Since(entry.startTime)
	}
	p.nativeStats[entry.op].Record(dur)
}

// ---- Sub-operation measurement methods

// BeginSubOp starts timing for a sub-operation within an opcode.
// Pass zero value SubOpContext{} if no context is needed, or use
// NewSubOpContext/NewSubOpContextWithVar/NewSubOpContextWithIndex constructors.
func (p *Profiler) BeginSubOp(op SubOp, ctx SubOpContext) {
	entry := &subOpStackEntry{op: op, ctx: ctx}
	if p.timingEnabled {
		entry.startTime = time.Now()
	}
	p.currentSubOp = entry
}

// EndSubOp stops timing for the current sub-operation and records the measurement.
//
// Unlike EndOp/EndStore/EndNative which panic without matching Begin calls,
// EndSubOp is tolerant and returns silently if no BeginSubOp was called.
// This design choice enables simpler conditional instrumentation patterns where
// BeginSubOp may be conditionally skipped (e.g., when profiling only certain
// sub-operation types), without requiring callers to track whether Begin was called.
func (p *Profiler) EndSubOp() {
	entry := p.currentSubOp
	if entry == nil {
		return
	}
	p.currentSubOp = nil

	var dur time.Duration
	if p.timingEnabled {
		dur = time.Since(entry.startTime)
	}
	p.subOpStats[entry.op].Record(dur)

	// Record per-variable stats if context was set
	if entry.ctx.File != "" && entry.ctx.Line > 0 {
		p.recordVarStat(entry.ctx, dur)
	}
}

// recordVarStat records count and optionally timing for a specific variable assignment.
// Pass zero duration when timing is disabled.
func (p *Profiler) recordVarStat(ctx SubOpContext, dur time.Duration) {
	key := ctx.File + ":" + strconv.Itoa(ctx.Line)
	if ctx.VarName != "" {
		key += ":" + ctx.VarName
	} else if ctx.Index >= 0 {
		key += ":#" + strconv.Itoa(ctx.Index)
	}

	stat := p.varStats[key]
	if stat == nil {
		stat = &VarStat{
			Name:  ctx.VarName,
			File:  ctx.File,
			Line:  ctx.Line,
			Index: ctx.Index,
		}
		p.varStats[key] = stat
	}
	stat.Record(dur)
}

// ---- Call stack tracking methods

// PushCall pushes a function call onto the call stack.
func (p *Profiler) PushCall(funcName, pkgPath, file string, line int) {
	if !p.stackEnabled {
		return
	}
	p.callStack = append(p.callStack, callFrame{
		funcName: funcName,
		pkgPath:  pkgPath,
		file:     file,
		line:     line,
	})
}

// PopCall pops the current function from the call stack.
func (p *Profiler) PopCall() {
	if !p.stackEnabled || len(p.callStack) == 0 {
		return
	}
	p.callStack = p.callStack[:len(p.callStack)-1]
}

// recordStackSample aggregates a gas and timing sample by the current call stack signature.
// Uses in-place aggregation to avoid memory growth from storing individual samples.
func (p *Profiler) recordStackSample(gas int64, dur time.Duration) {
	if len(p.callStack) == 0 {
		return
	}

	// Build stack key in reverse order (leaf-to-root for pprof)
	var keyBuilder strings.Builder
	keyBuilder.Grow(len(p.callStack) * estimatedFrameKeySize)
	for i := len(p.callStack) - 1; i >= 0; i-- {
		frame := p.callStack[i]
		if i < len(p.callStack)-1 {
			keyBuilder.WriteByte('|')
		}
		keyBuilder.WriteString(frame.funcName)
		keyBuilder.WriteByte('@')
		keyBuilder.WriteString(frame.file)
		keyBuilder.WriteByte(':')
		keyBuilder.WriteString(strconv.Itoa(frame.line))
	}

	key := keyBuilder.String()
	sample := p.stackSampleAgg[key]
	if sample == nil {
		sample = &stackSample{}
		p.stackSampleAgg[key] = sample
	}
	sample.gas += gas
	sample.durationNs += dur.Nanoseconds()
	sample.count++
}
