package benchops

import (
	"strconv"
	"time"
)

// ---- Op measurement methods

// BeginOp starts timing for an opcode.
func (p *Profiler) BeginOp(op Op) {
	entry := &opStackEntry{op: op}
	if p.timingEnabled {
		entry.startTime = time.Now()
	}
	p.currentOp = entry
}

// SetOpContext sets the source location context for the current opcode.
func (p *Profiler) SetOpContext(ctx OpContext) {
	if p.currentOp != nil {
		p.currentOp.ctx = ctx
	}
}

// EndOp stops timing for the current opcode and records the measurement.
func (p *Profiler) EndOp() {
	entry := p.currentOp
	if entry == nil {
		panic("benchops: EndOp called without matching BeginOp")
	}
	p.currentOp = nil

	gas := GetOpGas(entry.op)

	var dur time.Duration
	if p.timingEnabled {
		dur = entry.elapsed + time.Since(entry.startTime)
		p.opStatsTimed[entry.op].recordTimed(gas, dur)
	} else {
		p.opStats[entry.op].record(gas)
	}

	// Record location stats if context was set
	if entry.ctx.File != "" && entry.ctx.Line > 0 {
		p.recordLocation(entry.op, entry.ctx, dur)
	}
}

// recordLocation records gas and optionally timing for a specific source location.
// Pass zero duration when timing is disabled.
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
		panic("benchops: EndStore called without matching BeginStore")
	}

	// Pop and record the store entry
	idx := len(p.storeStack) - 1
	entry := p.storeStack[idx]
	p.storeStack = p.storeStack[:idx]

	if p.timingEnabled {
		dur := time.Since(entry.startTime)
		p.storeStatsTimed[entry.op].recordTimed(size, dur)
	} else {
		p.storeStats[entry.op].record(size)
	}

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

// BeginNative starts timing for a native function.
func (p *Profiler) BeginNative(op NativeOp) {
	entry := &nativeEntry{op: op}
	if p.timingEnabled {
		entry.startTime = time.Now()
	}
	p.currentNative = entry
}

// EndNative stops timing for the current native function and records the measurement.
func (p *Profiler) EndNative() {
	entry := p.currentNative
	if entry == nil {
		panic("benchops: EndNative called without matching BeginNative")
	}
	p.currentNative = nil

	if p.timingEnabled {
		dur := time.Since(entry.startTime)
		p.nativeStatsTimed[entry.op].recordTimed(dur)
	} else {
		p.nativeStats[entry.op].record()
	}
}
