package benchops

import (
	"encoding/csv"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
)

// WriteCSV writes the profiling results in CSV format.
// Each section is written as a separate table with a header row.
// Sections are separated by blank lines.
func (r *Results) WriteCSV(w io.Writer) error {
	if r == nil {
		return nil
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Write OpStats
	if len(r.OpStats) > 0 {
		if err := r.writeOpStatsCSV(cw); err != nil {
			return err
		}
		if err := cw.Write([]string{}); err != nil { // section separator
			return err
		}
	}

	// Write StoreStats
	if len(r.StoreStats) > 0 {
		if err := r.writeStoreStatsCSV(cw); err != nil {
			return err
		}
		if err := cw.Write([]string{}); err != nil {
			return err
		}
	}

	// Write NativeStats
	if len(r.NativeStats) > 0 {
		if err := r.writeNativeStatsCSV(cw); err != nil {
			return err
		}
		if err := cw.Write([]string{}); err != nil {
			return err
		}
	}

	// Write LocationStats
	if len(r.LocationStats) > 0 {
		if err := r.writeLocationStatsCSV(cw); err != nil {
			return err
		}
		if err := cw.Write([]string{}); err != nil {
			return err
		}
	}

	// Write SubOpStats
	if len(r.SubOpStats) > 0 {
		if err := r.writeSubOpStatsCSV(cw); err != nil {
			return err
		}
		if err := cw.Write([]string{}); err != nil {
			return err
		}
	}

	// Write VarStats
	if len(r.VarStats) > 0 {
		if err := r.writeVarStatsCSV(cw); err != nil {
			return err
		}
	}

	return cw.Error()
}

func (r *Results) writeOpStatsCSV(cw *csv.Writer) error {
	// Header
	header := []string{"opcode", "count", "gas"}
	if r.TimingEnabled {
		header = append(header, "total_ns", "avg_ns", "stddev_ns", "min_ns", "max_ns")
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	// Data rows sorted by name for determinism
	for _, name := range slices.Sorted(maps.Keys(r.OpStats)) {
		stat := r.OpStats[name]
		row := []string{
			name,
			strconv.FormatInt(stat.Count, 10),
			strconv.FormatInt(stat.Gas, 10),
		}
		if r.TimingEnabled {
			row = append(row, stat.CSVTimingFields()...)
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (r *Results) writeStoreStatsCSV(cw *csv.Writer) error {
	header := []string{"operation", "count", "bytes_read", "bytes_written"}
	if r.TimingEnabled {
		header = append(header, "total_ns", "avg_ns", "stddev_ns", "min_ns", "max_ns")
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, name := range slices.Sorted(maps.Keys(r.StoreStats)) {
		stat := r.StoreStats[name]
		row := []string{
			name,
			strconv.FormatInt(stat.Count, 10),
			strconv.FormatInt(stat.BytesRead, 10),
			strconv.FormatInt(stat.BytesWritten, 10),
		}
		if r.TimingEnabled {
			row = append(row, stat.CSVTimingFields()...)
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (r *Results) writeNativeStatsCSV(cw *csv.Writer) error {
	header := []string{"function", "count"}
	if r.TimingEnabled {
		header = append(header, "total_ns", "avg_ns", "stddev_ns", "min_ns", "max_ns")
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, name := range slices.Sorted(maps.Keys(r.NativeStats)) {
		stat := r.NativeStats[name]
		row := []string{
			name,
			strconv.FormatInt(stat.Count, 10),
		}
		if r.TimingEnabled {
			row = append(row, stat.CSVTimingFields()...)
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (r *Results) writeLocationStatsCSV(cw *csv.Writer) error {
	header := []string{"file", "line", "func", "pkg", "count", "gas"}
	if r.TimingEnabled {
		header = append(header, "total_ns")
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, stat := range r.LocationStats {
		row := []string{
			stat.File,
			strconv.Itoa(stat.Line),
			stat.FuncName,
			stat.PkgPath,
			strconv.FormatInt(stat.Count, 10),
			strconv.FormatInt(stat.Gas, 10),
		}
		if r.TimingEnabled {
			row = append(row, strconv.FormatInt(stat.TotalNs, 10))
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (r *Results) writeSubOpStatsCSV(cw *csv.Writer) error {
	header := []string{"subop", "count"}
	if r.TimingEnabled {
		header = append(header, "total_ns", "avg_ns", "stddev_ns", "min_ns", "max_ns")
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, name := range slices.Sorted(maps.Keys(r.SubOpStats)) {
		stat := r.SubOpStats[name]
		row := []string{
			name,
			strconv.FormatInt(stat.Count, 10),
		}
		if r.TimingEnabled {
			row = append(row, stat.CSVTimingFields()...)
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (r *Results) writeVarStatsCSV(cw *csv.Writer) error {
	header := []string{"file", "line", "name", "index", "count"}
	if r.TimingEnabled {
		header = append(header, "total_ns", "avg_ns", "stddev_ns", "min_ns", "max_ns")
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, stat := range r.VarStats {
		row := []string{
			stat.File,
			strconv.Itoa(stat.Line),
			stat.Name,
			strconv.Itoa(stat.Index),
			strconv.FormatInt(stat.Count, 10),
		}
		if r.TimingEnabled {
			row = append(row, stat.CSVTimingFields()...)
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// WriteCSVSection writes a single section of results in CSV format.
// This is useful when you want separate CSV files per section.
func (r *Results) WriteCSVSection(w io.Writer, section SectionFlags) error {
	if r == nil {
		return nil
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	switch section {
	case SectionOpcodes:
		return r.writeOpStatsCSV(cw)
	case SectionStore:
		return r.writeStoreStatsCSV(cw)
	case SectionNative:
		return r.writeNativeStatsCSV(cw)
	case SectionHotSpots:
		return r.writeLocationStatsCSV(cw)
	case SectionSubOps:
		return r.writeSubOpStatsCSV(cw)
	case SectionVars:
		return r.writeVarStatsCSV(cw)
	default:
		return fmt.Errorf("unknown section: %d", section)
	}
}
