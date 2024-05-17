package gno

import (
	"context"
	"errors"
	"strconv"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gno.me/state"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

const maxEventsRequestable uint64 = 100

var errEndingSequence = errors.New("ending sequence is less than starting sequence")

type EventRequest struct {
	StartingSequence string `json:"start"`
	EndingSequence   string `json:"end"`
	AppName          string `json:"app_name"`
}

func (r EventRequest) SequenceRange() (uint64, uint64, error) {
	start, err := strconv.ParseUint(r.StartingSequence, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	end, err := strconv.ParseUint(r.EndingSequence, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	if end < start && end != 0 {
		return 0, 0, errEndingSequence
	}

	if maxEnd := start + maxEventsRequestable - 1; maxEnd > end {
		end = maxEnd
	}

	return start, end, nil
}

func (v *VMKeeper) initEventStore() error {
	_, err := v.Create(context.Background(), eventStorageRealm, false, false)
	return err
}

const eventStorageRealm string = `
package events

import (
	"errors"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uintavl"
)

var store = avl.NewTree()

type event struct {
	sequence uint64
	funcName string
	args     string
}

// Sequences start from 1 in order to avoid confusion with zero values.
func NextSequence(pkgPath string) uint64 {
	eventTree, ok := store.Get(pkgPath)
	if !ok {
		return 1
	}

	return uint64(eventTree.(*uintavl.Tree).Size() + 1)
}

func Store(pkgPath string, sequence uint64, funcName, args string) uint64 {
	nextSequence := NextSequence(pkgPath)
	if sequence != nextSequence {
		panic("sequence out of order: expected " + strconv.FormatUint(nextSequence, 10) + " but got " + strconv.FormatUint(sequence, 10))
	}

	eventTree := uintavl.NewTree()
	createEventTree := true
	if tree, ok := store.Get(pkgPath); ok {
		eventTree = tree.(*uintavl.Tree)
		createEventTree = false
	}

	eventTree.Set(
		sequence,
		event{
			sequence: sequence,
			funcName: funcName,
			args:     args,
		},
	)

	if createEventTree {
		store.Set(pkgPath, eventTree)
	}

	return sequence
}

func Get(pkgPath string, start, end uint64) string {
	tree, ok := store.Get(pkgPath)
	if !ok {
		return ""
	}

	eventTree := tree.(*uintavl.Tree)
	if size := uint64(eventTree.Size()); size > end {
		end = size
	}

	if start > end || end == 0 {
		return ""
	}

	var sb strings.Builder
	pathParts := strings.Split(pkgPath, "/")
	appName := pathParts[len(pathParts)-1]

	sb.WriteString("[")
	first := true
	eventTree.Iterate(start, end+1, func(key uint64, value interface{}) bool {
		ev := value.(event)
		if !first {
			sb.WriteString(",")
			first = false
		}

		sb.WriteString("{\"sequence\":" + strconv.FormatUint(ev.sequence, 10) + ",\"app_name\":\"" + appName + "\",\"func\":\"" + ev.funcName + "\",\"args\":\"" + ev.args + "\"}")
		return false
	})

	sb.WriteString("]")
	return sb.String()
}
`

// ApplyEvent does two things: runs the event all and updates the event store.
func (v *VMKeeper) ApplyEvent(ctx context.Context, event *state.Event) error {
	v.Lock()
	defer v.Unlock()
	defer v.store.Commit()

	pkgPath := AppPrefix + event.AppName
	msg := vm.MsgCall{
		PkgPath: AppPrefix + "events",
		Func:    "Store",
		Args: []string{
			pkgPath,
			event.Sequence,
			event.Func,
			encodeArgs(event.Args),
		},
	}

	// TODO: parse error and return a special "out of order" error that is a struct
	// that specifies what the next messages should be.
	_, err := v.instance.Call(sdk.Context{}.WithContext(ctx), msg)
	if err != nil {
		return err
	}

	// The event was persisted using the given sequence number -- good!
	// Now run the event call to update the application state.
	msg = vm.MsgCall{
		PkgPath: pkgPath,
		Func:    event.Func,
		Args:    event.Args,
	}

	if _, err := v.instance.Call(sdk.Context{}.WithContext(ctx), msg); err != nil {
		return err
	}

	// TODO: rollback the event store if the event call fails.

	return nil
}
