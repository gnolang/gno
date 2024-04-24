package std

import (
	"encoding/json"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestEmit(t *testing.T) {
	m := gno.NewMachine("emit", nil)
	pkgPath := CurrentRealmPath(m)
	tests := []struct {
		name           string
		eventType      string
		attrs          []string
		expectedEvents []gnoEvent
		expectPanic    bool
	}{
		{
			name:      "SimpleValid",
			eventType: "test",
			attrs:     []string{"key1", "value1", "key2", "value2"},
			expectedEvents: []gnoEvent{
				{
					Type:       "test",
					PkgPath:    pkgPath,
					Identifier: "",
					Attributes: []gnoEventAttribute{
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
			expectedEvents: []gnoEvent{
				{
					Type:       "test",
					PkgPath:    pkgPath,
					Identifier: "",
					Attributes: []gnoEventAttribute{
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
			expectedEvents: []gnoEvent{
				{
					Type:       "",
					PkgPath:    pkgPath,
					Identifier: "",
					Attributes: []gnoEventAttribute{
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
			expectedEvents: []gnoEvent{
				{
					Type:       "test",
					PkgPath:    pkgPath,
					Identifier: "",
					Attributes: []gnoEventAttribute{
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

	expect := []gnoEvent{
		{
			Type:       "test1",
			PkgPath:    "",
			Identifier: "",
			Attributes: []gnoEventAttribute{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
		{
			Type:       "test2",
			PkgPath:    "",
			Identifier: "",
			Attributes: []gnoEventAttribute{
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

const (
	testFoo = "foo"
	testBar = "bar"
)

type contractA struct{}

func (c *contractA) foo(m *gno.Machine, cb func()) {
	subSender(m)
	cb()
}

func subSender(m *gno.Machine) {
	X_emit(m, testFoo, []string{"k1", "v1", "k2", "v2"})
}

type contractB struct{}

func (c *contractB) subReceiver(m *gno.Machine) {
	X_emit(m, testBar, []string{"bar", "baz"})
}

func TestEmit_ContractInteration(t *testing.T) {
	t.Parallel()
	m := gno.NewMachine("emit", nil)
	elgs := sdk.NewEventLogger()
	m.Context = ExecContext{EventLogger: elgs}

	a := &contractA{}
	b := &contractB{}

	a.sender(m, func() {
		b.subReceiver(m)
	})

	assert.Equal(t, 2, len(elgs.Events()))

	res, err := json.Marshal(elgs.Events())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, `[{"type":"foo","pkg_path":"","identifier":"","attributes":[{"key":"k1","value":"v1"},{"key":"k2","value":"v2"}]},{"type":"bar","pkg_path":"","identifier":"","attributes":[{"key":"bar","value":"baz"}]}]`, string(res))
}
