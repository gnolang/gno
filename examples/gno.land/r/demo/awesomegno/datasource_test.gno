package awesomegno

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
	"gno.land/p/jeronimoalbi/datasource"
	"gno.land/p/moul/md"
)

var (
	_ datasource.Datasource    = (*DappDatasource)(nil)
	_ datasource.Record        = (*dappRecord)(nil)
	_ datasource.ContentRecord = (*dappRecord)(nil)
	_ datasource.Iterator      = (*dappIterator)(nil)

	_ datasource.Datasource    = (*CategoryDatasource)(nil)
	_ datasource.Record        = (*categoryRecord)(nil)
	_ datasource.ContentRecord = (*categoryRecord)(nil)
	_ datasource.Iterator      = (*categoryIterator)(nil)
)

func TestDappDatasourceRecords(t *testing.T) {
	cases := []struct {
		name      string
		dapps     []*Dapp
		recordIDs []string
		options   []datasource.QueryOption
	}{
		{
			name:      "all dapps",
			dapps:     []*Dapp{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000001", "0000002", "0000003"},
		},
		{
			name:      "with offset",
			dapps:     []*Dapp{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000002", "0000003"},
			options:   []datasource.QueryOption{datasource.WithOffset(1)},
		},
		{
			name:      "with count",
			dapps:     []*Dapp{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000001", "0000002"},
			options:   []datasource.QueryOption{datasource.WithCount(2)},
		},
		{
			name:      "with offset and count",
			dapps:     []*Dapp{{ID: 1}, {ID: 2}, {ID: 3}},
			recordIDs: []string{"0000002"},
			options: []datasource.QueryOption{
				datasource.WithOffset(1),
				datasource.WithCount(1),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dapps := avl.NewTree()
			for _, dapp := range tc.dapps {
				dapps.Set(dapp.ID.String(), dapp)
			}

			ds := DappDatasource{dapps}
			query := datasource.NewQuery(tc.options...)
			iter := ds.Records(query)

			urequire.Equal(t, len(tc.dapps), ds.Size(), "datasource size")

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

func TestDappDatasourceRecord(t *testing.T) {
	cases := []struct {
		name  string
		dapps []*Dapp
		id    string
		err   string
	}{
		{
			name:  "found",
			dapps: []*Dapp{{ID: 1}, {ID: 2}, {ID: 3}},
			id:    "0000001",
		},
		{
			name:  "not found",
			dapps: []*Dapp{{ID: 1}, {ID: 2}, {ID: 3}},
			id:    "42",
			err:   "dapp not found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dapps := avl.NewTree()
			for _, dapp := range tc.dapps {
				dapps.Set(dapp.ID.String(), dapp)
			}

			ds := DappDatasource{dapps}
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

func TestDappRecord(t *testing.T) {
	addr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	categories := avl.NewTree()
	cat := &Category{ID: 1, Name: "Test Category"}
	categories.Set(cat.ID.String(), cat)

	dapp := Dapp{
		ID:          1,
		Title:       "Test Dapp",
		Description: "This is a test dapp",
		Author:      addr,
		URL:         "https://test.com",
		CreatedAt:   time.Now(),
		categories:  categories,
	}

	r := dappRecord{&dapp}

	uassert.Equal(t, "0000001", r.ID())
	uassert.Equal(t, "Test Dapp", r.String())

	fields, err := r.Fields()
	urequire.NoError(t, err)

	author, found := fields.Get("author")
	urequire.True(t, found, "author field")
	uassert.Equal(t, addr.String(), author)

	url, found := fields.Get("url")
	urequire.True(t, found, "url field")
	uassert.Equal(t, "https://test.com", url)

	content, err := r.Content()
	urequire.NoError(t, err)

	wantContent := md.H1(dapp.Title) +
		md.H2("Description") +
		dapp.Description +
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
	dapps := avl.NewTree()
	dapp := &Dapp{
		ID:    1,
		Title: "Test Dapp",
		URL:   "https://test.com",
	}
	dapps.Set(dapp.ID.String(), dapp)

	category := Category{
		ID:          1,
		Name:        "Test Category",
		Description: "This is a test category",
		CreatedAt:   time.Now(),
		dapps:       dapps,
	}

	r := categoryRecord{&category}

	uassert.Equal(t, "0000001", r.ID())
	uassert.Equal(t, "Test Category", r.String())

	fields, err := r.Fields()
	urequire.NoError(t, err)

	dappCount, found := fields.Get("dapp_count")
	urequire.True(t, found, "dapp_count field")
	uassert.Equal(t, "1 dapps", dappCount)

	content, err := r.Content()
	urequire.NoError(t, err)

	wantContent := md.H1(category.Name) +
		md.H2("Description") +
		category.Description +
		md.H2("Dapps in this category") +
		md.BulletItem(ufmt.Sprintf("[%s](%s)", dapp.Title, dapp.URL))

	uassert.Equal(t, wantContent, content)
}
