package gnolang

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/instrumentation"
	"github.com/gnolang/gno/gnovm/pkg/profiler"
)

type sampleCountingSink struct {
	samples int
}

func (s *sampleCountingSink) OnSample(*instrumentation.SampleContext) {
	s.samples++
}

func (*sampleCountingSink) OnAllocation(*instrumentation.AllocationEvent) {}
func (*sampleCountingSink) OnLineSample(*instrumentation.LineSample)      {}
func (*sampleCountingSink) WantsSamples() bool                            { return true }
func (*sampleCountingSink) WantsAllocations() bool                        { return false }
func (*sampleCountingSink) WantsLineSamples() bool                        { return false }

func TestMaybeEmitSampleRunsOnCPUSteps(t *testing.T) {
	m := &Machine{
		Frames: []Frame{{
			Func: &FuncValue{
				Name:     Name("loop"),
				FileName: "loop.gno",
				PkgPath:  "gno.land/p/demo",
			},
		}},
	}
	sink := &sampleCountingSink{}

	m.StartProfilingWithSink(sink, profiler.Options{Type: profiler.ProfileCPU, SampleRate: 1})

	m.incrCPU(1)

	if sink.samples == 0 {
		t.Fatalf("expected maybeEmitSample to trigger on CPU increments")
	}
}
