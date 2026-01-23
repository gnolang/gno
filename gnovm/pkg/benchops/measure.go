package benchops

import "time"

// BeginOp starts timing for an opcode.
func (p *Profiler) BeginOp(op Op) {
	if !p.config.EnableOps {
		return
	}

	p.currentOp = &opStackEntry{
		op:        op,
		startTime: time.Now(),
	}
}

// EndOp stops timing for the current opcode and records the measurement.
func (p *Profiler) EndOp() {
	if p.currentOp == nil {
		return
	}

	entry := p.currentOp
	p.currentOp = nil

	dur := entry.elapsed + time.Since(entry.startTime)
	p.opStats[entry.op].record(dur)
}

// BeginStore starts timing for a store operation.
// Automatically pauses the current opcode timing on the first nested call.
func (p *Profiler) BeginStore(op StoreOp) {
	if !p.config.EnableStore {
		return
	}

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
func (p *Profiler) EndStore(size int) {
	if len(p.storeStack) == 0 {
		return
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
func (p *Profiler) BeginNative(op NativeOp) {
	if !p.config.EnableNative {
		return
	}

	p.currentNative = &nativeEntry{
		op:        op,
		startTime: time.Now(),
	}
}

// EndNative stops timing for the current native function and records the measurement.
func (p *Profiler) EndNative() {
	if p.currentNative == nil {
		return
	}

	entry := p.currentNative
	p.currentNative = nil

	dur := time.Since(entry.startTime)
	p.nativeStats[entry.op].record(dur)
}
