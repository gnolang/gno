package chain

import (
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
	"unsafe"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pkgPath  = "emit_test"
	fileName = "emit_test.gno"
)

var line = 1

func pushFuncFrame(m *gno.Machine, name gno.Name) {
	fd := &gno.FuncDecl{}
	fd.SetLocation(gno.Location{
		PkgPath: pkgPath,
		File:    fileName,
		Span: gno.Span{ // fake unique span.
			Pos: gno.Pos{Line: line, Column: 0},
			End: gno.Pos{Line: line, Column: 100},
		},
	})
	line++
	fv := &gno.FuncValue{Name: name, PkgPath: m.Package.PkgPath, Source: fd}
	m.PushFrameCall(gno.Call(name), fv, gno.TypedValue{}, false) // fake frame
}

func TestEmit(t *testing.T) {
	m := gno.NewMachine(pkgPath, nil)
	m.Context = execctx.ExecContext{}
	m.Stage = gno.StageAdd
	pushFuncFrame(m, "main")
	pushFuncFrame(m, "Emit")
	_, pkgPath := execctx.GetRealm(m, 0)
	if m.Package.PkgPath != pkgPath {
		panic("inconsistent package paths")
	}
	tests := []struct {
		name           string
		eventType      string
		attrs          []string
		expectedEvents []Event
		expectPanic    bool
	}{
		{
			name:      "SimpleValid",
			eventType: "test",
			attrs:     []string{"key1", "value1", "key2", "value2"},
			expectedEvents: []Event{
				{
					Type:    "test",
					PkgPath: pkgPath,
					Attributes: []EventAttribute{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
			expectPanic: false,
		},
		{
			name:        "InvalidAttributes",
			eventType:   "test",
			attrs:       []string{"key1", "value1", "key2"},
			expectPanic: true,
		},
		{
			name:      "EmptyAttribute",
			eventType: "test",
			attrs:     []string{"key1", "", "key2", "value2"},
			expectedEvents: []Event{
				{
					Type:    "test",
					PkgPath: pkgPath,
					Attributes: []EventAttribute{
						{Key: "key1", Value: ""},
						{Key: "key2", Value: "value2"},
					},
				},
			},
			expectPanic: false,
		},
		{
			name:      "EmptyType",
			eventType: "",
			attrs:     []string{"key1", "value1", "key2", "value2"},
			expectedEvents: []Event{
				{
					Type:    "",
					PkgPath: pkgPath,
					Attributes: []EventAttribute{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
			expectPanic: false,
		},
		{
			name:      "EmptyAttributeKey",
			eventType: "test",
			attrs:     []string{"", "value1", "key2", "value2"},
			expectedEvents: []Event{
				{
					Type:    "test",
					PkgPath: pkgPath,
					Attributes: []EventAttribute{
						{Key: "", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
			expectPanic: false,
		},
		{
			// A value exactly at the cap is kept verbatim — never truncated.
			name:      "ValueAtCap",
			eventType: "test",
			attrs:     []string{"key1", strings.Repeat("v", MaxEventAttrLen)},
			expectedEvents: []Event{
				{
					Type:    "test",
					PkgPath: pkgPath,
					Attributes: []EventAttribute{
						{Key: "key1", Value: strings.Repeat("v", MaxEventAttrLen)},
					},
				},
			},
			expectPanic: false,
		},
		{
			// A value over the cap panics rather than silently truncating.
			name:        "ValueTooLong",
			eventType:   "test",
			attrs:       []string{"key1", strings.Repeat("v", MaxEventAttrLen+1)},
			expectPanic: true,
		},
		{
			name:        "KeyTooLong",
			eventType:   "test",
			attrs:       []string{strings.Repeat("k", MaxEventAttrLen+1), "value1"},
			expectPanic: true,
		},
		{
			name:        "TypeTooLong",
			eventType:   strings.Repeat("t", MaxEventAttrLen+1),
			attrs:       []string{"key1", "value1"},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elgs := sdk.NewEventLogger()
			m.Context = execctx.ExecContext{EventLogger: elgs}

			if tt.expectPanic {
				assert.Panics(t, func() {
					X_emit(m, tt.eventType, tt.attrs)
				})
				// X_emit() should m.Panic(), but it should not
				// set m.Exception. That happens after m.Run()
				// recovers and then calls m.pushPanic().
				// But stdlib should NOT call m.pushPanic().
				assert.Nil(t, m.Exception)
			} else {
				X_emit(m, tt.eventType, tt.attrs)
				assert.Equal(t, len(tt.expectedEvents), len(elgs.Events()))

				res, err := json.Marshal(elgs.Events())
				if err != nil {
					t.Fatal(err)
				}

				expectRes, err := json.Marshal(tt.expectedEvents)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, string(expectRes), string(res))
			}
		})
	}
}

func TestEmit_MultipleEvents(t *testing.T) {
	t.Parallel()
	m := gno.NewMachine(pkgPath, nil)
	pushFuncFrame(m, "main")
	pushFuncFrame(m, "Emit")

	elgs := sdk.NewEventLogger()
	m.Context = execctx.ExecContext{EventLogger: elgs}

	attrs1 := []string{"key1", "value1", "key2", "value2"}
	attrs2 := []string{"key3", "value3", "key4", "value4"}
	X_emit(m, "test1", attrs1)
	X_emit(m, "test2", attrs2)

	assert.Equal(t, 2, len(elgs.Events()))

	res, err := json.Marshal(elgs.Events())
	if err != nil {
		t.Fatal(err)
	}

	expect := []Event{
		{
			Type:    "test1",
			PkgPath: pkgPath,
			Attributes: []EventAttribute{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
		{
			Type:    "test2",
			PkgPath: pkgPath,
			Attributes: []EventAttribute{
				{Key: "key3", Value: "value3"},
				{Key: "key4", Value: "value4"},
			},
		},
	}

	expectRes, err := json.Marshal(expect)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(expectRes), string(res))
}

// eventSize must track amino's encoding of an Event, across varint length
// boundaries and nil vs empty attributes.
func TestEventSizeMatchesMarshalAny(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	lens := []int{0, 1, 127, 128, 4095, 4096}
	pick := func() string { return strings.Repeat("x", lens[r.Intn(len(lens))]) }
	for i := range 3000 {
		e := Event{Type: pick(), PkgPath: pick()}
		if np := r.Intn(MaxEventPairs + 1); np > 0 || r.Intn(2) == 0 {
			e.Attributes = make([]EventAttribute, np)
		}
		for j := range e.Attributes {
			e.Attributes[j] = EventAttribute{Key: pick(), Value: pick()}
		}
		require.Equal(t, len(amino.MustMarshalAny(e)), eventSize(e), "case %d", i)
	}
}

// An emitted substring must not keep its source string alive.
func TestEmitCopiesStrings(t *testing.T) {
	t.Parallel()
	m := gno.NewMachine(pkgPath, nil)
	pushFuncFrame(m, "main")
	pushFuncFrame(m, "Emit")
	elgs := sdk.NewEventLogger()
	m.Context = execctx.ExecContext{EventLogger: elgs}

	src := strings.Repeat("x", 1<<20)
	X_emit(m, src[:1], []string{src[1:2], src[2:3]})

	evt := elgs.Events()[0].(Event)
	for _, s := range []string{evt.Type, evt.Attributes[0].Key, evt.Attributes[0].Value} {
		p := uintptr(unsafe.Pointer(unsafe.StringData(s)))
		base := uintptr(unsafe.Pointer(unsafe.StringData(src)))
		assert.False(t, p >= base && p < base+uintptr(len(src)), "%q shares the source string's memory", s)
	}
}
