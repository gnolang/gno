package profiler

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

// ProfileType represents the type of profiling data that can be collected.
type ProfileType int

const (
	_ ProfileType = iota

	// ProfileCPU tracks CPU cycle consumption during program execution.
	// This is the default profiling mode and measures computational cost.
	ProfileCPU

	// ProfileMemory tracks memory allocations and usage patterns.
	// It helps identify memory-intensive operations and potential leaks.
	ProfileMemory

	// ProfileGas tracks gas consumption for blockchain operations.
	//
	// Note: Although gas price itself appears to have a one-to-one correspondence
	// with the number of CPU cycles, this ratio may change in the future
	// and it seems better to distinguish it from CPU, so it's added as a separate type.
	ProfileGas

	// ProfileGoroutine tracks goroutine creation and lifecycle.
	// TODO: not supported yet
	ProfileGoroutine
)

// ProfileSample represents a single sample in the profile
type ProfileSample struct {
	Location   []ProfileLocation   `json:"location"`
	Value      []int64             `json:"value"`
	Label      map[string][]string `json:"label,omitempty"`
	NumLabel   map[string][]int64  `json:"numLabel,omitempty"`
	SampleType ProfileType         `json:"sampleType"`
	GasUsed    int64               `json:"gasUsed,omitempty"` // Gas consumed at this sample
}

// ProfileLocation represents a location in the call stack
type ProfileLocation struct {
	Function   string  `json:"function"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	Column     int     `json:"column,omitempty"` // Column number for more precise location
	InlineCall bool    `json:"inlineCall,omitempty"`
	PC         uintptr `json:"pc,omitempty"` // Virtual machine program counter
}

// Profile represents collected profiling data
type Profile struct {
	Type          ProfileType     `json:"type"`
	TimeNanos     int64           `json:"timeNanos"`
	DurationNanos int64           `json:"durationNanos"`
	Samples       []ProfileSample `json:"samples"`

	// CPU specific
	CPUHz int64 `json:"cpuHz,omitempty"`

	// Memory specific
	DefaultSampleType string `json:"defaultSampleType,omitempty"`

	mu sync.RWMutex `json:"-"`
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

	// Line-level profiling support
	lineLevel     bool
	locationCache *locationCache
	lineSamples   map[string]map[int]*lineStats // file -> line -> stats
	locationPool  sync.Pool

	// Test file filtering
	excludeTests bool // whether to exclude *_test.gno files from profiling

	mu sync.Mutex
}

// stackSample represents a single stack trace sample
type stackSample struct {
	stack   []string
	cycles  int64
	gasUsed int64
}

// FuncProfile represents profiling data for a single function
type FuncProfile struct {
	Name         string
	CallCount    int64
	TotalCycles  int64 // Cumulative: includes time in called functions
	SelfCycles   int64 // Flat: only time spent in this function
	TotalGas     int64 // Cumulative: includes gas in called functions
	SelfGas      int64 // Flat: only gas spent in this function
	TotalTime    time.Duration
	SelfTime     time.Duration
	AllocBytes   int64
	AllocObjects int64
	Children     map[string]*FuncProfile

	// Track entry cycles and gas for calculating self time/gas
	entryCycles int64
	entryGas    int64
}

// Options for starting profiling
type Options struct {
	Type       ProfileType
	SampleRate int
}

// NewProfiler creates a new profiler instance
// Optional parameters: profileType (default: ProfileCPU), sampleRate (default: 1000)
func NewProfiler(params ...interface{}) *Profiler {
	// Default values
	profileType := ProfileCPU
	sampleRate := 1000

	// Parse optional parameters
	if len(params) > 0 {
		if pt, ok := params[0].(ProfileType); ok {
			profileType = pt
		}
	}
	if len(params) > 1 {
		if sr, ok := params[1].(int); ok {
			sampleRate = sr
		}
	}

	p := &Profiler{
		funcProfiles:  make(map[string]*FuncProfile),
		lineSamples:   make(map[string]map[int]*lineStats),
		locationCache: newLocationCache(),
		sampleRate:    sampleRate,
		excludeTests:  true, // Exclude test files by default
		locationPool: sync.Pool{
			New: func() interface{} {
				return &profileLocation{}
			},
		},
	}

	p.profile = &Profile{
		Type:    profileType,
		Samples: make([]ProfileSample, 0),
	}

	return p
}

// StartProfiling starts profiling with the given options
func (p *Profiler) StartProfiling(m MachineInfo, opts Options) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.enabled = true
	p.startTime = time.Now()
	p.sampleRate = opts.SampleRate
	if p.sampleRate <= 0 {
		p.sampleRate = 1000 // default sample rate
	}

	p.profile = &Profile{
		Type:      opts.Type,
		TimeNanos: p.startTime.UnixNano(),
		Samples:   make([]ProfileSample, 0),
	}

	// Reset state
	p.opCount = 0
	p.stackSamples = nil
	p.funcProfiles = make(map[string]*FuncProfile)
	p.callStack = nil
}

// StopProfiling stops profiling and returns the collected profile
func (p *Profiler) StopProfiling() *Profile {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled || p.profile == nil {
		return nil
	}

	p.enabled = false
	p.profile.DurationNanos = time.Since(p.startTime).Nanoseconds()

	// Build profile samples from collected data
	p.generateSamples()

	result := p.profile
	p.profile = nil
	return result
}

// IsEnabled returns whether profiling is enabled
func (p *Profiler) IsEnabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.enabled
}

// RecordSample records a profiling sample if it's time
func (p *Profiler) RecordSample(m MachineInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled {
		return
	}

	p.opCount++
	if p.opCount%p.sampleRate != 0 {
		return
	}

	// Build call stack
	stack := p.buildCallStack(m)
	if len(stack) == 0 {
		return
	}

	// Record stack sample
	p.stackSamples = append(p.stackSamples, stackSample{
		stack:   p.callStackToStrings(stack),
		cycles:  m.GetCycles(),
		gasUsed: m.GetGasUsed(),
	})

	// Update function profiles
	p.updateFunctionProfiles(stack, m.GetCycles())

	// For gas profiling, also track gas in function profiles
	if p.profile.Type == ProfileGas {
		p.updateFunctionGasProfiles(stack, m.GetGasUsed())
	}

	// Update line-level profiling if enabled
	if p.lineLevel && len(stack) > 0 {
		loc := stack[0]
		// Skip test files if excludeTests is enabled
		if loc.File != "" && loc.Line > 0 && !(p.excludeTests && strings.HasSuffix(loc.File, "_test.gno")) {
			if p.lineSamples[loc.File] == nil {
				p.lineSamples[loc.File] = make(map[int]*lineStats)
			}
			if p.lineSamples[loc.File][loc.Line] == nil {
				p.lineSamples[loc.File][loc.Line] = &lineStats{}
			}
			p.lineSamples[loc.File][loc.Line].count++
			p.lineSamples[loc.File][loc.Line].cycles += m.GetCycles()
			// Track gas for gas profiling
			if p.profile.Type == ProfileGas {
				p.lineSamples[loc.File][loc.Line].gas += m.GetGasUsed()
			}
		}
	}
}

// RecordFuncEnter records function entry
func (p *Profiler) RecordFuncEnter(m MachineInfo, funcName string) {
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
	prof.entryCycles = m.GetCycles() // Store entry cycles
	prof.entryGas = m.GetGasUsed()   // Store entry gas

	// Update parent-child relationships
	if len(p.callStack) > 1 {
		parentName := p.callStack[len(p.callStack)-2]
		if parentProf, ok := p.funcProfiles[parentName]; ok {
			if _, exists := parentProf.Children[funcName]; !exists {
				parentProf.Children[funcName] = prof
			}
		}
	}
}

// RecordFuncExit records function exit
func (p *Profiler) RecordFuncExit(m MachineInfo, funcName string, cycles int64) {
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
			stack:   stackCopy,
			cycles:  cycles,
			gasUsed: m.GetGasUsed(),
		})
	}

	if len(p.callStack) > 0 {
		p.callStack = p.callStack[:len(p.callStack)-1]
	}

	if prof, ok := p.funcProfiles[funcName]; ok {
		// Calculate self cycles and gas (flat time/gas)
		selfCycles := m.GetCycles() - prof.entryCycles
		selfGas := m.GetGasUsed() - prof.entryGas
		prof.SelfCycles += selfCycles
		prof.SelfGas += selfGas
		prof.TotalCycles += cycles // This should be the total including sub-calls
		prof.TotalGas += selfGas   // Add the gas consumed in this function
	}
}

// RecordAlloc records memory allocation
func (p *Profiler) RecordAlloc(m MachineInfo, size int64, count int64, allocType string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled || p.profile.Type != ProfileMemory {
		return
	}

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
func (p *Profiler) buildCallStack(m MachineInfo) []ProfileLocation {
	var stack []ProfileLocation

	if m == nil {
		return stack
	}

	frames := m.GetFrames()
	for i := len(frames) - 1; i >= 0; i-- {
		frame := frames[i]
		if !frame.IsCall() {
			continue
		}

		// Skip test files if excludeTests is enabled
		fileName := frame.GetFileName()
		if p.excludeTests && strings.HasSuffix(fileName, "_test.gno") {
			continue
		}

		loc := ProfileLocation{
			Function: frame.GetFuncName(),
			File:     fileName,
			Line:     0, // Default line
			Column:   0,
		}

		// Try to get line information from frame source
		if src := frame.GetSource(); src != nil {
			loc.Line = src.GetLine()
			loc.Column = src.GetColumn()
		}

		// Add full file path if available
		if pkgPath := frame.GetPkgPath(); pkgPath != "" {
			loc.Function = pkgPath + "." + loc.Function
			if loc.File == "" {
				// Use package path as file hint
				loc.File = pkgPath
			}
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
			GasUsed:    stackSample.gasUsed,
		}

		p.profile.Samples = append(p.profile.Samples, sample)
	}

	// Then add individual function summaries
	for _, prof := range p.funcProfiles {
		// Skip test functions when generating samples
		if p.excludeTests && strings.Contains(prof.Name, "_test.") {
			continue
		}

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

		switch p.profile.Type {
		case ProfileMemory:
			sample.NumLabel["bytes"] = []int64{prof.AllocBytes}
			sample.NumLabel["objects"] = []int64{prof.AllocObjects}
		case ProfileGas:
			// For gas profiling, use gas values instead of cycles
			sample.Value = []int64{prof.CallCount, prof.TotalGas}
			sample.NumLabel["gas"] = []int64{prof.TotalGas}
			sample.NumLabel["flat_gas"] = []int64{prof.SelfGas}
			sample.NumLabel["cum_gas"] = []int64{prof.TotalGas}
			sample.GasUsed = prof.TotalGas
		}

		p.profile.Samples = append(p.profile.Samples, sample)
	}
}

// countingWriter counts bytes written
type countingWriter struct {
	w   io.Writer
	n   int64
	err error
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	if cw.err != nil {
		return 0, cw.err
	}
	n, err := cw.w.Write(p)
	cw.n += int64(n)
	if err != nil {
		cw.err = err
	}
	return n, err
}

// WriteTo writes the profile in a human-readable format
// Implements io.WriterTo interface
func (p *Profile) WriteTo(w io.Writer) (int64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Use a counting writer to track bytes written
	cw := &countingWriter{w: w}

	fmt.Fprintf(cw, "Profile Type: %s\n", p.typeString())
	fmt.Fprintf(cw, "Duration: %s\n", time.Duration(p.DurationNanos))
	fmt.Fprintf(cw, "Samples: %d\n\n", len(p.Samples))

	// Sort samples by total cycles/value
	sort.Slice(p.Samples, func(i, j int) bool {
		if len(p.Samples[i].Value) > 1 && len(p.Samples[j].Value) > 1 {
			return p.Samples[i].Value[1] > p.Samples[j].Value[1]
		}
		return false
	})

	// Print top functions
	fmt.Fprintf(cw, "Top Functions:\n")
	if p.Type == ProfileGas {
		fmt.Fprintf(cw, "%-50s %12s %12s %12s %12s\n", "Function", "Flat Gas", "Flat%", "Cum Gas", "Cum%")
	} else {
		fmt.Fprintf(cw, "%-50s %12s %12s %12s %12s\n", "Function", "Flat", "Flat%", "Cum", "Cum%")
	}
	fmt.Fprintf(cw, "%s\n", strings.Repeat("-", 100))

	var totalValue int64
	if p.Type == ProfileGas {
		totalValue = p.totalGas()
	} else {
		totalValue = p.totalCycles()
	}

	for i, sample := range p.Samples {
		if i >= 20 { // Show top 20
			break
		}

		funcName := "unknown"
		if len(sample.Location) > 0 {
			// Skip samples from test files
			if len(sample.Location) > 0 && strings.HasSuffix(sample.Location[0].File, "_test.gno") {
				continue
			}
			funcName = sample.Location[0].Function
			if len(funcName) > 50 {
				funcName = funcName[:47] + "..."
			}
		}

		var flatValue, cumValue int64

		if p.Type == ProfileGas {
			if flatVal, ok := sample.NumLabel["flat_gas"]; ok && len(flatVal) > 0 {
				flatValue = flatVal[0]
			} else if gasVal, ok := sample.NumLabel["gas"]; ok && len(gasVal) > 0 {
				flatValue = gasVal[0]
			}

			if cumVal, ok := sample.NumLabel["cum_gas"]; ok && len(cumVal) > 0 {
				cumValue = cumVal[0]
			} else if gasVal, ok := sample.NumLabel["gas"]; ok && len(gasVal) > 0 {
				cumValue = gasVal[0]
			}
		} else {
			if flatVal, ok := sample.NumLabel["flat_cycles"]; ok && len(flatVal) > 0 {
				flatValue = flatVal[0]
			} else if cyclesVal, ok := sample.NumLabel["cycles"]; ok && len(cyclesVal) > 0 {
				flatValue = cyclesVal[0]
			}

			if cumVal, ok := sample.NumLabel["cum_cycles"]; ok && len(cumVal) > 0 {
				cumValue = cumVal[0]
			} else if cyclesVal, ok := sample.NumLabel["cycles"]; ok && len(cyclesVal) > 0 {
				cumValue = cyclesVal[0]
			}
		}

		flatPercent := float64(0)
		cumPercent := float64(0)
		if totalValue > 0 {
			flatPercent = float64(flatValue) / float64(totalValue) * 100
			cumPercent = float64(cumValue) / float64(totalValue) * 100
		}

		fmt.Fprintf(cw, "%-50s %12d %11.2f%% %12d %11.2f%%\n",
			funcName, flatValue, flatPercent, cumValue, cumPercent)
	}

	return cw.n, cw.err
}

// WriteJSON writes the profile in JSON format
func (p *Profile) WriteJSON(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(p)
}

func (p *Profile) typeString() string {
	switch p.Type {
	case ProfileCPU:
		return "CPU"
	case ProfileMemory:
		return "Memory"
	case ProfileGas:
		return "Gas"
	// TODO: gno does not support goroutine for now.
	case ProfileGoroutine:
		return "Goroutine"
	default:
		return "Unknown"
	}
}

// totalCycles calculates total cycles across all samples
func (p *Profile) totalCycles() int64 {
	total := int64(0)
	seen := make(map[string]bool)

	for _, sample := range p.Samples {
		if len(sample.Location) > 0 {
			funcName := sample.Location[0].Function
			if !seen[funcName] {
				seen[funcName] = true
				if cumVal, ok := sample.NumLabel["cum_cycles"]; ok && len(cumVal) > 0 {
					// For top-level functions only
					if len(sample.Location) == 1 {
						total += cumVal[0]
					}
				} else if len(sample.Value) > 1 && len(sample.Location) == 1 {
					total += sample.Value[1]
				}
			}
		}
	}

	return total
}

// totalGas calculates total gas across all samples
func (p *Profile) totalGas() int64 {
	total := int64(0)
	seen := make(map[string]bool)

	for _, sample := range p.Samples {
		if len(sample.Location) > 0 {
			funcName := sample.Location[0].Function
			if !seen[funcName] {
				seen[funcName] = true
				if cumVal, ok := sample.NumLabel["cum_gas"]; ok && len(cumVal) > 0 {
					// For top-level functions only
					if len(sample.Location) == 1 {
						total += cumVal[0]
					}
				} else if sample.GasUsed > 0 && len(sample.Location) == 1 {
					total += sample.GasUsed
				}
			}
		}
	}

	return total
}

// GetProfile returns the current profile without stopping profiling
func (p *Profiler) GetProfile() *Profile {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled || p.profile == nil {
		return nil
	}

	// Create a copy of the current profile
	profileCopy := &Profile{
		Type:          p.profile.Type,
		TimeNanos:     p.profile.TimeNanos,
		DurationNanos: time.Since(p.startTime).Nanoseconds(),
		Samples:       make([]ProfileSample, 0),
		CPUHz:         p.profile.CPUHz,
	}

	// Build samples from current data
	savedProfile := p.profile
	p.profile = profileCopy
	p.generateSamples()
	p.profile = savedProfile

	return profileCopy
}

// EnableLineProfiling enables line-level profiling
func (p *Profiler) EnableLineProfiling() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lineLevel = true
}

// DisableLineProfiling disables line-level profiling
func (p *Profiler) DisableLineProfiling() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lineLevel = false
}

// IsLineProfilingEnabled returns whether line-level profiling is enabled
func (p *Profiler) IsLineProfilingEnabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lineLevel
}

// RecordLineSample records a line-level profiling sample
// This method is called from the VM's main execution loop when line-level profiling is enabled
// It tracks the exact source location and cycles spent at each line
func (p *Profiler) RecordLineSample(funcName, file string, line int, cycles int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled || !p.lineLevel {
		return
	}

	// Skip test files if excludeTests is enabled
	if p.excludeTests && strings.HasSuffix(file, "_test.gno") {
		return
	}

	// Create or update line statistics
	if p.lineSamples[file] == nil {
		p.lineSamples[file] = make(map[int]*lineStats)
	}
	if p.lineSamples[file][line] == nil {
		p.lineSamples[file][line] = &lineStats{}
	}
	p.lineSamples[file][line].count++
	p.lineSamples[file][line].cycles += cycles

	// Also record as a sample for function matching
	// This allows the WriteFunctionList method to find and display these line-level samples
	sample := ProfileSample{
		Location: []ProfileLocation{{
			Function: funcName,
			File:     file,
			Line:     line,
		}},
		Value: []int64{1, cycles},
	}
	p.profile.Samples = append(p.profile.Samples, sample)
}

// Start starts profiling
func (p *Profiler) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.enabled = true
	p.startTime = time.Now()

	if p.profile != nil {
		// Update start time if profile already exists
		p.profile.TimeNanos = p.startTime.UnixNano()
	}

	// Reset state
	p.opCount = 0
	p.stackSamples = nil
	// Don't reset funcProfiles and lineSamples to preserve data
	p.callStack = nil
}

// Stop stops profiling and returns the collected profile
func (p *Profiler) Stop() *Profile {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled || p.profile == nil {
		return nil
	}

	p.enabled = false
	p.profile.DurationNanos = time.Since(p.startTime).Nanoseconds()

	// Build profile samples from collected data
	p.generateSamples()

	result := p.profile
	// Don't reset p.profile to nil here - keep it for GetLineStats
	return result
}

// callStackToStrings converts a call stack to string slice
func (p *Profiler) callStackToStrings(stack []ProfileLocation) []string {
	result := make([]string, len(stack))
	for i, loc := range stack {
		result[i] = loc.Function
	}
	return result
}

// updateFunctionProfiles updates function profiling data
func (p *Profiler) updateFunctionProfiles(stack []ProfileLocation, cycles int64) {
	seen := make(map[string]bool)

	for i, loc := range stack {
		funcName := loc.Function
		if seen[funcName] {
			continue
		}
		seen[funcName] = true

		prof := p.funcProfiles[funcName]
		if prof == nil {
			prof = &FuncProfile{
				Name:     funcName,
				Children: make(map[string]*FuncProfile),
			}
			p.funcProfiles[funcName] = prof
		}

		// Update cumulative cycles (appears anywhere in stack)
		prof.TotalCycles += cycles
		prof.CallCount++

		// Update self cycles (only for the top of stack)
		if i == 0 {
			prof.SelfCycles += cycles
		}
	}
}

// updateFunctionGasProfiles updates function gas profiling data
func (p *Profiler) updateFunctionGasProfiles(stack []ProfileLocation, gasUsed int64) {
	seen := make(map[string]bool)

	for i, loc := range stack {
		funcName := loc.Function
		if seen[funcName] {
			continue
		}
		seen[funcName] = true

		prof := p.funcProfiles[funcName]
		if prof == nil {
			prof = &FuncProfile{
				Name:     funcName,
				Children: make(map[string]*FuncProfile),
			}
			p.funcProfiles[funcName] = prof
		}

		// Update cumulative gas (appears anywhere in stack)
		prof.TotalGas += gasUsed

		// Update self gas (only for the top of stack)
		if i == 0 {
			prof.SelfGas += gasUsed
		}
	}
}
