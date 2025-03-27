package awesomegno

import (
	"errors"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/jeronimoalbi/datasource"
	"gno.land/p/moul/md"
)

// Dapp datasource
func NewDappsDatasource() *DappsDatasource {
	return &DappsDatasource{}
}

type DappsDatasource struct{}

func (ds *DappsDatasource) Size() int {
	return dapps.Size()
}

func (ds *DappsDatasource) Records(q datasource.Query) datasource.Iterator {
	return &dappsIterator{
		index:    q.Offset,
		maxIndex: q.Offset + q.Count,
	}
}

func (ds *DappsDatasource) Record(id string) (datasource.Record, error) {
	dappID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New("invalid dapp ID")
	}

	dapp, exists := GetDapp(dappID)
	if !exists {
		return nil, errors.New("dapp not found")
	}

	return &dappRecord{dapp: dapp}, nil
}

type dappRecord struct {
	dapp Dapp
}

func (r *dappRecord) ID() string {
	return strconv.FormatUint(r.dapp.ID, 10)
}

func (r *dappRecord) String() string {
	return r.dapp.Title
}

func (r *dappRecord) Fields() (datasource.Fields, error) {
	fields := avl.NewTree()

	fields.Set("title", r.dapp.Title)
	fields.Set("description", r.dapp.Description)
	fields.Set("author", r.dapp.Author.String())
	fields.Set("url", r.dapp.URL)
	fields.Set("approved", r.dapp.Approved)
	fields.Set("created_at", r.dapp.CreatedAt)

	categoryStr := ""
	for i, cat := range r.dapp.Categories {
		if i > 0 {
			categoryStr += ", "
		}
		categoryStr += cat
	}
	fields.Set("categories", categoryStr)

	return fields, nil
}

func (r *dappRecord) Tags() []string {
	println(r.dapp)
	return r.dapp.Categories
}

func (r *dappRecord) Content() (string, error) {
	content := md.H1(r.dapp.Title)
	content += r.dapp.Description + "\n\n"

	if len(r.dapp.Categories) > 0 {
		categoryStr := ""
		for i, cat := range r.dapp.Categories {
			if i > 0 {
				categoryStr += ", "
			}
			categoryStr += cat
		}
		content += md.Bold("Categories:") + " " + categoryStr + "\n\n"
	}

	content += md.Bold("URL:") + " " + md.Link(r.dapp.URL, r.dapp.URL) + "\n\n"
	content += md.Bold("Author:") + " " + r.dapp.Author.String() + "\n\n"
	content += md.Bold("Created:") + " " + r.dapp.CreatedAt.String() + "\n\n"

	return content, nil
}

type dappsIterator struct {
	index, maxIndex int
	record          *dappRecord
	err             error
}

func (it *dappsIterator) Next() bool {
	if it.index >= it.maxIndex || it.index >= dapps.Size() {
		return false
	}

	_, value := dapps.GetByIndex(it.index)
	it.record = &dappRecord{dapp: value.(Dapp)}
	it.index++
	return true
}

func (it *dappsIterator) Record() datasource.Record {
	return it.record
}

func (it *dappsIterator) Err() error {
	return it.err
}

// Categories datasource
func NewCategoriesDatasource() *CategoriesDatasource {
	return &CategoriesDatasource{}
}

type CategoriesDatasource struct{}

func (ds *CategoriesDatasource) Size() int {
	return categories.Size()
}

func (ds *CategoriesDatasource) Records(q datasource.Query) datasource.Iterator {
	return &categoriesIterator{
		index:    q.Offset,
		maxIndex: q.Offset + q.Count,
	}
}

func (ds *CategoriesDatasource) Record(id string) (datasource.Record, error) {
	category, err := GetCategory(id)
	if err != nil {
		return nil, err
	}

	return &categoryRecord{category: category}, nil
}

type categoryRecord struct {
	category Category
}

func (r *categoryRecord) ID() string {
	return r.category.Name
}

func (r *categoryRecord) String() string {
	return r.category.Name
}

func (r *categoryRecord) Fields() (datasource.Fields, error) {
	fields := avl.NewTree()

	fields.Set("id", r.category.ID)
	fields.Set("name", r.category.Name)
	fields.Set("description", r.category.Description)
	fields.Set("created_at", r.category.CreatedAt)
	fields.Set("dapp_count", len(r.category.DappIDs))

	return fields, nil
}

func (r *categoryRecord) Content() (string, error) {
	content := md.H1("Category: " + r.category.Name)

	if r.category.Description != "" {
		content += r.category.Description + "\n\n"
	}

	content += md.Bold("Dapps:") + " " + strconv.Itoa(len(r.category.DappIDs)) + "\n\n"
	content += md.Bold("Created:") + " " + r.category.CreatedAt.String() + "\n\n"

	return content, nil
}

type categoriesIterator struct {
	index, maxIndex int
	record          *categoryRecord
	err             error
}

func (it *categoriesIterator) Next() bool {
	if it.index >= it.maxIndex || it.index >= categories.Size() {
		return false
	}

	_, value := categories.GetByIndex(it.index)
	it.record = &categoryRecord{category: value.(Category)}
	it.index++
	return true
}

func (it *categoriesIterator) Record() datasource.Record {
	return it.record
}

func (it *categoriesIterator) Err() error {
	return it.err
}

// Proposals datasource
func NewProposalsDatasource() *ProposalsDatasource {
	return &ProposalsDatasource{}
}

type ProposalsDatasource struct{}

func (ds *ProposalsDatasource) Size() int {
	return proposals.Size()
}

func (ds *ProposalsDatasource) Records(q datasource.Query) datasource.Iterator {
	return &proposalsIterator{
		index:    q.Offset,
		maxIndex: q.Offset + q.Count,
	}
}

func (ds *ProposalsDatasource) Record(id string) (datasource.Record, error) {
	proposalID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New("invalid proposal ID")
	}

	proposal, exists := GetProposal(proposalID)
	if !exists {
		return nil, errors.New("proposal not found")
	}

	return &proposalRecord{proposal: proposal}, nil
}

type proposalRecord struct {
	proposal Proposal
}

func (r *proposalRecord) ID() string {
	return strconv.FormatUint(r.proposal.ID, 10)
}

func (r *proposalRecord) String() string {
	return r.proposal.Dapp.Title
}

func (r *proposalRecord) Fields() (datasource.Fields, error) {
	fields := avl.NewTree()

	fields.Set("id", r.proposal.ID)
	fields.Set("title", r.proposal.Dapp.Title)
	fields.Set("description", r.proposal.Dapp.Description)
	fields.Set("proposer", r.proposal.Proposer.String())
	fields.Set("status", r.proposal.Status)
	fields.Set("reason", r.proposal.Reason)
	fields.Set("created_at", r.proposal.CreatedAt)
	fields.Set("url", r.proposal.Dapp.URL)

	categoryStr := ""
	for i, cat := range r.proposal.Dapp.Categories {
		if i > 0 {
			categoryStr += ", "
		}
		categoryStr += cat
	}
	fields.Set("categories", categoryStr)

	return fields, nil
}

func (r *proposalRecord) Tags() []string {
	tags := append(r.proposal.Dapp.Categories, r.proposal.Status)
	return tags
}

func (r *proposalRecord) Content() (string, error) {
	content := md.H1("Proposal: " + r.proposal.Dapp.Title)
	content += md.Bold("Status: ") + r.proposal.Status + "\n\n"

	content += r.proposal.Dapp.Description + "\n\n"

	if len(r.proposal.Dapp.Categories) > 0 {
		categoryStr := ""
		for i, cat := range r.proposal.Dapp.Categories {
			if i > 0 {
				categoryStr += ", "
			}
			categoryStr += cat
		}
		content += md.Bold("Categories: ") + categoryStr + "\n\n"
	}

	content += md.Bold("URL: ") + md.Link(r.proposal.Dapp.URL, r.proposal.Dapp.URL) + "\n\n"
	content += md.Bold("Proposer: ") + r.proposal.Proposer.String() + "\n\n"

	if r.proposal.Reason != "" {
		content += md.Bold("Reason: ") + r.proposal.Reason + "\n\n"
	}

	content += md.Bold("Created: ") + r.proposal.CreatedAt.String() + "\n\n"

	return content, nil
}

type proposalsIterator struct {
	index, maxIndex int
	record          *proposalRecord
	err             error
}

func (it *proposalsIterator) Next() bool {
	if it.index >= it.maxIndex || it.index >= proposals.Size() {
		return false
	}

	_, value := proposals.GetByIndex(it.index)
	it.record = &proposalRecord{proposal: value.(Proposal)}
	it.index++
	return true
}

func (it *proposalsIterator) Record() datasource.Record {
	return it.record
}

func (it *proposalsIterator) Err() error {
	return it.err
}
