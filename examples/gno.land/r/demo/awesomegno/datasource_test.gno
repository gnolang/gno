package awesomegno

import (
	"std"
	"strconv"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/jeronimoalbi/datasource"
)

var (
	_ datasource.Datasource    = (*DappsDatasource)(nil)
	_ datasource.Record        = (*dappRecord)(nil)
	_ datasource.ContentRecord = (*dappRecord)(nil)
	_ datasource.Iterator      = (*dappsIterator)(nil)

	_ datasource.Datasource    = (*CategoriesDatasource)(nil)
	_ datasource.Record        = (*categoryRecord)(nil)
	_ datasource.ContentRecord = (*categoryRecord)(nil)
	_ datasource.Iterator      = (*categoriesIterator)(nil)

	_ datasource.Datasource    = (*ProposalsDatasource)(nil)
	_ datasource.Record        = (*proposalRecord)(nil)
	_ datasource.ContentRecord = (*proposalRecord)(nil)
	_ datasource.Iterator      = (*proposalsIterator)(nil)
)

func setupTestData() {
	dapps = avl.NewTree()
	categories = avl.NewTree()
	proposals = avl.NewTree()

	testAddress := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	cat1 := Category{
		ID:          1,
		Name:        "Games",
		Description: "Gaming applications",
		CreatedAt:   time.Now(),
		DappIDs:     []uint64{1},
	}
	cat2 := Category{
		ID:          2,
		Name:        "Finance",
		Description: "Financial applications",
		CreatedAt:   time.Now(),
		DappIDs:     []uint64{2},
	}
	categories.Set(cat1.Name, cat1)
	categories.Set(cat2.Name, cat2)

	dapp1 := Dapp{
		ID:          1,
		Title:       "Game Dapp",
		Description: "A cool game on Gno.land",
		Author:      testAddress,
		URL:         "https://example.com/game",
		Categories:  []string{"Games"},
		Approved:    true,
		CreatedAt:   time.Now(),
	}
	dapp2 := Dapp{
		ID:          2,
		Title:       "Finance Dapp",
		Description: "A DeFi application",
		Author:      testAddress,
		URL:         "https://example.com/finance",
		Categories:  []string{"Finance"},
		Approved:    false,
		CreatedAt:   time.Now(),
	}
	dapps.Set(strconv.FormatUint(dapp1.ID, 10), dapp1)
	dapps.Set(strconv.FormatUint(dapp2.ID, 10), dapp2)

	proposal1 := Proposal{
		ID:        1,
		Dapp:      dapp1,
		Proposer:  testAddress,
		Status:    StatusApproved,
		Reason:    "",
		CreatedAt: time.Now(),
	}
	proposal2 := Proposal{
		ID:        2,
		Dapp:      dapp2,
		Proposer:  testAddress,
		Status:    StatusPending,
		Reason:    "",
		CreatedAt: time.Now(),
	}
	proposals.Set(strconv.FormatUint(proposal1.ID, 10), proposal1)
	proposals.Set(strconv.FormatUint(proposal2.ID, 10), proposal2)
}

func TestDappsDatasourceRecords(t *testing.T) {
	setupTestData()

	cases := []struct {
		name      string
		recordIDs []string
		options   []datasource.QueryOption
	}{
		{
			name:      "all dapps",
			recordIDs: []string{"1", "2"},
		},
		{
			name:      "with offset",
			recordIDs: []string{"2"},
			options:   []datasource.QueryOption{datasource.WithOffset(1)},
		},
		{
			name:      "with count",
			recordIDs: []string{"1"},
			options:   []datasource.QueryOption{datasource.WithCount(1)},
		},
		{
			name:      "with offset and count",
			recordIDs: []string{"2"},
			options: []datasource.QueryOption{
				datasource.WithOffset(1),
				datasource.WithCount(1),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ds := NewDappsDatasource()
			query := datasource.NewQuery(tc.options...)
			iter := ds.Records(query)

			urequire.Equal(t, dapps.Size(), ds.Size(), "datasource size")

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

func TestDappsDatasourceRecord(t *testing.T) {
	setupTestData()

	cases := []struct {
		name string
		id   string
		err  string
	}{
		{
			name: "found",
			id:   "1",
		},
		{
			name: "not found",
			id:   "999",
			err:  "dapp not found",
		},
		{
			name: "invalid id",
			id:   "invalid",
			err:  "invalid dapp ID",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ds := NewDappsDatasource()
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
	setupTestData()

	dapp, exists := GetDapp(1)
	urequire.True(t, exists, "dapp exists")

	r := dappRecord{dapp: dapp}

	uassert.Equal(t, "1", r.ID())
	uassert.Equal(t, dapp.Title, r.String())

	fields, err := r.Fields()
	urequire.NoError(t, err, "fields error")

	title, found := fields.Get("title")
	urequire.True(t, found, "title field exists")
	uassert.Equal(t, dapp.Title, title)

	url, found := fields.Get("url")
	urequire.True(t, found, "url field exists")
	uassert.Equal(t, dapp.URL, url)

	content, err := r.Content()
	urequire.NoError(t, err, "content error")
	uassert.True(t, strings.Contains(content, dapp.Title), "content contains title")
	uassert.True(t, strings.Contains(content, dapp.URL), "content contains URL")
}

func TestCategoriesDatasourceRecords(t *testing.T) {
	setupTestData()

	cases := []struct {
		name      string
		recordIDs []string
		options   []datasource.QueryOption
	}{
		{
			name:      "all categories",
			recordIDs: []string{"Finance", "Games"},
		},
		{
			name:      "with offset",
			recordIDs: []string{"Games"},
			options:   []datasource.QueryOption{datasource.WithOffset(1)},
		},
		{
			name:      "with count",
			recordIDs: []string{"Finance"},
			options:   []datasource.QueryOption{datasource.WithCount(1)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ds := NewCategoriesDatasource()
			query := datasource.NewQuery(tc.options...)
			iter := ds.Records(query)

			urequire.Equal(t, categories.Size(), ds.Size(), "datasource size")

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

func TestCategoriesDatasourceRecord(t *testing.T) {
	setupTestData()

	cases := []struct {
		name string
		id   string
		err  string
	}{
		{
			name: "found",
			id:   "Games",
		},
		{
			name: "not found",
			id:   "Unknown",
			err:  "category not found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ds := NewCategoriesDatasource()
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

func TestCategoryRecord(t *testing.T) {
	setupTestData()

	category, err := GetCategory("Games")
	urequire.NoError(t, err, "category exists")

	r := categoryRecord{category: category}

	uassert.Equal(t, "Games", r.ID())
	uassert.Equal(t, category.Name, r.String())

	fields, err := r.Fields()
	urequire.NoError(t, err, "fields error")

	name, found := fields.Get("name")
	urequire.True(t, found, "name field exists")
	uassert.Equal(t, category.Name, name)

	desc, found := fields.Get("description")
	urequire.True(t, found, "description field exists")
	uassert.Equal(t, category.Description, desc)

	content, err := r.Content()
	urequire.NoError(t, err, "content error")
	uassert.True(t, strings.Contains(content, category.Name), "content contains name")
	uassert.True(t, strings.Contains(content, category.Description), "content contains description")
}

func TestProposalsDatasourceRecords(t *testing.T) {
	setupTestData()

	cases := []struct {
		name      string
		recordIDs []string
		options   []datasource.QueryOption
	}{
		{
			name:      "all proposals",
			recordIDs: []string{"1", "2"},
		},
		{
			name:      "with offset",
			recordIDs: []string{"2"},
			options:   []datasource.QueryOption{datasource.WithOffset(1)},
		},
		{
			name:      "with count",
			recordIDs: []string{"1"},
			options:   []datasource.QueryOption{datasource.WithCount(1)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ds := NewProposalsDatasource()
			query := datasource.NewQuery(tc.options...)
			iter := ds.Records(query)

			urequire.Equal(t, proposals.Size(), ds.Size(), "datasource size")

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

func TestProposalsDatasourceRecord(t *testing.T) {
	setupTestData()

	cases := []struct {
		name string
		id   string
		err  string
	}{
		{
			name: "found",
			id:   "1",
		},
		{
			name: "not found",
			id:   "999",
			err:  "proposal not found",
		},
		{
			name: "invalid id",
			id:   "invalid",
			err:  "invalid proposal ID",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ds := NewProposalsDatasource()
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

func TestProposalRecord(t *testing.T) {
	setupTestData()

	proposal, exists := GetProposal(1)
	urequire.True(t, exists, "proposal exists")

	r := proposalRecord{proposal: proposal}

	uassert.Equal(t, "1", r.ID())
	uassert.Equal(t, proposal.Dapp.Title, r.String())

	fields, err := r.Fields()
	urequire.NoError(t, err, "fields error")

	title, found := fields.Get("title")
	urequire.True(t, found, "title field exists")
	uassert.Equal(t, proposal.Dapp.Title, title)

	status, found := fields.Get("status")
	urequire.True(t, found, "status field exists")
	uassert.Equal(t, proposal.Status, status)

	content, err := r.Content()
	urequire.NoError(t, err, "content error")
	uassert.True(t, strings.Contains(content, proposal.Dapp.Title), "content contains title")
	uassert.True(t, strings.Contains(content, proposal.Status), "content contains status")

	tags := r.Tags()
	uassert.Equal(t, len(proposal.Dapp.Categories)+1, len(tags), "tags length")
	uassert.Equal(t, proposal.Status, tags[len(tags)-1], "status tag")
}
