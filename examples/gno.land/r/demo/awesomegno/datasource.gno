package awesomegno

import (
	"errors"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/jeronimoalbi/datasource"
	"gno.land/p/moul/md"
)

// Dapps

func NewDappDatasource() DappDatasource {
	return DappDatasource{dapps}
}

type DappDatasource struct {
	dapps *avl.Tree
}

func (ds DappDatasource) Size() int { return ds.dapps.Size() }

func (ds DappDatasource) Records(q datasource.Query) datasource.Iterator {
	return &dappIterator{
		dapps:    ds.dapps,
		index:    q.Offset,
		maxIndex: q.Offset + q.Count,
	}
}

func (ds DappDatasource) Record(id string) (datasource.Record, error) {
	v, found := ds.dapps.Get(id)
	if !found {
		return nil, errors.New("dapp not found")
	}
	return dappRecord{v.(*Dapp)}, nil
}

type dappRecord struct {
	dapp *Dapp
}

func (r dappRecord) ID() string     { return r.dapp.ID.String() }
func (r dappRecord) String() string { return r.dapp.Title }

func (r dappRecord) Fields() (datasource.Fields, error) {
	fields := avl.NewTree()
	fields.Set("author", r.dapp.Author.String())
	fields.Set("url", r.dapp.URL)
	return fields, nil
}

func (r dappRecord) Content() (string, error) {
	content := md.H1(r.dapp.Title)
	content += md.H2("Description")
	content += r.dapp.Description
	content += md.H2("Categories")

	r.dapp.categories.Iterate("", "", func(key string, value interface{}) bool {
		cat := value.(*Category)
		content += md.BulletItem(cat.Name)
		return false
	})

	return content, nil
}

type dappIterator struct {
	dapps           *avl.Tree
	index, maxIndex int
	record          *dappRecord
}

func (it dappIterator) Record() datasource.Record { return it.record }
func (it dappIterator) Err() error                { return nil }

func (it *dappIterator) Next() bool {
	if it.index >= it.maxIndex || it.index >= it.dapps.Size() {
		return false
	}

	_, v := it.dapps.GetByIndex(it.index)
	it.record = &dappRecord{v.(*Dapp)}
	it.index++
	return true
}

// Categories

func NewCategoryDatasource() CategoryDatasource {
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
		return nil, errors.New("category not found")
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
	fields.Set("dapp_count", ufmt.Sprintf("%d dapps", r.category.dapps.Size()))
	return fields, nil
}

func (r categoryRecord) Content() (string, error) {
	content := md.H1(r.category.Name)
	content += md.H2("Description")
	content += r.category.Description
	content += md.H2("Dapps in this category")

	r.category.dapps.Iterate("", "", func(key string, value interface{}) bool {
		dapp := value.(*Dapp)
		content += md.BulletItem(ufmt.Sprintf("[%s](%s)", dapp.Title, dapp.URL))
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
