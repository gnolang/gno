package stdlibs

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// These tests run crypto/secp256k1.RecoverPubkey in a Machine through the
// generated binding and the production gas table: the Gno wrapper, the native
// argument conversion and the flat native charge. The native itself is tested
// on the whole recovery corpus in crypto/secp256k1/recover_test.go.

const secp256k1Path = "crypto/secp256k1"

// recoverPubkeyCharge returns the flat gas charge of the recoverPubkey native.
func recoverPubkeyCharge(t *testing.T) int64 {
	t.Helper()
	for _, e := range calibratedNativeGas {
		if e.Pkg != secp256k1Path || e.Fn != "recoverPubkey" {
			continue
		}
		if e.SlopeKind != SizeFlat || e.Slope != 0 || e.Slope2 != 0 ||
			e.PostBase != 0 || e.PostSlope != 0 || e.PostSlope2 != 0 {
			t.Fatal("the recoverPubkey row must be a flat charge")
		}
		return e.Base * gno.GasFactorCPU
	}
	t.Fatal("no gas row for crypto/secp256k1.recoverPubkey")
	return 0
}

// recoveryMeter is a real gas meter that also records how much gas was used
// when the native's flat charge began.
type recoveryMeter struct {
	stypes.GasMeter
	charge        int64
	beforeNative  int64
	nativeCharges int
}

func (m *recoveryMeter) ConsumeGas(amount int64, descriptor string) {
	if descriptor == "CPUCycles" && amount == m.charge {
		m.beforeNative = m.GasConsumed()
		m.nativeCharges++
	}
	m.GasMeter.ConsumeGas(amount, descriptor)
}

// newRecoveryMachine loads crypto/secp256k1 unmetered, then attaches meter.
func newRecoveryMachine(t *testing.T, meter stypes.GasMeter) *gno.Machine {
	t.Helper()
	mpkg, err := gno.ReadMemPackage(secp256k1Path, secp256k1Path, gno.MPStdlibProd)
	if err != nil {
		t.Fatal(err)
	}
	m := gno.NewMachineWithOptions(gno.MachineOptions{PkgPath: secp256k1Path, MaxAllocBytes: 64 << 20})
	t.Cleanup(m.Release)
	m.Store.SetNativeResolver(NativeResolver)
	m.RunMemPackage(mpkg, false)
	if meter != nil {
		m.GasMeter = meter
		m.Alloc.SetGasMeter(meter)
	}
	return m
}

// evalRecoverPubkey evaluates RecoverPubkey(args) and returns its results, or
// the panic it raised.
func evalRecoverPubkey(m *gno.Machine, args string) (key [65]byte, ok bool, panicked any) {
	defer func() { panicked = recover() }()
	res := m.Eval(m.MustParseExpr("RecoverPubkey(" + args + ")"))
	gno.Gno2GoValue(&res[0], reflect.ValueOf(&key).Elem())
	return key, res[1].GetBool(), nil
}

func byteSliceLiteral(t *testing.T, hexStr string) string {
	t.Helper()
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatal(err)
	}
	elems := make([]string, len(b))
	for i, v := range b {
		elems[i] = fmt.Sprint(v)
	}
	return "[]byte{" + strings.Join(elems, ",") + "}"
}

// countRecoverPubkeyCalls wraps the native, before any Machine resolves it,
// to count its calls and check that both parameters are fixed-size arrays.
func countRecoverPubkeyCalls(t *testing.T) *int {
	t.Helper()
	nf := FindNative(secp256k1Path, "recoverPubkey")
	if nf == nil {
		t.Fatal("no generated binding for crypto/secp256k1.recoverPubkey")
	}
	orig, calls := nf.f, new(int)
	nf.f = func(m *gno.Machine) {
		*calls++
		params := m.LastBlock().Values
		if params[0].T.Kind() != gno.ArrayKind || params[0].GetLength() != 32 ||
			params[1].T.Kind() != gno.ArrayKind || params[1].GetLength() != 65 {
			t.Errorf("native called with %v and %v", params[0].T, params[1].T)
		}
		orig(m)
	}
	t.Cleanup(func() { nf.f = orig })
	return calls
}

type recoveryCase struct {
	Name      string `json:"name"`
	Digest    string `json:"digest_hex"`
	Signature string `json:"signature_hex"`
}

type recoveryOutcome struct {
	Status    string `json:"status"`
	PublicKey string `json:"public_key_hex"`
}

func loadRecoveryCorpus(t *testing.T) ([]recoveryCase, map[string]recoveryOutcome) {
	t.Helper()
	var vectors []recoveryCase
	var want map[string]recoveryOutcome
	for name, v := range map[string]any{"vectors.json": &vectors, "proposed-boundary.json": &want} {
		b, err := os.ReadFile(secp256k1Path + "/testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b, v); err != nil {
			t.Fatal(err)
		}
	}
	if len(vectors) != 28 || len(want) != 28 {
		t.Fatal("recovery corpus is incomplete")
	}
	return vectors, want
}

// The native runs only for a 32-byte digest and a 65-byte signature, and
// always receives fixed-size arrays, so the conversion it pays for is bounded.
func TestRecoverPubkeyNativeEntry(t *testing.T) {
	calls := countRecoverPubkeyCalls(t)
	vectors, want := loadRecoveryCorpus(t)
	for _, v := range vectors {
		t.Run(v.Name, func(t *testing.T) {
			m := newRecoveryMachine(t, nil)
			before := *calls
			key, ok, panicked := evalRecoverPubkey(m, byteSliceLiteral(t, v.Digest)+","+byteSliceLiteral(t, v.Signature))
			if panicked != nil {
				t.Fatalf("panic: %v", panicked)
			}
			w := want[v.Name]
			if ok != (w.Status == "recovered") || (ok && hex.EncodeToString(key[:]) != w.PublicKey) || (!ok && key != [65]byte{}) {
				t.Errorf("got %x, %v; want %+v", key, ok, w)
			}
			entries := 0
			if len(v.Digest) == 64 && len(v.Signature) == 130 {
				entries = 1
			}
			if got := *calls - before; got != entries {
				t.Errorf("native ran %d times, want %d", got, entries)
			}
		})
	}
	// Large slices: wrong lengths stop in Gno, and a large capacity behind the
	// right lengths still reaches the native as two fixed-size arrays (whose
	// zero r it then refuses).
	for _, c := range []struct {
		args    string
		entries int
	}{
		{"nil, nil", 0},
		{"make([]byte, 1<<20), make([]byte, 65)", 0},
		{"make([]byte, 32), make([]byte, 1<<20)", 0},
		{"make([]byte, 32, 1<<20), make([]byte, 65, 1<<20)", 1},
	} {
		m := newRecoveryMachine(t, nil)
		before := *calls
		key, ok, panicked := evalRecoverPubkey(m, c.args)
		if panicked != nil || ok || key != [65]byte{} || *calls-before != c.entries {
			t.Errorf("RecoverPubkey(%s): got %x, %v, panic %v, %d native calls", c.args, key, ok, panicked, *calls-before)
		}
	}
}

// With a finite gas meter, a call funded one unit short of the native charge
// runs out of gas before the native starts; at the charge, the native runs.
func TestRecoverPubkeyGasBoundary(t *testing.T) {
	charge := recoverPubkeyCharge(t)
	calls := countRecoverPubkeyCalls(t)
	vectors, _ := loadRecoveryCorpus(t)
	args := byteSliceLiteral(t, vectors[0].Digest) + "," + byteSliceLiteral(t, vectors[0].Signature)

	probe := &recoveryMeter{GasMeter: stypes.NewGasMeter(math.MaxInt64), charge: charge}
	if _, ok, panicked := evalRecoverPubkey(newRecoveryMachine(t, probe), args); panicked != nil || !ok || probe.nativeCharges != 1 {
		t.Fatalf("unmetered run: ok %v, panic %v, %d native charges", ok, panicked, probe.nativeCharges)
	}
	nativeStart := probe.beforeNative + charge
	full := probe.GasConsumed()

	for _, limit := range []int64{nativeStart - 1, nativeStart, full - 1, full} {
		meter := &recoveryMeter{GasMeter: stypes.NewGasMeter(limit), charge: charge}
		before := *calls
		_, ok, panicked := evalRecoverPubkey(newRecoveryMachine(t, meter), args)
		wantCalls := 1
		if limit < nativeStart {
			wantCalls = 0
		}
		if got := *calls - before; got != wantCalls {
			t.Errorf("limit %d: native ran %d times, want %d", limit, got, wantCalls)
		}
		if limit < full {
			if _, oog := panicked.(stypes.OutOfGasError); !oog {
				t.Errorf("limit %d: got %v (%T), want out of gas", limit, panicked, panicked)
			}
		} else if panicked != nil || !ok {
			t.Errorf("limit %d: ok %v, panic %v", limit, ok, panicked)
		}
	}
}

// Every 32/65-byte input pays the same flat charge, whether the native
// recovers a key or refuses the signature, early or late, and the whole call
// costs the same.
func TestRecoverPubkeyFlatCharge(t *testing.T) {
	charge := recoverPubkeyCharge(t)
	vectors, want := loadRecoveryCorpus(t)
	type input struct {
		name, digest, signature string
		ok                      bool
	}
	var inputs []input
	for _, v := range vectors {
		if len(v.Digest) == 64 && len(v.Signature) == 130 {
			inputs = append(inputs, input{v.Name, v.Digest, v.Signature, want[v.Name].Status == "recovered"})
		}
	}
	const one = "0000000000000000000000000000000000000000000000000000000000000001"
	inputs = append(inputs,
		// The native returns before recovery.
		input{"recovery ID 2", vectors[0].Digest, vectors[0].Signature[:128] + "02", false},
		// Refused after both multiplications: the key would be G - G.
		input{"point at infinity", one, "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798" + one + "00", false},
	)
	var firstGas int64
	for i, in := range inputs {
		meter := &recoveryMeter{GasMeter: stypes.NewGasMeter(math.MaxInt64), charge: charge}
		_, ok, panicked := evalRecoverPubkey(newRecoveryMachine(t, meter), byteSliceLiteral(t, in.digest)+","+byteSliceLiteral(t, in.signature))
		if panicked != nil || ok != in.ok || meter.nativeCharges != 1 {
			t.Fatalf("%s: ok %v, panic %v, %d native charges", in.name, ok, panicked, meter.nativeCharges)
		}
		if i == 0 {
			firstGas = meter.GasConsumed()
		} else if got := meter.GasConsumed(); got != firstGas {
			t.Errorf("%s: used %d gas, %s used %d", in.name, got, inputs[0].name, firstGas)
		}
	}
}
