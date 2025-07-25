package gnolang

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// ProfileFormat represents the output format for profile data
type ProfileFormat int

const (
	FormatText ProfileFormat = iota
	FormatFlameGraph
	FormatCallTree
	FormatTopList
)

// WriteFormat writes the profile in the specified format
func (p *Profile) WriteFormat(w io.Writer, format ProfileFormat) error {
	switch format {
	case FormatFlameGraph:
		return p.WriteFlameGraph(w)
	case FormatCallTree:
		return p.WriteCallTree(w)
	case FormatTopList:
		return p.WriteTopList(w)
	default:
		return p.WriteTo(w)
	}
}

// WriteFlameGraph writes profile data in flame graph format
// Format: function;caller;caller cycles
func (p *Profile) WriteFlameGraph(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Build call stacks
	stacks := make(map[string]int64)

	for _, sample := range p.Samples {
		// Only process samples with multiple locations (stack traces)
		if len(sample.Location) <= 1 {
			continue
		}

		// Build stack string (already in correct order)
		var stack []string
		for _, loc := range sample.Location {
			stack = append(stack, loc.Function)
		}
		stackStr := strings.Join(stack, ";")

		// Add cycles
		if len(sample.Value) > 1 {
			stacks[stackStr] += sample.Value[1] // cycles
		}
	}

	// Write in flame graph format
	for stack, cycles := range stacks {
		fmt.Fprintf(w, "%s %d\n", stack, cycles)
	}

	return nil
}

// WriteCallTree writes a hierarchical call tree
func (p *Profile) WriteCallTree(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	fmt.Fprintf(w, "Call Tree (CPU Cycles)\n")
	fmt.Fprintf(w, "======================\n\n")

	// Build tree structure
	type node struct {
		name     string
		cycles   int64
		calls    int64
		children map[string]*node
	}

	root := &node{
		name:     "root",
		children: make(map[string]*node),
	}

	// Build tree from samples
	for _, sample := range p.Samples {
		// Only process samples with multiple locations (stack traces)
		if len(sample.Location) <= 1 {
			continue
		}

		current := root
		// Traverse from root to leaf (already in correct order)
		for i := 0; i < len(sample.Location); i++ {
			funcName := sample.Location[i].Function

			if _, exists := current.children[funcName]; !exists {
				current.children[funcName] = &node{
					name:     funcName,
					children: make(map[string]*node),
				}
			}

			child := current.children[funcName]

			// Add cycles to each node in the path
			if len(sample.Value) > 1 {
				child.cycles += sample.Value[1]
				child.calls += sample.Value[0]
			}

			current = child
		}
	}

	// Print tree
	var printNode func(*node, int)
	printNode = func(n *node, depth int) {
		if n.name != "root" {
			indent := strings.Repeat("  ", depth)
			percentage := float64(n.cycles) / float64(p.totalCycles()) * 100
			fmt.Fprintf(w, "%s%s: %d cycles (%.1f%%), %d calls\n",
				indent, n.name, n.cycles, percentage, n.calls)
		}

		// Sort children by cycles
		var children []*node
		for _, child := range n.children {
			children = append(children, child)
		}
		sort.Slice(children, func(i, j int) bool {
			return children[i].cycles > children[j].cycles
		})

		for _, child := range children {
			printNode(child, depth+1)
		}
	}

	printNode(root, -1)
	return nil
}

// WriteTopList writes a sorted list of top functions with visual bars
func (p *Profile) WriteTopList(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Adjust title based on profile type
	if p.Type == ProfileMemory {
		fmt.Fprintf(w, "Top Functions by Memory Allocation\n")
		fmt.Fprintf(w, "==================================\n\n")
	} else {
		fmt.Fprintf(w, "Top Functions by CPU Cycles\n")
		fmt.Fprintf(w, "===========================\n\n")
	}

	// Sort samples by primary metric (cycles for CPU, bytes for memory)
	sortedSamples := make([]ProfileSample, len(p.Samples))
	copy(sortedSamples, p.Samples)

	sort.Slice(sortedSamples, func(i, j int) bool {
		if len(sortedSamples[i].Value) > 1 && len(sortedSamples[j].Value) > 1 {
			return sortedSamples[i].Value[1] > sortedSamples[j].Value[1]
		}
		return false
	})

	totalCycles := p.totalCycles()
	if p.Type == ProfileCPU && totalCycles == 0 {
		fmt.Fprintf(w, "No CPU cycles recorded\n")
		return nil
	}
	if p.Type == ProfileMemory && p.totalMemory() == 0 {
		fmt.Fprintf(w, "No memory allocations recorded\n")
		return nil
	}

	maxBarWidth := 50

	// Print header based on profile type
	if p.Type == ProfileMemory {
		fmt.Fprintf(w, "%-40s %12s %12s %12s %8s %-10s %s\n", "Function", "Bytes", "Bytes%", "Allocs", "Allocs%", "Type", "Graph")
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", 130))
	} else {
		fmt.Fprintf(w, "%-40s %12s %12s %12s %12s %8s %s\n", "Function", "Flat", "Flat%", "Cum", "Cum%", "Calls", "Graph")
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", 130))
	}

	// Print top 20 functions
	for i, sample := range sortedSamples {
		if i >= 20 {
			break
		}

		funcName := "unknown"
		if len(sample.Location) > 0 {
			funcName = sample.Location[0].Function
			if len(funcName) > 40 {
				funcName = funcName[:37] + "..."
			}
		}

		if p.Type == ProfileMemory {
			// Memory profile display
			bytes := int64(0)
			allocs := int64(0)
			allocType := "unknown"

			if bytesVal, ok := sample.NumLabel["bytes"]; ok && len(bytesVal) > 0 {
				bytes = bytesVal[0]
			}
			if allocsVal, ok := sample.NumLabel["allocations"]; ok && len(allocsVal) > 0 {
				allocs = allocsVal[0]
			}
			if typeLabels, ok := sample.Label["type"]; ok && len(typeLabels) > 0 {
				allocType = typeLabels[0]
			}

			totalMem := p.totalMemory()
			totalAllocs := p.totalAllocations()

			bytesPercent := float64(0)
			allocsPercent := float64(0)
			if totalMem > 0 {
				bytesPercent = float64(bytes) / float64(totalMem) * 100
			}
			if totalAllocs > 0 {
				allocsPercent = float64(allocs) / float64(totalAllocs) * 100
			}

			// Create visual bar based on bytes percentage
			barWidth := int(bytesPercent * float64(maxBarWidth) / 100)
			if barWidth < 1 && bytesPercent > 0 {
				barWidth = 1
			}
			bar := strings.Repeat("█", barWidth)

			fmt.Fprintf(w, "%-40s %12d %11.2f%% %12d %7.2f%% %-10s %s\n",
				funcName, bytes, bytesPercent, allocs, allocsPercent, allocType, bar)
		} else {
			// CPU profile display (existing code)
			calls := int64(0)
			flatCycles := int64(0)
			cumCycles := int64(0)

			if callsVal, ok := sample.NumLabel["calls"]; ok && len(callsVal) > 0 {
				calls = callsVal[0]
			}
			if flatVal, ok := sample.NumLabel["flat_cycles"]; ok && len(flatVal) > 0 {
				flatCycles = flatVal[0]
			} else if cyclesVal, ok := sample.NumLabel["cycles"]; ok && len(cyclesVal) > 0 {
				// Fallback for compatibility
				flatCycles = cyclesVal[0]
			}
			if cumVal, ok := sample.NumLabel["cum_cycles"]; ok && len(cumVal) > 0 {
				cumCycles = cumVal[0]
			} else if cyclesVal, ok := sample.NumLabel["cycles"]; ok && len(cyclesVal) > 0 {
				// Fallback for compatibility
				cumCycles = cyclesVal[0]
			}

			flatPercent := float64(flatCycles) / float64(totalCycles) * 100
			cumPercent := float64(cumCycles) / float64(totalCycles) * 100

			// Create visual bar based on cumulative percentage
			barWidth := int(cumPercent * float64(maxBarWidth) / 100)
			if barWidth < 1 && cumPercent > 0 {
				barWidth = 1
			}
			bar := strings.Repeat("█", barWidth)

			fmt.Fprintf(w, "%-40s %12d %11.2f%% %12d %11.2f%% %12d %s\n",
				funcName, flatCycles, flatPercent, cumCycles, cumPercent, calls, bar)
		}
	}

	// Print summary
	if p.Type == ProfileMemory {
		fmt.Fprintf(w, "\nTotal bytes: %d\n", p.totalMemory())
		fmt.Fprintf(w, "Total allocations: %d\n", p.totalAllocations())
	} else {
		fmt.Fprintf(w, "\nTotal cycles: %d\n", totalCycles)
		fmt.Fprintf(w, "Total samples: %d\n", len(p.Samples))
	}

	return nil
}

// totalCycles calculates total CPU cycles across all samples
func (p *Profile) totalCycles() int64 {
	var total int64
	for _, sample := range p.Samples {
		if len(sample.Value) > 1 {
			total += sample.Value[1]
		}
	}
	return total
}

// totalMemory calculates total memory bytes across all samples
func (p *Profile) totalMemory() int64 {
	var total int64
	for _, sample := range p.Samples {
		if bytes, ok := sample.NumLabel["bytes"]; ok && len(bytes) > 0 {
			total += bytes[0]
		}
	}
	return total
}

// totalAllocations calculates total allocation count across all samples
func (p *Profile) totalAllocations() int64 {
	var total int64
	for _, sample := range p.Samples {
		if allocs, ok := sample.NumLabel["allocations"]; ok && len(allocs) > 0 {
			total += allocs[0]
		}
	}
	return total
}

// WriteProfileComparison writes a comparison between two profiles
func WriteProfileComparison(w io.Writer, before, after *Profile) error {
	if before.Type != after.Type {
		return fmt.Errorf("cannot compare profiles of different types")
	}

	fmt.Fprintf(w, "Profile Comparison\n")
	fmt.Fprintf(w, "==================\n\n")
	fmt.Fprintf(w, "Type: %s\n", before.typeString())
	fmt.Fprintf(w, "Before Duration: %s\n", time.Duration(before.DurationNanos))
	fmt.Fprintf(w, "After Duration: %s\n", time.Duration(after.DurationNanos))
	fmt.Fprintf(w, "\n")

	// Build function maps
	beforeFuncs := make(map[string]*ProfileSample)
	afterFuncs := make(map[string]*ProfileSample)

	for i := range before.Samples {
		sample := &before.Samples[i]
		if len(sample.Location) > 0 {
			beforeFuncs[sample.Location[0].Function] = sample
		}
	}

	for i := range after.Samples {
		sample := &after.Samples[i]
		if len(sample.Location) > 0 {
			afterFuncs[sample.Location[0].Function] = sample
		}
	}

	// Find all functions
	allFuncs := make(map[string]bool)
	for f := range beforeFuncs {
		allFuncs[f] = true
	}
	for f := range afterFuncs {
		allFuncs[f] = true
	}

	// Create comparison data
	type comparison struct {
		function     string
		beforeCycles int64
		afterCycles  int64
		deltaCycles  int64
		deltaPercent float64
	}

	var comparisons []comparison

	for funcName := range allFuncs {
		comp := comparison{function: funcName}

		if before, ok := beforeFuncs[funcName]; ok && len(before.Value) > 1 {
			comp.beforeCycles = before.Value[1]
		}

		if after, ok := afterFuncs[funcName]; ok && len(after.Value) > 1 {
			comp.afterCycles = after.Value[1]
		}

		comp.deltaCycles = comp.afterCycles - comp.beforeCycles
		if comp.beforeCycles > 0 {
			comp.deltaPercent = float64(comp.deltaCycles) / float64(comp.beforeCycles) * 100
		} else if comp.afterCycles > 0 {
			comp.deltaPercent = 100.0
		}

		comparisons = append(comparisons, comp)
	}

	// Sort by absolute delta
	sort.Slice(comparisons, func(i, j int) bool {
		absI := comparisons[i].deltaCycles
		if absI < 0 {
			absI = -absI
		}
		absJ := comparisons[j].deltaCycles
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})

	// Print comparison table
	fmt.Fprintf(w, "%-50s %12s %12s %12s %8s\n",
		"Function", "Before", "After", "Delta", "Change")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 100))

	for i, comp := range comparisons {
		if i >= 20 { // Show top 20
			break
		}

		funcName := comp.function
		if len(funcName) > 50 {
			funcName = funcName[:47] + "..."
		}

		changeStr := fmt.Sprintf("%+.1f%%", comp.deltaPercent)
		if comp.beforeCycles == 0 && comp.afterCycles > 0 {
			changeStr = "NEW"
		} else if comp.beforeCycles > 0 && comp.afterCycles == 0 {
			changeStr = "REMOVED"
		}

		fmt.Fprintf(w, "%-50s %12d %12d %+12d %8s\n",
			funcName, comp.beforeCycles, comp.afterCycles, comp.deltaCycles, changeStr)
	}

	return nil
}
