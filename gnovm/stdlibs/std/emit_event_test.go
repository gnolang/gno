package std

import (
	"encoding/json"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/stretchr/testify/assert"
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
	m.Context = ExecContext{}
	m.Stage = gno.StageAdd
	pushFuncFrame(m, "main")
	pushFuncFrame(m, "Emit")
	_, pkgPath := X_getRealm(m, 0)
	if pkgPath != pkgPath || m.Package.PkgPath != pkgPath {
		panic("inconsistent package paths")
	}
	tests := []struct {
		name           string
		eventType      string
		attrs          []string
		expectedEvents []GnoEvent
		expectPanic    bool
	}{
		{
			name:      "SimpleValid",
			eventType: "test",
			attrs:     []string{"key1", "value1", "key2", "value2"},
			expectedEvents: []GnoEvent{
				{
					Type:    "test",
					PkgPath: pkgPath,
					Attributes: []GnoEventAttribute{
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
			expectedEvents: []GnoEvent{
				{
					Type:    "test",
					PkgPath: pkgPath,
					Attributes: []GnoEventAttribute{
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
			expectedEvents: []GnoEvent{
				{
					Type:    "",
					PkgPath: pkgPath,
					Attributes: []GnoEventAttribute{
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
			expectedEvents: []GnoEvent{
				{
					Type:    "test",
					PkgPath: pkgPath,
					Attributes: []GnoEventAttribute{
						{Key: "", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elgs := sdk.NewEventLogger()
			m.Context = ExecContext{EventLogger: elgs}

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
	m.Context = ExecContext{EventLogger: elgs}

	attrs1 := []string{"key1", "value1", "key2", "value2"}
	attrs2 := []string{"key3", "value3", "key4", "value4"}
	X_emit(m, "test1", attrs1)
	X_emit(m, "test2", attrs2)

	assert.Equal(t, 2, len(elgs.Events()))

	res, err := json.Marshal(elgs.Events())
	if err != nil {
		t.Fatal(err)
	}

	expect := []GnoEvent{
		{
			Type:    "test1",
			PkgPath: pkgPath,
			Attributes: []GnoEventAttribute{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
		{
			Type:    "test2",
			PkgPath: pkgPath,
			Attributes: []GnoEventAttribute{
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
