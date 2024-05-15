package gno

import (
	"context"
	"errors"
)

const maxEventsRequestable uint64 = 100

var errEndingSequence = errors.New("ending sequence is less than starting sequence")

type Event struct {
	MsgCall
	Sequence uint64 `json:"sequence"`
}

type EventRequest struct {
	StartingSequence uint64 `json:"start"`
	EndingSequence   uint64 `json:"end"`
	AppName          string `json:"app_name"`
}

func (r EventRequest) SequenceRange() (uint64, uint64, error) {
	if r.EndingSequence < r.StartingSequence && r.EndingSequence != 0 {
		return 0, 0, errEndingSequence
	}

	end := r.EndingSequence
	if maxEnd := r.StartingSequence + maxEventsRequestable - 1; maxEnd > r.EndingSequence {
		end = maxEnd
	}

	return r.StartingSequence, end, nil
}

func (v *VMKeeper) initEventStore() error {
	return v.Create(context.Background(), eventStorageRealm, false)
}

const eventStorageRealm string = `
package main

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

func Store(pkgPath string, sequence uint64, funcName, args string) (uint64, error) {
	nextSequence := NextSequence(pkgPath)
	if sequence != nextSequence {
		return 0, errors.New("expected sequence " + strconv.FormatUint(nextSequence, 10) + " but got " + strconv.FormatUint(sequence, 10))
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

	return sequence, nil
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
