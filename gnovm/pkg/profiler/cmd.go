package profiler

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// profilerTop shows top functions
func profilerTop(p *ProfilerCLI, arg string) error {
	n := 10
	if arg != "" {
		var err error
		n, err = strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("invalid number: %s", arg)
		}
	}

	// Get filtered samples
	samples := p.getFilteredSamples()
	if len(samples) == 0 {
		fmt.Fprintln(p.out, "No samples found after filtering")
		return nil
	}

	// Aggregate by function
	type funcInfo struct {
		name       string
		flat       int64
		cumulative int64
	}

	funcMap := make(map[string]*funcInfo)
	totalSamples := int64(0)

	for _, sample := range samples {
		if len(sample.Location) == 0 || len(sample.Value) < 2 {
			continue
		}

		value := sample.Value[p.sampleIndex]
		if p.sampleIndex >= len(sample.Value) {
			value = sample.Value[0]
		}

		totalSamples += value

		// Add to flat count for the top function
		topFunc := sample.Location[0].Function
		if info, ok := funcMap[topFunc]; ok {
			info.flat += value
		} else {
			funcMap[topFunc] = &funcInfo{
				name: topFunc,
				flat: value,
			}
		}

		// Add to cumulative for all functions in stack
		seen := make(map[string]bool)
		for _, loc := range sample.Location {
			if !seen[loc.Function] {
				seen[loc.Function] = true
				if info, ok := funcMap[loc.Function]; ok {
					info.cumulative += value
				} else {
					// Only create entry if function is not hidden
					hidden := false
					for _, hide := range p.hideFunc {
						if strings.Contains(loc.Function, hide) {
							hidden = true
							break
						}
					}
					if !hidden {
						funcMap[loc.Function] = &funcInfo{
							name:       loc.Function,
							cumulative: value,
						}
					}
				}
			}
		}
	}

	// Convert to slice and sort, filtering out hidden functions
	funcs := make([]*funcInfo, 0, len(funcMap))
	for _, info := range funcMap {
		// Skip functions that are in the hide list
		hidden := false
		for _, hide := range p.hideFunc {
			if strings.Contains(info.name, hide) {
				hidden = true
				break
			}
		}
		if !hidden {
			funcs = append(funcs, info)
		}
	}

	if p.cumulative && !p.flat {
		sort.Slice(funcs, func(i, j int) bool {
			return funcs[i].cumulative > funcs[j].cumulative
		})
	} else {
		sort.Slice(funcs, func(i, j int) bool {
			return funcs[i].flat > funcs[j].flat
		})
	}

	// Print header
	fmt.Fprintf(p.out, "Showing nodes accounting for %s, %s of %s total\n",
		p.formatValue(totalSamples), "100%", p.formatValue(totalSamples))

	if p.focusFunc != "" {
		fmt.Fprintf(p.out, "Focused on: %s\n", p.focusFunc)
	}
	if len(p.hideFunc) > 0 {
		fmt.Fprintf(p.out, "Hiding: %s\n", strings.Join(p.hideFunc, ", "))
	}
	if len(p.ignoreFunc) > 0 {
		fmt.Fprintf(p.out, "Ignoring: %s\n", strings.Join(p.ignoreFunc, ", "))
	}

	fmt.Fprintf(p.out, "      flat  flat%%   sum%%        cum   cum%%\n")

	// Print functions
	sum := int64(0)
	for i, info := range funcs {
		if i >= n {
			break
		}

		sum += info.flat
		flatPct := float64(info.flat) / float64(totalSamples) * 100
		sumPct := float64(sum) / float64(totalSamples) * 100
		cumPct := float64(info.cumulative) / float64(totalSamples) * 100

		fmt.Fprintf(p.out, "%10s %6.2f%% %6.2f%% %10s %6.2f%%  %s\n",
			p.formatValue(info.flat), flatPct, sumPct,
			p.formatValue(info.cumulative), cumPct,
			p.truncateFuncName(info.name, 50))
	}

	return nil
}

// profilerList shows annotated source for a function
func profilerList(p *ProfilerCLI, arg string) error {
	if arg == "" {
		return errors.New("function name required")
	}

	// If arg is ".", use the function from the last list command
	if arg == "." && p.lastCmd == "list" && p.lastArg != "" && p.lastArg != "." {
		arg = p.lastArg
	}

	return p.profile.WriteFunctionList(p.out, arg, p.store)
}

// profilerTree shows the call tree
// func profilerTree(p *ProfilerCLI, arg string) error {
// 	samples := p.getFilteredSamples()
// 	p.profile.mu.Lock()
// 	p.profile.Samples = samples
// 	p.profile.mu.Unlock()

// 	return p.profile.WriteCallTree(p.out)
// }

// profilerFocus sets function focus
func profilerFocus(p *ProfilerCLI, arg string) error {
	if arg == "" {
		return errors.New("function name required")
	}
	p.focusFunc = arg
	fmt.Fprintf(p.out, "Focused on: %s\n", arg)
	return nil
}

// profilerIgnore adds function to ignore list
func profilerIgnore(p *ProfilerCLI, arg string) error {
	if arg == "" {
		return errors.New("function name required")
	}
	p.ignoreFunc = append(p.ignoreFunc, arg)
	fmt.Fprintf(p.out, "Ignoring: %s\n", arg)
	return nil
}

// profilerHide adds function to hide list
func profilerHide(p *ProfilerCLI, arg string) error {
	if arg == "" {
		return errors.New("function name required")
	}
	p.hideFunc = append(p.hideFunc, arg)
	fmt.Fprintf(p.out, "Hiding: %s\n", arg)
	return nil
}

// profilerShow shows current settings
func profilerShow(p *ProfilerCLI, arg string) error {
	fmt.Fprintf(p.out, "Current settings:\n")
	fmt.Fprintf(p.out, "  Focus: %s\n", p.focusFunc)
	fmt.Fprintf(p.out, "  Hide: %s\n", strings.Join(p.hideFunc, ", "))
	fmt.Fprintf(p.out, "  Ignore: %s\n", strings.Join(p.ignoreFunc, ", "))
	fmt.Fprintf(p.out, "  Cumulative: %v\n", p.cumulative)
	fmt.Fprintf(p.out, "  Flat: %v\n", p.flat)
	fmt.Fprintf(p.out, "  Addresses: %v\n", p.addresses)
	fmt.Fprintf(p.out, "  Lines: %v\n", p.lines)
	fmt.Fprintf(p.out, "  Node count: %d\n", p.nodeCount)
	fmt.Fprintf(p.out, "  Unit: %s\n", p.unit)
	fmt.Fprintf(p.out, "  Sample index: %d\n", p.sampleIndex)
	return nil
}

// profilerReset resets all settings
func profilerReset(p *ProfilerCLI, arg string) error {
	p.focusFunc = ""
	p.hideFunc = []string{}
	p.ignoreFunc = []string{}
	p.minSamples = 0
	fmt.Fprintln(p.out, "Reset all focus/ignore/hide settings")
	return nil
}

// profilerSample selects sample type
func profilerSample(p *ProfilerCLI, arg string) error {
	if arg == "" {
		return errors.New("sample index required")
	}

	index, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("invalid sample index: %s", arg)
	}

	p.sampleIndex = index
	fmt.Fprintf(p.out, "Selected sample index: %d\n", index)
	return nil
}

// profilerSave saves profile to file
func profilerSave(p *ProfilerCLI, arg string) error {
	if arg == "" {
		return errors.New("filename required")
	}

	file, err := os.Create(arg)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Determine format from extension
	ext := filepath.Ext(arg)
	switch ext {
	case ".html":
		_, err = p.profile.WriteTo(file) // Default format for .html files
	// case ".tree":
	// 	err = p.profile.WriteCallTree(file)
	// case ".top":
	// 	err = p.profile.WriteTopList(file)
	default:
		_, err = p.profile.WriteTo(file)
	}

	if err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}

	fmt.Fprintf(p.out, "Saved profile to %s\n", arg)
	return nil
}

// profilerHelp shows help
func profilerHelp(p *ProfilerCLI, arg string) error {
	if arg != "" {
		c, ok := profilerCmds[arg]
		if !ok {
			return fmt.Errorf("unknown command: %s", arg)
		}

		fmt.Fprintf(p.out, "%-25s %s\n", c.usage, c.short)
		if c.long != "" {
			fmt.Fprintf(p.out, "\n%s\n", c.long)
		}
		return nil
	}

	fmt.Fprintln(p.out, "Commands:")
	for _, name := range profilerCmdNames {
		c := profilerCmds[name]
		fmt.Fprintf(p.out, "  %-25s %s\n", c.usage, c.short)
	}
	fmt.Fprintln(p.out, "\nType 'help <command>' for more information")
	return nil
}

// profilerQuit exits the profiler
func profilerQuit(p *ProfilerCLI, arg string) error {
	return io.EOF
}

// Toggle commands
func profilerToggleCum(p *ProfilerCLI, arg string) error {
	p.cumulative = !p.cumulative
	fmt.Fprintf(p.out, "Cumulative mode: %v\n", p.cumulative)
	return nil
}

func profilerToggleFlat(p *ProfilerCLI, arg string) error {
	p.flat = !p.flat
	fmt.Fprintf(p.out, "Flat mode: %v\n", p.flat)
	return nil
}

func profilerToggleAddr(p *ProfilerCLI, arg string) error {
	p.addresses = !p.addresses
	fmt.Fprintf(p.out, "Show addresses: %v\n", p.addresses)
	return nil
}

func profilerToggleLines(p *ProfilerCLI, arg string) error {
	p.lines = !p.lines
	fmt.Fprintf(p.out, "Show lines: %v\n", p.lines)
	return nil
}

func profilerNodeCount(p *ProfilerCLI, arg string) error {
	if arg == "" {
		return errors.New("node count required")
	}

	n, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("invalid node count: %s", arg)
	}

	p.nodeCount = n
	fmt.Fprintf(p.out, "Node count set to: %d\n", n)
	return nil
}

func profilerUnit(p *ProfilerCLI, arg string) error {
	if arg == "" {
		return errors.New("unit required")
	}

	p.unit = arg
	fmt.Fprintf(p.out, "Unit set to: %s\n", arg)
	return nil
}

// Helper functions

// getFilteredSamples returns samples after applying focus/ignore/hide filters
func (p *ProfilerCLI) getFilteredSamples() []ProfileSample {
	p.profile.mu.RLock()
	defer p.profile.mu.RUnlock()

	filtered := make([]ProfileSample, 0)

	for _, sample := range p.profile.Samples {
		// Apply focus filter
		if p.focusFunc != "" {
			found := false
			for _, loc := range sample.Location {
				if strings.Contains(loc.Function, p.focusFunc) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Apply ignore filter
		skip := false
		for _, ignore := range p.ignoreFunc {
			for _, loc := range sample.Location {
				if strings.Contains(loc.Function, ignore) {
					skip = true
					break
				}
			}
			if skip {
				break
			}
		}
		if skip {
			continue
		}

		// Apply hide filter - remove hidden functions from stack
		newLocs := make([]ProfileLocation, 0, len(sample.Location))
		for _, loc := range sample.Location {
			hidden := false
			for _, hide := range p.hideFunc {
				if strings.Contains(loc.Function, hide) {
					hidden = true
					break
				}
			}
			if !hidden {
				newLocs = append(newLocs, loc)
			}
		}

		if len(newLocs) > 0 {
			newSample := sample
			newSample.Location = newLocs
			filtered = append(filtered, newSample)
		}
	}

	return filtered
}

// formatValue formats a value based on the current unit setting
func (p *ProfilerCLI) formatValue(v int64) string {
	switch p.unit {
	case "ms":
		return fmt.Sprintf("%.2fms", float64(v)/1e6)
	case "us":
		return fmt.Sprintf("%.2fus", float64(v)/1e3)
	case "cycles":
		return fmt.Sprintf("%d", v)
	case "auto":
		if v > 1e9 {
			return fmt.Sprintf("%.2fs", float64(v)/1e9)
		} else if v > 1e6 {
			return fmt.Sprintf("%.2fms", float64(v)/1e6)
		} else if v > 1e3 {
			return fmt.Sprintf("%.2fus", float64(v)/1e3)
		}
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%d", v)
	}
}

// truncateFuncName truncates long function names
func (p *ProfilerCLI) truncateFuncName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}

	// Try to keep the most important part (function name at the end)
	parts := strings.Split(name, ".")
	if len(parts) > 1 {
		funcName := parts[len(parts)-1]
		if len(funcName) < maxLen-3 {
			prefix := name[:maxLen-len(funcName)-3]
			return prefix + "..." + funcName
		}
	}

	return name[:maxLen-3] + "..."
}
