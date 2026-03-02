package config

import (
	"strconv"

	"gno.land/p/moul/md"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
	p "gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/ufmt/v0"
)

var (
	banner = "---\n[[Leon's Home page]](/r/leon/home) | [[Leon's snippets]](/r/leon/config) | [[GitHub: @leohhhn]](https://github.com/leohhhn)\n\n---"
	pager  = p.NewPager(configs, 10, true)
)

func Banner() string {
	return banner
}

func Render(path string) (out string) {
	req := realmpath.Parse(path)
	if req.Path == "" {
		out += md.H1("Leon's configs & snippets")

		out += ufmt.Sprintf("Leon's main address: %s\n\n", OwnableMain.Owner().String())
		out += ufmt.Sprintf("Leon's backup address: %s\n\n", OwnableBackup.Owner().String())

		out += md.H2("Snippets")

		if configs.Size() == 0 {
			out += "No configs yet :c\n\n"
		} else {
			page := pager.MustGetPageByPath(path)
			for _, item := range page.Items {
				out += ufmt.Sprintf("- [%s](%s:%s)\n\n", item.Value.(Config).name, absPath, item.Key)
			}

			out += page.Picker(path)
			out += "\n\n"
			out += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"
		}

		out += Banner()

		return out
	}

	return renderConfPage(req.Path)
}

func renderConfPage(id string) (out string) {
	raw, ok := configs.Get(id)
	if !ok {
		out += md.H1("404")
		out += "That config does not exist :/"
		return out
	}

	conf := raw.(Config)
	out += md.H1(conf.name)
	out += ufmt.Sprintf("```\n%s\n```\n\n", conf.lines)
	out += ufmt.Sprintf("_Last updated on %s_\n\n", conf.updated.Format("02 Jan, 2006"))
	out += md.HorizontalRule()
	out += ufmt.Sprintf("[[EDIT]](%s) - [[DELETE]](%s)", txlink.Call("EditConfig", "id", conf.id.String()), txlink.Call("RemoveConfig", "id", conf.id.String()))

	return out
}
