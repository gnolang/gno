package benchops

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// Results contains the profiling results after Stop() is called.
type Results struct {
	Duration      time.Duration
	StartTime     time.Time
	EndTime       time.Time
	TimingEnabled bool // Whether timing was enabled for this run
	OpStats       map[string]*OpStat
	StoreStats    map[string]*StoreStat
	NativeStats   map[string]*TimingStat
	LocationStats []*LocationStat        `json:"LocationStats,omitempty"`
	SubOpStats    map[string]*TimingStat `json:"SubOpStats,omitempty"`
	VarStats      []*VarStat             `json:"VarStats,omitempty"`
	StackSamples  []*StackSample         `json:"StackSamples,omitempty"`
}

// SectionFlags controls which sections are included in WriteGolden output.
type SectionFlags uint8

const (
	SectionOpcodes SectionFlags = 1 << iota
	SectionStore
	SectionNative
	SectionHotSpots
	SectionSubOps
	SectionVars

	SectionAll = SectionOpcodes | SectionStore | SectionNative | SectionHotSpots | SectionSubOps | SectionVars
)

// Has returns true if the flag is set.
func (s SectionFlags) Has(flag SectionFlags) bool {
	if s == 0 {
		return true
	}
	return s&flag != 0
}

// ParseSectionFlags parses a comma-separated list of section names.
func ParseSectionFlags(s string) (SectionFlags, error) {
	if s == "" || s == "all" {
		return 0, nil
	}

	var flags SectionFlags
	for _, name := range strings.Split(s, ",") {
		name = strings.TrimSpace(strings.ToLower(name))
		switch name {
		case "opcodes":
			flags |= SectionOpcodes
		case "store":
			flags |= SectionStore
		case "native":
			flags |= SectionNative
		case "hotspots":
			flags |= SectionHotSpots
		case "subops":
			flags |= SectionSubOps
		case "vars":
			flags |= SectionVars
		case "all":
			return 0, nil
		default:
			return 0, fmt.Errorf("unknown section %q", name)
		}
	}
	return flags, nil
}

// Capacity hints for result maps, based on typical number of defined operations.
const (
	opStatsHint     = 48 // ~40 defined opcodes + headroom
	storeStatsHint  = 16 // ~10 store operations
	nativeStatsHint = 16 // ~10 native operations
	subOpStatsHint  = 16 // ~13 sub-operations
)

// buildResults creates Results from the profiler's internal state.
func (p *Profiler) buildResults() *Results {
	r := &Results{
		Duration:      p.stopTime.Sub(p.startTime),
		StartTime:     p.startTime,
		EndTime:       p.stopTime,
		TimingEnabled: p.timingEnabled,
		OpStats:       make(map[string]*OpStat, opStatsHint),
		StoreStats:    make(map[string]*StoreStat, storeStatsHint),
		NativeStats:   make(map[string]*TimingStat, nativeStatsHint),
		SubOpStats:    make(map[string]*TimingStat, subOpStatsHint),
	}

	// Build op stats (copy non-zero entries)
	for i := range maxOpCodes {
		s := &p.opStats[i]
		if s.Count == 0 {
			continue
		}
		name := Op(i).String()
		r.OpStats[name] = &OpStat{
			TimingStat: s.TimingStat,
			Gas:        s.Gas,
		}
	}

	// Build store stats
	for i := range maxOpCodes {
		s := &p.storeStats[i]
		if s.Count == 0 {
			continue
		}
		name := StoreOp(i).String()
		r.StoreStats[name] = &StoreStat{
			TimingStat: s.TimingStat,
			TotalSize:  s.TotalSize,
		}
	}

	// Build native stats
	for i := range maxOpCodes {
		s := &p.nativeStats[i]
		if s.Count == 0 {
			continue
		}
		name := NativeOp(i).String()
		clone := *s
		r.NativeStats[name] = &clone
	}

	// Build location stats
	if len(p.locationStats) > 0 {
		r.LocationStats = make([]*LocationStat, 0, len(p.locationStats))
		for _, s := range p.locationStats {
			r.LocationStats = append(r.LocationStats, &LocationStat{
				File:     s.File,
				Line:     s.Line,
				FuncName: s.FuncName,
				PkgPath:  s.PkgPath,
				Count:    s.Count,
				TotalNs:  s.TotalNs,
				Gas:      s.Gas,
			})
		}
		slices.SortFunc(r.LocationStats, func(a, b *LocationStat) int {
			return cmp.Compare(b.Gas, a.Gas) // descending
		})
	}

	// Build sub-op stats
	for i := range maxSubOps {
		s := &p.subOpStats[i]
		if s.Count == 0 {
			continue
		}
		name := SubOp(i).String()
		clone := *s
		r.SubOpStats[name] = &clone
	}

	// Build variable stats
	if len(p.varStats) > 0 {
		r.VarStats = make([]*VarStat, 0, len(p.varStats))
		for _, s := range p.varStats {
			r.VarStats = append(r.VarStats, &VarStat{
				Name:       s.Name,
				File:       s.File,
				Line:       s.Line,
				Index:      s.Index,
				TimingStat: s.TimingStat,
			})
		}
		slices.SortFunc(r.VarStats, func(a, b *VarStat) int {
			return cmp.Compare(b.TotalNs, a.TotalNs) // descending
		})
	}

	// Build stack samples from aggregated map
	if p.stackEnabled && len(p.stackSampleAgg) > 0 {
		r.StackSamples = make([]*StackSample, 0, len(p.stackSampleAgg))
		for keyStr, sample := range p.stackSampleAgg {
			var frames []StackFrame
			for _, part := range strings.Split(keyStr, "|") {
				frame := parseFrameFromKey(part)
				frames = append(frames, frame)
			}
			r.StackSamples = append(r.StackSamples, &StackSample{
				Stack:      frames,
				Gas:        sample.gas,
				DurationNs: sample.durationNs,
				Count:      sample.count,
			})
		}
	}

	return r
}

// WriteJSON writes the results as compact JSON.
func (r *Results) WriteJSON(w io.Writer) error {
	if r == nil {
		return nil
	}
	return json.NewEncoder(w).Encode(r)
}

// WriteReport writes a human-readable summary.
func (r *Results) WriteReport(w io.Writer, topN int) error {
	if r == nil {
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)

	// Calculate totals for summary
	var totalGas, totalOps int64
	for _, stat := range r.OpStats {
		totalGas += stat.Gas
		totalOps += stat.Count
	}
	var totalStoreOps int64
	for _, stat := range r.StoreStats {
		totalStoreOps += stat.Count
	}

	fmt.Fprintf(tw, "Profiling Results\n")
	fmt.Fprintf(tw, "═════════════════\n")
	fmt.Fprintf(tw, "Duration:\t%v\t\n", r.Duration)
	fmt.Fprintf(tw, "Start:\t%v\t\n", r.StartTime.Format(time.RFC3339))
	fmt.Fprintf(tw, "End:\t%v\t\n", r.EndTime.Format(time.RFC3339))
	fmt.Fprintf(tw, "Timing:\t%v\t\n", r.TimingEnabled)
	fmt.Fprintf(tw, "Total Gas:\t%d\t\n", totalGas)
	fmt.Fprintf(tw, "Total Ops:\t%d\t\n", totalOps)
	fmt.Fprintf(tw, "Total Store Ops:\t%d\t\n\n", totalStoreOps)

	r.writeOpStats(tw, topN)
	r.writeStoreStats(tw, topN)
	r.writeNativeStats(tw, topN)
	r.writeHotSpots(tw, topN)
	r.writeSubOpStats(tw, topN)
	r.writeVarStats(tw, topN)

	return tw.Flush()
}

func (r *Results) writeOpStats(tw *tabwriter.Writer, topN int) {
	if len(r.OpStats) == 0 {
		return
	}

	// Calculate total gas for percentage
	var totalGas int64
	for _, stat := range r.OpStats {
		totalGas += stat.Gas
	}

	if r.TimingEnabled {
		fmt.Fprintf(tw, "Opcode Statistics (by total time)\n")
		fmt.Fprintf(tw, "──────────────────────────────────\n")
		fmt.Fprintf(tw, "Opcode\tCount\tTotal\tAvg\tStdDev\tMin\tMax\tGas\t%%\t\n")

		sorted := sortedMap(r.OpStats, func(s *OpStat) int64 { return s.TotalNs })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			pct := float64(0)
			if totalGas > 0 {
				pct = float64(kv.val.Gas) * 100 / float64(totalGas)
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%v\t%d\t%.1f%%\t\n",
				kv.key, kv.val.Count,
				time.Duration(kv.val.TotalNs), time.Duration(kv.val.AvgNs()),
				time.Duration(kv.val.StdDevNs()),
				time.Duration(kv.val.MinNs), time.Duration(kv.val.MaxNs),
				kv.val.Gas, pct)
		}
	} else {
		fmt.Fprintf(tw, "Opcode Statistics (by gas)\n")
		fmt.Fprintf(tw, "──────────────────────────\n")
		fmt.Fprintf(tw, "Opcode\tCount\tGas\t%%\t\n")

		sorted := sortedMap(r.OpStats, func(s *OpStat) int64 { return s.Gas })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			pct := float64(0)
			if totalGas > 0 {
				pct = float64(kv.val.Gas) * 100 / float64(totalGas)
			}
			fmt.Fprintf(tw, "%s\t%d\t%d\t%.1f%%\t\n", kv.key, kv.val.Count, kv.val.Gas, pct)
		}
	}
	fmt.Fprintln(tw)
}

func (r *Results) writeStoreStats(tw *tabwriter.Writer, topN int) {
	if len(r.StoreStats) == 0 {
		return
	}
	if r.TimingEnabled {
		fmt.Fprintf(tw, "Store Statistics (by total time)\n")
		fmt.Fprintf(tw, "─────────────────────────────────\n")
		fmt.Fprintf(tw, "Operation\tCount\tTotal\tAvg\tStdDev\tMin\tMax\tTotal Size\tAvg Size\t\n")

		sorted := sortedMap(r.StoreStats, func(s *StoreStat) int64 { return s.TotalNs })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%v\t%d\t%d\t\n",
				kv.key, kv.val.Count,
				time.Duration(kv.val.TotalNs), time.Duration(kv.val.AvgNs()),
				time.Duration(kv.val.StdDevNs()),
				time.Duration(kv.val.MinNs), time.Duration(kv.val.MaxNs),
				kv.val.TotalSize, kv.val.AvgSize())
		}
	} else {
		fmt.Fprintf(tw, "Store Statistics (by count)\n")
		fmt.Fprintf(tw, "───────────────────────────\n")
		fmt.Fprintf(tw, "Operation\tCount\tTotal Size\tAvg Size\t\n")

		sorted := sortedMap(r.StoreStats, func(s *StoreStat) int64 { return s.Count })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t\n",
				kv.key, kv.val.Count, kv.val.TotalSize, kv.val.AvgSize())
		}
	}
	fmt.Fprintln(tw)
}

func (r *Results) writeNativeStats(tw *tabwriter.Writer, topN int) {
	if len(r.NativeStats) == 0 {
		return
	}
	if r.TimingEnabled {
		fmt.Fprintf(tw, "Native Statistics (by total time)\n")
		fmt.Fprintf(tw, "──────────────────────────────────\n")
		fmt.Fprintf(tw, "Function\tCount\tTotal\tAvg\tStdDev\tMin\tMax\t\n")

		sorted := sortedMap(r.NativeStats, func(s *TimingStat) int64 { return s.TotalNs })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%v\t\n",
				kv.key, kv.val.Count,
				time.Duration(kv.val.TotalNs), time.Duration(kv.val.AvgNs()),
				time.Duration(kv.val.StdDevNs()),
				time.Duration(kv.val.MinNs), time.Duration(kv.val.MaxNs))
		}
	} else {
		fmt.Fprintf(tw, "Native Statistics (by count)\n")
		fmt.Fprintf(tw, "────────────────────────────\n")
		fmt.Fprintf(tw, "Function\tCount\t\n")

		sorted := sortedMap(r.NativeStats, func(s *TimingStat) int64 { return s.Count })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t\n", kv.key, kv.val.Count)
		}
	}
	fmt.Fprintln(tw)
}

func (r *Results) writeHotSpots(tw *tabwriter.Writer, topN int) {
	if len(r.LocationStats) == 0 {
		return
	}

	// Calculate total gas for percentage
	var totalGas int64
	for _, loc := range r.LocationStats {
		totalGas += loc.Gas
	}

	fmt.Fprintf(tw, "Hot Spots (by gas)\n")
	fmt.Fprintf(tw, "──────────────────\n")
	if r.TimingEnabled {
		fmt.Fprintf(tw, "Package\tLocation\tFunc\tCount\tTotal\tGas\t%%\t\n")
	} else {
		fmt.Fprintf(tw, "Package\tLocation\tFunc\tCount\tGas\t%%\t\n")
	}

	for i, loc := range r.LocationStats {
		if topN > 0 && i >= topN {
			break
		}
		location := fmt.Sprintf("%s:%d", loc.File, loc.Line)
		funcName := cmp.Or(loc.FuncName, "-")
		pkgPath := cmp.Or(abbreviatePkgPath(loc.PkgPath), "-")

		// Calculate percentage
		pct := float64(0)
		if totalGas > 0 {
			pct = float64(loc.Gas) * 100 / float64(totalGas)
		}

		if r.TimingEnabled {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%v\t%d\t%.1f%%\t\n",
				pkgPath, location, funcName, loc.Count,
				time.Duration(loc.TotalNs), loc.Gas, pct)
		} else {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\t%.1f%%\t\n",
				pkgPath, location, funcName, loc.Count, loc.Gas, pct)
		}
	}
	fmt.Fprintln(tw)
}

// abbreviatePkgPath shortens a package path for display.
// e.g., "gno.land/r/demo/boards" -> "r/demo/boards"
func abbreviatePkgPath(pkg string) string {
	for _, prefix := range []string{"gno.land/", "github.com/"} {
		if rest, ok := strings.CutPrefix(pkg, prefix); ok {
			return rest
		}
	}
	return pkg
}

func (r *Results) writeSubOpStats(tw *tabwriter.Writer, topN int) {
	if len(r.SubOpStats) == 0 {
		return
	}
	if r.TimingEnabled {
		fmt.Fprintf(tw, "Sub-Operation Statistics (by total time)\n")
		fmt.Fprintf(tw, "─────────────────────────────────────────\n")
		fmt.Fprintf(tw, "SubOp\tCount\tTotal\tAvg\tStdDev\tMin\tMax\t\n")

		sorted := sortedMap(r.SubOpStats, func(s *TimingStat) int64 { return s.TotalNs })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%v\t\n",
				kv.key, kv.val.Count,
				time.Duration(kv.val.TotalNs), time.Duration(kv.val.AvgNs()),
				time.Duration(kv.val.StdDevNs()),
				time.Duration(kv.val.MinNs), time.Duration(kv.val.MaxNs))
		}
	} else {
		fmt.Fprintf(tw, "Sub-Operation Statistics (by count)\n")
		fmt.Fprintf(tw, "────────────────────────────────────\n")
		fmt.Fprintf(tw, "SubOp\tCount\t\n")

		sorted := sortedMap(r.SubOpStats, func(s *TimingStat) int64 { return s.Count })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t\n", kv.key, kv.val.Count)
		}
	}
	fmt.Fprintln(tw)
}

func (r *Results) writeVarStats(tw *tabwriter.Writer, topN int) {
	if len(r.VarStats) == 0 {
		return
	}
	if r.TimingEnabled {
		fmt.Fprintf(tw, "Variable Statistics (by total time)\n")
		fmt.Fprintf(tw, "────────────────────────────────────\n")
		fmt.Fprintf(tw, "Location\tVar\tCount\tTotal\tAvg\tStdDev\tMin\tMax\t\n")

		for i, v := range r.VarStats {
			if topN > 0 && i >= topN {
				break
			}
			location := fmt.Sprintf("%s:%d", v.File, v.Line)
			fmt.Fprintf(tw, "%s\t%s\t%d\t%v\t%v\t%v\t%v\t%v\t\n",
				location, v.DisplayName(), v.Count,
				time.Duration(v.TotalNs), time.Duration(v.AvgNs()),
				time.Duration(v.StdDevNs()),
				time.Duration(v.MinNs), time.Duration(v.MaxNs))
		}
	} else {
		fmt.Fprintf(tw, "Variable Statistics (by count)\n")
		fmt.Fprintf(tw, "───────────────────────────────\n")
		fmt.Fprintf(tw, "Location\tVar\tCount\t\n")

		for i, v := range r.VarStats {
			if topN > 0 && i >= topN {
				break
			}
			location := fmt.Sprintf("%s:%d", v.File, v.Line)
			fmt.Fprintf(tw, "%s\t%s\t%d\t\n", location, v.DisplayName(), v.Count)
		}
	}
	fmt.Fprintln(tw)
}

// WriteGolden writes a deterministic summary for test golden files.
func (r *Results) WriteGolden(w io.Writer, sections SectionFlags) {
	if r == nil {
		return
	}

	// Opcodes
	if sections.Has(SectionOpcodes) && len(r.OpStats) > 0 {
		fmt.Fprintln(w, "Opcodes:")
		for _, name := range slices.Sorted(maps.Keys(r.OpStats)) {
			stat := r.OpStats[name]
			fmt.Fprintf(w, "  %s: count=%d gas=%d\n", name, stat.Count, stat.Gas)
		}
	}

	// Store
	if sections.Has(SectionStore) && len(r.StoreStats) > 0 {
		fmt.Fprintln(w, "Store:")
		for _, name := range slices.Sorted(maps.Keys(r.StoreStats)) {
			stat := r.StoreStats[name]
			fmt.Fprintf(w, "  %s: count=%d size=%d\n", name, stat.Count, stat.TotalSize)
		}
	}

	// Native
	if sections.Has(SectionNative) && len(r.NativeStats) > 0 {
		fmt.Fprintln(w, "Native:")
		for _, name := range slices.Sorted(maps.Keys(r.NativeStats)) {
			stat := r.NativeStats[name]
			fmt.Fprintf(w, "  %s: count=%d\n", name, stat.Count)
		}
	}

	// HotSpots
	if sections.Has(SectionHotSpots) && len(r.LocationStats) > 0 {
		fmt.Fprintln(w, "HotSpots:")
		locs := slices.Clone(r.LocationStats)
		slices.SortFunc(locs, func(a, b *LocationStat) int {
			return cmp.Or(
				cmp.Compare(a.File, b.File),
				cmp.Compare(a.Line, b.Line),
			)
		})
		for _, loc := range locs {
			funcName := cmp.Or(loc.FuncName, "-")
			fmt.Fprintf(w, "  %s:%d %s: count=%d gas=%d\n",
				loc.File, loc.Line, funcName, loc.Count, loc.Gas)
		}
	}

	// SubOps
	if sections.Has(SectionSubOps) && len(r.SubOpStats) > 0 {
		fmt.Fprintln(w, "SubOps:")
		for _, name := range slices.Sorted(maps.Keys(r.SubOpStats)) {
			stat := r.SubOpStats[name]
			fmt.Fprintf(w, "  %s: count=%d\n", name, stat.Count)
		}
	}

	// Variables
	if sections.Has(SectionVars) && len(r.VarStats) > 0 {
		fmt.Fprintln(w, "Variables:")
		vars := slices.Clone(r.VarStats)
		slices.SortFunc(vars, func(a, b *VarStat) int {
			return cmp.Or(
				cmp.Compare(a.File, b.File),
				cmp.Compare(a.Line, b.Line),
				cmp.Compare(a.Name, b.Name),
				cmp.Compare(a.Index, b.Index),
			)
		})
		for _, v := range vars {
			fmt.Fprintf(w, "  %s:%d %s: count=%d\n", v.File, v.Line, v.DisplayName(), v.Count)
		}
	}
}

// ---- Generic sort helper

// kvPair holds a key-value pair for sorted iteration.
type kvPair[V any] struct {
	key string
	val V
}

// sortedMap returns map entries sorted descending by the given key function.
func sortedMap[V any](m map[string]V, keyFn func(V) int64) []kvPair[V] {
	pairs := make([]kvPair[V], 0, len(m))
	for k, v := range m {
		pairs = append(pairs, kvPair[V]{k, v})
	}
	slices.SortFunc(pairs, func(a, b kvPair[V]) int {
		return cmp.Compare(keyFn(b.val), keyFn(a.val)) // descending
	})
	return pairs
}

// MergeResults combines multiple Results into a single aggregated Result.
// This is useful for combining profiling data from multiple test runs.
//
// Aggregation behavior:
//   - OpStats, StoreStats, NativeStats, SubOpStats: Merged by name with counts/gas/timing combined
//   - LocationStats, StackSamples: Appended without deduplication (may contain duplicates from
//     different runs; consumers should aggregate if needed)
//   - Duration: Summed across all results
//   - TimingEnabled: True if any input had timing enabled
func MergeResults(results ...*Results) *Results {
	if len(results) == 0 {
		return nil
	}
	if len(results) == 1 {
		return results[0]
	}

	merged := &Results{
		OpStats:     make(map[string]*OpStat),
		StoreStats:  make(map[string]*StoreStat),
		NativeStats: make(map[string]*TimingStat),
		SubOpStats:  make(map[string]*TimingStat),
	}

	for _, r := range results {
		if r == nil {
			continue
		}

		merged.Duration += r.Duration
		merged.TimingEnabled = merged.TimingEnabled || r.TimingEnabled

		// Merge OpStats
		for name, stat := range r.OpStats {
			if existing, ok := merged.OpStats[name]; ok {
				existing.TimingStat.Merge(&stat.TimingStat)
				existing.Gas += stat.Gas
			} else {
				merged.OpStats[name] = &OpStat{
					TimingStat: stat.TimingStat,
					Gas:        stat.Gas,
				}
			}
		}

		// Merge StoreStats
		for name, stat := range r.StoreStats {
			if existing, ok := merged.StoreStats[name]; ok {
				existing.TimingStat.Merge(&stat.TimingStat)
				existing.TotalSize += stat.TotalSize
			} else {
				merged.StoreStats[name] = &StoreStat{
					TimingStat: stat.TimingStat,
					TotalSize:  stat.TotalSize,
				}
			}
		}

		// Merge NativeStats
		for name, stat := range r.NativeStats {
			if existing, ok := merged.NativeStats[name]; ok {
				existing.Merge(stat)
			} else {
				clone := *stat
				merged.NativeStats[name] = &clone
			}
		}

		// Merge SubOpStats
		for name, stat := range r.SubOpStats {
			if existing, ok := merged.SubOpStats[name]; ok {
				existing.Merge(stat)
			} else {
				clone := *stat
				merged.SubOpStats[name] = &clone
			}
		}

		// Merge LocationStats (append all)
		merged.LocationStats = append(merged.LocationStats, r.LocationStats...)

		// Merge StackSamples (append all)
		merged.StackSamples = append(merged.StackSamples, r.StackSamples...)
	}

	return merged
}

// parseFrameFromKey parses a frame from a key part like "funcName@file:line".
func parseFrameFromKey(part string) StackFrame {
	atIdx := strings.IndexByte(part, '@')
	if atIdx == -1 {
		return StackFrame{Func: part}
	}

	funcName := part[:atIdx]
	rest := part[atIdx+1:]

	colonIdx := strings.LastIndexByte(rest, ':')
	if colonIdx == -1 {
		return StackFrame{Func: funcName, File: rest}
	}

	// Parse line number. Keys are created by buildFromStacks using strconv.Itoa,
	// so this should never fail for valid keys. If malformed, line defaults to 0.
	line, err := strconv.Atoi(rest[colonIdx+1:])
	if err != nil {
		line = 0
	}
	return StackFrame{
		Func: funcName,
		File: rest[:colonIdx],
		Line: line,
	}
}
