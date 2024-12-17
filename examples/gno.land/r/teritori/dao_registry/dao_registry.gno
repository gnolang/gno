package dao_registry

import (
	"encoding/binary"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/teritori/dao_interfaces"
	"gno.land/p/teritori/jsonutil"
	"gno.land/p/demo/json"
)

type Registration struct {
	PkgPath     string      `json:"pkgPath"`
	Addr        std.Address `json:"addr"`
	CreatedAt   time.Time   `json:"createdAt"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ImageURI    string      `json:"imageURI"`
	Getter      DAOGetter   `json:"-"`
}

var (
	byCreatedAt = avl.NewTree()
	byPkgPath   = avl.NewTree()
)

// TODO: use profile realm for info

type DAOGetter func() dao_interfaces.IDAOCore

func Register(getter DAOGetter, name string, description string, imageURI string) {
	realm := std.PrevRealm()
	pkgPath := realm.PkgPath()
	if pkgPath == "" {
		panic("caller is not a realm")
	}

	doRegister(getter, pkgPath, realm.Addr(), name, description, imageURI)
}

// splitted for tests
func doRegister(getter DAOGetter, pkgPath string, addr std.Address, name string, description string, imageURI string) {
	if byPkgPath.Has(pkgPath) {
		panic("already registered")
	}

	registration := &Registration{
		PkgPath:     pkgPath,
		Addr:        addr,
		CreatedAt:   time.Now(),
		Name:        name,
		Description: description,
		ImageURI:    imageURI,
		Getter:      getter,
	}
	byPkgPath.Set(pkgPath, registration)

	keyBytes := []byte{}
	keyBytes = binary.BigEndian.AppendUint64(keyBytes, uint64(registration.CreatedAt.UnixMilli())) // won't be before unix epoch
	keyBytes = append(keyBytes, []byte(pkgPath)...)
	key := string(keyBytes)
	byCreatedAt.Set(key, registration)
}

func GetCore(pkgPath string) dao_interfaces.IDAOCore {
	r, ok := byPkgPath.Get(pkgPath)
	if !ok {
		panic("not found")
	}

	return r.(*Registration).Getter()
}

func GetJSON(pkgPath string) string {
	r, ok := byPkgPath.Get(pkgPath)
	if !ok {
		panic("not found")
	}

	return r.(*Registration).ToJSON().String()
}

func List(start string, end string, limit int, reverse bool) []Registration {
	regs := []Registration{}
	cb := func(key string, value interface{}) bool {
		regs = append(regs, *value.(*Registration))
		if limit == 0 {
			return false
		}
		return len(regs) >= limit
	}

	if reverse {
		byCreatedAt.ReverseIterate(start, end, cb)
	} else {
		byCreatedAt.Iterate(start, end, cb)
	}

	return regs
}

func ListJSON(start string, end string, limit int, reverse bool) string {
	regs := List(start, end, limit, reverse)
	iSlice := make([]*json.Node, len(regs))
	for i, v := range regs {
		iSlice[i] = v.ToJSON()
	}

	return json.ArrayNode("", iSlice).String()
}

func (r Registration) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"pkgPath":     json.StringNode("", r.PkgPath),
		"addr":        jsonutil.AddressNode(r.Addr),
		"createdAt":   jsonutil.TimeNode(r.CreatedAt),
		"name":        json.StringNode("", r.Name),
		"description": json.StringNode("", r.Description),
		"imageURI":    json.StringNode("", r.ImageURI),
	})
}
