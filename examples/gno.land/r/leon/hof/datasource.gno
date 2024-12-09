package hof

import (
	"errors"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnome/datasource"
)

type Datasource struct{}

func (ds Datasource) GetRecords(offset, _ int, _ string) datasource.RecordIterator {
	return &iterator{index: offset}
}

func (ds Datasource) GetContent(rawID string, fn func(content string)) error {
	id, err := seqid.FromString(rawID)
	if err != nil {
		return err
	}

	v, ok := exhibition.itemsSorted.Get(id.String())
	if !ok {
		return errors.New("realm submission not found")
	}

	content := ufmt.Sprintf("# Submission #%d\n\n", int(id))
	content += v.(*Item).Render(false)

	fn(content)
	return nil
}

type record struct {
	item *Item
}

func (r record) ID() string     { return r.item.id.String() }
func (r record) Title() string  { return r.item.pkgpath }
func (r record) Tags() []string { return nil }

func (r record) Details() string {
	return ufmt.Sprintf("Votes: ⏶ %d - ⏷ %d", r.item.upvote.Size(), r.item.downvote.Size())
}

type iterator struct {
	index  int
	record *record
}

func (it iterator) Record() datasource.Record { return it.record }
func (it iterator) Err() error                { return nil }

func (it *iterator) Next() bool {
	if it.index >= exhibition.itemsSorted.Size() {
		return false
	}

	_, v := exhibition.itemsSorted.GetByIndex(it.index)
	it.record = &record{v.(*Item)}
	it.index++
	return true
}
