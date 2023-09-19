package dao_registry

import (
	"encoding/binary"
	"std"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/teritori/ujson"
)

// TODO: rewrite

type Registration struct {
	PkgPath     string      `json:"pkgPath"`
	Addr        std.Address `json:"addr"`
	CreatedAt   uint64      `json:"createdAt"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ImageURI    string      `json:"imageURI"`
}

var (
	byCreatedAt = avl.NewTree()
	byPkgPath   = avl.NewTree()
)

// TODO: use profile for info
// TODO: make core accessible

func Register(name string, description string, imageURI string) {
	realm := std.PrevRealm()
	pkgPath := realm.PkgPath()
	if pkgPath == "" {
		panic("no pkg path in prev realm")
	}
	doRegister(pkgPath, realm.Addr(), name, description, imageURI)
}

// splitted for tests
func doRegister(pkgPath string, addr std.Address, name string, description string, imageURI string) {
	if byPkgPath.Has(pkgPath) {
		panic("already registered")
	}
	registration := &Registration{
		PkgPath:     pkgPath,
		Addr:        addr,
		CreatedAt:   uint64(time.Now().UnixMilli()),
		Name:        name,
		Description: description,
		ImageURI:    imageURI,
	}
	byPkgPath.Set(pkgPath, registration)

	keyBytes := []byte{}
	keyBytes = binary.BigEndian.AppendUint64(keyBytes, uint64(registration.CreatedAt))
	keyBytes = append(keyBytes, []byte(pkgPath)...)
	key := string(keyBytes)
	byCreatedAt.Set(key, registration)
}

func GetJSON(pkgPath string) string {
	r, ok := byPkgPath.Get(pkgPath)
	if !ok {
		panic("not found")
	}
	return r.(*Registration).ToJSON()
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
	return ujson.FormatSlice(regs)
}

func (r Registration) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "pkgPath", Value: r.PkgPath},
		{Key: "addr", Value: r.Addr},
		{Key: "createdAt", Value: r.CreatedAt},
		{Key: "name", Value: r.Name},
		{Key: "description", Value: r.Description},
		{Key: "imageURI", Value: r.ImageURI},
	})
}
