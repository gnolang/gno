package gnolang

import (
	"bytes"
	"testing"
)

// TestDebugEmptyOps verifies that Debug() does not panic when m.Ops is empty.
// This covers the guard at the end of Debug() that checks len(m.Ops) == 0
// before accessing m.Ops[len(m.Ops)-1].
func TestDebugEmptyOps(t *testing.T) {
	// Set up a minimal machine with the debugger enabled but in DebugAtRun
	// state with the debugger disabled, so the loop exits immediately and
	// we reach the post-loop ops access.
	var out bytes.Buffer
	m := &Machine{
		Ops: []Op{}, // empty ops stack
	}
	m.Debugger.enabled = false
	m.Debugger.state = DebugAtRun
	m.Debugger.out = &out
	// Provide minimal block chain so debugUpdateLocation doesn't panic.
	m.Blocks = []*Block{{Source: &PackageNode{
		StaticBlock: StaticBlock{
			Block: Block{Source: nil},
		},
	}}}

	// Debug() should return without panicking when Ops is empty.
	m.Debug()
}

// TestDebugEmptyCallStackOnReturn verifies that Debug() does not panic when
// m.Ops contains OpReturn but m.Debugger.call is empty.
// This covers the guard that checks len(m.Debugger.call) > 0 before popping.
func TestDebugEmptyCallStackOnReturn(t *testing.T) {
	var out bytes.Buffer
	m := &Machine{
		Ops: []Op{OpReturn},
	}
	m.Debugger.enabled = false
	m.Debugger.state = DebugAtRun
	m.Debugger.out = &out
	m.Debugger.call = nil // empty call stack
	m.Blocks = []*Block{{Source: &PackageNode{
		StaticBlock: StaticBlock{
			Block: Block{Source: nil},
		},
	}}}

	// Debug() should handle OpReturn with empty call stack without panicking.
	m.Debug()

	if len(m.Debugger.call) != 0 {
		t.Errorf("expected call stack to remain empty, got length %d", len(m.Debugger.call))
	}
}

// TestDebugEmptyCallStackOnReturnFromBlock is the same as above but with
// OpReturnFromBlock instead of OpReturn.
func TestDebugEmptyCallStackOnReturnFromBlock(t *testing.T) {
	var out bytes.Buffer
	m := &Machine{
		Ops: []Op{OpReturnFromBlock},
	}
	m.Debugger.enabled = false
	m.Debugger.state = DebugAtRun
	m.Debugger.out = &out
	m.Debugger.call = nil
	m.Blocks = []*Block{{Source: &PackageNode{
		StaticBlock: StaticBlock{
			Block: Block{Source: nil},
		},
	}}}

	m.Debug()

	if len(m.Debugger.call) != 0 {
		t.Errorf("expected call stack to remain empty, got length %d", len(m.Debugger.call))
	}
}

// TestDebugFrameLocExceedsCallStack verifies that debugFrameLoc does not
// panic when n exceeds len(m.Debugger.call). It should return m.Debugger.loc.
func TestDebugFrameLocExceedsCallStack(t *testing.T) {
	m := &Machine{}
	m.Debugger.loc = Location{PkgPath: "test", File: "test.gno", Span: Span{Pos: Pos{Line: 42, Column: 1}}}
	m.Debugger.call = nil // empty call stack

	// n=1 exceeds len(call)=0, should fall back to m.Debugger.loc.
	if loc := debugFrameLoc(m, 1); loc != m.Debugger.loc {
		t.Errorf("expected %v, got %v", m.Debugger.loc, loc)
	}

	// n=2 exceeds len(call)=1, should still fall back.
	m.Debugger.call = []Location{{PkgPath: "other"}}
	if loc := debugFrameLoc(m, 2); loc != m.Debugger.loc {
		t.Errorf("expected %v, got %v", m.Debugger.loc, loc)
	}
}
