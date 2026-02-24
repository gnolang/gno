package hor

import (
	"testing"

	"gno.land/p/jeronimoalbi/datasource"
	"gno.land/p/moul/addrset"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
	uinit "gno.land/r/sys/users/init"
)

func init() {
	// Pre-register "demo" user for render tests
	uinit.RegisterUser(cross, "demo", address("g13ek2zz9qurzynzvssyc4sthwppnruhnp0gdz8n"))
}

var (
	_ datasource.Datasource    = (*Datasource)(nil)
	_ datasource.Record        = (*record)(nil)
	_ datasource.ContentRecord = (*record)(nil)
	_ datasource.Iterator      = (*iterator)(nil)
)

func TestDatasourceRecords(t *testing.T) {
	cases := []struct {
		name      string
		items     []*Item
		recordIDs []string
		options   []datasource.QueryOption
	}{
		{
			name:      "all items",
			items:     []*Item{{id: 1}, {id: 2}, {id: 3}},
			recordIDs: []string{"0000001", "0000002", "0000003"},
		},
		{
			name:      "with offset",
			items:     []*Item{{id: 1}, {id: 2}, {id: 3}},
			recordIDs: []string{"0000002", "0000003"},
			options:   []datasource.QueryOption{datasource.WithOffset(1)},
		},
		{
			name:      "with count",
			items:     []*Item{{id: 1}, {id: 2}, {id: 3}},
			recordIDs: []string{"0000001", "0000002"},
			options:   []datasource.QueryOption{datasource.WithCount(2)},
		},
		{
			name:      "with offset and count",
			items:     []*Item{{id: 1}, {id: 2}, {id: 3}},
			recordIDs: []string{"0000002"},
			options: []datasource.QueryOption{
				datasource.WithOffset(1),
				datasource.WithCount(1),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize a local instance of exhibition
			exhibition := &Exhibition{items: avl.NewTree()}
			for _, item := range tc.items {
				exhibition.items.Set(item.id.String(), item)
			}

			// Get a records iterator
			ds := Datasource{exhibition}
			query := datasource.NewQuery(tc.options...)
			iter := ds.Records(query)

			// Start asserting
			urequire.Equal(t, len(tc.items), ds.Size(), "datasource size")

			var records []datasource.Record
			for iter.Next() {
				records = append(records, iter.Record())
			}
			urequire.Equal(t, len(tc.recordIDs), len(records), "record count")

			for i, r := range records {
				uassert.Equal(t, tc.recordIDs[i], r.ID())
			}
		})
	}
}

func TestDatasourceRecord(t *testing.T) {
	cases := []struct {
		name  string
		items []*Item
		id    string
		err   string
	}{
		{
			name:  "found",
			items: []*Item{{id: 1}, {id: 2}, {id: 3}},
			id:    "0000001",
		},
		{
			name:  "no found",
			items: []*Item{{id: 1}, {id: 2}, {id: 3}},
			id:    "42",
			err:   "realm submission not found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize a local instance of exhibition
			exhibition := &Exhibition{items: avl.NewTree()}
			for _, item := range tc.items {
				exhibition.items.Set(item.id.String(), item)
			}

			// Get a single record
			ds := Datasource{exhibition}
			r, err := ds.Record(tc.id)

			// Start asserting
			if tc.err != "" {
				uassert.ErrorContains(t, err, tc.err)
				return
			}

			urequire.NoError(t, err, "no error")
			urequire.NotEqual(t, nil, r, "record not nil")
			uassert.Equal(t, tc.id, r.ID())
		})
	}
}

func TestItemRecord(t *testing.T) {
	pkgpath := "gno.land/r/demo/test"
	item := Item{
		id:          1,
		pkgpath:     pkgpath,
		title:       "Test Realm",
		description: "This is a test realm in the Hall of Fame",
		blockNum:    42,
		upvote:      &addrset.Set{},
		downvote:    &addrset.Set{},
	}
	item.downvote.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	item.upvote.Add("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
	item.upvote.Add("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")

	r := record{&item}

	uassert.Equal(t, "0000001", r.ID())
	uassert.Equal(t, pkgpath, r.String())

	fields, _ := r.Fields()
	details, found := fields.Get("details")
	urequire.True(t, found, "details field")
	uassert.Equal(t, "Votes: ⏶ 2 - ⏷ 1", details)

	content, _ := r.Content()
	wantContent := `### [Test Realm](/r/demo/test)
This is a test realm in the Hall of Fame

by [@demo](/u/demo)

Submitted at Block #42

**[2👍](/r/leon/hor$help&func=Upvote&pkgpath=gno.land%2Fr%2Fdemo%2Ftest) - [1👎](/r/leon/hor$help&func=Downvote&pkgpath=gno.land%2Fr%2Fdemo%2Ftest)**`
	uassert.Equal(t, wantContent, content)
}
