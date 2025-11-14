package gnolang

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/instrumentation"
)

type mockInstrumentationSink struct {
	allocCount int
	lastAlloc  *instrumentation.AllocationEvent
}

func (m *mockInstrumentationSink) OnSample(*instrumentation.SampleContext) {}

func (m *mockInstrumentationSink) OnAllocation(ev *instrumentation.AllocationEvent) {
	m.allocCount++
	m.lastAlloc = ev
}

func (m *mockInstrumentationSink) OnLineSample(*instrumentation.LineSample) {}

func TestMachineWithInstrumentationPropagatesToAllocator(t *testing.T) {
	sink := &mockInstrumentationSink{}
	m := NewMachineWithOptions(MachineOptions{
		MaxAllocBytes:   1024,
		Instrumentation: sink,
	})
	defer m.Release()

	if m.Alloc == nil {
		t.Fatalf("expected allocator to be initialized")
	}

	m.Alloc.AllocatePointer()

	if sink.allocCount == 0 {
		t.Fatalf("expected allocation event to reach sink")
	}
	if sink.lastAlloc == nil || sink.lastAlloc.Bytes == 0 {
		t.Fatalf("invalid allocation event %+v", sink.lastAlloc)
	}
}

func TestAllocatorEmitsAllocationEvents(t *testing.T) {
	alloc := NewAllocator(1024)
	sink := &mockInstrumentationSink{}
	alloc.SetInstrumentationSink(sink)

	const size int64 = 64
	alloc.Allocate(size)

	if sink.allocCount != 1 {
		t.Fatalf("expected 1 allocation event, got %d", sink.allocCount)
	}
	if sink.lastAlloc == nil || sink.lastAlloc.Bytes != size {
		t.Fatalf("expected allocation size %d, got %+v", size, sink.lastAlloc)
	}
}
