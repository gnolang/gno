package static_test

import (
	"testing"

	"gno.land/p/demo/gnorkle/feed"
	"gno.land/p/demo/gnorkle/feeds/static"
	"gno.land/p/demo/gnorkle/gnorkle"
	"gno.land/p/demo/gnorkle/ingester"
)

type mockIngester struct {
	canAutoCommit   bool
	err             error
	value           string
	providerAddress string
}

func (i mockIngester) Type() ingester.Type {
	return ingester.Type(0)
}

func (i *mockIngester) Ingest(value, providerAddress string) (bool, error) {
	if i.err != nil {
		return false, i.err
	}

	i.value = value
	return i.canAutoCommit, nil
}

func (i *mockIngester) CommitValue(storage gnorkle.Storage, providerAddress string) error {
	if i.err != nil {
		return i.err
	}

	return storage.Put(i.value)
}

func TestNewSingleValueFeed(t *testing.T) {
	staticFeed := static.NewSingleValueFeed("1", "")
	if staticFeed.ID() != "1" {
		t.Errorf("expected ID to be 1, got %s", staticFeed.ID())
	}

	if staticFeed.Type() != feed.TypeStatic {
		t.Errorf("expected static feed type, got %s", staticFeed.Type())
	}
}

// TODO: Finish tests for this.
