package benchops

import (
	"strconv"
	"time"
)

// BeginOp starts timing for an opcode.
// If the profiler is not running, measurements are still recorded but will be
// cleared on the next Start() call. No state check is performed for performance.
func (p *Profiler) BeginOp(op Op) {
	p.currentOp = &opStackEntry{
		op:        op,
		startTime: time.Now(),
	}
}

// SetOpContext sets the source location context for the current opcode.
// Must be called after BeginOp and before EndOp.
func (p *Profiler) SetOpContext(ctx OpContext) {
	if p.currentOp != nil {
		p.currentOp.ctx = ctx
	}
}

// EndOp stops timing for the current opcode and records the measurement.
// Panics if called without a matching BeginOp.
func (p *Profiler) EndOp() {
	entry := p.currentOp
	if entry == nil {
		panic("benchops: EndOp called without matching BeginOp")
	}
	p.currentOp = nil

	dur := entry.elapsed + time.Since(entry.startTime)
	p.opStats[entry.op].record(dur)

	// Record location stats if context was set
	if entry.ctx.File != "" && entry.ctx.Line > 0 {
		p.recordLocation(entry.op, entry.ctx, dur)
	}
}

// recordLocation records timing for a specific source location.
func (p *Profiler) recordLocation(op Op, ctx OpContext, dur time.Duration) {
	key := ctx.File + ":" + strconv.Itoa(ctx.Line)
	stat := p.locationStats[key]
	if stat == nil {
		stat = &locationStat{
			file:     ctx.File,
			line:     ctx.Line,
			funcName: ctx.FuncName,
			pkgPath:  ctx.PkgPath,
		}
		p.locationStats[key] = stat
	}
	stat.count++
	stat.totalDur += dur
	stat.gasTotal += GetOpGas(op)
}

// BeginStore starts timing for a store operation.
// Automatically pauses the current opcode timing on the first nested call.
// If the profiler is not running, measurements are recorded but cleared on next Start().
func (p *Profiler) BeginStore(op StoreOp) {
	// Pause current opcode timing on first store call
	if len(p.storeStack) == 0 && p.currentOp != nil {
		p.currentOp.elapsed += time.Since(p.currentOp.startTime)
		p.opStack = append(p.opStack, *p.currentOp)
		p.currentOp = nil
	}

	p.storeStack = append(p.storeStack, storeStackEntry{
		op:        op,
		startTime: time.Now(),
	})
}

// EndStore stops timing for the current store operation and records the measurement.
// Automatically resumes opcode timing when the store stack empties.
// Panics if called without a matching BeginStore.
func (p *Profiler) EndStore(size int) {
	if len(p.storeStack) == 0 {
		panic("benchops: EndStore called without matching BeginStore")
	}

	// Pop and record the store entry
	idx := len(p.storeStack) - 1
	entry := p.storeStack[idx]
	p.storeStack = p.storeStack[:idx]

	dur := time.Since(entry.startTime)
	p.storeStats[entry.op].record(dur, size)

	// Resume opcode timing when store stack empties
	if len(p.storeStack) == 0 && len(p.opStack) > 0 {
		idx := len(p.opStack) - 1
		restored := p.opStack[idx]
		p.opStack = p.opStack[:idx]

		restored.startTime = time.Now()
		p.currentOp = &restored
	}
}

// BeginNative starts timing for a native function.
// If the profiler is not running, measurements are recorded but cleared on next Start().
func (p *Profiler) BeginNative(op NativeOp) {
	p.currentNative = &nativeEntry{
		op:        op,
		startTime: time.Now(),
	}
}

// EndNative stops timing for the current native function and records the measurement.
// Panics if called without a matching BeginNative.
func (p *Profiler) EndNative() {
	entry := p.currentNative
	if entry == nil {
		panic("benchops: EndNative called without matching BeginNative")
	}
	p.currentNative = nil

	dur := time.Since(entry.startTime)
	p.nativeStats[entry.op].record(dur)
}
