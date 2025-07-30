package gnolang

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// LocationKey is used for caching profile locations
type LocationKey struct {
	PkgPath  string
	Function string
	File     string
	Line     int
}

// profileLocation represents an enhanced location in the call stack
type profileLocation struct {
	function   string
	file       string
	line       int
	column     int
	inlineCall bool
	pc         uintptr
}

// newProfileLocation creates a new profile location
func newProfileLocation(function, file string, line, column int) *profileLocation {
	return &profileLocation{
		function: function,
		file:     file,
		line:     line,
		column:   column,
	}
}

func (pl *profileLocation) Function() string   { return pl.function }
func (pl *profileLocation) File() string       { return pl.file }
func (pl *profileLocation) Line() int          { return pl.line }
func (pl *profileLocation) Column() int        { return pl.column }
func (pl *profileLocation) PC() uintptr        { return pl.pc }
func (pl *profileLocation) IsInlineCall() bool { return pl.inlineCall }

func (pl *profileLocation) SetPC(pc uintptr)          { pl.pc = pc }
func (pl *profileLocation) SetInlineCall(inline bool) { pl.inlineCall = inline }

// setValues sets all values at once (useful for pooling)
func (pl *profileLocation) setValues(function, file string, line, column int) {
	pl.function = function
	pl.file = file
	pl.line = line
	pl.column = column
	pl.pc = 0
	pl.inlineCall = false
}

// reset clears all values (for pool reuse)
func (pl *profileLocation) reset() {
	pl.function = ""
	pl.file = ""
	pl.line = 0
	pl.column = 0
	pl.pc = 0
	pl.inlineCall = false
}

// Convert to public ProfileLocation for API compatibility
func (pl *profileLocation) toPublic() ProfileLocation {
	return ProfileLocation{
		Function:   pl.function,
		File:       pl.file,
		Line:       pl.line,
		Column:     pl.column,
		InlineCall: pl.inlineCall,
		PC:         pl.pc,
	}
}

// locationCache provides efficient location deduplication
type locationCache struct {
	locations map[LocationKey]*profileLocation
	mu        sync.RWMutex
}

func newLocationCache() *locationCache {
	return &locationCache{
		locations: make(map[LocationKey]*profileLocation),
	}
}

func (lc *locationCache) getOrCreate(key LocationKey) *profileLocation {
	lc.mu.RLock()
	if loc, exists := lc.locations[key]; exists {
		lc.mu.RUnlock()
		return loc
	}
	lc.mu.RUnlock()

	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Double-check after acquiring write lock
	if loc, exists := lc.locations[key]; exists {
		return loc
	}

	loc := newProfileLocation(key.Function, key.File, key.Line, 0)
	lc.locations[key] = loc
	return loc
}

func (lc *locationCache) size() int {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return len(lc.locations)
}

// lineStats tracks statistics for a single line
type lineStats struct {
	count       int64
	cycles      int64
	allocations int64
	allocBytes  int64
	mu          sync.Mutex
}

func newLineStats() *lineStats {
	return &lineStats{}
}

func (ls *lineStats) addSample(cycles int64, allocations, allocBytes int64) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.count++
	ls.cycles += cycles
	ls.allocations += allocations
	ls.allocBytes += allocBytes
}

// Getters (thread-safe)
func (ls *lineStats) GetCount() int64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.count
}

func (ls *lineStats) GetCycles() int64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.cycles
}

func (ls *lineStats) GetAllocations() int64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.allocations
}

func (ls *lineStats) GetAllocBytes() int64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.allocBytes
}

// EnableLineLevel enables or disables line-level profiling
func (p *Profiler) EnableLineLevel(enable bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.lineLevel = enable
	if enable && p.locationCache == nil {
		p.locationCache = newLocationCache()
	}
}

// RecordLineLevel records a sample with line-level information
func (p *Profiler) RecordLineLevel(m *Machine, loc *profileLocation, cycles int64) {
	if !p.enabled || !p.lineLevel {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.recordLineLevelUnlocked(loc, cycles)
}

// recordLineLevelUnlocked records line-level profiling information without locking
// Must be called with mutex already held
func (p *Profiler) recordLineLevelUnlocked(loc *profileLocation, cycles int64) {
	// Update line statistics
	p.updateLineStatsUnlocked(loc, cycles, 0, 0)

	// Record in regular samples for compatibility
	sample := ProfileSample{
		Location:   []ProfileLocation{loc.toPublic()},
		Value:      []int64{1, cycles},
		Label:      make(map[string][]string),
		NumLabel:   make(map[string][]int64),
		SampleType: p.profile.Type,
	}

	p.profile.Samples = append(p.profile.Samples, sample)
}

// updateLineStats updates line-level statistics
func (p *Profiler) updateLineStats(loc *profileLocation, cycles, allocations, allocBytes int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.updateLineStatsUnlocked(loc, cycles, allocations, allocBytes)
}

// updateLineStatsUnlocked updates line-level statistics without locking
// Must be called with mutex already held
func (p *Profiler) updateLineStatsUnlocked(loc *profileLocation, cycles, allocations, allocBytes int64) {
	file := loc.File()
	line := loc.Line()

	if p.lineSamples[file] == nil {
		p.lineSamples[file] = make(map[int]*lineStats)
	}

	stats := p.lineSamples[file][line]
	if stats == nil {
		stats = newLineStats()
		p.lineSamples[file][line] = stats
	}

	stats.addSample(cycles, allocations, allocBytes)
}

// GetLineStats returns line statistics for a file
func (p *Profiler) GetLineStats(filename string) map[int]*lineStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.lineSamples[filename]
}

// Memory pool methods
func (p *Profiler) getLocationFromPool() *profileLocation {
	loc := p.locationPool.Get().(*profileLocation)
	loc.reset()
	return loc
}

func (p *Profiler) putLocationToPool(loc *profileLocation) {
	loc.reset()
	p.locationPool.Put(loc)
}

// WriteSourceAnnotated writes source code with profiling annotations
func (p *Profile) WriteSourceAnnotated(w io.Writer, filename string, source io.Reader) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Collect line stats for this file
	lineStats := make(map[int]*lineStat)
	totalCycles := int64(0)

	// Get profiler instance to access line stats
	// This is a bit tricky as Profile doesn't directly reference Profiler
	// For now, we'll extract from samples
	for _, sample := range p.Samples {
		if len(sample.Location) > 0 && sample.Location[0].File == filename {
			line := sample.Location[0].Line
			if line > 0 {
				if lineStats[line] == nil {
					lineStats[line] = &lineStat{}
				}
				if len(sample.Value) > 1 {
					lineStats[line].cycles += sample.Value[1]
					totalCycles += sample.Value[1]
				}
				if len(sample.Value) > 0 {
					lineStats[line].count += sample.Value[0]
				}
			}
		}
	}

	// Read source content
	content, err := io.ReadAll(source)
	if err != nil {
		return fmt.Errorf("failed to read source: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Write header
	fmt.Fprintf(w, "File: %s\n", filename)
	fmt.Fprintf(w, "Total cycles: %d\n\n", totalCycles)
	fmt.Fprintf(w, "%-8s %-8s | Source\n", "Cycles", "Count")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 80))

	// Write each line with annotations
	for i, line := range lines {
		lineNum := i + 1
		if stats, exists := lineStats[lineNum]; exists {
			percentage := float64(0)
			if totalCycles > 0 {
				percentage = float64(stats.cycles) / float64(totalCycles) * 100
			}

			fmt.Fprintf(w, "%7d %7d | %4d: %s\n",
				stats.cycles, stats.count, lineNum, line)

			// Mark hot spots (>10% of total cycles)
			if percentage > 10.0 {
				fmt.Fprintf(w, "%16s | %s^ HOT (%.1f%%)\n",
					"", strings.Repeat(" ", 6), percentage)
			}
		} else {
			// No profiling data for this line
			fmt.Fprintf(w, "%7s %7s | %4d: %s\n",
				".", ".", lineNum, line)
		}
	}

	return nil
}

// lineStat is a simplified version for WriteSourceAnnotated
type lineStat struct {
	count  int64
	cycles int64
}
