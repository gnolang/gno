package releases

import (
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/ufmt/v0"
)

type changelog struct {
	name     string
	releases []release
}

func NewChangelog(name string) *changelog {
	return &changelog{
		name:     name,
		releases: make([]release, 0),
	}
}

func (c *changelog) NewRelease(version, url, notes string) {
	if latest := c.Latest(); latest != nil {
		latest.isLatest = false
	}

	release := release{
		// manual
		version: version,
		url:     url,
		notes:   notes,

		// internal
		changelog: c,
		isLatest:  true,
	}

	c.releases = append(c.releases, release)
}

func (c changelog) RenderAsTable(max int) string {
	out := md.H1(c.name)

	if len(c.releases) == 0 {
		out += "No releases."
		return out
	}

	out += ufmt.Sprintf("See the %s changelog below.\n\n", c.name)
	table := mdtable.Table{
		Headers: []string{"Version", "Link", "Notes"},
	}

	if max > len(c.releases) {
		max = len(c.releases)
	}

	for i := len(c.releases) - 1; i >= len(c.releases)-max; i-- {
		r := c.releases[i]
		table.Append([]string{r.version, r.Link(), r.notes})
	}

	out += table.String()
	return out
}

func (c *changelog) Render(path string) string {
	if path == "" {
		output := "# " + c.name + "\n\n"
		max := len(c.releases) - 1
		min := 0
		if max-min > 10 {
			min = max - 10
		}
		for i := max; i >= min; i-- {
			release := c.releases[i]
			output += release.Render()
		}
		return output
	}

	release := c.ByVersion(path)
	if release != nil {
		return release.Render()
	}

	return "no such release"
}

func (c *changelog) Latest() *release {
	if len(c.releases) > 0 {
		pos := len(c.releases) - 1
		return &c.releases[pos]
	}
	return nil
}

func (c *changelog) ByVersion(version string) *release {
	for _, release := range c.releases {
		if release.version == version {
			return &release
		}
	}
	return nil
}
