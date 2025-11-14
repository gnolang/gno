package profiler

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

// lineStats tracks statistics for a single line
type lineStats struct {
	lineStat
	allocations int64
	allocBytes  int64
	mu          sync.Mutex
}

// lineStat is a simplified version for WriteSourceAnnotated
type lineStat struct {
	count  int64
	cycles int64
	gas    int64
}

func (ls *lineStats) Count() int64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.count
}

func (ls *lineStats) Cycles() int64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.cycles
}

func (ls *lineStats) Allocations() int64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.allocations
}

func (ls *lineStats) AllocBytes() int64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.allocBytes
}

// LineStats returns line statistics for a file
func (p *Profiler) LineStats(filename string) map[int]*lineStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.lineSamples[filename]
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
			if line <= 0 {
				continue
			}

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
