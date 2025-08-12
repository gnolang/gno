package profiler

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// ProfileFormat represents the output format for profile data
type ProfileFormat int

const (
	FormatText ProfileFormat = iota
	FormatCallTree
	FormatTopList
	FormatJSON
)

// node represents a call tree node
type node struct {
	name     string
	cycles   int64
	calls    int64
	children map[string]*node
}

// WriteFormat writes the profile in the specified format
func (p *Profile) WriteFormat(w io.Writer, format ProfileFormat) error {
	switch format {
	case FormatCallTree:
		return p.WriteCallTree(w)
	case FormatTopList:
		return p.WriteTopList(w)
	case FormatJSON:
		return p.WriteJSON(w)
	default:
		return p.WriteTo(w)
	}
}

// WriteCallTree writes a hierarchical call tree
func (p *Profile) WriteCallTree(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	root := &node{
		name:     "<root>",
		children: make(map[string]*node),
	}

	// Process stack samples to build tree
	for _, sample := range p.Samples {
		if len(sample.Location) <= 1 {
			continue // Skip non-stack samples
		}

		cycles := int64(0)
		if len(sample.Value) > 1 {
			cycles = sample.Value[1]
		}

		// Traverse/build tree from root to leaf
		current := root
		for i := len(sample.Location) - 1; i >= 0; i-- {
			funcName := sample.Location[i].Function
			if child, ok := current.children[funcName]; ok {
				child.cycles += cycles
				child.calls++
				current = child
			} else {
				newNode := &node{
					name:     funcName,
					cycles:   cycles,
					calls:    1,
					children: make(map[string]*node),
				}
				current.children[funcName] = newNode
				current = newNode
			}
		}
	}

	// Print tree
	fmt.Fprintf(w, "Call Tree (CPU Cycles)\n")
	fmt.Fprintf(w, "======================\n\n")

	totalCycles := p.totalCycles()
	printNode(w, root, "", true, totalCycles, 0)

	return nil
}

// printNode recursively prints a call tree node
func printNode(w io.Writer, n *node, prefix string, isLast bool, totalCycles int64, depth int) {
	if n.name != "<root>" {
		// Print current node
		percent := float64(0)
		if totalCycles > 0 {
			percent = float64(n.cycles) / float64(totalCycles) * 100
		}

		connector := "├─"
		if isLast {
			connector = "└─"
		}
		if depth == 0 {
			connector = ""
		}

		fmt.Fprintf(w, "%s%s %s: %d cycles (%.1f%%), %d calls\n",
			prefix, connector, n.name, n.cycles, percent, n.calls)
	}

	// Sort children by cycles
	children := make([]*node, 0, len(n.children))
	for _, child := range n.children {
		children = append(children, child)
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].cycles > children[j].cycles
	})

	// Print children
	childPrefix := prefix
	if n.name != "<root>" {
		if isLast {
			childPrefix += "  "
		} else {
			childPrefix += "│ "
		}
	}

	for i, child := range children {
		isLastChild := i == len(children)-1
		printNode(w, child, childPrefix, isLastChild, totalCycles, depth+1)
	}
}

// WriteTopList writes a sorted list of top functions
func (p *Profile) WriteTopList(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Aggregate by function
	type funcStat struct {
		name       string
		flat       int64
		cumulative int64
		calls      int64
	}

	funcMap := make(map[string]*funcStat)

	// Process samples
	for _, sample := range p.Samples {
		if len(sample.Location) == 0 {
			continue
		}

		// Get metrics
		cycles := int64(0)
		if len(sample.Value) > 1 {
			cycles = sample.Value[1]
		}

		calls := int64(1)
		if callsVal, ok := sample.NumLabel["calls"]; ok && len(callsVal) > 0 {
			calls = callsVal[0]
		}

		flatCycles := int64(0)
		if flatVal, ok := sample.NumLabel["flat_cycles"]; ok && len(flatVal) > 0 {
			flatCycles = flatVal[0]
		} else {
			flatCycles = cycles
		}

		cumCycles := int64(0)
		if cumVal, ok := sample.NumLabel["cum_cycles"]; ok && len(cumVal) > 0 {
			cumCycles = cumVal[0]
		} else {
			cumCycles = cycles
		}

		funcName := sample.Location[0].Function
		if stat, ok := funcMap[funcName]; ok {
			// Update existing
			stat.flat = max(stat.flat, flatCycles)
			stat.cumulative = max(stat.cumulative, cumCycles)
			stat.calls = max(stat.calls, calls)
		} else {
			// Create new
			funcMap[funcName] = &funcStat{
				name:       funcName,
				flat:       flatCycles,
				cumulative: cumCycles,
				calls:      calls,
			}
		}
	}

	// Convert to slice and sort
	funcs := make([]*funcStat, 0, len(funcMap))
	for _, stat := range funcMap {
		funcs = append(funcs, stat)
	}
	sort.Slice(funcs, func(i, j int) bool {
		return funcs[i].cumulative > funcs[j].cumulative
	})

	// Print header
	totalCycles := p.totalCycles()
	fmt.Fprintf(w, "Top Functions by Cumulative Time\n")
	fmt.Fprintf(w, "Total cycles: %d\n\n", totalCycles)
	fmt.Fprintf(w, "%-6s %-12s %-8s %-12s %-8s %-8s %s\n",
		"Rank", "Cumulative", "Cum%", "Flat", "Flat%", "Calls", "Function")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 80))

	// Print functions
	for i, stat := range funcs {
		if i >= 50 { // Limit to top 50
			break
		}

		cumPercent := float64(0)
		flatPercent := float64(0)
		if totalCycles > 0 {
			cumPercent = float64(stat.cumulative) / float64(totalCycles) * 100
			flatPercent = float64(stat.flat) / float64(totalCycles) * 100
		}

		fmt.Fprintf(w, "%-6d %-12d %-7.2f%% %-12d %-7.2f%% %-8d %s\n",
			i+1, stat.cumulative, cumPercent, stat.flat, flatPercent, stat.calls, stat.name)
	}

	// Print summary
	fmt.Fprintf(w, "\nShowing top %d of %d functions\n", min(50, len(funcs)), len(funcs))

	return nil
}
