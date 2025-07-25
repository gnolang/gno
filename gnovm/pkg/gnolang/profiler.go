package gnolang

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

// ProfileType represents the type of profiling data
type ProfileType int

const (
	ProfileCPU ProfileType = iota
	ProfileMemory
	ProfileGoroutine
)

// ProfileSample represents a single sample in the profile
type ProfileSample struct {
	Location   []ProfileLocation
	Value      []int64
	Label      map[string][]string
	NumLabel   map[string][]int64
	SampleType ProfileType
}

// ProfileLocation represents a location in the call stack
type ProfileLocation struct {
	Function   string
	File       string
	Line       int
	InlineCall bool
}

// Profile represents collected profiling data
type Profile struct {
	Type          ProfileType
	TimeNanos     int64
	DurationNanos int64
	Samples       []ProfileSample

	// CPU specific
	CPUHz int64

	// Memory specific
	DefaultSampleType string

	mu sync.RWMutex
}

// Profiler manages profiling data collection
type Profiler struct {
	enabled    bool
	profile    *Profile
	startTime  time.Time
	sampleRate int // sample every N operations
	opCount    int

	// Function profiling
	funcProfiles map[string]*FuncProfile
	callStack    []string

	// Stack samples for call tree
	stackSamples []stackSample

	mu sync.Mutex
}

// stackSample represents a single stack trace sample
type stackSample struct {
	stack  []string
	cycles int64
}

// FuncProfile represents profiling data for a single function
type FuncProfile struct {
	Name         string
	CallCount    int64
	TotalCycles  int64 // Cumulative: includes time in called functions
	SelfCycles   int64 // Flat: only time spent in this function
	TotalTime    time.Duration
	SelfTime     time.Duration
	AllocBytes   int64
	AllocObjects int64
	Children     map[string]*FuncProfile

	// Track entry cycles for calculating self time
	entryCycles int64
}

// NewProfiler creates a new profiler instance
func NewProfiler(profileType ProfileType, sampleRate int) *Profiler {
	return &Profiler{
		enabled:      false,
		sampleRate:   sampleRate,
		funcProfiles: make(map[string]*FuncProfile),
		profile: &Profile{
			Type:      profileType,
			TimeNanos: time.Now().UnixNano(),
			Samples:   make([]ProfileSample, 0),
		},
	}
}

// Start begins profiling
func (p *Profiler) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.enabled = true
	p.startTime = time.Now()
	p.opCount = 0
}

// Stop ends profiling and returns the profile
func (p *Profiler) Stop() *Profile {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.enabled = false
	p.profile.DurationNanos = time.Since(p.startTime).Nanoseconds()

	// Convert funcProfiles to samples
	p.generateSamples()

	return p.profile
}

// RecordOp records an operation execution (called from Machine.Run)
func (p *Profiler) RecordOp(m *Machine, op Op, cycles int64) {
	if !p.enabled {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.opCount++
	if p.opCount%p.sampleRate != 0 {
		return
	}

	// Build call stack from machine frames
	stack := p.buildCallStack(m)
	if len(stack) == 0 {
		return
	}

	// Update function profiles
	funcName := stack[0].Function
	prof, ok := p.funcProfiles[funcName]
	if !ok {
		prof = &FuncProfile{
			Name:     funcName,
			Children: make(map[string]*FuncProfile),
		}
		p.funcProfiles[funcName] = prof
	}

	prof.CallCount++
	prof.TotalCycles += cycles
}

// RecordFuncEnter records function entry
func (p *Profiler) RecordFuncEnter(m *Machine, funcName string) {
	if !p.enabled {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.callStack = append(p.callStack, funcName)

	prof, ok := p.funcProfiles[funcName]
	if !ok {
		prof = &FuncProfile{
			Name:     funcName,
			Children: make(map[string]*FuncProfile),
		}
		p.funcProfiles[funcName] = prof
	}
	prof.CallCount++
	prof.entryCycles = m.Cycles // Store entry cycles

	// Update parent-child relationships
	if len(p.callStack) > 1 {
		parentName := p.callStack[len(p.callStack)-2]
		if parentProf, ok := p.funcProfiles[parentName]; ok {
			if _, exists := parentProf.Children[funcName]; !exists {
				parentProf.Children[funcName] = prof
			}
		}
	}

	// Debug: print function entry
	// fmt.Printf("PROFILE: Enter %s (count: %d)\n", funcName, prof.CallCount)
}

// RecordFuncExit records function exit
func (p *Profiler) RecordFuncExit(m *Machine, funcName string, cycles int64) {
	if !p.enabled {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Record stack sample
	if len(p.callStack) > 0 && p.opCount%p.sampleRate == 0 {
		// Make a copy of the current call stack
		stackCopy := make([]string, len(p.callStack))
		copy(stackCopy, p.callStack)

		p.stackSamples = append(p.stackSamples, stackSample{
			stack:  stackCopy,
			cycles: cycles,
		})
	}

	if len(p.callStack) > 0 {
		p.callStack = p.callStack[:len(p.callStack)-1]
	}

	if prof, ok := p.funcProfiles[funcName]; ok {
		// Calculate self cycles (flat time)
		selfCycles := m.Cycles - prof.entryCycles
		prof.SelfCycles += selfCycles
		prof.TotalCycles += cycles // This should be the total including sub-calls
	}
}

// RecordAlloc records memory allocation
func (p *Profiler) RecordAlloc(m *Machine, size int64, count int64, allocType string) {
	if !p.enabled || p.profile.Type != ProfileMemory {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Sampling for memory allocations
	p.opCount++
	if p.opCount%p.sampleRate != 0 {
		return
	}

	stack := p.buildCallStack(m)
	if len(stack) == 0 {
		// If no call stack, create a synthetic entry
		stack = []ProfileLocation{{
			Function: "<allocation>",
		}}
	}

	// Create memory allocation sample
	sample := ProfileSample{
		Location:   stack,
		Value:      []int64{count, size},
		Label:      make(map[string][]string),
		NumLabel:   make(map[string][]int64),
		SampleType: ProfileMemory,
	}

	// Add type information if available
	if allocType != "" {
		sample.Label["type"] = []string{allocType}
	}

	sample.NumLabel["allocations"] = []int64{count}
	sample.NumLabel["bytes"] = []int64{size}

	p.profile.Samples = append(p.profile.Samples, sample)

	// Update function profile stats
	funcName := stack[0].Function
	prof, ok := p.funcProfiles[funcName]
	if !ok {
		prof = &FuncProfile{
			Name:     funcName,
			Children: make(map[string]*FuncProfile),
		}
		p.funcProfiles[funcName] = prof
	}
	prof.AllocBytes += size
	prof.AllocObjects += count
}

// buildCallStack builds a call stack from machine frames
func (p *Profiler) buildCallStack(m *Machine) []ProfileLocation {
	var stack []ProfileLocation

	if m == nil {
		return stack
	}

	for i := len(m.Frames) - 1; i >= 0; i-- {
		frame := &m.Frames[i]
		if !frame.IsCall() {
			continue
		}

		loc := ProfileLocation{
			Function: string(frame.Func.Name),
			File:     frame.Func.FileName,
			Line:     0, // Line number not available from FuncValue
		}

		// Add package path to function name
		if frame.Func.PkgPath != "" {
			loc.Function = frame.Func.PkgPath + "." + loc.Function
		}

		stack = append(stack, loc)
	}

	return stack
}

// generateSamples converts function profiles to profile samples
func (p *Profiler) generateSamples() {
	// First, add stack samples for call tree visualization
	for _, stackSample := range p.stackSamples {
		if len(stackSample.stack) == 0 {
			continue
		}

		// Build locations from stack (reverse order for proper hierarchy)
		locations := make([]ProfileLocation, 0, len(stackSample.stack))
		for i := len(stackSample.stack) - 1; i >= 0; i-- {
			locations = append(locations, ProfileLocation{
				Function: stackSample.stack[i],
			})
		}

		sample := ProfileSample{
			Location:   locations,
			Value:      []int64{1, stackSample.cycles}, // 1 sample, N cycles
			Label:      make(map[string][]string),
			NumLabel:   make(map[string][]int64),
			SampleType: p.profile.Type,
		}

		p.profile.Samples = append(p.profile.Samples, sample)
	}

	// Then add individual function summaries
	for _, prof := range p.funcProfiles {
		sample := ProfileSample{
			Location: []ProfileLocation{{
				Function: prof.Name,
			}},
			Value: []int64{prof.CallCount, prof.TotalCycles},
			Label: make(map[string][]string),
			NumLabel: map[string][]int64{
				"calls":       {prof.CallCount},
				"cycles":      {prof.TotalCycles},
				"flat_cycles": {prof.SelfCycles},
				"cum_cycles":  {prof.TotalCycles},
			},
			SampleType: p.profile.Type,
		}

		if p.profile.Type == ProfileMemory {
			sample.NumLabel["bytes"] = []int64{prof.AllocBytes}
			sample.NumLabel["objects"] = []int64{prof.AllocObjects}
		}

		p.profile.Samples = append(p.profile.Samples, sample)
	}
}

// WriteTo writes the profile in a human-readable format
func (p *Profile) WriteTo(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	fmt.Fprintf(w, "Profile Type: %s\n", p.typeString())
	fmt.Fprintf(w, "Duration: %s\n", time.Duration(p.DurationNanos))
	fmt.Fprintf(w, "Samples: %d\n\n", len(p.Samples))

	// Sort samples by total cycles/value
	sort.Slice(p.Samples, func(i, j int) bool {
		if len(p.Samples[i].Value) > 1 && len(p.Samples[j].Value) > 1 {
			return p.Samples[i].Value[1] > p.Samples[j].Value[1]
		}
		return false
	})

	// Print top functions
	fmt.Fprintf(w, "Top Functions:\n")
	fmt.Fprintf(w, "%-50s %12s %12s %12s %12s\n", "Function", "Flat", "Flat%", "Cum", "Cum%")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 100))

	totalCycles := p.totalCycles()

	for i, sample := range p.Samples {
		if i >= 20 { // Show top 20
			break
		}

		funcName := "unknown"
		if len(sample.Location) > 0 {
			funcName = sample.Location[0].Function
			if len(funcName) > 50 {
				funcName = funcName[:47] + "..."
			}
		}

		flatCycles := int64(0)
		cumCycles := int64(0)

		if flatVal, ok := sample.NumLabel["flat_cycles"]; ok && len(flatVal) > 0 {
			flatCycles = flatVal[0]
		} else if cyclesVal, ok := sample.NumLabel["cycles"]; ok && len(cyclesVal) > 0 {
			flatCycles = cyclesVal[0]
		}

		if cumVal, ok := sample.NumLabel["cum_cycles"]; ok && len(cumVal) > 0 {
			cumCycles = cumVal[0]
		} else if cyclesVal, ok := sample.NumLabel["cycles"]; ok && len(cyclesVal) > 0 {
			cumCycles = cyclesVal[0]
		}

		flatPercent := float64(0)
		cumPercent := float64(0)
		if totalCycles > 0 {
			flatPercent = float64(flatCycles) / float64(totalCycles) * 100
			cumPercent = float64(cumCycles) / float64(totalCycles) * 100
		}

		fmt.Fprintf(w, "%-50s %12d %11.2f%% %12d %11.2f%%\n",
			funcName, flatCycles, flatPercent, cumCycles, cumPercent)
	}

	return nil
}

func (p *Profile) typeString() string {
	switch p.Type {
	case ProfileCPU:
		return "CPU"
	case ProfileMemory:
		return "Memory"
	// TODO: gno does not support goroutine for now.
	case ProfileGoroutine:
		return "Goroutine"
	default:
		return "Unknown"
	}
}

// MachineProfiler extension for Machine
func (m *Machine) StartProfiling(profileType ProfileType, sampleRate int) {
	if m.Profiler == nil {
		m.Profiler = NewProfiler(profileType, sampleRate)
	}
	m.Profiler.Start()
}

func (m *Machine) StopProfiling() *Profile {
	if m.Profiler == nil {
		return nil
	}
	return m.Profiler.Stop()
}

func (m *Machine) IsProfilingEnabled() bool {
	return m.Profiler != nil && m.Profiler.enabled
}
