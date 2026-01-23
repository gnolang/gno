package benchops

import (
	"sync"
	"time"
)

// State represents the profiler's current state.
type State int

const (
	StateIdle    State = iota // Not started
	StateRunning              // Actively profiling
	StateStopped              // Stopped, results available
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// Profiler collects timing statistics for GnoVM operations.
type Profiler struct {
	mu        sync.Mutex
	config    Config
	state     State
	startTime time.Time
	stopTime  time.Time

	// Op statistics
	opStats   [256]opStat
	opStack   []opStackEntry
	currentOp *opStackEntry

	// Store statistics
	storeStats [256]storeStat
	storeStack []storeStackEntry

	// Native statistics
	nativeStats   [256]nativeStat
	currentNative *nativeEntry
}

// New creates a new Profiler with the given configuration.
func New(cfg Config) *Profiler {
	return &Profiler{
		config:     cfg,
		state:      StateIdle,
		opStack:    make([]opStackEntry, 0, 16),
		storeStack: make([]storeStackEntry, 0, 16),
	}
}

// Start begins profiling. Panics if not in StateIdle.
func (p *Profiler) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != StateIdle {
		panic("benchops: Start called on non-idle profiler")
	}

	p.state = StateRunning
	p.startTime = time.Now()
}

// Stop ends profiling and returns the results. Panics if not in StateRunning.
func (p *Profiler) Stop() *Results {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != StateRunning {
		panic("benchops: Stop called on non-running profiler")
	}

	p.stopTime = time.Now()
	p.state = StateStopped

	return p.buildResults()
}

// Reset clears all collected data and returns to idle state.
func (p *Profiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.opStats = [256]opStat{}
	p.opStack = p.opStack[:0]
	p.currentOp = nil

	p.storeStats = [256]storeStat{}
	p.storeStack = p.storeStack[:0]

	p.nativeStats = [256]nativeStat{}
	p.currentNative = nil

	p.state = StateIdle
}

// Recovery resets internal state after a panic without changing profiler state.
// Call this from a recover block to ensure the profiler can continue.
func (p *Profiler) Recovery() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.opStack = p.opStack[:0]
	p.currentOp = nil
	p.storeStack = p.storeStack[:0]
	p.currentNative = nil
}

// State returns the current profiler state.
func (p *Profiler) State() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// Config returns the profiler's configuration.
func (p *Profiler) Config() Config {
	return p.config
}
