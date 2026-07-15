package gasprof

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/require"
)

// pprof value indices (must match the emission order in WritePprof).
const (
	viCPU = iota
	viAlloc
	viStore
	viOther
	viRefund
	viTotal
)

// sim drives a Profiler exactly like the Machine does: cursor Enter/SyncDepth
// events plus a wrapped GasMeter that observes charges.
type sim struct {
	p *Profiler
	m store.GasMeter
}

func newSim() *sim {
	p := New()
	return &sim{p: p, m: WrapMeter(store.NewInfiniteGasMeter(), p)}
}

func (s *sim) enter(name string)   { s.p.Enter(Frame{Func: name}) }
func (s *sim) pop()                { s.p.Pop() }
func (s *sim) sync(callFrames int) { s.p.SyncDepth(callFrames) }
func (s *sim) cpu(a int64)         { s.m.ConsumeGas(store.Gas(a), "CPUCycles") }
func (s *sim) alloc(a int64)       { s.m.ConsumeGas(store.Gas(a), "memory allocation") }
func (s *sim) storeGas(a int64)    { s.m.ConsumeGas(store.Gas(a), "DepthSet") }
func (s *sim) other(a int64)       { s.m.ConsumeGas(store.Gas(a), "txSize") }
func (s *sim) refund(a int64)      { s.m.RefundGas(store.Gas(a), "Refund") }

func TestCursor_descendAscendAndFolded(t *testing.T) {
	s := newSim()
	require.True(t, s.p.Empty())

	// Run() { cpu 3; Insert() { cpu 8 }; cpu 2 }
	s.enter("Run")
	s.cpu(3)
	s.enter("Insert")
	s.cpu(8)
	s.sync(1) // Insert returns -> back to Run (1 call frame)
	s.cpu(2)
	s.sync(0) // Run returns

	require.False(t, s.p.Empty())
	var b bytes.Buffer
	require.NoError(t, s.p.WriteFolded(&b))
	// Root has no gas (not emitted). Run flat=3+2=5, Insert flat=8.
	require.Equal(t, "(root);Run 5\n(root);Run;Insert 8\n", b.String())
}

func TestCursor_popAscendsOne(t *testing.T) {
	s := newSim()
	s.enter("Run")
	s.enter("f")
	s.cpu(8)
	s.pop()  // f returns via the O(1) single-frame path
	s.cpu(2) // back in Run
	s.pop()  // Run returns
	s.pop()  // extra pop at root is a no-op (must not underflow)
	s.cpu(1) // now at root

	var b bytes.Buffer
	require.NoError(t, s.p.WriteFolded(&b))
	require.Equal(t, "(root) 1\n(root);Run 2\n(root);Run;f 8\n", b.String())
}

// The O(1) Pop path (used by PopFrame) must produce the exact same tree as the
// absolute SyncDepth path (used on the revive/bulk path). Drive both with an
// identical enter/charge/return script and assert byte-identical output — this
// locks the equivalence the O(1) optimization depends on, which the
// cursor-blind reconciliation invariant cannot.
func TestCursor_popEqualsSyncDepth(t *testing.T) {
	// (op, arg): "e"=enter arg, "c"=charge arg-as-gas, "r"=return (one frame).
	script := []struct {
		op  string
		arg string
	}{
		{"e", "Run"}, {"c", "1"},
		{"e", "a"}, {"c", "2"}, {"e", "b"}, {"c", "3"}, {"r", ""}, {"r", ""},
		{"e", "a"}, {"c", "4"}, {"r", ""}, // re-enter a (same node)
		{"c", "5"},
		{"e", "c"}, {"e", "c"}, {"c", "6"}, {"r", ""}, {"r", ""}, // recursion
		{"r", ""},
	}

	run := func(useSyncDepth bool) string {
		p := New()
		w := WrapMeter(store.NewInfiniteGasMeter(), p)
		depth := 0
		for _, s := range script {
			switch s.op {
			case "e":
				p.Enter(Frame{Func: s.arg})
				depth++
			case "c":
				n, _ := parseInt(s.arg)
				w.ConsumeGas(store.Gas(n), "CPUCycles")
			case "r":
				depth--
				if useSyncDepth {
					p.SyncDepth(depth)
				} else {
					p.Pop()
				}
			}
		}
		var b bytes.Buffer
		require.NoError(t, p.WriteFolded(&b))
		return b.String()
	}

	require.Equal(t, run(true), run(false), "Pop and SyncDepth must yield identical trees")
	require.NotEmpty(t, run(false))
}

func parseInt(s string) (int64, error) {
	var n int64
	for _, c := range s {
		n = n*10 + int64(c-'0')
	}
	return n, nil
}

func TestCursor_recursionBuildsTower(t *testing.T) {
	s := newSim()
	s.enter("Run")
	// fib(3)-ish: Run -> fib -> fib -> fib, each charging on the way down.
	s.enter("fib")
	s.cpu(1)
	s.enter("fib")
	s.cpu(2)
	s.enter("fib")
	s.cpu(4)
	s.sync(0) // unwind everything at once (like an abort)

	var b bytes.Buffer
	require.NoError(t, s.p.WriteFolded(&b))
	require.Equal(t,
		"(root);Run;fib 1\n(root);Run;fib;fib 2\n(root);Run;fib;fib;fib 4\n",
		b.String())
}

func TestCursor_reEntryAccumulates(t *testing.T) {
	s := newSim()
	s.enter("Run")
	s.enter("f")
	s.cpu(10)
	s.sync(1)    // f returns
	s.enter("f") // Run calls f again — same tree node
	s.cpu(5)
	s.sync(1)

	require.Equal(t, int64(15), sumChild(t, s.p, "(root)", "Run", "f"))
}

func TestCursor_resetStrandedCursor(t *testing.T) {
	s := newSim()
	s.enter("Run")
	s.enter("deep")
	s.cpu(7)
	// Simulate a hard abort that never popped: Reset returns cursor to root.
	s.p.Reset()
	s.enter("Next")
	s.cpu(3)
	// "Next" must hang off root, not off the stranded "deep".
	require.Equal(t, int64(3), sumChild(t, s.p, "(root)", "Next"))
}

// dimensionOf is a hand-maintained allow-list decoupled from the gas-config
// descriptor constants. This pins the current mapping so a regression in the
// switch (or a descriptor silently reclassifying) fails loudly. New descriptors
// intentionally fall through to dimOther (reconciliation stays exact; only the
// store/cpu split would be off — see the ADR).
func TestDimensionOf_classifiesKnownDescriptors(t *testing.T) {
	cases := map[string]int{
		"CPUCycles":          dimCPU,
		"GC":                 dimCPU,
		"parsing":            dimCPU,
		"ComputeMapKey":      dimCPU,
		"stream output":      dimCPU,
		"memory allocation":  dimAlloc,
		"DepthReadFlat":      dimStore,
		"DepthSet":           dimStore,
		"DepthDelete":        dimStore,
		"ReadFlat":           dimStore,
		"ReadPerByte":        dimStore,
		"WriteFlat":          dimStore,
		"Delete":             dimStore,
		"IterNextFlat":       dimStore,
		"ValuePerByte":       dimStore,
		"AminoEncodePerByte": dimStore,
		"AminoDecodePerByte": dimStore,
		"txSize":             dimOther,
		"ante verify: x":     dimOther,
		"some-new-descr":     dimOther, // unknown -> other (documented fallback)
	}
	for desc, want := range cases {
		require.Equal(t, want, dimensionOf(desc), "descriptor %q", desc)
	}
}

func TestBook_anteSnapshotReconciles(t *testing.T) {
	s := newSim()
	// Simulate an install-time ante snapshot booked into the tree, then normal
	// post-install charges from execution.
	s.p.Book("(ante)", 11_000)
	s.enter("pkg.Run")
	s.cpu(50)
	s.storeGas(30)

	tot := s.p.Totals()
	require.Equal(t, int64(11_000), tot.Other, "(ante) booked in other dimension")
	require.Equal(t, int64(50), tot.CPU)
	require.Equal(t, int64(30), tot.Store)

	prof := encodeDecode(t, s.p)
	require.Equal(t, int64(11_000), prof.valueOfStack(viOther, "(ante)", "(root)"))
}

func TestChargesBeforeAnyFrameGoToRoot(t *testing.T) {
	s := newSim()
	s.cpu(9) // charged before entering any call frame
	s.enter("Run")
	s.cpu(1)
	var b bytes.Buffer
	require.NoError(t, s.p.WriteFolded(&b))
	require.Equal(t, "(root) 9\n(root);Run 1\n", b.String())
}

func TestDimensions_bucketedByDescriptor(t *testing.T) {
	s := newSim()
	s.enter("Run")
	s.cpu(100)
	s.alloc(40)
	s.storeGas(25)
	s.other(5)
	s.refund(10) // clamped to <= consumed; here consumed=170 so full 10

	tot := s.p.Totals()
	require.Equal(t, Totals{CPU: 100, Alloc: 40, Store: 25, Other: 5, Refund: 10}, tot)

	// And through pprof value indices.
	prof := encodeDecode(t, s.p)
	require.Equal(t, int64(100), prof.valueOfStack(viCPU, "Run", "(root)"))
	require.Equal(t, int64(40), prof.valueOfStack(viAlloc, "Run", "(root)"))
	require.Equal(t, int64(25), prof.valueOfStack(viStore, "Run", "(root)"))
	require.Equal(t, int64(5), prof.valueOfStack(viOther, "Run", "(root)"))
	require.Equal(t, int64(10), prof.valueOfStack(viRefund, "Run", "(root)"))
	require.Equal(t, int64(170), prof.valueOfStack(viTotal, "Run", "(root)")) // cpu+alloc+store+other
}

func TestMeter_delegatesAndIsObservationOnly(t *testing.T) {
	inner := store.NewInfiniteGasMeter()
	p := New()
	w := WrapMeter(inner, p)
	p.Enter(Frame{Func: "Run"})

	w.ConsumeGas(30, "CPUCycles")
	w.ConsumeGas(20, "memory allocation")
	// The wrapped meter's accounting must be identical to charging inner directly.
	require.Equal(t, store.Gas(50), w.GasConsumed())
	require.Equal(t, store.Gas(50), inner.GasConsumed())

	w.RefundGas(5, "Refund")
	require.Equal(t, store.Gas(45), inner.GasConsumed())

	// Unwrap returns the original meter.
	require.Equal(t, inner, w.(interface{ Unwrap() store.GasMeter }).Unwrap())
}

func TestWritePprof_structureAndAttribution(t *testing.T) {
	s := newSim()
	s.enter("pkg.Run")
	s.cpu(7)
	s.enter("pkg.g")
	s.cpu(13)
	s.alloc(4)
	s.sync(1)

	prof := encodeDecode(t, s.p)

	// Six sample types; default = total_gas.
	require.Equal(t, []valueType{
		{"cpu_gas", "gas"}, {"alloc_gas", "gas"}, {"store_gas", "gas"},
		{"other_gas", "gas"}, {"refund_gas", "gas"}, {"total_gas", "gas"},
	}, prof.sampleTypes)
	require.Equal(t, "total_gas", prof.defaultSampleType)

	// Leaf-first attribution: pkg.g's stack is [g, Run, root].
	require.Equal(t, int64(13), prof.valueOfStack(viCPU, "pkg.g", "pkg.Run", "(root)"))
	require.Equal(t, int64(4), prof.valueOfStack(viAlloc, "pkg.g", "pkg.Run", "(root)"))
	require.Equal(t, int64(7), prof.valueOfStack(viCPU, "pkg.Run", "(root)"))

	// Functions: (root), pkg.Run, pkg.g.
	require.Len(t, prof.funcs, 3)
	prof.requireIntegrity(t)
}

func TestWritePprof_refundRoundTrip(t *testing.T) {
	s := newSim()
	s.enter("R")
	s.cpu(10)
	s.refund(3)
	prof := encodeDecode(t, s.p)
	require.Equal(t, int64(10), prof.valueOfStack(viCPU, "R", "(root)"))
	require.Equal(t, int64(3), prof.valueOfStack(viRefund, "R", "(root)"))
	// total is gross (pre-refund); refund is its own index.
	require.Equal(t, int64(10), prof.valueOfStack(viTotal, "R", "(root)"))
}

func TestWritePprof_empty(t *testing.T) {
	p := New()
	var buf bytes.Buffer
	require.NoError(t, p.WritePprof(&buf))
	prof := decode(t, buf.Bytes())
	require.Len(t, prof.sampleTypes, 6)
	require.Empty(t, prof.samples)
	require.Equal(t, "", prof.strings[0])
}

func TestWritePprof_deepRecursionValid(t *testing.T) {
	s := newSim()
	s.enter("Run")
	for i := range 200 {
		s.enter("rec")
		s.cpu(int64(i + 1))
	}
	s.sync(0)
	prof := encodeDecode(t, s.p)
	require.Len(t, prof.funcs, 3) // (root), Run, rec
	require.Positive(t, len(prof.samples))
	prof.requireIntegrity(t)
}

// Conformance smoke test: real go tool pprof must accept the profile.
func TestWritePprof_goToolPprofAccepts(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	s := newSim()
	s.enter("pkg.Run")
	s.cpu(7)
	s.enter("pkg.hot")
	s.cpu(93)
	s.alloc(20)

	dir := t.TempDir()
	path := filepath.Join(dir, "gas.pprof")
	f, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, s.p.WritePprof(f))
	require.NoError(t, f.Close())

	out, err := exec.Command(goBin, "tool", "pprof", "-top", "-nodecount=10", path).CombinedOutput()
	require.NoError(t, err, "go tool pprof rejected the profile:\n%s", out)
	require.Contains(t, string(out), "total_gas")
	require.Contains(t, string(out), "pkg.hot")

	// A specific dimension is selectable by name.
	out, err = exec.Command(goBin, "tool", "pprof", "-top", "-sample_index=alloc_gas", path).CombinedOutput()
	require.NoError(t, err, "%s", out)
	require.Contains(t, string(out), "alloc_gas")
}

// ---------------------------------------------------------------------------
// minimal pprof reader for validation (no external dependency)
// ---------------------------------------------------------------------------

type valueType struct{ typ, unit string }
type decodedFunc struct {
	name, file string
	line       int64
}
type decodedSample struct {
	locIDs []uint64 // leaf-first
	values []int64
}
type decodedProfile struct {
	strings           []string
	sampleTypes       []valueType
	funcs             map[uint64]decodedFunc
	locs              map[uint64]uint64 // location id -> function id
	samples           []decodedSample
	defaultSampleType string
}

// valueOfStack returns value index vi for the sample whose leaf-first function
// chain equals namesLeafFirst; a sentinel if no exact match, so a wrong stack
// fails equality asserts loudly.
func (p *decodedProfile) valueOfStack(vi int, namesLeafFirst ...string) int64 {
	for _, s := range p.samples {
		if len(s.locIDs) != len(namesLeafFirst) {
			continue
		}
		ok := true
		for i, lid := range s.locIDs {
			if p.funcs[p.locs[lid]].name != namesLeafFirst[i] {
				ok = false
				break
			}
		}
		if ok {
			return s.values[vi]
		}
	}
	return 1 << 62
}

func (p *decodedProfile) requireIntegrity(t *testing.T) {
	t.Helper()
	for _, s := range p.samples {
		require.NotEmpty(t, s.locIDs)
		require.Len(t, s.values, len(p.sampleTypes))
		for _, lid := range s.locIDs {
			fid, ok := p.locs[lid]
			require.True(t, ok, "sample references unknown location %d", lid)
			_, ok = p.funcs[fid]
			require.True(t, ok, "location %d references unknown function %d", lid, fid)
		}
	}
}

func encodeDecode(t *testing.T, p *Profiler) *decodedProfile {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, p.WritePprof(&buf))
	return decode(t, buf.Bytes())
}

func decode(t *testing.T, gzbytes []byte) *decodedProfile {
	t.Helper()
	raw := gunzip(t, gzbytes)
	top := parseFields(raw)
	p := &decodedProfile{funcs: map[uint64]decodedFunc{}, locs: map[uint64]uint64{}}

	for _, e := range top[6] {
		p.strings = append(p.strings, string(e.data))
	}
	str := func(i uint64) string {
		require.Less(t, int(i), len(p.strings), "string index out of range")
		return p.strings[i]
	}
	for _, e := range top[1] {
		vt := parseFields(e.data)
		p.sampleTypes = append(p.sampleTypes, valueType{str(vt[1][0].v), str(vt[2][0].v)})
	}
	for _, e := range top[5] {
		fn := parseFields(e.data)
		p.funcs[fn[1][0].v] = decodedFunc{str(fn[2][0].v), str(fn[4][0].v), int64(fn[5][0].v)}
	}
	for _, e := range top[4] {
		loc := parseFields(e.data)
		ln := parseFields(loc[4][0].data)
		p.locs[loc[1][0].v] = ln[1][0].v
	}
	for _, e := range top[2] {
		sm := parseFields(e.data)
		var s decodedSample
		for _, l := range sm[1] {
			s.locIDs = append(s.locIDs, l.v)
		}
		for _, val := range sm[2] {
			s.values = append(s.values, int64(val.v))
		}
		p.samples = append(p.samples, s)
	}
	if len(top[14]) == 1 {
		p.defaultSampleType = str(top[14][0].v)
	}
	return p
}

// sumChild returns the flat cpu gas of the node at the given root->leaf path,
// by decoding the profile.
func sumChild(t *testing.T, p *Profiler, pathRootFirst ...string) int64 {
	t.Helper()
	prof := encodeDecode(t, p)
	leafFirst := make([]string, len(pathRootFirst))
	for i, n := range pathRootFirst {
		leafFirst[len(pathRootFirst)-1-i] = n
	}
	return prof.valueOfStack(viCPU, leafFirst...)
}

type pfield struct {
	v    uint64
	data []byte
}
type pfields map[int][]pfield

func parseFields(b []byte) pfields {
	f := pfields{}
	i := 0
	for i < len(b) {
		tag, n := binary.Uvarint(b[i:])
		if n <= 0 {
			panic("bad tag varint")
		}
		i += n
		fn := int(tag >> 3)
		switch tag & 7 {
		case 0:
			v, n := binary.Uvarint(b[i:])
			if n <= 0 {
				panic("bad varint")
			}
			i += n
			f[fn] = append(f[fn], pfield{v: v})
		case 2:
			l, n := binary.Uvarint(b[i:])
			if n <= 0 {
				panic("bad length")
			}
			i += n
			f[fn] = append(f[fn], pfield{data: b[i : i+int(l)]})
			i += int(l)
		case 5:
			i += 4
		case 1:
			i += 8
		default:
			panic("unsupported wire type")
		}
	}
	return f
}

func gunzip(t *testing.T, b []byte) []byte {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(b))
	require.NoError(t, err)
	out, err := io.ReadAll(gz)
	require.NoError(t, err)
	require.NoError(t, gz.Close())
	return out
}
