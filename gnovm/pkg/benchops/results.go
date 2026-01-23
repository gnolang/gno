package benchops

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
	"time"
)

// Results contains the profiling results after Stop() is called.
type Results struct {
	Duration    time.Duration
	StartTime   time.Time
	EndTime     time.Time
	OpStats     map[string]*OpStatJSON
	StoreStats  map[string]*StoreStatJSON
	NativeStats map[string]*NativeStatJSON
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
	for i := range 256 {
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
	for i := range 256 {
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
	for i := range 256 {
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

	return r
}

// WriteJSON writes the results as JSON to the given writer.
func (r *Results) WriteJSON(w io.Writer) error {
	if r == nil {
		return nil
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// WriteReport writes a human-readable summary to the given writer.
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

	return tw.Flush()
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
