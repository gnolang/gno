package multi_test

import (
	"strings"
	"testing"

	"gno.land/p/demo/gnorkle/ingester"
	"gno.land/p/demo/gnorkle/ingesters/multi"
	"gno.land/p/demo/gnorkle/storage/simple"
	"gno.land/p/demo/uassert"
)

func TestValueIngester(t *testing.T) {
	storage := simple.NewStorage(1)

	var undefinedIngester *multi.ValueIngester
	_, err := undefinedIngester.Ingest("asdf", "gno11111")
	uassert.ErrorIs(t, err, ingester.ErrUndefined, "undefined ingester call to Ingest should return ingester.ErrUndefined")

	err = undefinedIngester.CommitValue(storage, "gno11111")
	uassert.ErrorIs(t, err, ingester.ErrUndefined, "undefined ingester call to CommitValue should return ingester.ErrUndefined")

	var valueIngester multi.ValueIngester
	typ := valueIngester.Type()
	uassert.Equal(t, int(ingester.TypeMulti), int(typ), "multi value ingester should return type ingester.TypMulti")

	ingestValue := "value"
	autocommit, err := valueIngester.Ingest(ingestValue, "gno11111")
	uassert.False(t, autocommit, "multi value ingester should return autocommit false")
	uassert.NoError(t, err)

	ingestValue2 := "value2"
	autocommit2, err := valueIngester.Ingest(ingestValue2, "gno11112")
	uassert.False(t, autocommit2, "multi value ingester should return autocommit false")
	uassert.NoError(t, err)

	err = valueIngester.CommitValue(storage, "gno11111")
	uassert.NoError(t, err)

	latestValue := storage.GetLatest()
	uassert.Equal(t, ingestValue.concat(",", ingestvalue2), latestValue.String)
}
