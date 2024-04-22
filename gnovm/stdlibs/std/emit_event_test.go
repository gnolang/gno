package std

import (
	"encoding/json"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestEmit_SimpleValid(t *testing.T) {
	m := gno.NewMachine("emit", nil)

	elgs := sdk.NewEventLogger()
	m.Context = ExecContext{EventLogger: elgs}

	attrs := []string{"key1", "value1", "key2", "value2"}
	X_emit(m, "test", attrs)

	assert.Equal(t, 1, len(elgs.Events()))

	res, err := json.Marshal(elgs.Events())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, `["{\"type\":\"test\",\"pkg_path\":\"\",\"identifier\":\"\",\"timestamp\":0,\"attributes\":[{\"key\":\"key1\",\"value\":\"value1\"},{\"key\":\"key2\",\"value\":\"value2\"}]}"]`, string(res))
}

func TestEmit_OddNumberAttrs(t *testing.T) {
	m := gno.NewMachine("emit", nil)

	elgs := sdk.NewEventLogger()
	m.Context = ExecContext{EventLogger: elgs}

	attrs := []string{"key1", "value1", "key2"}

	assert.Panics(t, func() {
		X_emit(m, "test", attrs)
	})
}

func TestNewGnoEventString(t *testing.T) {
	eventType := "test"
	pkgPath := "p/demo/foo"
	ident := "Receiver"
	timestamp := int64(0)
	attrs := []gnoEventAttribute{
		{
			Key:   "key1",
			Value: "value1",
		},
		{
			Key:   "key2",
			Value: "value2",
		},
	}

	expectedEvent := gnoEvent{
		Type:       eventType,
		PkgPath:    pkgPath,
		Identifier: ident,
		Timestamp:  timestamp,
		Attributes: attrs,
	}

	expected, err := json.Marshal(expectedEvent)
	if err != nil {
		t.Fatal(err)
	}

	result := NewGnoEventString(eventType, pkgPath, ident, timestamp, attrs...)
	assert.Equal(t, abci.EventString(expected), result)
}

const (
	sender   = "sender"
	receiver = "receiver"
)

type contractA struct{}

func (c *contractA) sender(m *gno.Machine, cb func()) {
	subSender(m)
	cb()
}

func subSender(m *gno.Machine) {
	X_emit(m, sender, []string{"k1", "v1", "k2", "v2"})
}

type contractB struct{}

func (c *contractB) subReceiver(m *gno.Machine) {
	X_emit(m, receiver, []string{"bar", "baz"})
}

func TestEmit_ContractInteration(t *testing.T) {
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

	assert.Equal(t, `["{\"type\":\"sender\",\"pkg_path\":\"\",\"identifier\":\"\",\"timestamp\":0,\"attributes\":[{\"key\":\"k1\",\"value\":\"v1\"},{\"key\":\"k2\",\"value\":\"v2\"}]}","{\"type\":\"receiver\",\"pkg_path\":\"\",\"identifier\":\"\",\"timestamp\":0,\"attributes\":[{\"key\":\"bar\",\"value\":\"baz\"}]}"]`, string(res))
}
