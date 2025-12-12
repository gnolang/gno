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

func TestAllocationStackInjectorAddsMachineFrames(t *testing.T) {
	m := &Machine{
		Frames: []Frame{
			{
				Func: &FuncValue{
					Name:     Name("parent"),
					FileName: "parent.gno",
					PkgPath:  "gno.land/p/demo",
				},
			},
			{
				Func: &FuncValue{
					Name:     Name("leaf"),
					FileName: "leaf.gno",
					PkgPath:  "gno.land/p/demo",
				},
			},
		},
	}
	sink := &mockInstrumentationSink{}
	injector := &allocationStackInjector{
		machine: m,
		sink:    sink,
	}

	injector.OnAllocation(&instrumentation.AllocationEvent{
		Bytes:   64,
		Objects: 1,
	})

	if sink.allocCount != 1 {
		t.Fatalf("expected allocation to reach sink, got %d", sink.allocCount)
	}
	if sink.lastAlloc == nil {
		t.Fatalf("expected allocation event to be recorded")
	}
	if len(sink.lastAlloc.Stack) != 2 {
		t.Fatalf("expected stack of length 2, got %d", len(sink.lastAlloc.Stack))
	}
	if sink.lastAlloc.Stack[0].FuncName != "leaf" {
		t.Fatalf("expected top frame leaf, got %s", sink.lastAlloc.Stack[0].FuncName)
	}
	if sink.lastAlloc.Stack[1].FuncName != "parent" {
		t.Fatalf("expected parent frame, got %s", sink.lastAlloc.Stack[1].FuncName)
	}
	if sink.lastAlloc.Stack[0].PkgPath != "gno.land/p/demo" {
		t.Fatalf("expected pkg path propagated, got %+v", sink.lastAlloc.Stack[0])
	}
}
