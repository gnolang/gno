package awesomegno

import (
	"std"
	"testing"
	"time"

	"gno.land/p/jeronimoalbi/datasource"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
	"gno.land/p/nt/urequire"
)

var (
	_ datasource.Datasource    = (*ItemDatasource)(nil)
	_ datasource.Record        = (*itemRecord)(nil)
	_ datasource.ContentRecord = (*itemRecord)(nil)
	_ datasource.Iterator      = (*itemIterator)(nil)

	_ datasource.Datasource    = (*CategoryDatasource)(nil)
	_ datasource.Record        = (*categoryRecord)(nil)
	_ datasource.ContentRecord = (*categoryRecord)(nil)
	_ datasource.Iterator      = (*categoryIterator)(nil)
)

func TestItemDatasourceRecords(t *testing.T) {
	cases := []struct {
		name      string
		items     []*Item
		recordIDs []string
		options   []datasource.QueryOption
	}{
		{
			name:      "all items",
			items:     []*Item{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000001", "0000002", "0000003"},
		},
		{
			name:      "with offset",
			items:     []*Item{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000002", "0000003"},
			options:   []datasource.QueryOption{datasource.WithOffset(1)},
		},
		{
			name:      "with count",
			items:     []*Item{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000001", "0000002"},
			options:   []datasource.QueryOption{datasource.WithCount(2)},
		},
		{
			name:      "with offset and count",
			items:     []*Item{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000002"},
			options: []datasource.QueryOption{
				datasource.WithOffset(1),
				datasource.WithCount(1),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			items := avl.NewTree()
			for _, item := range tc.items {
				items.Set(item.ID.String(), item)
			}

			ds := ItemDatasource{items}
			query := datasource.NewQuery(tc.options...)
			iter := ds.Records(query)

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

func TestItemDatasourceRecord(t *testing.T) {
	cases := []struct {
		name  string
		items []*Item
		id    string
		err   string
	}{
		{
			name:  "found",
			items: []*Item{{ID: 1}, {ID: 2}, {ID: 3}},
			id:    "0000001",
		},
		{
			name:  "not found",
			items: []*Item{{ID: 1}, {ID: 2}, {ID: 3}},
			id:    "42",
			err:   ErrItemNotFound + ": 42",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			items := avl.NewTree()
			for _, item := range tc.items {
				items.Set(item.ID.String(), item)
			}

			ds := ItemDatasource{items}
			r, err := ds.Record(tc.id)

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
	addr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	catID := seqid.ID(1)

	cat := &Category{ID: catID, Name: "Test Category"}
	categories.Set(cat.ID.String(), cat)

	item := Item{
		ID:          1,
		Title:       "Test Item",
		Description: "This is a test item",
		Authors:     []std.Address{addr},
		URL:         "https://test.com",
		CreatedAt:   time.Now(),
		Categories:  []seqid.ID{catID},
	}

	r := itemRecord{&item}

	uassert.Equal(t, "0000001", r.ID())
	uassert.Equal(t, "Test Item", r.String())

	fields, err := r.Fields()
	urequire.NoError(t, err)

	_, found := fields.Get("authors")
	urequire.True(t, found, "authors field")

	url, found := fields.Get("url")
	urequire.True(t, found, "url field")
	uassert.Equal(t, "https://test.com", url)

	content, err := r.Content()
	urequire.NoError(t, err)

	wantContent := md.H1(item.Title) +
		md.H2("Description") +
		item.Description +
		md.H2("Categories") +
		md.BulletItem("Test Category")

	uassert.Equal(t, wantContent, content)
}

func TestCategoryDatasourceRecords(t *testing.T) {
	cases := []struct {
		name      string
		cats      []*Category
		recordIDs []string
		options   []datasource.QueryOption
	}{
		{
			name:      "all categories",
			cats:      []*Category{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000001", "0000002", "0000003"},
		},
		{
			name:      "with offset",
			cats:      []*Category{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000002", "0000003"},
			options:   []datasource.QueryOption{datasource.WithOffset(1)},
		},
		{
			name:      "with count",
			cats:      []*Category{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000001", "0000002"},
			options:   []datasource.QueryOption{datasource.WithCount(2)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			categories := avl.NewTree()
			for _, cat := range tc.cats {
				categories.Set(cat.ID.String(), cat)
			}

			ds := CategoryDatasource{categories}
			query := datasource.NewQuery(tc.options...)
			iter := ds.Records(query)

			urequire.Equal(t, len(tc.cats), ds.Size(), "datasource size")

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

func TestCategoryRecord(t *testing.T) {
	items := avl.NewTree()
	item := &Item{
		ID:    1,
		Title: "Test Item",
		URL:   "https://test.com",
	}
	items.Set(item.ID.String(), item)

	category := Category{
		ID:          1,
		Name:        "Test Category",
		Description: "This is a test category",
		CreatedAt:   time.Now(),
		items:       items,
	}

	r := categoryRecord{&category}

	uassert.Equal(t, "0000001", r.ID())
	uassert.Equal(t, "Test Category", r.String())

	fields, err := r.Fields()
	urequire.NoError(t, err)

	itemCount, found := fields.Get("item_count")
	urequire.True(t, found, "item_count field")
	uassert.Equal(t, "1 items", itemCount)

	content, err := r.Content()
	urequire.NoError(t, err)

	wantContent := md.H1(category.Name) +
		md.H2("Description") +
		category.Description +
		md.H2("Items in this category") +
		md.BulletItem(ufmt.Sprintf("[%s](%s)", item.Title, item.URL))

	uassert.Equal(t, wantContent, content)
}
