package std

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestEmit(t *testing.T) {
	m := gno.NewMachine("emit", nil)

	m.Context = ExecContext{}

	_, pkgPath := X_getRealm(m, 0)
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
					Func:    "",
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
					Func:    "",
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
					Func:    "",
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
					Func:    "",
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
	m := gno.NewMachine("emit", nil)

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
			PkgPath: "",
			Func:    "",
			Attributes: []GnoEventAttribute{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
		{
			Type:    "test2",
			PkgPath: "",
			Func:    "",
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

func TestEmit_ContractInteraction(t *testing.T) {
	const (
		testFoo = "foo"
		testQux = "qux"
	)

	type (
		contractA struct {
			foo func(*gno.Machine, func())
		}

		contractB struct {
			qux func(m *gno.Machine)
		}
	)

	t.Parallel()
	m := gno.NewMachine("emit", nil)
	elgs := sdk.NewEventLogger()
	m.Context = ExecContext{EventLogger: elgs}

	baz := func(m *gno.Machine) {
		X_emit(m, testFoo, []string{"k1", "v1", "k2", "v2"})
	}

	a := &contractA{
		foo: func(m *gno.Machine, cb func()) {
			baz(m)
			cb()
		},
	}
	b := &contractB{
		qux: func(m *gno.Machine) {
			X_emit(m, testQux, []string{"bar", "baz"})
		},
	}

	a.foo(m, func() {
		b.qux(m)
	})

	assert.Equal(t, 2, len(elgs.Events()))

	res, err := json.Marshal(elgs.Events())
	if err != nil {
		t.Fatal(err)
	}

	expected := `[{"type":"foo","attrs":[{"key":"k1","value":"v1"},{"key":"k2","value":"v2"}],"pkg_path":"","func":""},{"type":"qux","attrs":[{"key":"bar","value":"baz"}],"pkg_path":"","func":""}]`

	assert.Equal(t, expected, string(res))
}

func TestEmit_Iteration(t *testing.T) {
	const testBar = "bar"
	m := gno.NewMachine("emit", nil)

	elgs := sdk.NewEventLogger()
	m.Context = ExecContext{EventLogger: elgs}

	iterEvent := func(m *gno.Machine) {
		for i := 0; i < 10; i++ {
			X_emit(m, testBar, []string{"qux", "value1"})
		}
	}
	iterEvent(m)
	assert.Equal(t, 10, len(elgs.Events()))

	res, err := json.Marshal(elgs.Events())
	if err != nil {
		t.Fatal(err)
	}

	var builder strings.Builder
	builder.WriteString("[")
	for i := 0; i < 10; i++ {
		builder.WriteString(`{"type":"bar","attrs":[{"key":"qux","value":"value1"}],"pkg_path":"","func":""},`)
	}
	expected := builder.String()[:builder.Len()-1] + "]"

	assert.Equal(t, expected, string(res))
}

func complexInteraction(m *gno.Machine) {
	deferEmitExample(m)
}

func deferEmitExample(m *gno.Machine) {
	defer func() {
		X_emit(m, "DeferEvent", []string{"key1", "value1", "key2", "value2"})
	}()

	forLoopEmitExample(m, 3, func(i int) {
		X_emit(m, "ForLoopEvent", []string{"iteration", strconv.Itoa(i), "key", "value"})
	})

	callbackEmitExample(m, func() {
		X_emit(m, "CallbackEvent", []string{"key1", "value1", "key2", "value2"})
	})
}

func forLoopEmitExample(m *gno.Machine, count int, callback func(int)) {
	defer func() {
		X_emit(m, "ForLoopCompletionEvent", []string{"count", strconv.Itoa(count)})
	}()

	for i := 0; i < count; i++ {
		callback(i)
	}
}

func callbackEmitExample(m *gno.Machine, callback func()) {
	defer func() {
		X_emit(m, "CallbackCompletionEvent", []string{"key", "value"})
	}()

	callback()
}

func TestEmit_ComplexInteraction(t *testing.T) {
	m := gno.NewMachine("emit", nil)

	elgs := sdk.NewEventLogger()
	m.Context = ExecContext{EventLogger: elgs}

	complexInteraction(m)

	assert.Equal(t, 7, len(elgs.Events()))

	res, err := json.Marshal(elgs.Events())
	if err != nil {
		t.Fatal(err)
	}

	expected := `[{"type":"ForLoopEvent","attrs":[{"key":"iteration","value":"0"},{"key":"key","value":"value"}],"pkg_path":"","func":""},{"type":"ForLoopEvent","attrs":[{"key":"iteration","value":"1"},{"key":"key","value":"value"}],"pkg_path":"","func":""},{"type":"ForLoopEvent","attrs":[{"key":"iteration","value":"2"},{"key":"key","value":"value"}],"pkg_path":"","func":""},{"type":"ForLoopCompletionEvent","attrs":[{"key":"count","value":"3"}],"pkg_path":"","func":""},{"type":"CallbackEvent","attrs":[{"key":"key1","value":"value1"},{"key":"key2","value":"value2"}],"pkg_path":"","func":""},{"type":"CallbackCompletionEvent","attrs":[{"key":"key","value":"value"}],"pkg_path":"","func":""},{"type":"DeferEvent","attrs":[{"key":"key1","value":"value1"},{"key":"key2","value":"value2"}],"pkg_path":"","func":""}]`
	assert.Equal(t, expected, string(res))
}
