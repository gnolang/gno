package config

import (
	"errors"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	p "gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/realmpath"
)

var (
	configs = avl.NewTree()
	pager   = p.NewPager(configs, 10, false)
	banner  = "---\n[[mouss's Home page]](/r/mouss/home) | [[GitHub: @mouss1985]](https://github.com/1985)\n\n---"
	absPath = strings.TrimPrefix(std.CurrentRealm().PkgPath(), std.GetChainDomain())

	// SafeObjects
	OwnableMain   = ownable.NewWithAddress("g1hrfvdh7jdvnlxpk2y20tp3scj9jqal3zzu7wjz")
	OwnableBackup = ownable.NewWithAddress("g1wq2h93ppkf2gkgncz5unayrsmt7pl8npktnznd")

	ErrUnauthorized = errors.New("mouss's config: unauthorized")
)

func Address() std.Address {
	return OwnableMain.Owner()
}

func Backup() std.Address {
	return OwnableBackup.Owner()
}

type Config struct {
	lines   string
	updated time.Time
}

func AddConfig(name, lines string) {
	if !IsAuthorized(std.PrevRealm().Addr()) {
		panic(ErrUnauthorized)
	}

	if _, ok := configs.Get(name); ok {
		panic("config with that name already exists")
	}

	configs.Set(name, Config{
		lines:   lines,
		updated: time.Now(),
	})
}

func RemoveConfig(name string) {
	if !IsAuthorized(std.PrevRealm().Addr()) {
		panic(ErrUnauthorized)
	}

	if _, ok := configs.Remove(name); !ok {
		panic("no config with that name")
	}
}

func UpdateBanner(newBanner string) {
	if !IsAuthorized(std.PrevRealm().Addr()) {
		panic(ErrUnauthorized)
	}

	banner = newBanner
}

func IsAuthorized(addr std.Address) bool {
	return addr == Address() || addr == Backup()
}

func Banner() string {
	return banner
}

func Render(path string) (out string) {
	req := realmpath.Parse(path)
	if req.Path == "" {
		out += md.H1("mouss's config package")

		out += ufmt.Sprintf("mouss's main address: %s\n\n", Address().String())
		out += ufmt.Sprintf("mouss's backup address: %s\n\n", Backup().String())

		out += md.H2("mouss's configs")

		if configs.Size() == 0 {
			out += "No configs yet :c\n\n"
		}

		page := pager.MustGetPageByPath(path)
		for _, item := range page.Items {
			out += ufmt.Sprintf("- [%s](%s:%s)\n\n", item.Key, absPath, item.Key)
		}

		out += page.Picker()
		out += "\n\n"
		out += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"

		out += Banner()

		return out
	}

	return renderConfPage(req.Path)
}

func renderConfPage(confName string) (out string) {
	raw, ok := configs.Get(confName)
	if !ok {
		out += md.H1("404")
		out += "That config does not exist :/"
		return out
	}

	conf := raw.(Config)
	out += md.H1(confName)
	out += ufmt.Sprintf("```\n%s\n```\n\n", conf.lines)
	out += ufmt.Sprintf("_Last updated on %s_", conf.updated.Format("02 Jan, 2006"))

	return out
}
