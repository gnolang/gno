package datasource

import (
	"errors"
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

func TestNewIterator(t *testing.T) {
	cases := []struct {
		name    string
		records []Record
		err     error
	}{
		{
			name: "ok",
			records: []Record{
				testRecord{id: "1"},
				testRecord{id: "2"},
				testRecord{id: "3"},
			},
		},
		{
			name: "error",
			err:  errors.New("test"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ds := testDatasource{
				records: tc.records,
				err:     tc.err,
			}

			// Act
			iter := NewIterator(ds)

			// Assert
			if tc.err != nil {
				uassert.ErrorIs(t, tc.err, iter.Err())
				return
			}

			uassert.NoError(t, iter.Err())

			for i := 0; iter.Next(); i++ {
				r := iter.Record()
				urequire.NotEqual(t, nil, r, "valid record")
				urequire.True(t, i < len(tc.records), "iteration count")
				uassert.Equal(t, tc.records[i].ID(), r.ID())
			}
		})
	}
}

func TestQueryRecords(t *testing.T) {
	cases := []struct {
		name        string
		records     []Record
		recordCount int
		options     []QueryOption
		err         error
	}{
		{
			name: "ok",
			records: []Record{
				testRecord{id: "1"},
				testRecord{id: "2"},
				testRecord{id: "3"},
			},
			recordCount: 3,
		},
		{
			name:    "with count",
			options: []QueryOption{WithCount(2)},
			records: []Record{
				testRecord{id: "1"},
				testRecord{id: "2"},
				testRecord{id: "3"},
			},
			recordCount: 2,
		},
		{
			name: "invalid record",
			records: []Record{
				testRecord{id: "1"},
				nil,
				testRecord{id: "3"},
			},
			err: ErrInvalidRecord,
		},
		{
			name: "iterator error",
			records: []Record{
				testRecord{id: "1"},
				testRecord{id: "3"},
			},
			err: errors.New("test"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ds := testDatasource{
				records: tc.records,
				err:     tc.err,
			}

			// Act
			records, err := QueryRecords(ds, tc.options...)

			// Assert
			if tc.err != nil {
				uassert.ErrorIs(t, tc.err, err)
				return
			}

			uassert.NoError(t, err)

			urequire.Equal(t, tc.recordCount, len(records), "record count")
			for i, r := range records {
				urequire.NotEqual(t, nil, r, "valid record")
				uassert.Equal(t, tc.records[i].ID(), r.ID())
			}
		})
	}
}

type testDatasource struct {
	records []Record
	err     error
}

func (testDatasource) Size() int                     { return -1 }
func (testDatasource) Record(string) (Record, error) { return nil, nil }
func (ds testDatasource) Records(Query) Iterator     { return &testIter{records: ds.records, err: ds.err} }

type testRecord struct {
	id     string
	fields Fields
	err    error
}

func (r testRecord) ID() string              { return r.id }
func (r testRecord) String() string          { return "str" + r.id }
func (r testRecord) Fields() (Fields, error) { return r.fields, r.err }

type testIter struct {
	index   int
	records []Record
	current Record
	err     error
}

func (it testIter) Err() error     { return it.err }
func (it testIter) Record() Record { return it.current }

func (it *testIter) Next() bool {
	count := len(it.records)
	if it.err != nil || count == 0 || it.index >= count {
		return false
	}
	it.current = it.records[it.index]
	it.index++
	return true
}
