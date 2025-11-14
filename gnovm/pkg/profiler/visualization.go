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
	_ ProfileFormat = iota
	FormatText
	FormatCallTree
	FormatTopList
	FormatJSON
)

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
		_, err := p.WriteTo(w)
		return err
	}
}

// WriteCallTree writes a hierarchical call tree
func (p *Profile) WriteCallTree(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.Type == ProfileGas {
		fmt.Fprintf(w, "Call Tree (Gas Usage)\n")
		fmt.Fprintf(w, "=====================\n\n")
	} else {
		fmt.Fprintf(w, "Call Tree (CPU Cycles)\n")
		fmt.Fprintf(w, "======================\n\n")
	}

	if p.CallTree == nil {
		fmt.Fprintf(w, "No call tree data available.\n")
		return nil
	}

	totalCycles := p.totalCycles()
	totalGas := p.totalGas()
	printVisualizationNode(w, p, p.CallTree, "", true, totalCycles, totalGas, 0)

	return nil
}

func printVisualizationNode(w io.Writer, p *Profile, node *CallTreeNode, prefix string, isLast bool, totalCycles, totalGas int64, depth int) {
	if node == nil {
		return
	}
	name := frameName(p, node.FrameID)
	if node.FrameID == invalidFrameID {
		name = "<root>"
	}

	if name != "<root>" {
		connector := "├─"
		if isLast {
			connector = "└─"
		}
		if depth == 0 {
			connector = ""
		}

		if p.Type == ProfileGas {
			percent := percent(node.TotalGas, totalGas)
			fmt.Fprintf(w, "%s%s %s: %d gas (%.1f%%), %d calls\n",
				prefix, connector, name, node.TotalGas, percent, node.Calls)
		} else {
			percent := percent(node.TotalCycles, totalCycles)
			fmt.Fprintf(w, "%s%s %s: %d cycles (%.1f%%), %d calls\n",
				prefix, connector, name, node.TotalCycles, percent, node.Calls)
		}
	}

	if len(node.Children) == 0 {
		return
	}

	childPrefix := prefix
	if name != "<root>" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		printVisualizationNode(w, p, child, childPrefix, isLastChild, totalCycles, totalGas, depth+1)
	}
}

// WriteTopList writes a sorted list of top functions
func (p *Profile) WriteTopList(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.Functions) == 0 {
		fmt.Fprintf(w, "No function statistics available.\n")
		return nil
	}

	funcs := append([]*FunctionStat(nil), p.Functions...)

	var total int64
	var label string
	var flatMetric func(*FunctionStat) int64
	var cumMetric func(*FunctionStat) int64

	switch p.Type {
	case ProfileGas:
		label = "Gas"
		total = p.totalGas()
		flatMetric = func(stat *FunctionStat) int64 { return stat.SelfGas }
		cumMetric = func(stat *FunctionStat) int64 { return stat.TotalGas }
	default:
		label = "CPU Cycles"
		total = p.totalCycles()
		flatMetric = func(stat *FunctionStat) int64 { return stat.SelfCycles }
		cumMetric = func(stat *FunctionStat) int64 { return stat.TotalCycles }
	}

	sort.Slice(funcs, func(i, j int) bool {
		return cumMetric(funcs[i]) > cumMetric(funcs[j])
	})

	fmt.Fprintf(w, "Top Functions by %s\n", label)
	fmt.Fprintf(w, "Total: %d\n\n", total)
	fmt.Fprintf(w, "%-6s %-12s %-8s %-12s %-8s %-8s %-20s %s\n",
		"Rank", "Cumulative", "Cum%", "Flat", "Flat%", "Calls", "Bar", "Function")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 100))

	for i, stat := range funcs {
		if i >= 50 { // Limit to top 50
			break
		}

		flat := flatMetric(stat)
		cum := cumMetric(stat)
		cumPercent := percent(cum, total)
		flatPercent := percent(flat, total)

		// Create visual bar
		barLength := min(int(cumPercent/5), 20) // 20 chars max
		bar := strings.Repeat("█", barLength)

		// Truncate function name if too long
		funcName := stat.Name
		if len(funcName) > 50 {
			funcName = funcName[:47] + "..."
		}

		fmt.Fprintf(w, "%-6d %-12d %-7.2f%% %-12d %-7.2f%% %-8d %-20s %s\n",
			i+1, cum, cumPercent, flat, flatPercent, stat.CallCount, bar, funcName)
	}

	// Print summary
	fmt.Fprintf(w, "\nShowing top %d of %d functions\n", min(50, len(funcs)), len(funcs))

	return nil
}

func frameName(p *Profile, id FrameID) string {
	if id < 0 || int(id) >= len(p.Frames) {
		return ""
	}
	return p.Frames[int(id)].Function
}
