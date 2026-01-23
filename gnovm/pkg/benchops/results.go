package benchops

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

// Results contains the profiling results after Stop() is called.
type Results struct {
	Duration      time.Duration
	StartTime     time.Time
	EndTime       time.Time
	OpStats       map[string]*OpStatJSON
	StoreStats    map[string]*StoreStatJSON
	NativeStats   map[string]*NativeStatJSON
	LocationStats []*LocationStatJSON `json:"LocationStats,omitempty"`
}

// OpStatJSON is the JSON-serializable form of opcode statistics.
type OpStatJSON struct {
	Count   int64 `json:"count"`
	TotalNs int64 `json:"total_ns"`
	AvgNs   int64 `json:"avg_ns"`
	MinNs   int64 `json:"min_ns"`
	MaxNs   int64 `json:"max_ns"`
	Gas     int64 `json:"gas"`
}

// StoreStatJSON is the JSON-serializable form of store statistics.
type StoreStatJSON struct {
	Count     int64 `json:"count"`
	TotalNs   int64 `json:"total_ns"`
	AvgNs     int64 `json:"avg_ns"`
	MinNs     int64 `json:"min_ns"`
	MaxNs     int64 `json:"max_ns"`
	TotalSize int64 `json:"total_size"`
	AvgSize   int64 `json:"avg_size"`
}

// NativeStatJSON is the JSON-serializable form of native statistics.
type NativeStatJSON struct {
	Count   int64 `json:"count"`
	TotalNs int64 `json:"total_ns"`
	AvgNs   int64 `json:"avg_ns"`
	MinNs   int64 `json:"min_ns"`
	MaxNs   int64 `json:"max_ns"`
}

// LocationStatJSON aggregates stats by source location.
type LocationStatJSON struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	FuncName string `json:"func,omitempty"`
	PkgPath  string `json:"pkg"`
	Count    int64  `json:"count"`
	TotalNs  int64  `json:"total_ns"`
	Gas      int64  `json:"gas"`
}

// SectionFlags controls which sections are included in WriteGolden output.
// Use bitwise OR to combine flags. Zero is equivalent to SectionAll.
type SectionFlags uint8

const (
	SectionOpcodes  SectionFlags = 1 << iota // 1: Include Opcodes section
	SectionStore                             // 2: Include Store section
	SectionNative                            // 4: Include Native section
	SectionHotSpots                          // 8: Include HotSpots section

	SectionAll = SectionOpcodes | SectionStore | SectionNative | SectionHotSpots // 15: All sections
)

// Has returns true if the flag is set. If s is 0 (SectionAll), returns true for any flag.
func (s SectionFlags) Has(flag SectionFlags) bool {
	if s == 0 {
		return true
	}
	return s&flag != 0
}

// ParseSectionFlags parses a comma-separated list of section names.
// Valid names: opcodes, store, native, hotspots, all
// Empty string or "all" returns 0 (all sections).
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
		Duration:    p.stopTime.Sub(p.startTime),
		StartTime:   p.startTime,
		EndTime:     p.stopTime,
		OpStats:     make(map[string]*OpStatJSON),
		StoreStats:  make(map[string]*StoreStatJSON),
		NativeStats: make(map[string]*NativeStatJSON),
	}

	// Build op stats
	for i := range maxOpCodes {
		s := &p.opStats[i]
		if s.count == 0 {
			continue
		}
		op := Op(i)
		r.OpStats[op.String()] = &OpStatJSON{
			Count:   s.count,
			TotalNs: s.totalDur.Nanoseconds(),
			AvgNs:   s.totalDur.Nanoseconds() / s.count,
			MinNs:   s.minDur.Nanoseconds(),
			MaxNs:   s.maxDur.Nanoseconds(),
			Gas:     GetOpGas(op),
		}
	}

	// Build store stats
	for i := range maxOpCodes {
		s := &p.storeStats[i]
		if s.count == 0 {
			continue
		}
		op := StoreOp(i)
		r.StoreStats[op.String()] = &StoreStatJSON{
			Count:     s.count,
			TotalNs:   s.totalDur.Nanoseconds(),
			AvgNs:     s.totalDur.Nanoseconds() / s.count,
			MinNs:     s.minDur.Nanoseconds(),
			MaxNs:     s.maxDur.Nanoseconds(),
			TotalSize: s.totalSize,
			AvgSize:   s.totalSize / s.count,
		}
	}

	// Build native stats
	for i := range maxOpCodes {
		s := &p.nativeStats[i]
		if s.count == 0 {
			continue
		}
		op := NativeOp(i)
		r.NativeStats[op.String()] = &NativeStatJSON{
			Count:   s.count,
			TotalNs: s.totalDur.Nanoseconds(),
			AvgNs:   s.totalDur.Nanoseconds() / s.count,
			MinNs:   s.minDur.Nanoseconds(),
			MaxNs:   s.maxDur.Nanoseconds(),
		}
	}

	// Build location stats (sorted by gas, hot spots first)
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
		// Sort by gas (descending) for hot spots analysis
		sort.Slice(r.LocationStats, func(i, j int) bool {
			return r.LocationStats[i].Gas > r.LocationStats[j].Gas
		})
	}

	return r
}

// ---- Output Methods
//
// Results provides three output formats:
//
//   - WriteJSON: Machine-readable JSON with all data
//   - WriteReport: Human-readable with timing (non-deterministic, for CLI)
//   - WriteGolden: Human-readable without timing (deterministic, for tests)
//
// Use WriteReport for dynamic output (gno run --bench, gno test --bench)
// where timing information is useful. Use WriteGolden for test golden files
// where reproducibility across runs is required.

// WriteJSON writes the results as compact JSON to the given writer.
// Includes all data (timing, sizes). Use for machine processing.
func (r *Results) WriteJSON(w io.Writer) error {
	if r == nil {
		return nil
	}

	return json.NewEncoder(w).Encode(r)
}

// WriteReport writes a human-readable summary for dynamic/interactive use.
// Includes timing data and sorts by total time (slowest operations first).
// Use this for CLI output where timing information is valuable.
// topN limits how many entries to show per category (0 = all).
func (r *Results) WriteReport(w io.Writer, topN int) error {
	if r == nil {
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)

	fmt.Fprintf(tw, "Profiling Results\n")
	fmt.Fprintf(tw, "═════════════════\n")
	fmt.Fprintf(tw, "Duration:\t%v\t\n", r.Duration)
	fmt.Fprintf(tw, "Start:\t%v\t\n", r.StartTime.Format(time.RFC3339))
	fmt.Fprintf(tw, "End:\t%v\t\n\n", r.EndTime.Format(time.RFC3339))

	// Op stats
	if len(r.OpStats) > 0 {
		fmt.Fprintf(tw, "Opcode Statistics (by total time)\n")
		fmt.Fprintf(tw, "──────────────────────────────────\n")
		fmt.Fprintf(tw, "Opcode\tCount\tTotal\tAvg\tMin\tMax\tGas\t\n")

		sorted := sortedOpStats(r.OpStats)
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%d\t\n",
				kv.name,
				kv.stat.Count,
				time.Duration(kv.stat.TotalNs),
				time.Duration(kv.stat.AvgNs),
				time.Duration(kv.stat.MinNs),
				time.Duration(kv.stat.MaxNs),
				kv.stat.Gas,
			)
		}
		fmt.Fprintln(tw)
	}

	// Store stats
	if len(r.StoreStats) > 0 {
		fmt.Fprintf(tw, "Store Statistics (by total time)\n")
		fmt.Fprintf(tw, "─────────────────────────────────\n")
		fmt.Fprintf(tw, "Operation\tCount\tTotal\tAvg\tMin\tMax\tTotal Size\tAvg Size\t\n")

		sorted := sortedStoreStats(r.StoreStats)
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%d\t%d\t\n",
				kv.name,
				kv.stat.Count,
				time.Duration(kv.stat.TotalNs),
				time.Duration(kv.stat.AvgNs),
				time.Duration(kv.stat.MinNs),
				time.Duration(kv.stat.MaxNs),
				kv.stat.TotalSize,
				kv.stat.AvgSize,
			)
		}
		fmt.Fprintln(tw)
	}

	// Native stats
	if len(r.NativeStats) > 0 {
		fmt.Fprintf(tw, "Native Statistics (by total time)\n")
		fmt.Fprintf(tw, "──────────────────────────────────\n")
		fmt.Fprintf(tw, "Function\tCount\tTotal\tAvg\tMin\tMax\t\n")

		sorted := sortedNativeStats(r.NativeStats)
		for i, kv := range sorted {
			if topN > 0 && i >= topN {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t\n",
				kv.name,
				kv.stat.Count,
				time.Duration(kv.stat.TotalNs),
				time.Duration(kv.stat.AvgNs),
				time.Duration(kv.stat.MinNs),
				time.Duration(kv.stat.MaxNs),
			)
		}
		fmt.Fprintln(tw)
	}

	// Hot spots (by gas)
	if len(r.LocationStats) > 0 {
		fmt.Fprintf(tw, "Hot Spots (by gas)\n")
		fmt.Fprintf(tw, "──────────────────\n")
		fmt.Fprintf(tw, "Location\tFunc\tCount\tTotal\tGas\t\n")

		for i, loc := range r.LocationStats {
			if topN > 0 && i >= topN {
				break
			}
			location := fmt.Sprintf("%s:%d", loc.File, loc.Line)
			funcName := loc.FuncName
			if funcName == "" {
				funcName = "-"
			}
			fmt.Fprintf(tw, "%s\t%s\t%d\t%v\t%d\t\n",
				location,
				funcName,
				loc.Count,
				time.Duration(loc.TotalNs),
				loc.Gas,
			)
		}
		fmt.Fprintln(tw)
	}

	return tw.Flush()
}

// WriteGolden writes a deterministic summary for test golden file comparison.
// Unlike WriteReport, this:
//   - Excludes timing data (which varies between runs)
//   - Sorts alphabetically by name (not by time)
//   - Uses a simple format without tabwriter
//
// The sections parameter controls which sections to include (0 = all).
// Use this for filetest golden output where reproducibility is required.
func (r *Results) WriteGolden(w io.Writer, sections SectionFlags) {
	if r == nil {
		return
	}

	// Opcodes (sorted alphabetically)
	if sections.Has(SectionOpcodes) && len(r.OpStats) > 0 {
		fmt.Fprintln(w, "Opcodes:")
		names := make([]string, 0, len(r.OpStats))
		for name := range r.OpStats {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			stat := r.OpStats[name]
			fmt.Fprintf(w, "  %s: count=%d gas=%d\n", name, stat.Count, stat.Gas)
		}
	}

	// Store operations (sorted alphabetically)
	if sections.Has(SectionStore) && len(r.StoreStats) > 0 {
		fmt.Fprintln(w, "Store:")
		names := make([]string, 0, len(r.StoreStats))
		for name := range r.StoreStats {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			stat := r.StoreStats[name]
			fmt.Fprintf(w, "  %s: count=%d size=%d\n", name, stat.Count, stat.TotalSize)
		}
	}

	// Native functions (sorted alphabetically)
	if sections.Has(SectionNative) && len(r.NativeStats) > 0 {
		fmt.Fprintln(w, "Native:")
		names := make([]string, 0, len(r.NativeStats))
		for name := range r.NativeStats {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			stat := r.NativeStats[name]
			fmt.Fprintf(w, "  %s: count=%d\n", name, stat.Count)
		}
	}

	// Hot spots by location (sorted by file:line for determinism)
	if sections.Has(SectionHotSpots) && len(r.LocationStats) > 0 {
		fmt.Fprintln(w, "HotSpots:")
		// Sort by file:line for deterministic output
		locs := make([]*LocationStatJSON, len(r.LocationStats))
		copy(locs, r.LocationStats)
		sort.Slice(locs, func(i, j int) bool {
			if locs[i].File != locs[j].File {
				return locs[i].File < locs[j].File
			}
			return locs[i].Line < locs[j].Line
		})
		for _, loc := range locs {
			funcName := loc.FuncName
			if funcName == "" {
				funcName = "-"
			}
			fmt.Fprintf(w, "  %s:%d %s: count=%d gas=%d\n",
				loc.File, loc.Line, funcName, loc.Count, loc.Gas)
		}
	}
}

type opStatPair struct {
	name string
	stat *OpStatJSON
}

func sortedOpStats(m map[string]*OpStatJSON) []opStatPair {
	pairs := make([]opStatPair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, opStatPair{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].stat.TotalNs > pairs[j].stat.TotalNs
	})
	return pairs
}

type storeStatPair struct {
	name string
	stat *StoreStatJSON
}

func sortedStoreStats(m map[string]*StoreStatJSON) []storeStatPair {
	pairs := make([]storeStatPair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, storeStatPair{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].stat.TotalNs > pairs[j].stat.TotalNs
	})
	return pairs
}

type nativeStatPair struct {
	name string
	stat *NativeStatJSON
}

func sortedNativeStats(m map[string]*NativeStatJSON) []nativeStatPair {
	pairs := make([]nativeStatPair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, nativeStatPair{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].stat.TotalNs > pairs[j].stat.TotalNs
	})
	return pairs
}
