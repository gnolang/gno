package debug

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/umd"
)

type Debug struct {
	Logs         []string
	HideMetadata bool
}

func (d *Debug) Log(line string) {
	d.Logs = append(d.Logs, line)
}

func (d Debug) Render(path string) string {
	if realmpath.Parse(path).Query.Get("debug") != "1" {
		return ""
	}

	var content string

	if d.Logs != nil {
		content += umd.H3("Logs")
		content += umd.BulletList(d.Logs)
	}

	if !d.HideMetadata {
		content += umd.H3("Metadata")
		table := mdtable.Table{
			Headers: []string{"Key", "Value"},
		}
		table.Append([]string{"`std.CurrentRealm().PkgPath()`", string(std.CurrentRealm().PkgPath())})
		table.Append([]string{"`std.CurrentRealm().Addr()`", string(std.CurrentRealm().Addr())})
		table.Append([]string{"`std.GetHeight()`", ufmt.Sprintf("%d", std.GetHeight())})
		content += table.String()
	}

	if content == "" {
		return ""
	}

	return umd.CollapsibleSection("debug", content)
}

// top-level Render returns a default Debug struct's Render.
func Render(path string) string {
	return Debug{}.Render(path)
}
