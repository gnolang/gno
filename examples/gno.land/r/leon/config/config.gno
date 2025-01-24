package config

import (
	"errors"
	"std"
	"strconv"
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
	banner  = "---\n[[Leon's Home page]](/r/leon/home) | [[GitHub: @leohhhn]](https://github.com/leohhhn)\n\n---"

	// SafeObjects
	OwnableMain   = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
	OwnableBackup = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

	ErrUnauthorized = errors.New("leon's config: unauthorized")
)

type Config struct {
	lines   string
	updated time.Time
}

func AddConfig(name, lines string) {
	if !IsAuthorized(std.PrevRealm().Addr()) {
		panic(ErrUnauthorized)
	}

	configs.Set(name, Config{
		lines:   lines,
		updated: time.Now(),
	}) // no overwrite check
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
	return addr == OwnableMain.Owner() || addr == OwnableBackup.Owner()
}

func Banner() string {
	return banner
}

func Render(path string) (out string) {
	req := realmpath.Parse(path)
	if req.Path == "" {
		out += md.H1("Leon's config package")

		out += ufmt.Sprintf("Leon's main address: %s\n\n", OwnableMain.Owner().String())
		out += ufmt.Sprintf("Leon's backup address: %s\n\n", OwnableBackup.Owner().String())

		out += md.H2("Leon's configs")

		if configs.Size() == 0 {
			out += "No configs yet :c\n\n"
		}

		page := pager.MustGetPageByPath(path)
		for _, item := range page.Items {
			out += ufmt.Sprintf("- [%s](/r/leon/config:%s)\n\n", item.Key, item.Key)
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
