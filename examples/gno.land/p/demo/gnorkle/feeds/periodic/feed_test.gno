package periodic_test

import (
	"errors"
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/gnorkle/feed"
	"gno.land/p/demo/gnorkle/feeds/periodic"
	"gno.land/p/demo/gnorkle/gnorkle"
	"gno.land/p/demo/gnorkle/ingester"
	"gno.land/p/demo/gnorkle/message"
	"gno.land/p/demo/gnorkle/storage/simple"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

type mockIngester struct {
	canAutoCommit   bool
	ingestErr       error
	commitErr       error
	values          []string
	providerAddress string
}

func (i mockIngester) Type() ingester.Type {
	return ingester.Type(0)
}

func (i *mockIngester) Ingest(value, providerAddress string) (bool, error) {
	if i.ingestErr != nil {
		return false, i.ingestErr
	}

	i.values = append(i.values, value)
	i.providerAddress = providerAddress
	return i.canAutoCommit, nil
}

func (i *mockIngester) CommitValue(storage gnorkle.Storage, providerAddress string) error {
	if i.commitErr != nil {
		return i.commitErr
	}

	return storage.Put(strings.Join(i.values, ","))
}

func TestNewMultiValueFeed(t *testing.T) {
	periodicFeed := periodic.NewMultiValueFeed("1", "", 100)

	uassert.Equal(t, "1", periodicFeed.ID())
	uassert.Equal(t, int(feed.TypePeriodic), int(periodicFeed.Type()))
}

func TestFeed_Ingest(t *testing.T) {
	var undefinedFeed *periodic.Feed
	err := undefinedFeed.Ingest("", "", "")
	uassert.ErrorIs(t, err, feed.ErrUndefined)

	tests := []struct {
		name               string
		ingester           *mockIngester
		verifyIsLocked     bool
		doCommit           bool
		timeAfterIngest    int64
		funcType           message.FuncType
		msg                string
		providerAddress    string
		expFeedValueString string
		expErrText         string
		expCommitErrText   string
		firstCommit        bool
	}{
		{
			name:           "func invalid error",
			ingester:       &mockIngester{},
			funcType:       message.FuncType("derp"),
			expErrText:     "invalid message function derp",
			verifyIsLocked: true,
		},
		{
			name:            "only ingest",
			ingester:        &mockIngester{},
			funcType:        message.FuncTypeIngest,
			msg:             "still active feed",
			providerAddress: "gno1234",
			verifyIsLocked:  true,
		},
		{
			name:               "ingest and commit first value",
			ingester:           &mockIngester{},
			funcType:           message.FuncTypeIngest,
			msg:                "blahblah",
			providerAddress:    "gno1234",
			verifyIsLocked:     false,
			expFeedValueString: "blahblah",
			doCommit:           true,
		},
		{
			name:               "ingest and commit with enough time passed",
			ingester:           &mockIngester{},
			funcType:           message.FuncTypeIngest,
			msg:                "blahblah",
			providerAddress:    "gno1234",
			verifyIsLocked:     false,
			expFeedValueString: "blahblah",
			timeAfterIngest:    101,
			doCommit:           true,
			firstCommit:        true,
		},
		{
			name:             "ingest and commit without time passed",
			ingester:         &mockIngester{},
			funcType:         message.FuncTypeIngest,
			msg:              "still active feed",
			providerAddress:  "gno1234",
			verifyIsLocked:   false,
			doCommit:         true,
			expCommitErrText: "time between commit not reached",
			firstCommit:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			periodicFeed := periodic.NewFeed(
				"1",
				"string",
				100,
				tt.ingester,
				simple.NewStorage(1),
				nil,
			)

			if tt.firstCommit {
				periodicFeed.Ingest(message.FuncTypeCommit, "", "")
			}

			var errText string
			if err := periodicFeed.Ingest(tt.funcType, tt.msg, tt.providerAddress); err != nil {
				errText = err.Error()
			}
			urequire.Equal(t, tt.expErrText, errText)

			if tt.timeAfterIngest > 0 {
				std.TestSkipHeights(tt.timeAfterIngest)
			}

			if tt.doCommit {
				var errText string
				if err := periodicFeed.Ingest(message.FuncTypeCommit, "", ""); err != nil {
					errText = err.Error()
				}
				urequire.Equal(t, errText, tt.expCommitErrText)
			}

			uassert.Equal(t, tt.providerAddress, tt.ingester.providerAddress)

			feedValue, dataType, isLocked := periodicFeed.Value()
			uassert.Equal(t, tt.expFeedValueString, feedValue.String)
			uassert.Equal(t, "string", dataType)
			uassert.Equal(t, tt.verifyIsLocked, isLocked)
			uassert.Equal(t, true, periodicFeed.IsActive())
		})
	}
}

type mockTask struct {
	err   error
	value string
}

func (t mockTask) MarshalJSON() ([]byte, error) {
	if t.err != nil {
		return nil, t.err
	}

	return []byte(`{"value":"` + t.value + `"}`), nil
}

func TestFeed_Tasks(t *testing.T) {
	id := "99"
	valueDataType := "int"

	tests := []struct {
		name       string
		tasks      []feed.Task
		expErrText string
		expJSON    string
	}{
		{
			name:    "no tasks",
			expJSON: `{"id":"99","type":"2","value_type":"int","tasks":[]}`,
		},
		{
			name: "marshal error",
			tasks: []feed.Task{
				mockTask{err: errors.New("marshal error")},
			},
			expErrText: "marshal error",
		},
		{
			name: "one task",
			tasks: []feed.Task{
				mockTask{value: "single"},
			},
			expJSON: `{"id":"99","type":"2","value_type":"int","tasks":[{"value":"single"}]}`,
		},
		{
			name: "two tasks",
			tasks: []feed.Task{
				mockTask{value: "first"},
				mockTask{value: "second"},
			},
			expJSON: `{"id":"99","type":"2","value_type":"int","tasks":[{"value":"first"},{"value":"second"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			periodicFeed := periodic.NewMultiValueFeed(
				id,
				valueDataType,
				100,
				tt.tasks...,
			)

			urequire.Equal(t, len(tt.tasks), len(periodicFeed.Tasks()))

			var errText string
			json, err := gnorkle.MarshalFeedJSON(periodicFeed)
			if err != nil {
				errText = err.Error()
			}

			urequire.Equal(t, tt.expErrText, errText)
			urequire.Equal(t, tt.expJSON, string(json))
		})
	}
}
