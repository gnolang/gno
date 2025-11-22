package test

import (
	"bytes"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestProfileConfigInitializeFinalize(t *testing.T) {
	pc := &ProfileConfig{
		Enabled:       true,
		PrintToStdout: true,
		Type:          "cpu",
	}

	if err := pc.initialize(); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if pc.profiler == nil {
		t.Fatalf("expected profiler to be initialized")
	}
	if pc.sink == nil {
		t.Fatalf("expected instrumentation sink to be initialized")
	}

	opts := &TestOptions{
		Output:    &bytes.Buffer{},
		Error:     &bytes.Buffer{},
		TestStore: gno.NewStore(nil, nil, nil),
	}

	if err := pc.finalize(opts, &DefaultProfileWriter{}); err != nil {
		t.Fatalf("finalize failed: %v", err)
	}
}
