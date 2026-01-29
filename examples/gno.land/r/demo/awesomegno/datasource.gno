package awesomegno

import (
	"errors"

	"gno.land/p/jeronimoalbi/datasource"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

// Items

func NewItemDatasource(cur realm) ItemDatasource {
	return ItemDatasource{items}
}

type ItemDatasource struct {
	items *avl.Tree
}

func (ds ItemDatasource) Size() int { return ds.items.Size() }

func (ds ItemDatasource) Records(q datasource.Query) datasource.Iterator {
	return &itemIterator{
		items:    ds.items,
		index:    q.Offset,
		maxIndex: q.Offset + q.Count,
	}
}

func (ds ItemDatasource) Record(id string) (datasource.Record, error) {
	v, found := ds.items.Get(id)
	if !found {
		return nil, errors.New(ErrItemNotFound + ": " + id)
	}
	return itemRecord{v.(*Item)}, nil
}

type itemRecord struct {
	item *Item
}

func (r itemRecord) ID() string     { return r.item.ID.String() }
func (r itemRecord) String() string { return r.item.Title }

func (r itemRecord) Fields() (datasource.Fields, error) {
	fields := avl.NewTree()
	fields.Set("authors", r.item.Authors)
	fields.Set("url", r.item.URL)
	return fields, nil
}

func (r itemRecord) Content() (string, error) {
	content := md.H1(r.item.Title)
	content += md.H2("Description")
	content += r.item.Description
	content += md.H2("Categories")

	for _, catID := range r.item.Categories {
		if cat, found := categories.Get(catID.String()); found {
			category := cat.(*Category)
			content += md.BulletItem(category.Name)
		}
	}

	return content, nil
}

type itemIterator struct {
	items           *avl.Tree
	index, maxIndex int
	record          *itemRecord
}

func (it itemIterator) Record() datasource.Record { return it.record }
func (it itemIterator) Err() error                { return nil }

func (it *itemIterator) Next() bool {
	if it.index >= it.maxIndex || it.index >= it.items.Size() {
		return false
	}

	_, v := it.items.GetByIndex(it.index)
	it.record = &itemRecord{v.(*Item)}
	it.index++
	return true
}

// Categories

func NewCategoryDatasource(cur realm) CategoryDatasource {
	return CategoryDatasource{categories}
}

type CategoryDatasource struct {
	categories *avl.Tree
}

func (ds CategoryDatasource) Size() int { return ds.categories.Size() }

func (ds CategoryDatasource) Records(q datasource.Query) datasource.Iterator {
	return &categoryIterator{
		categories: ds.categories,
		index:      q.Offset,
		maxIndex:   q.Offset + q.Count,
	}
}

func (ds CategoryDatasource) Record(id string) (datasource.Record, error) {
	v, found := ds.categories.Get(id)
	if !found {
		return nil, errors.New(ErrCategoryNotFound + ": " + id)
	}
	return categoryRecord{v.(*Category)}, nil
}

type categoryRecord struct {
	category *Category
}

func (r categoryRecord) ID() string     { return r.category.ID.String() }
func (r categoryRecord) String() string { return r.category.Name }

func (r categoryRecord) Fields() (datasource.Fields, error) {
	fields := avl.NewTree()
	fields.Set("item_count", ufmt.Sprintf("%d items", r.category.items.Size()))
	return fields, nil
}

func (r categoryRecord) Content() (string, error) {
	content := md.H1(r.category.Name)
	content += md.H2("Description")
	content += r.category.Description
	content += md.H2("Items in this category")

	r.category.items.Iterate("", "", func(key string, value interface{}) bool {
		item := value.(*Item)
		content += md.BulletItem(ufmt.Sprintf("[%s](%s)", item.Title, item.URL))
		return false
	})

	return content, nil
}

type categoryIterator struct {
	categories      *avl.Tree
	index, maxIndex int
	record          *categoryRecord
}

func (it categoryIterator) Record() datasource.Record { return it.record }
func (it categoryIterator) Err() error                { return nil }

func (it *categoryIterator) Next() bool {
	if it.index >= it.maxIndex || it.index >= it.categories.Size() {
		return false
	}

	_, v := it.categories.GetByIndex(it.index)
	it.record = &categoryRecord{v.(*Category)}
	it.index++
	return true
}
