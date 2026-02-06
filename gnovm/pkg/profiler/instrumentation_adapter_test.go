package profiler

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/instrumentation"
)

func TestSinkAdapterOnSample(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	p.StartProfiling(nil, Options{Type: ProfileCPU, SampleRate: 1})

	sink := NewSinkAdapter(p, Options{Type: ProfileCPU, SampleRate: 1})

	ctx := &instrumentation.SampleContext{
		Frames: []instrumentation.FrameSnapshot{
			{
				FuncName: "foo",
				File:     "pkg/foo.gno",
				PkgPath:  "pkg/foo",
				Line:     42,
				IsCall:   true,
			},
		},
		Cycles:  100,
		GasUsed: 50,
	}

	sink.OnSample(ctx)

	profile := p.StopProfiling()
	if profile == nil || len(profile.Functions) == 0 {
		t.Fatalf("expected profile sample to be recorded")
	}
}

func TestSinkAdapterOnAllocation(t *testing.T) {
	p := NewProfiler(ProfileMemory, 1)
	p.StartProfiling(nil, Options{Type: ProfileMemory, SampleRate: 1})
	sink := NewSinkAdapter(p, Options{Type: ProfileMemory, SampleRate: 1})

	event := &instrumentation.AllocationEvent{
		Bytes:   64,
		Objects: 1,
		Kind:    "struct",
		Stack: []instrumentation.FrameSnapshot{
			{
				FuncName: "bar",
				File:     "pkg/bar.gno",
				PkgPath:  "pkg/bar",
				Line:     21,
				IsCall:   true,
			},
		},
	}

	sink.OnAllocation(event)
	profile := p.StopProfiling()
	if profile == nil || len(profile.Functions) == 0 {
		t.Fatalf("expected allocation sample to be recorded")
	}
	if profile.Functions[0].AllocBytes == 0 {
		t.Fatalf("expected allocation bytes to be tracked")
	}
}

func TestSinkAdapterCapabilities(t *testing.T) {
	p := NewProfiler(ProfileCPU, 1)
	sink := NewSinkAdapter(p, Options{Type: ProfileCPU})

	if !sink.WantsSamples() {
		t.Fatalf("expected CPU profile to want samples")
	}
	if sink.WantsAllocations() {
		t.Fatalf("did not expect CPU profile to want allocations")
	}

	p.EnableLineProfiling()
	if !sink.WantsLineSamples() {
		t.Fatalf("expected line profiling to request line samples")
	}
}
