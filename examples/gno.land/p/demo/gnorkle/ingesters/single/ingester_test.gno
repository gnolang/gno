package single_test

import (
	"testing"

	"gno.land/p/demo/gnorkle/ingester"
	"gno.land/p/demo/gnorkle/ingesters/single"
	"gno.land/p/demo/gnorkle/storage/simple"
)

func TestValueIngester(t *testing.T) {
	storage := simple.NewStorage(1)

	var undefinedIngester *single.ValueIngester
	if _, err := undefinedIngester.Ingest("asdf", "gno11111"); err != ingester.ErrUndefined {
		t.Error("undefined ingester call to Ingest should return ingester.ErrUndefined")
	}
	if err := undefinedIngester.CommitValue(storage, "gno11111"); err != ingester.ErrUndefined {
		t.Error("undefined ingester call to CommitValue should return ingester.ErrUndefined")
	}

	var valueIngester single.ValueIngester
	if typ := valueIngester.Type(); typ != ingester.TypeSingle {
		t.Error("single value ingester should return type ingester.TypeSingle")
	}

	ingestValue := "value"
	autocommit, err := valueIngester.Ingest(ingestValue, "gno11111")
	if !autocommit {
		t.Error("single value ingester should return autocommit true")
	}
	if err != nil {
		t.Errorf("unexpected ingest error %s", err.Error())
	}

	if err := valueIngester.CommitValue(storage, "gno11111"); err != nil {
		t.Errorf("unexpected commit error %s", err.Error())
	}

	if latestValue := storage.GetLatest(); latestValue.String != ingestValue {
		t.Errorf("expected latest value of %s, got %s", ingestValue, latestValue.String)
	}
}
