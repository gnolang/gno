package hof

import (
	"strings"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
	"gno.land/p/jeronimoalbi/datasource"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
)

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
			exhibition := &Exhibition{itemsSorted: avl.NewTree()}
			for _, item := range tc.items {
				exhibition.itemsSorted.Set(item.id.String(), item)
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
			exhibition := &Exhibition{itemsSorted: avl.NewTree()}
			for _, item := range tc.items {
				exhibition.itemsSorted.Set(item.id.String(), item)
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
		upvote:      avl.NewTree(),
		downvote:    avl.NewTree(),
	}
	item.downvote.Set("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", struct{}{})
	item.upvote.Set("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", struct{}{})
	item.upvote.Set("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", struct{}{})

	r := record{&item}

	uassert.Equal(t, "0000001", r.ID())
	uassert.Equal(t, pkgpath, r.String())

	fields, _ := r.Fields()
	details, found := fields.Get("details")
	urequire.True(t, found, "details field")
	uassert.Equal(t, "Votes: ⏶ 2 - ⏷ 1", details)

	content, _ := r.Content()
	wantContent := md.H1(item.title) +
		md.H2(r.item.description) +
		ufmt.Sprintf("\n%s\n\n", md.CodeBlock(item.pkgpath)) +
		ufmt.Sprintf("%s\n\n", item.description) +
		ufmt.Sprintf("by %s\n\n", strings.Split(item.pkgpath, "/")[2]) +
		md.Link("View Realm", strings.TrimPrefix(item.pkgpath, "gno.land")) + "\n\n" +
		ufmt.Sprintf("Submitted at Block #%d\n\n", item.blockNum) +
		md.Bold(ufmt.Sprintf("[%d👍](%s) - [%d👎](%s)",
			item.upvote.Size(), txlink.Call("Upvote", "pkgpath", item.pkgpath),
			item.downvote.Size(), txlink.Call("Downvote", "pkgpath", item.pkgpath),
		))
	uassert.Equal(t, wantContent, content)
}
