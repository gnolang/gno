package test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/profiler"
	"github.com/gnolang/gno/tm2/pkg/std"
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

type recordingProfileWriter struct {
	count   int
	profile *profiler.Profile
}

func (w *recordingProfileWriter) WriteProfile(profile *profiler.Profile, _ *ProfileConfig, _ io.Writer, _ io.Writer, _ gno.Store) error {
	w.count++
	w.profile = profile
	return nil
}

func TestProfileConfigStartIsIdempotent(t *testing.T) {
	pc := &ProfileConfig{
		Enabled: true,
		Type:    "cpu",
	}

	started, err := pc.Start()
	if err != nil {
		t.Fatalf("first start failed: %v", err)
	}
	if !started {
		t.Fatalf("expected first start to initialize profiling")
	}
	firstProfiler := pc.profiler

	started, err = pc.Start()
	if err != nil {
		t.Fatalf("second start failed: %v", err)
	}
	if started {
		t.Fatalf("expected second start to be a no-op when profiler is already running")
	}
	if pc.profiler != firstProfiler {
		t.Fatalf("profiler was reinitialized on second start")
	}

	opts := &TestOptions{
		Output:    io.Discard,
		Error:     io.Discard,
		TestStore: gno.NewStore(nil, nil, nil),
	}
	w := &recordingProfileWriter{}
	if err := pc.Stop(opts, w); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if w.count != 1 || w.profile == nil {
		t.Fatalf("expected writer to receive exactly one profile")
	}
}

func TestProfilingSpansMultiplePackages(t *testing.T) {
	pc := &ProfileConfig{
		Enabled: true,
		Type:    "cpu",
	}
	started, err := pc.Start()
	if err != nil || !started {
		t.Fatalf("failed to start profiling: %v", err)
	}

	rootDir := gnoenv.RootDir()
	opts := NewTestOptions(rootDir, io.Discard, io.Discard, nil)
	opts.Profile = pc

	pkgs := []*std.MemPackage{
		newMemTestPackage("pkgone"),
		newMemTestPackage("pkgtwo"),
	}

	for _, mpkg := range pkgs {
		if err := Test(mpkg, mpkg.Path, opts); err != nil {
			t.Fatalf("Test() failed for %s: %v", mpkg.Path, err)
		}
	}
	if pc.profiler == nil {
		t.Fatalf("profiling was stopped unexpectedly during package loop")
	}

	writer := &recordingProfileWriter{}
	if err := pc.Stop(opts, writer); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if writer.count != 1 {
		t.Fatalf("expected a single profile write, got %d", writer.count)
	}
	if writer.profile == nil || pc.LastProfile() != writer.profile {
		t.Fatalf("profile result not recorded on ProfileConfig")
	}
}

func newMemTestPackage(name string) *std.MemPackage {
	body := fmt.Sprintf(`package %s

import "testing"

func TestValue(t *testing.T) {
	_ = 1 + 1
}
`, name)

	return &std.MemPackage{
		Name: name,
		Path: "gno.land/r/" + name,
		Type: gno.MPUserTest,
		Files: []*std.MemFile{
			{
				Name: name + "_test.gno",
				Body: body,
			},
		},
	}
}
