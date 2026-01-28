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

// FrameID uniquely identifies a deduplicated call frame.
type FrameID int

const invalidFrameID FrameID = -1

// frameKey is used to intern logical call frames.
type frameKey struct {
	function   string
	file       string
	line       int
	column     int
	inlineCall bool
	pc         uintptr
}

// frameStore interns ProfileLocation entries so call stacks can be represented
// by compact FrameIDs.
type frameStore struct {
	index  map[frameKey]FrameID
	frames []ProfileLocation
}

func newFrameStore() frameStore {
	return frameStore{
		index:  make(map[frameKey]FrameID),
		frames: make([]ProfileLocation, 0, 128),
	}
}

func (fs *frameStore) intern(loc ProfileLocation) FrameID {
	key := frameKey{
		function:   loc.Function,
		file:       loc.File,
		line:       loc.Line,
		column:     loc.Column,
		inlineCall: loc.InlineCall,
		pc:         loc.PC,
	}
	if id, ok := fs.index[key]; ok {
		return id
	}
	id := FrameID(len(fs.frames))
	fs.frames = append(fs.frames, loc)
	fs.index[key] = id
	return id
}

func (fs *frameStore) reset() {
	fs.index = make(map[frameKey]FrameID)
	fs.frames = fs.frames[:0]
}

func (fs *frameStore) framesSnapshot() []ProfileLocation {
	out := make([]ProfileLocation, len(fs.frames))
	copy(out, fs.frames)
	return out
}

// callTreeNode accumulates aggregated stack information.
type callTreeNode struct {
	frameID FrameID

	calls int64

	totalCycles int64
	selfCycles  int64

	totalGas int64
	selfGas  int64

	allocBytes   int64
	allocObjects int64

	children map[FrameID]*callTreeNode
}

func newCallTreeNode(frameID FrameID) *callTreeNode {
	return &callTreeNode{
		frameID:  frameID,
		children: make(map[FrameID]*callTreeNode),
	}
}

// FunctionStat represents aggregated information for a single function.
type FunctionStat struct {
	Name         string `json:"name"`
	CallCount    int64  `json:"callCount"`
	TotalCycles  int64  `json:"totalCycles,omitempty"`
	SelfCycles   int64  `json:"selfCycles,omitempty"`
	TotalGas     int64  `json:"totalGas,omitempty"`
	SelfGas      int64  `json:"selfGas,omitempty"`
	AllocBytes   int64  `json:"allocBytes,omitempty"`
	AllocObjects int64  `json:"allocObjects,omitempty"`
}

// CallTreeNode is the exported representation of aggregated call stacks.
type CallTreeNode struct {
	FrameID     FrameID         `json:"frameId"`
	Calls       int64           `json:"calls"`
	TotalCycles int64           `json:"totalCycles,omitempty"`
	SelfCycles  int64           `json:"selfCycles,omitempty"`
	TotalGas    int64           `json:"totalGas,omitempty"`
	SelfGas     int64           `json:"selfGas,omitempty"`
	AllocBytes  int64           `json:"allocBytes,omitempty"`
	AllocObjs   int64           `json:"allocObjs,omitempty"`
	Children    []*CallTreeNode `json:"children,omitempty"`
}

// functionLineData tracks per-line stats for a function across files.
type functionLineData struct {
	funcName    string
	fileSamples map[string]map[int]*lineStat
	totalCycles int64
	totalGas    int64
	totalAllocs int64
	totalAllocB int64
}

func (fld *functionLineData) clone() *functionLineData {
	if fld == nil {
		return nil
	}
	clone := &functionLineData{
		funcName:    fld.funcName,
		totalCycles: fld.totalCycles,
		totalGas:    fld.totalGas,
		totalAllocs: fld.totalAllocs,
		totalAllocB: fld.totalAllocB,
		fileSamples: make(map[string]map[int]*lineStat, len(fld.fileSamples)),
	}
	for file, lines := range fld.fileSamples {
		lineCopy := make(map[int]*lineStat, len(lines))
		for line, stat := range lines {
			if stat == nil {
				continue
			}
			lineCopy[line] = &lineStat{
				count:  stat.count,
				cycles: stat.cycles,
				gas:    stat.gas,
			}
		}
		clone.fileSamples[file] = lineCopy
	}
	return clone
}

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
	Type          ProfileType       `json:"type"`
	TimeNanos     int64             `json:"timeNanos"`
	DurationNanos int64             `json:"durationNanos"`
	TotalCycles   int64             `json:"totalCycles,omitempty"`
	TotalGas      int64             `json:"totalGas,omitempty"`
	Frames        []ProfileLocation `json:"frames,omitempty"`
	Functions     []*FunctionStat   `json:"functions,omitempty"`
	CallTree      *CallTreeNode     `json:"callTree,omitempty"`

	// Aggregated per-function line statistics (not serialized directly).
	FunctionLines map[string]*functionLineData `json:"-"`

	// Captured line-level profiling data keyed by file -> line.
	LineStats map[string]map[int]*lineStats `json:"-"`

	// CPU specific
	CPUHz int64 `json:"cpuHz,omitempty"`

	// Memory specific
	DefaultSampleType string `json:"defaultSampleType,omitempty"`

	mu sync.RWMutex `json:"-"`
}

type sampleBaseline struct {
	prevCycles     int64
	prevGas        int64
	prevLineCycles int64
}

// Profiler manages profiling data collection
type Profiler struct {
	enabled    bool
	profile    *Profile
	startTime  time.Time
	sampleRate int // sample every N operations
	opCount    int

	baselines map[uintptr]*sampleBaseline // keyed by machine identity; 0 for unknown

	// Function profiling
	funcStats map[string]*FunctionStat

	// Aggregated call tree rooted at sentinel node.
	callRoot *callTreeNode

	// Line-level profiling support
	lineLevel   bool
	lineSamples map[string]map[int]*lineStats // file -> line -> stats

	// Function line stats aggregated per sample.
	functionLines map[string]*functionLineData

	// Test file filtering
	excludeTests bool // whether to exclude *_test.gno files from profiling

	totalCycles int64
	totalGas    int64

	frameStore frameStore

	mu sync.Mutex
}

// Options for starting profiling
type Options struct {
	Type       ProfileType
	SampleRate int
}

// NewProfiler creates a new profiler instance
// Optional parameters: profileType (default: ProfileCPU), sampleRate (default: 1000)
func NewProfiler(params ...any) *Profiler {
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
		funcStats:     make(map[string]*FunctionStat),
		lineSamples:   make(map[string]map[int]*lineStats),
		functionLines: make(map[string]*functionLineData),
		sampleRate:    sampleRate,
		excludeTests:  true, // Exclude test files by default
		frameStore:    newFrameStore(),
		callRoot:      newCallTreeNode(invalidFrameID),
	}

	p.profile = &Profile{
		Type:          profileType,
		FunctionLines: make(map[string]*functionLineData),
		LineStats:     make(map[string]map[int]*lineStats),
	}

	return p
}

// StartProfiling starts profiling with the given options
func (p *Profiler) StartProfiling(m MachineInfo, opts Options) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.enabled {
		// Avoid resetting an active session; callers should StopProfiling first.
		return
	}

	p.enabled = true
	p.startTime = time.Now()
	p.sampleRate = opts.SampleRate
	if p.sampleRate <= 0 {
		p.sampleRate = 1000 // default sample rate
	}

	p.profile = &Profile{
		Type:          opts.Type,
		TimeNanos:     p.startTime.UnixNano(),
		FunctionLines: make(map[string]*functionLineData),
		LineStats:     make(map[string]map[int]*lineStats),
	}

	// Reset state
	p.opCount = 0
	p.funcStats = make(map[string]*FunctionStat)
	p.functionLines = make(map[string]*functionLineData)
	p.callRoot = newCallTreeNode(invalidFrameID)
	p.frameStore.reset()
	p.lineSamples = make(map[string]map[int]*lineStats)
	p.totalCycles = 0
	p.totalGas = 0
	p.baselines = make(map[uintptr]*sampleBaseline)

	// Initialize previous cycle/gas counters to current values
	// so the first sample computes delta from this baseline
	if m != nil {
		key := machineKey(m)
		p.baselines[key] = &sampleBaseline{
			prevCycles:     m.GetCycles(),
			prevGas:        m.GetGasUsed(),
			prevLineCycles: m.GetCycles(),
		}
	}
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

	// VM already handles sampling rate via maybeEmitSample(),
	// so we record every callback we receive
	p.opCount++

	// Compute deltas since last sample (VM passes cumulative totals)
	baseline := p.ensureBaseline(machineKey(m))
	currentCycles := m.GetCycles()
	currentGas := m.GetGasUsed()
	deltaCycles := currentCycles - baseline.prevCycles
	if deltaCycles < 0 || currentCycles <= baseline.prevCycles {
		deltaCycles = currentCycles
	}
	deltaGas := currentGas - baseline.prevGas
	if deltaGas < 0 || currentGas <= baseline.prevGas {
		deltaGas = currentGas
	}
	// Update previous values immediately so that even skipped samples keep
	// counters in sync (e.g. when the stack is empty)
	baseline.prevCycles = currentCycles
	baseline.prevGas = currentGas

	// Build call stack
	stack := p.buildCallStack(m)
	if len(stack) == 0 {
		return
	}

	p.totalCycles += deltaCycles
	p.totalGas += deltaGas

	frameIDs := p.stackFrameIDs(stack)
	p.updateCallTree(frameIDs, deltaCycles, deltaGas, 0, 0)
	p.updateFunctionStats(stack, deltaCycles, deltaGas)
	if !p.lineLevel {
		p.updateFunctionLineStats(stack, deltaCycles, deltaGas)
	}

	// Update line-level profiling if enabled
	if p.lineLevel && len(stack) > 0 {
		loc := stack[0]
		// Some frames only record basenames (e.g. "basename.gno"), which caused
		// `source not available` when profiling code outside the current package.
		// Canonicalize the path so cross-package references can be resolved later.
		file := canonicalFilePath(loc.File, loc.Function)
		// Skip test files if excludeTests is enabled
		if file != "" && loc.Line > 0 && !(p.excludeTests && strings.HasSuffix(file, "_test.gno")) {
			if p.lineSamples[file] == nil {
				p.lineSamples[file] = make(map[int]*lineStats)
			}
			if p.lineSamples[file][loc.Line] == nil {
				p.lineSamples[file][loc.Line] = &lineStats{}
			}
			p.lineSamples[file][loc.Line].count++
			p.lineSamples[file][loc.Line].cycles += deltaCycles
			// Track gas for gas profiling
			if p.profile.Type == ProfileGas {
				p.lineSamples[file][loc.Line].gas += deltaGas
				p.updateFunctionLineGas(loc, deltaGas)
			}
		}
	}
}

// RecordAlloc records memory allocation
func (p *Profiler) RecordAlloc(m MachineInfo, size, count int64, allocType string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled || p.profile.Type != ProfileMemory {
		return
	}

	// VM already handles sampling rate for allocations,
	// so we record every callback we receive
	p.opCount++

	stack := p.buildCallStack(m)
	if len(stack) == 0 {
		// If no call stack, create a synthetic entry
		stack = []ProfileLocation{{
			Function: "<allocation>",
		}}
	}

	frameIDs := p.stackFrameIDs(stack)
	p.updateCallTree(frameIDs, 0, 0, size, count)
	p.updateAllocationLineStats(stack, size, count)

	// Update function profile stats
	funcName := stack[0].Function
	stat := p.getFunctionStat(funcName)
	stat.CallCount++
	stat.AllocBytes += size
	stat.AllocObjects += count
}

// buildCallStack builds a call stack from machine frames
func (p *Profiler) buildCallStack(m MachineInfo) []ProfileLocation {
	var stack []ProfileLocation

	if m == nil {
		return stack
	}

	frames := m.GetFrames()

	// preserve the VM's top-first frame order.
	// This ensures that the `updateFunctionStats` increments
	// `SelfCycles` for the actual top-of-stack frame and
	// `updateCallTree` still walks from root to leaf correctly.
	for i := 0; i < len(frames); i++ {
		frame := frames[i]
		if !frame.IsCall() {
			continue
		}

		// Skip test files if excludeTests is enabled
		fileName := frame.GetFileName()
		if p.excludeTests && strings.HasSuffix(fileName, "_test.gno") {
			continue
		}

		// fallback: anonymous or package-level init function.
		funcName := frame.GetFuncName()
		if funcName == "" {
			funcName = "<anonymous>"
		}

		loc := ProfileLocation{
			Function: funcName,
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
			} else if loc.File != "" && !strings.Contains(loc.File, "/") {
				// If File is just a filename (no path), prepend the package path
				// Convert package path to file path format
				// e.g., "gno.land/p/nt/ufmt" -> "gno.land/p/nt/ufmt/ufmt.gno"
				loc.File = pkgPath + "/" + loc.File
			}
		}

		stack = append(stack, loc)
	}

	return stack
}

// generateSamples finalizes aggregated profiling data.
func (p *Profiler) generateSamples() {
	if p.profile == nil {
		return
	}

	p.profile.TotalCycles = p.totalCycles
	p.profile.TotalGas = p.totalGas
	p.profile.Frames = p.frameStore.framesSnapshot()
	p.profile.Functions = p.collectFunctionStats()
	p.profile.CallTree = p.buildExportCallTree()
	p.profile.FunctionLines = make(map[string]*functionLineData, len(p.functionLines))
	for name, data := range p.functionLines {
		p.profile.FunctionLines[name] = data.clone()
	}
	p.profile.LineStats = cloneLineSamples(p.lineSamples)
}

func (p *Profiler) collectFunctionStats() []*FunctionStat {
	out := make([]*FunctionStat, 0, len(p.funcStats))
	for _, stat := range p.funcStats {
		clone := *stat
		out = append(out, &clone)
	}
	sort.Slice(out, func(i, j int) bool {
		if p.profile != nil && p.profile.Type == ProfileMemory {
			if out[i].AllocBytes == out[j].AllocBytes {
				return out[i].AllocObjects > out[j].AllocObjects
			}
			return out[i].AllocBytes > out[j].AllocBytes
		}
		if p.profile != nil && p.profile.Type == ProfileGas {
			if out[i].TotalGas == out[j].TotalGas {
				return out[i].Name < out[j].Name
			}
			return out[i].TotalGas > out[j].TotalGas
		}
		if out[i].TotalCycles == out[j].TotalCycles {
			return out[i].Name < out[j].Name
		}
		return out[i].TotalCycles > out[j].TotalCycles
	})
	return out
}

func (p *Profiler) buildExportCallTree() *CallTreeNode {
	if p.callRoot == nil {
		return nil
	}
	preferGas := p.profile != nil && p.profile.Type == ProfileGas
	preferAlloc := p.profile != nil && p.profile.Type == ProfileMemory
	return convertCallTreeNode(p.callRoot, preferGas, preferAlloc)
}

func convertCallTreeNode(node *callTreeNode, preferGas, preferAlloc bool) *CallTreeNode {
	if node == nil {
		return nil
	}

	export := &CallTreeNode{
		FrameID:     node.frameID,
		Calls:       node.calls,
		TotalCycles: node.totalCycles,
		SelfCycles:  node.selfCycles,
		TotalGas:    node.totalGas,
		SelfGas:     node.selfGas,
		AllocBytes:  node.allocBytes,
		AllocObjs:   node.allocObjects,
	}

	if len(node.children) > 0 {
		children := make([]*CallTreeNode, 0, len(node.children))
		ids := make([]FrameID, 0, len(node.children))
		for id := range node.children {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool {
			left := node.children[ids[i]]
			right := node.children[ids[j]]
			if preferAlloc {
				if left.allocBytes == right.allocBytes {
					return left.allocObjects > right.allocObjects
				}
				return left.allocBytes > right.allocBytes
			}
			if preferGas {
				if left.totalGas == right.totalGas {
					return left.totalCycles > right.totalCycles
				}
				return left.totalGas > right.totalGas
			}
			if left.totalCycles == right.totalCycles {
				return left.totalGas > right.totalGas
			}
			return left.totalCycles > right.totalCycles
		})
		for _, id := range ids {
			children = append(children, convertCallTreeNode(node.children[id], preferGas, preferAlloc))
		}
		export.Children = children
	}

	return export
}

func cloneLineSamples(src map[string]map[int]*lineStats) map[string]map[int]*lineStats {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]map[int]*lineStats, len(src))
	for file, lines := range src {
		lineCopy := make(map[int]*lineStats, len(lines))
		for line, stat := range lines {
			if stat == nil {
				continue
			}
			copied := &lineStats{
				lineStat: lineStat{
					count:       stat.count,
					cycles:      stat.cycles,
					gas:         stat.gas,
					allocations: stat.allocations,
					allocBytes:  stat.allocBytes,
				},
			}
			lineCopy[line] = copied
		}
		dst[file] = lineCopy
	}
	return dst
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
	fmt.Fprintf(cw, "Functions Tracked: %d\n\n", len(p.Functions))

	switch p.Type {
	case ProfileGas:
		writeTopGasFunctions(cw, p.Functions, p.TotalGas)
	case ProfileMemory:
		writeTopMemoryFunctions(cw, p.Functions, p.totalAllocBytes(), p.totalAllocObjects())
	default:
		writeTopCPUFunctions(cw, p.Functions, p.TotalCycles)
	}

	return cw.n, cw.err
}

func writeTopCPUFunctions(w io.Writer, stats []*FunctionStat, total int64) {
	fmt.Fprintf(w, "Top Functions (CPU Cycles):\n")
	fmt.Fprintf(w, "%-50s %12s %12s %12s %12s\n", "Function", "Flat", "Flat%", "Cum", "Cum%")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 100))
	for i, stat := range stats {
		if i >= 20 {
			break
		}
		flat := stat.SelfCycles
		cum := stat.TotalCycles
		flatPercent := percent(flat, total)
		cumPercent := percent(cum, total)
		fmt.Fprintf(w, "%-50s %12d %11.2f%% %12d %11.2f%%\n",
			shortenName(stat.Name, 50), flat, flatPercent, cum, cumPercent)
	}
	fmt.Fprintln(w)
}

func writeTopGasFunctions(w io.Writer, stats []*FunctionStat, total int64) {
	fmt.Fprintf(w, "Top Functions (Gas):\n")
	fmt.Fprintf(w, "%-50s %12s %12s %12s %12s\n", "Function", "Flat Gas", "Flat%", "Cum Gas", "Cum%")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 100))
	for i, stat := range stats {
		if i >= 20 {
			break
		}
		flat := stat.SelfGas
		cum := stat.TotalGas
		flatPercent := percent(flat, total)
		cumPercent := percent(cum, total)
		fmt.Fprintf(w, "%-50s %12d %11.2f%% %12d %11.2f%%\n",
			shortenName(stat.Name, 50), flat, flatPercent, cum, cumPercent)
	}
	fmt.Fprintln(w)
}

func writeTopMemoryFunctions(w io.Writer, stats []*FunctionStat, totalBytes, totalObjects int64) {
	fmt.Fprintf(w, "Top Functions (Memory Allocations):\n")
	fmt.Fprintf(w, "%-50s %12s %12s %12s %12s\n", "Function", "Bytes", "Bytes%", "Objects", "Obj%")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 100))
	for i, stat := range stats {
		if i >= 20 {
			break
		}
		bytesPercent := percent(stat.AllocBytes, totalBytes)
		objPercent := percent(stat.AllocObjects, totalObjects)
		fmt.Fprintf(w, "%-50s %12d %11.2f%% %12d %11.2f%%\n",
			shortenName(stat.Name, 50), stat.AllocBytes, bytesPercent, stat.AllocObjects, objPercent)
	}
	fmt.Fprintln(w)
}

func percent(value, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total) * 100
}

func shortenName(name string, maxlen int) string {
	if len(name) <= maxlen {
		return name
	}
	return name[:maxlen-3] + "..."
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
	if p.TotalCycles > 0 {
		return p.TotalCycles
	}
	var total int64
	for _, stat := range p.Functions {
		total += stat.SelfCycles
	}
	return total
}

// totalGas calculates total gas across all samples
func (p *Profile) totalGas() int64 {
	if p.TotalGas > 0 {
		return p.TotalGas
	}
	var total int64
	for _, stat := range p.Functions {
		total += stat.SelfGas
	}
	return total
}

func (p *Profile) totalAllocBytes() int64 {
	var total int64
	for _, stat := range p.Functions {
		total += stat.AllocBytes
	}
	return total
}

func (p *Profile) totalAllocObjects() int64 {
	var total int64
	for _, stat := range p.Functions {
		total += stat.AllocObjects
	}
	return total
}

// EnableLineProfiling enables line-level profiling
func (p *Profiler) EnableLineProfiling() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lineLevel = true
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
func (p *Profiler) RecordLineSample(funcName, file string, line int, cycles int64, machineID uintptr) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled || !p.lineLevel {
		return
	}

	// Compute delta since last line sample; fall back to current cycles when the
	// counter resets (e.g. switching to a different machine).
	baseline := p.ensureBaseline(machineID)
	deltaCycles := cycles - baseline.prevLineCycles
	if deltaCycles < 0 || cycles <= baseline.prevLineCycles {
		deltaCycles = cycles
	}
	baseline.prevLineCycles = cycles

	file = canonicalFilePath(file, funcName)

	// Skip test files if excludeTests is enabled
	if file == "" || (p.excludeTests && strings.HasSuffix(file, "_test.gno")) {
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
	p.lineSamples[file][line].cycles += deltaCycles

	p.updateFunctionLineStatsFromLineSample(funcName, file, line, deltaCycles)
}

func (p *Profiler) stackFrameIDs(stack []ProfileLocation) []FrameID {
	ids := make([]FrameID, len(stack))
	for i, loc := range stack {
		ids[i] = p.frameStore.intern(loc)
	}
	return ids
}

func (p *Profiler) updateCallTree(ids []FrameID, cycles, gas, allocBytes, allocObjs int64) {
	if len(ids) == 0 {
		return
	}

	if p.callRoot == nil {
		p.callRoot = newCallTreeNode(invalidFrameID)
	}

	node := p.callRoot
	for i := len(ids) - 1; i >= 0; i-- {
		id := ids[i]
		child, ok := node.children[id]
		if !ok {
			child = newCallTreeNode(id)
			node.children[id] = child
		}
		child.calls++
		child.totalCycles += cycles
		if i == 0 {
			child.selfCycles += cycles
		}
		child.totalGas += gas
		if i == 0 {
			child.selfGas += gas
		}
		child.allocBytes += allocBytes
		child.allocObjects += allocObjs
		node = child
	}
}

func (p *Profiler) updateFunctionStats(stack []ProfileLocation, cycles, gas int64) {
	seen := make(map[string]bool)

	for i, loc := range stack {
		funcName := loc.Function
		if funcName == "" || seen[funcName] || isFilteredFunction(funcName) {
			continue
		}
		seen[funcName] = true

		stat := p.getFunctionStat(funcName)
		stat.CallCount++
		stat.TotalCycles += cycles
		if i == 0 {
			stat.SelfCycles += cycles
		}
		stat.TotalGas += gas
		if i == 0 {
			stat.SelfGas += gas
		}
	}
}

func (p *Profiler) updateFunctionLineStats(stack []ProfileLocation, cycles, gas int64) {
	seen := make(map[string]bool)

	for _, loc := range stack {
		funcName := loc.Function
		if funcName == "" || seen[funcName] || isFilteredFunction(funcName) {
			continue
		}
		seen[funcName] = true

		file := canonicalFilePath(loc.File, funcName)
		line := loc.Line
		if file == "" || line <= 0 {
			continue
		}
		if p.excludeTests && strings.HasSuffix(file, "_test.gno") {
			continue
		}

		info, ok := p.functionLines[funcName]
		if !ok {
			info = &functionLineData{
				funcName:    funcName,
				fileSamples: make(map[string]map[int]*lineStat),
			}
			p.functionLines[funcName] = info
		}
		if info.fileSamples[file] == nil {
			info.fileSamples[file] = make(map[int]*lineStat)
		}
		stat := info.fileSamples[file][line]
		if stat == nil {
			stat = &lineStat{}
			info.fileSamples[file][line] = stat
		}
		stat.count++
		stat.cycles += cycles
		if gas > 0 {
			stat.gas += gas
		}
		info.totalCycles += cycles
		info.totalGas += gas
	}
}

func (p *Profiler) updateAllocationLineStats(stack []ProfileLocation, size, count int64) {
	seen := make(map[string]bool)
	for _, loc := range stack {
		funcName := loc.Function
		if funcName == "" || seen[funcName] || isFilteredFunction(funcName) {
			continue
		}
		seen[funcName] = true

		file := canonicalFilePath(loc.File, funcName)
		line := loc.Line
		if file == "" || line <= 0 {
			continue
		}
		if p.excludeTests && strings.HasSuffix(file, "_test.gno") {
			continue
		}

		if p.lineSamples[file] == nil {
			p.lineSamples[file] = make(map[int]*lineStats)
		}
		if p.lineSamples[file][line] == nil {
			p.lineSamples[file][line] = &lineStats{}
		}
		lstat := p.lineSamples[file][line]
		lstat.allocations += count
		lstat.allocBytes += size

		info, ok := p.functionLines[funcName]
		if !ok {
			info = &functionLineData{
				funcName:    funcName,
				fileSamples: make(map[string]map[int]*lineStat),
			}
			p.functionLines[funcName] = info
		}
		if info.fileSamples[file] == nil {
			info.fileSamples[file] = make(map[int]*lineStat)
		}
		stat := info.fileSamples[file][line]
		if stat == nil {
			stat = &lineStat{}
			info.fileSamples[file][line] = stat
		}
		stat.allocations += count
		stat.allocBytes += size
		info.totalAllocs += count
		info.totalAllocB += size
	}
}

func (p *Profiler) updateFunctionLineStatsFromLineSample(funcName, file string, line int, cycles int64) {
	file = canonicalFilePath(file, funcName)
	if funcName == "" || file == "" || line <= 0 || isFilteredFunction(funcName) {
		return
	}

	info, ok := p.functionLines[funcName]
	if !ok {
		info = &functionLineData{
			funcName:    funcName,
			fileSamples: make(map[string]map[int]*lineStat),
		}
		p.functionLines[funcName] = info
	}

	if info.fileSamples[file] == nil {
		info.fileSamples[file] = make(map[int]*lineStat)
	}

	stat := info.fileSamples[file][line]
	if stat == nil {
		stat = &lineStat{}
		info.fileSamples[file][line] = stat
	}

	stat.count++
	stat.cycles += cycles
	info.totalCycles += cycles
}

// updateFunctionLineGas records gas for a specific line without altering cycle-based
// statistics. This keeps gas totals in sync for line-level profiles where cycle
// attribution is handled by dedicated line samples.
func (p *Profiler) updateFunctionLineGas(loc ProfileLocation, gas int64) {
	if gas <= 0 {
		return
	}

	funcName := loc.Function
	if funcName == "" || isFilteredFunction(funcName) {
		return
	}

	file := canonicalFilePath(loc.File, funcName)
	line := loc.Line
	if file == "" || line <= 0 {
		return
	}
	if p.excludeTests && strings.HasSuffix(file, "_test.gno") {
		return
	}

	info, ok := p.functionLines[funcName]
	if !ok {
		info = &functionLineData{
			funcName:    funcName,
			fileSamples: make(map[string]map[int]*lineStat),
		}
		p.functionLines[funcName] = info
	}
	if info.fileSamples[file] == nil {
		info.fileSamples[file] = make(map[int]*lineStat)
	}
	stat := info.fileSamples[file][line]
	if stat == nil {
		stat = &lineStat{}
		info.fileSamples[file][line] = stat
	}
	stat.gas += gas
	info.totalGas += gas
}

func (p *Profiler) getFunctionStat(name string) *FunctionStat {
	if name == "" {
		name = "<unknown>"
	}
	stat, ok := p.funcStats[name]
	if !ok {
		stat = &FunctionStat{Name: name}
		p.funcStats[name] = stat
	}
	return stat
}

func isFilteredFunction(name string) bool {
	// The Go `testing` harness dominates samples when running `gno test`.
	// Filter those frames out so the toplist highlights user code instead.
	return strings.HasPrefix(name, "testing.")
}

// canonicalFilePath reconstructs a fully-qualified file path when the VM only
// reported a basename. This fixes `source not available` errors that previously
// occurred for packages imported from other directories.
func canonicalFilePath(file, funcName string) string {
	if file == "" {
		return ""
	}
	if strings.Contains(file, "/") {
		return file
	}
	if pkg := packageFromFunction(funcName); pkg != "" {
		return pkg + "/" + file
	}
	return file
}

func packageFromFunction(funcName string) string {
	if funcName == "" {
		return ""
	}

	// Handle method calls like "pkg.(*Type).Method" or "pkg.Type.Method"
	// We need to extract just the package path, not the type information

	// First, check if this is a method (contains parentheses or multiple dots after the package)
	if strings.Contains(funcName, "(") {
		// This is a pointer receiver method like "pkg.(*Type).Method"
		// Find the first dot to get the package name
		if idx := strings.Index(funcName, "."); idx > 0 {
			return funcName[:idx]
		}
	} else {
		// Count dots to distinguish between "pkg.Func" and "pkg.Type.Method"
		parts := strings.Split(funcName, ".")
		if len(parts) >= 2 {
			// Check if the second part starts with uppercase (likely a type)
			if len(parts) == 3 && len(parts[1]) > 0 && parts[1][0] >= 'A' && parts[1][0] <= 'Z' {
				// This is likely "pkg.Type.Method"
				return parts[0]
			}
			// This is likely "pkg.subpkg.Func" or just "pkg.Func"
			if idx := strings.LastIndex(funcName, "."); idx > 0 {
				return funcName[:idx]
			}
		}
	}

	return ""
}

func (p *Profiler) ensureBaseline(key uintptr) *sampleBaseline {
	if p.baselines == nil {
		p.baselines = make(map[uintptr]*sampleBaseline)
	}
	if b := p.baselines[key]; b != nil {
		return b
	}
	b := &sampleBaseline{}
	p.baselines[key] = b
	return b
}

func machineKey(m MachineInfo) uintptr {
	if m == nil {
		return 0
	}
	type identifier interface {
		Identity() uintptr
	}
	if idm, ok := m.(identifier); ok {
		if id := idm.Identity(); id != 0 {
			return id
		}
	}
	return 0
}
