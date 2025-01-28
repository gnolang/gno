package single_test

import (
	"testing"

	"gno.land/p/demo/gnorkle/ingester"
	"gno.land/p/demo/gnorkle/ingesters/single"
	"gno.land/p/demo/gnorkle/storage/simple"
	"gno.land/p/nt/uassert"
)

func TestValueIngester(t *testing.T) {
	storage := simple.NewStorage(1)

	var undefinedIngester *single.ValueIngester
	_, err := undefinedIngester.Ingest("asdf", "gno11111")
	uassert.ErrorIs(t, err, ingester.ErrUndefined, "undefined ingester call to Ingest should return ingester.ErrUndefined")

	err = undefinedIngester.CommitValue(storage, "gno11111")
	uassert.ErrorIs(t, err, ingester.ErrUndefined, "undefined ingester call to CommitValue should return ingester.ErrUndefined")

	var valueIngester single.ValueIngester
	typ := valueIngester.Type()
	uassert.Equal(t, int(ingester.TypeSingle), int(typ), "single value ingester should return type ingester.TypeSingle")

	ingestValue := "value"
	autocommit, err := valueIngester.Ingest(ingestValue, "gno11111")
	uassert.True(t, autocommit, "single value ingester should return autocommit true")
	uassert.NoError(t, err)

	err = valueIngester.CommitValue(storage, "gno11111")
	uassert.NoError(t, err)

	latestValue := storage.GetLatest()
	uassert.Equal(t, ingestValue, latestValue.String)
}
