//go:build gnobench

package benchops

import (
	"io"
	"strconv"
	"strings"

	"github.com/google/pprof/profile"
)

// PprofOption configures pprof output generation.
type PprofOption func(*pprofConfig)

type pprofConfig struct {
	includeDuration bool // Include duration sample type
	includeCount    bool // Include count sample type
	includeLabels   bool // Include filtering labels (op, pkg, depth)
}

// WithDuration includes duration (nanoseconds) as a sample type in pprof output.
func WithDuration() PprofOption {
	return func(c *pprofConfig) { c.includeDuration = true }
}

// WithCount includes sample count as a sample type in pprof output.
func WithCount() PprofOption {
	return func(c *pprofConfig) { c.includeCount = true }
}

// WithLabels includes filtering labels (op, pkg, depth) in pprof samples.
func WithLabels() PprofOption {
	return func(c *pprofConfig) { c.includeLabels = true }
}

// WritePprof writes the profiling results in pprof protobuf format.
// The output is gzip-compressed and compatible with `go tool pprof`.
//
// The profile uses gas as the sample value, which is deterministic.
// If stack samples are available (via WithStacks option), the output
// includes full call stack information for flame graph visualization.
// Otherwise, it falls back to flat location-based output.
//
// For more control over output, use WritePprofWithOptions.
func (r *Results) WritePprof(w io.Writer) error {
	return r.WritePprofWithOptions(w)
}

// WritePprofWithOptions writes the profiling results in pprof protobuf format
// with configurable sample types and labels.
//
// By default, only gas is included as a sample type. Use options to include:
//   - WithDuration(): Add duration in nanoseconds (requires timing to be enabled)
//   - WithCount(): Add sample count
//   - WithLabels(): Add filtering labels (pkg, depth)
//
// Example:
//
//	results.WritePprofWithOptions(w, WithDuration(), WithCount(), WithLabels())
func (r *Results) WritePprofWithOptions(w io.Writer, opts ...PprofOption) error {
	if r == nil {
		return nil
	}

	cfg := &pprofConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Build sample types based on options
	sampleTypes := []*profile.ValueType{{Type: "gas", Unit: "units"}}
	if cfg.includeDuration && r.TimingEnabled {
		sampleTypes = append(sampleTypes, &profile.ValueType{Type: "duration", Unit: "nanoseconds"})
	}
	if cfg.includeCount {
		sampleTypes = append(sampleTypes, &profile.ValueType{Type: "count", Unit: "samples"})
	}

	p := &profile.Profile{
		SampleType: sampleTypes,
	}

	builder := &pprofBuilder{
		profile:       p,
		funcByID:      make(map[funcKey]*profile.Function),
		locByID:       make(map[funcKey]*profile.Location),
		cfg:           cfg,
		timingEnabled: r.TimingEnabled,
	}

	// Use stack-based output if we have stack samples
	if len(r.StackSamples) > 0 {
		builder.buildFromStacks(r.StackSamples)
	} else if len(r.LocationStats) > 0 {
		builder.buildFromLocations(r.LocationStats)
	}

	return p.Write(w)
}

// pprofBuilder helps construct a pprof profile with deduplication.
type pprofBuilder struct {
	profile       *profile.Profile
	funcByID      map[funcKey]*profile.Function
	locByID       map[funcKey]*profile.Location
	cfg           *pprofConfig
	timingEnabled bool
}

// funcKey identifies a unique function/location by name, file, and line.
type funcKey struct {
	name string
	file string
	line int
}

// buildFromLocations populates the profile from flat location stats.
func (b *pprofBuilder) buildFromLocations(locs []*LocationStat) {
	for _, loc := range locs {
		funcName := loc.FuncName
		if funcName == "" {
			funcName = extractFilename(loc.File) + ":" + strconv.Itoa(loc.Line)
		}

		// Build system name with package path if available
		systemName := funcName
		if loc.PkgPath != "" {
			systemName = loc.PkgPath + "." + funcName
		}

		fn := &profile.Function{
			ID:         uint64(len(b.profile.Function) + 1),
			Name:       funcName,
			SystemName: systemName,
			Filename:   loc.File,
			StartLine:  int64(loc.Line),
		}
		b.profile.Function = append(b.profile.Function, fn)

		location := &profile.Location{
			ID:   uint64(len(b.profile.Location) + 1),
			Line: []profile.Line{{Function: fn, Line: int64(loc.Line)}},
		}
		b.profile.Location = append(b.profile.Location, location)

		// Build sample values based on configured sample types
		values := []int64{loc.Gas}
		if b.cfg.includeDuration && b.timingEnabled {
			values = append(values, loc.TotalNs)
		}
		if b.cfg.includeCount {
			values = append(values, loc.Count)
		}

		sample := &profile.Sample{
			Location: []*profile.Location{location},
			Value:    values,
		}

		// Add labels if configured
		if b.cfg.includeLabels {
			if loc.PkgPath != "" {
				sample.Label = map[string][]string{
					"pkg": {loc.PkgPath},
				}
			}
		}

		b.profile.Sample = append(b.profile.Sample, sample)
	}
}

// buildFromStacks populates the profile from stack samples with full call stacks.
// Samples are already aggregated by stack signature in buildResults(), so we
// simply convert them to pprof format without re-aggregation.
func (b *pprofBuilder) buildFromStacks(samples []*StackSample) {
	for _, s := range samples {
		var locations []*profile.Location
		for _, frame := range s.Stack {
			locations = append(locations, b.getOrCreateLoc(frame))
		}

		// Build sample values based on configured sample types
		values := []int64{s.Gas}
		if b.cfg.includeDuration && b.timingEnabled {
			values = append(values, s.DurationNs)
		}
		if b.cfg.includeCount {
			values = append(values, s.Count)
		}

		sample := &profile.Sample{
			Location: locations,
			Value:    values,
		}

		// Add labels if configured
		if b.cfg.includeLabels {
			sample.Label = make(map[string][]string)
			if len(s.Stack) > 0 && s.Stack[0].PkgPath != "" {
				sample.Label["pkg"] = []string{s.Stack[0].PkgPath}
			}
			sample.NumLabel = map[string][]int64{
				"depth": {int64(len(s.Stack))},
			}
		}

		b.profile.Sample = append(b.profile.Sample, sample)
	}
}

// getOrCreateLoc returns an existing location or creates a new one.
func (b *pprofBuilder) getOrCreateLoc(frame StackFrame) *profile.Location {
	funcName := frame.Func
	if funcName == "" {
		funcName = extractFilename(frame.File) + ":" + strconv.Itoa(frame.Line)
	}

	key := funcKey{name: funcName, file: frame.File, line: frame.Line}
	if loc, ok := b.locByID[key]; ok {
		return loc
	}

	// Get or create function
	fn, ok := b.funcByID[key]
	if !ok {
		// Build system name with package path if available
		systemName := funcName
		if frame.PkgPath != "" {
			systemName = frame.PkgPath + "." + funcName
		}

		fn = &profile.Function{
			ID:         uint64(len(b.profile.Function) + 1),
			Name:       funcName,
			SystemName: systemName,
			Filename:   frame.File,
			StartLine:  int64(frame.Line),
		}
		b.profile.Function = append(b.profile.Function, fn)
		b.funcByID[key] = fn
	}

	loc := &profile.Location{
		ID:   uint64(len(b.profile.Location) + 1),
		Line: []profile.Line{{Function: fn, Line: int64(frame.Line)}},
	}
	b.profile.Location = append(b.profile.Location, loc)
	b.locByID[key] = loc
	return loc
}

// extractFilename returns just the filename without the directory path.
func extractFilename(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}
