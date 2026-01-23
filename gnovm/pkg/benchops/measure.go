package benchops

import "time"

// BeginOp starts timing for an opcode.
// Panics if profiler is not running (Start() must be called first).
func (p *Profiler) BeginOp(op Op) {
	p.currentOp = &opStackEntry{
		op:        op,
		startTime: time.Now(),
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
}

// BeginStore starts timing for a store operation.
// Automatically pauses the current opcode timing on the first nested call.
// Panics if profiler is not running.
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
// Panics if profiler is not running.
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
