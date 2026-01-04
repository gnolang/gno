package hor

import (
	"errors"

	"gno.land/p/jeronimoalbi/datasource"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

func NewDatasource() Datasource {
	return Datasource{exhibition}
}

type Datasource struct {
	exhibition *Exhibition
}

func (ds Datasource) Size() int { return ds.exhibition.items.Size() }

func (ds Datasource) Records(q datasource.Query) datasource.Iterator {
	return &iterator{
		exhibition: ds.exhibition,
		index:      q.Offset,
		maxIndex:   q.Offset + q.Count,
	}
}

func (ds Datasource) Record(id string) (datasource.Record, error) {
	v, found := ds.exhibition.items.Get(id)
	if !found {
		return nil, errors.New("realm submission not found")
	}
	return record{v.(*Item)}, nil
}

type record struct {
	item *Item
}

func (r record) ID() string     { return r.item.id.String() }
func (r record) String() string { return r.item.pkgpath }

func (r record) Fields() (datasource.Fields, error) {
	fields := avl.NewTree()
	fields.Set(
		"details",
		ufmt.Sprintf("Votes: ⏶ %d - ⏷ %d", r.item.upvote.Size(), r.item.downvote.Size()),
	)
	return fields, nil
}

func (r record) Content() (string, error) {
	content := r.item.Render(false)
	return content, nil
}

type iterator struct {
	exhibition      *Exhibition
	index, maxIndex int
	record          *record
}

func (it iterator) Record() datasource.Record { return it.record }
func (it iterator) Err() error                { return nil }

func (it *iterator) Next() bool {
	if it.index >= it.maxIndex || it.index >= it.exhibition.items.Size() {
		return false
	}

	_, v := it.exhibition.items.GetByIndex(it.index)
	it.record = &record{v.(*Item)}
	it.index++
	return true
}
