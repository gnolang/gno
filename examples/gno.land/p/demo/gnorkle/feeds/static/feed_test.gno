package static_test

import (
	"errors"
	"testing"

	"gno.land/p/demo/gnorkle/feed"
	"gno.land/p/demo/gnorkle/feeds/static"
	"gno.land/p/demo/gnorkle/gnorkle"
	"gno.land/p/demo/gnorkle/ingester"
	"gno.land/p/demo/gnorkle/message"
	"gno.land/p/demo/gnorkle/storage/simple"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

type mockIngester struct {
	canAutoCommit   bool
	ingestErr       error
	commitErr       error
	value           string
	providerAddress string
}

func (i mockIngester) Type() ingester.Type {
	return ingester.Type(0)
}

func (i *mockIngester) Ingest(value, providerAddress string) (bool, error) {
	if i.ingestErr != nil {
		return false, i.ingestErr
	}

	i.value = value
	i.providerAddress = providerAddress
	return i.canAutoCommit, nil
}

func (i *mockIngester) CommitValue(storage gnorkle.Storage, providerAddress string) error {
	if i.commitErr != nil {
		return i.commitErr
	}

	return storage.Put(i.value)
}

func TestNewSingleValueFeed(t *testing.T) {
	staticFeed := static.NewSingleValueFeed("1", "")

	uassert.Equal(t, "1", staticFeed.ID())
	uassert.Equal(t, int(feed.TypeStatic), int(staticFeed.Type()))
}

func TestFeed_Ingest(t *testing.T) {
	var undefinedFeed *static.Feed
	err := undefinedFeed.Ingest("", "", "")
	uassert.ErrorIs(t, err, feed.ErrUndefined)

	tests := []struct {
		name               string
		ingester           *mockIngester
		verifyIsLocked     bool
		doCommit           bool
		funcType           message.FuncType
		msg                string
		providerAddress    string
		expFeedValueString string
		expErrText         string
		expIsActive        bool
	}{
		{
			name:        "func invalid error",
			ingester:    &mockIngester{},
			funcType:    message.FuncType("derp"),
			expErrText:  "invalid message function derp",
			expIsActive: true,
		},
		{
			name: "func ingest ingest error",
			ingester: &mockIngester{
				ingestErr: errors.New("ingest error"),
			},
			funcType:    message.FuncTypeIngest,
			expErrText:  "ingest error",
			expIsActive: true,
		},
		{
			name: "func ingest commit error",
			ingester: &mockIngester{
				commitErr:     errors.New("commit error"),
				canAutoCommit: true,
			},
			funcType:    message.FuncTypeIngest,
			expErrText:  "commit error",
			expIsActive: true,
		},
		{
			name: "func commit commit error",
			ingester: &mockIngester{
				commitErr:     errors.New("commit error"),
				canAutoCommit: true,
			},
			funcType:    message.FuncTypeCommit,
			expErrText:  "commit error",
			expIsActive: true,
		},
		{
			name:            "only ingest",
			ingester:        &mockIngester{},
			funcType:        message.FuncTypeIngest,
			msg:             "still active feed",
			providerAddress: "gno1234",
			expIsActive:     true,
		},
		{
			name:               "ingest autocommit",
			ingester:           &mockIngester{canAutoCommit: true},
			funcType:           message.FuncTypeIngest,
			msg:                "still active feed",
			providerAddress:    "gno1234",
			expFeedValueString: "still active feed",
			verifyIsLocked:     true,
		},
		{
			name:           "commit no value",
			ingester:       &mockIngester{},
			funcType:       message.FuncTypeCommit,
			msg:            "shouldn't be stored",
			verifyIsLocked: true,
		},
		{
			name:               "ingest then commmit",
			ingester:           &mockIngester{},
			funcType:           message.FuncTypeIngest,
			msg:                "blahblah",
			doCommit:           true,
			expFeedValueString: "blahblah",
			verifyIsLocked:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			staticFeed := static.NewFeed(
				"1",
				"string",
				tt.ingester,
				simple.NewStorage(1),
				nil,
			)

			var errText string
			if err := staticFeed.Ingest(tt.funcType, tt.msg, tt.providerAddress); err != nil {
				errText = err.Error()
			}

			urequire.Equal(t, tt.expErrText, errText)

			if tt.doCommit {
				err := staticFeed.Ingest(message.FuncTypeCommit, "", "")
				urequire.NoError(t, err, "follow up commit failed")
			}

			if tt.verifyIsLocked {
				errText = ""
				if err := staticFeed.Ingest(tt.funcType, tt.msg, tt.providerAddress); err != nil {
					errText = err.Error()
				}

				urequire.Equal(t, "feed locked", errText)
			}

			uassert.Equal(t, tt.providerAddress, tt.ingester.providerAddress)

			feedValue, dataType, isLocked := staticFeed.Value()
			uassert.Equal(t, tt.expFeedValueString, feedValue.String)
			uassert.Equal(t, "string", dataType)
			uassert.Equal(t, tt.verifyIsLocked, isLocked)
			uassert.Equal(t, tt.expIsActive, staticFeed.IsActive())
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
			expJSON: `{"id":"99","type":"0","value_type":"int","tasks":[]}`,
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
			expJSON: `{"id":"99","type":"0","value_type":"int","tasks":[{"value":"single"}]}`,
		},
		{
			name: "two tasks",
			tasks: []feed.Task{
				mockTask{value: "first"},
				mockTask{value: "second"},
			},
			expJSON: `{"id":"99","type":"0","value_type":"int","tasks":[{"value":"first"},{"value":"second"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			staticFeed := static.NewSingleValueFeed(
				id,
				valueDataType,
				tt.tasks...,
			)

			urequire.Equal(t, len(tt.tasks), len(staticFeed.Tasks()))

			var errText string
			json, err := staticFeed.MarshalJSON()
			if err != nil {
				errText = err.Error()
			}

			urequire.Equal(t, tt.expErrText, errText)
			urequire.Equal(t, tt.expJSON, string(json))
		})
	}
}
