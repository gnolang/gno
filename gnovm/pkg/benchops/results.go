package benchops

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
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
	OpStats       map[string]*OpStatJSON
	StoreStats    map[string]*StoreStatJSON
	NativeStats   map[string]*NativeStatJSON
	LocationStats []*LocationStatJSON `json:"LocationStats,omitempty"`
}

// OpStatJSON is the JSON-serializable form of opcode statistics.
type OpStatJSON struct {
	Count   int64 `json:"count"`
	TotalNs int64 `json:"total_ns,omitempty"`
	AvgNs   int64 `json:"avg_ns,omitempty"`
	MinNs   int64 `json:"min_ns,omitempty"`
	MaxNs   int64 `json:"max_ns,omitempty"`
	Gas     int64 `json:"gas"`
}

// StoreStatJSON is the JSON-serializable form of store statistics.
type StoreStatJSON struct {
	Count     int64 `json:"count"`
	TotalNs   int64 `json:"total_ns,omitempty"`
	AvgNs     int64 `json:"avg_ns,omitempty"`
	MinNs     int64 `json:"min_ns,omitempty"`
	MaxNs     int64 `json:"max_ns,omitempty"`
	TotalSize int64 `json:"total_size"`
	AvgSize   int64 `json:"avg_size"`
}

// NativeStatJSON is the JSON-serializable form of native statistics.
type NativeStatJSON struct {
	Count   int64 `json:"count"`
	TotalNs int64 `json:"total_ns,omitempty"`
	AvgNs   int64 `json:"avg_ns,omitempty"`
	MinNs   int64 `json:"min_ns,omitempty"`
	MaxNs   int64 `json:"max_ns,omitempty"`
}

// LocationStatJSON aggregates stats by source location.
type LocationStatJSON struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	FuncName string `json:"func,omitempty"`
	PkgPath  string `json:"pkg"`
	Count    int64  `json:"count"`
	TotalNs  int64  `json:"total_ns,omitempty"`
	Gas      int64  `json:"gas"`
}

// SectionFlags controls which sections are included in WriteGolden output.
type SectionFlags uint8

const (
	SectionOpcodes SectionFlags = 1 << iota
	SectionStore
	SectionNative
	SectionHotSpots

	SectionAll = SectionOpcodes | SectionStore | SectionNative | SectionHotSpots
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
		case "all":
			return 0, nil
		default:
			return 0, fmt.Errorf("unknown section %q (valid: opcodes, store, native, hotspots, all)", name)
		}
	}
	return flags, nil
}

// buildResults creates Results from the profiler's internal state.
func (p *Profiler) buildResults() *Results {
	r := &Results{
		Duration:      p.stopTime.Sub(p.startTime),
		StartTime:     p.startTime,
		EndTime:       p.stopTime,
		TimingEnabled: p.timingEnabled,
		OpStats:       make(map[string]*OpStatJSON),
		StoreStats:    make(map[string]*StoreStatJSON),
		NativeStats:   make(map[string]*NativeStatJSON),
	}

	// Build op stats
	for i := range maxOpCodes {
		name := Op(i).String()
		if p.timingEnabled {
			s := &p.opStatsTimed[i]
			if s.count == 0 {
				continue
			}
			r.OpStats[name] = &OpStatJSON{
				Count:   s.count,
				TotalNs: s.totalDur.Nanoseconds(),
				AvgNs:   s.totalDur.Nanoseconds() / s.count,
				MinNs:   s.minDur.Nanoseconds(),
				MaxNs:   s.maxDur.Nanoseconds(),
				Gas:     s.gas,
			}
		} else {
			s := &p.opStats[i]
			if s.count == 0 {
				continue
			}
			r.OpStats[name] = &OpStatJSON{
				Count: s.count,
				Gas:   s.gas,
			}
		}
	}

	// Build store stats
	for i := range maxOpCodes {
		name := StoreOp(i).String()
		if p.timingEnabled {
			s := &p.storeStatsTimed[i]
			if s.count == 0 {
				continue
			}
			r.StoreStats[name] = &StoreStatJSON{
				Count:     s.count,
				TotalNs:   s.totalDur.Nanoseconds(),
				AvgNs:     s.totalDur.Nanoseconds() / s.count,
				MinNs:     s.minDur.Nanoseconds(),
				MaxNs:     s.maxDur.Nanoseconds(),
				TotalSize: s.totalSize,
				AvgSize:   s.totalSize / s.count,
			}
		} else {
			s := &p.storeStats[i]
			if s.count == 0 {
				continue
			}
			r.StoreStats[name] = &StoreStatJSON{
				Count:     s.count,
				TotalSize: s.totalSize,
				AvgSize:   s.totalSize / s.count,
			}
		}
	}

	// Build native stats
	for i := range maxOpCodes {
		name := NativeOp(i).String()
		if p.timingEnabled {
			s := &p.nativeStatsTimed[i]
			if s.count == 0 {
				continue
			}
			r.NativeStats[name] = &NativeStatJSON{
				Count:   s.count,
				TotalNs: s.totalDur.Nanoseconds(),
				AvgNs:   s.totalDur.Nanoseconds() / s.count,
				MinNs:   s.minDur.Nanoseconds(),
				MaxNs:   s.maxDur.Nanoseconds(),
			}
		} else {
			s := &p.nativeStats[i]
			if s.count == 0 {
				continue
			}
			r.NativeStats[name] = &NativeStatJSON{
				Count: s.count,
			}
		}
	}

	// Build location stats
	if len(p.locationStats) > 0 {
		r.LocationStats = make([]*LocationStatJSON, 0, len(p.locationStats))
		for _, s := range p.locationStats {
			r.LocationStats = append(r.LocationStats, &LocationStatJSON{
				File:     s.file,
				Line:     s.line,
				FuncName: s.funcName,
				PkgPath:  s.pkgPath,
				Count:    s.count,
				TotalNs:  s.totalDur.Nanoseconds(),
				Gas:      s.gasTotal,
			})
		}
		slices.SortFunc(r.LocationStats, func(a, b *LocationStatJSON) int {
			return cmp.Compare(b.Gas, a.Gas) // descending
		})
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
		fmt.Fprintf(tw, "Opcode\tCount\tTotal\tAvg\tMin\tMax\tGas\t%%\t\n")

		sorted := sortedMap(r.OpStats, func(s *OpStatJSON) int64 { return s.TotalNs })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			pct := float64(0)
			if totalGas > 0 {
				pct = float64(kv.val.Gas) * 100 / float64(totalGas)
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%d\t%.1f%%\t\n",
				kv.key, kv.val.Count,
				time.Duration(kv.val.TotalNs), time.Duration(kv.val.AvgNs),
				time.Duration(kv.val.MinNs), time.Duration(kv.val.MaxNs),
				kv.val.Gas, pct)
		}
	} else {
		fmt.Fprintf(tw, "Opcode Statistics (by gas)\n")
		fmt.Fprintf(tw, "──────────────────────────\n")
		fmt.Fprintf(tw, "Opcode\tCount\tGas\t%%\t\n")

		sorted := sortedMap(r.OpStats, func(s *OpStatJSON) int64 { return s.Gas })
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
		fmt.Fprintf(tw, "Operation\tCount\tTotal\tAvg\tMin\tMax\tTotal Size\tAvg Size\t\n")

		sorted := sortedMap(r.StoreStats, func(s *StoreStatJSON) int64 { return s.TotalNs })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%d\t%d\t\n",
				kv.key, kv.val.Count,
				time.Duration(kv.val.TotalNs), time.Duration(kv.val.AvgNs),
				time.Duration(kv.val.MinNs), time.Duration(kv.val.MaxNs),
				kv.val.TotalSize, kv.val.AvgSize)
		}
	} else {
		fmt.Fprintf(tw, "Store Statistics (by count)\n")
		fmt.Fprintf(tw, "───────────────────────────\n")
		fmt.Fprintf(tw, "Operation\tCount\tTotal Size\tAvg Size\t\n")

		sorted := sortedMap(r.StoreStats, func(s *StoreStatJSON) int64 { return s.Count })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t\n",
				kv.key, kv.val.Count, kv.val.TotalSize, kv.val.AvgSize)
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
		fmt.Fprintf(tw, "Function\tCount\tTotal\tAvg\tMin\tMax\t\n")

		sorted := sortedMap(r.NativeStats, func(s *NativeStatJSON) int64 { return s.TotalNs })
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t\n",
				kv.key, kv.val.Count,
				time.Duration(kv.val.TotalNs), time.Duration(kv.val.AvgNs),
				time.Duration(kv.val.MinNs), time.Duration(kv.val.MaxNs))
		}
	} else {
		fmt.Fprintf(tw, "Native Statistics (by count)\n")
		fmt.Fprintf(tw, "────────────────────────────\n")
		fmt.Fprintf(tw, "Function\tCount\t\n")

		sorted := sortedMap(r.NativeStats, func(s *NativeStatJSON) int64 { return s.Count })
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
	if pkg == "" {
		return ""
	}
	// Remove common prefixes
	for _, prefix := range []string{"gno.land/", "github.com/"} {
		if strings.HasPrefix(pkg, prefix) {
			return pkg[len(prefix):]
		}
	}
	return pkg
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
		slices.SortFunc(locs, func(a, b *LocationStatJSON) int {
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

// ---- Merge helper types and functions

// timedStat defines the interface for stats that can be merged with timing data.
type timedStat interface {
	getCount() int64
	getTotalNs() int64
	getMinNs() int64
	getMaxNs() int64
}

func (s *OpStatJSON) getCount() int64     { return s.Count }
func (s *OpStatJSON) getTotalNs() int64   { return s.TotalNs }
func (s *OpStatJSON) getMinNs() int64     { return s.MinNs }
func (s *OpStatJSON) getMaxNs() int64     { return s.MaxNs }
func (s *NativeStatJSON) getCount() int64 { return s.Count }
func (s *NativeStatJSON) getTotalNs() int64 {
	return s.TotalNs
}
func (s *NativeStatJSON) getMinNs() int64 { return s.MinNs }
func (s *NativeStatJSON) getMaxNs() int64 { return s.MaxNs }

// mergeTimedStats updates dst with values from src using standard min/max/avg logic.
func mergeTimedStats(dstCount, dstTotalNs, dstMinNs, dstMaxNs *int64, src timedStat) {
	*dstCount += src.getCount()
	*dstTotalNs += src.getTotalNs()
	srcMin := src.getMinNs()
	if srcMin > 0 && (*dstMinNs == 0 || srcMin < *dstMinNs) {
		*dstMinNs = srcMin
	}
	if src.getMaxNs() > *dstMaxNs {
		*dstMaxNs = src.getMaxNs()
	}
}

// MergeResults combines multiple Results into a single aggregated Result.
// This is useful for combining profiling data from multiple test runs.
func MergeResults(results ...*Results) *Results {
	if len(results) == 0 {
		return nil
	}
	if len(results) == 1 {
		return results[0]
	}

	merged := &Results{
		OpStats:     make(map[string]*OpStatJSON),
		StoreStats:  make(map[string]*StoreStatJSON),
		NativeStats: make(map[string]*NativeStatJSON),
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
				mergeTimedStats(&existing.Count, &existing.TotalNs, &existing.MinNs, &existing.MaxNs, stat)
				existing.Gas += stat.Gas
				if existing.Count > 0 {
					existing.AvgNs = existing.TotalNs / existing.Count
				}
			} else {
				merged.OpStats[name] = &OpStatJSON{
					Count: stat.Count, TotalNs: stat.TotalNs, AvgNs: stat.AvgNs,
					MinNs: stat.MinNs, MaxNs: stat.MaxNs, Gas: stat.Gas,
				}
			}
		}

		// Merge StoreStats
		for name, stat := range r.StoreStats {
			if existing, ok := merged.StoreStats[name]; ok {
				existing.Count += stat.Count
				existing.TotalNs += stat.TotalNs
				existing.TotalSize += stat.TotalSize
				if stat.MinNs > 0 && (existing.MinNs == 0 || stat.MinNs < existing.MinNs) {
					existing.MinNs = stat.MinNs
				}
				if stat.MaxNs > existing.MaxNs {
					existing.MaxNs = stat.MaxNs
				}
				if existing.Count > 0 {
					existing.AvgNs = existing.TotalNs / existing.Count
					existing.AvgSize = existing.TotalSize / existing.Count
				}
			} else {
				merged.StoreStats[name] = &StoreStatJSON{
					Count: stat.Count, TotalNs: stat.TotalNs, AvgNs: stat.AvgNs,
					MinNs: stat.MinNs, MaxNs: stat.MaxNs,
					TotalSize: stat.TotalSize, AvgSize: stat.AvgSize,
				}
			}
		}

		// Merge NativeStats
		for name, stat := range r.NativeStats {
			if existing, ok := merged.NativeStats[name]; ok {
				mergeTimedStats(&existing.Count, &existing.TotalNs, &existing.MinNs, &existing.MaxNs, stat)
				if existing.Count > 0 {
					existing.AvgNs = existing.TotalNs / existing.Count
				}
			} else {
				merged.NativeStats[name] = &NativeStatJSON{
					Count: stat.Count, TotalNs: stat.TotalNs, AvgNs: stat.AvgNs,
					MinNs: stat.MinNs, MaxNs: stat.MaxNs,
				}
			}
		}

		// Merge LocationStats (append all)
		merged.LocationStats = append(merged.LocationStats, r.LocationStats...)
	}

	return merged
}
