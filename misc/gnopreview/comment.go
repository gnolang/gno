package main

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// CommentMarker identifies the sticky PR comment so the publishing job updates
// it in place instead of adding one per push.
const CommentMarker = "<!-- gnoweb-pr-preview -->"

// Comment renders the sticky PR comment body. baseURL is where the snapshot is
// served from (with or without a trailing slash); when it is empty the comment
// still lists what was rendered, just without links.
func Comment(p *Plan, baseURL, pr string) string {
	var b strings.Builder
	b.WriteString(CommentMarker + "\n")
	b.WriteString("### 🖼️ gnoweb preview\n\n")

	// Nothing to preview means no comment: Comment is not called in that case,
	// and an empty string here keeps a stray call from posting a useless one.
	if p.Empty() {
		return ""
	}

	base := strings.TrimSuffix(baseURL, "/")
	link := func(urlPath, label string) string {
		if base == "" {
			return "`" + label + "`"
		}
		return fmt.Sprintf("[`%s`](%s%s/)", label, base, urlPath)
	}

	switch p.Mode() {
	case "gnoweb":
		b.WriteString("This PR changes **gnoweb itself**, so the preview is a sample of pages rendered with it:\n\n")
		if base != "" {
			b.WriteString(fmt.Sprintf("**[Open the preview homepage](%s/)**\n\n", base))
		}
		b.WriteString(shotGrid(p.Shots, base))
		for _, r := range p.Realms {
			b.WriteString("- " + link(urlOf(r), r) + "\n")
		}
	default:
		if p.Gnoweb {
			b.WriteString("This PR changes **gnoweb** and realm sources. ")
			if base != "" {
				b.WriteString(fmt.Sprintf("**[Open the preview homepage](%s/)**\n", base))
			}
			b.WriteString("\n")
			b.WriteString(shotGrid(p.Shots, base))
		}
		direct := map[string]bool{}
		for _, r := range p.ChangedRealms {
			direct[r] = true
		}
		if len(p.ChangedRealms) > 0 {
			b.WriteString(fmt.Sprintf("**Changed realms (%d)**\n\n", len(p.ChangedRealms)))
			for _, r := range p.ChangedRealms {
				b.WriteString("- " + link(urlOf(r), r) + tabs(base, r) + "\n")
			}
			b.WriteString("\n")
		}
		var indirect []string
		for _, r := range p.Realms {
			if !direct[r] && !isSeed(r, p) {
				indirect = append(indirect, r)
			}
		}
		if len(indirect) > 0 {
			sort.Strings(indirect)
			b.WriteString(fmt.Sprintf("**Realms affected through a changed package (%d)**\n\n", len(indirect)))
			if len(p.ChangedPkgs) > 0 {
				b.WriteString("_changed: " + "`" + strings.Join(p.ChangedPkgs, "`, `") + "`_\n\n")
			}
			for _, r := range indirect {
				b.WriteString("- " + link(urlOf(r), r) + tabs(base, r) + "\n")
			}
		}
	}

	if p.Dropped > 0 {
		b.WriteString(fmt.Sprintf("\n⚠️ %d more affected realm(s) were **not** rendered (cap reached) — the changed realms are always kept.\n", p.Dropped))
	}
	b.WriteString("\n<sub>Static snapshot of a fresh chain: realms render their post-`init` state, transactions and search do not work, and links out of the preview go to the live site.")
	if pr != "" {
		b.WriteString(fmt.Sprintf(" Rebuilt on every push to this PR; removed when PR #%s closes.", pr))
	}
	b.WriteString("</sub>\n")
	return b.String()
}

// shotGrid embeds the screenshots two per row. They are served from the
// published snapshot itself, so nothing has to be uploaded anywhere.
func shotGrid(shots []Shot, base string) string {
	if len(shots) == 0 || base == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("<table><tr>")
	for i, s := range shots {
		if i > 0 && i%2 == 0 {
			b.WriteString("</tr><tr>")
		}
		b.WriteString(fmt.Sprintf(
			`<td width="50%%"><a href="%s/%s"><img src="%s/%s" width="100%%" alt="%s"></a><br><sub>%s</sub></td>`,
			base, s.File, base, s.File, s.Label, s.Label))
	}
	b.WriteString("</tr></table>\n\n")
	return b.String()
}

// tabs adds the source/help shortcuts next to a realm link.
func tabs(base, pkgPath string) string {
	if base == "" {
		return ""
	}
	u := urlOf(pkgPath)
	return fmt.Sprintf(" · [source](%s%s/_t/source/) · [help](%s%s/_t/help/)", base, u, base, u)
}

func isSeed(pkgPath string, p *Plan) bool {
	if !p.Gnoweb {
		return false
	}
	return contains(gnowebSeedRealms, pkgPath)
}

// Index is the landing page of the snapshot: every captured render page, so a
// reviewer who opens the root URL still finds their way around.
func Index(p *Plan, c *Crawler) string {
	var rows strings.Builder
	for _, r := range p.Realms {
		u := urlOf(r)
		if _, ok := c.pages[u]; !ok {
			continue
		}
		rows.WriteString(fmt.Sprintf(
			`    <li><a href="%s/"><code>gno.land%s</code></a> <a href="%s/_t/source/">source</a> <a href="%s/_t/help/">help</a></li>`+"\n",
			strings.TrimPrefix(path.Dir(urlToFile(u)), "/"), u,
			strings.TrimPrefix(path.Dir(urlToFile(u)), "/"), strings.TrimPrefix(path.Dir(urlToFile(u)), "/")))
	}
	return fmt.Sprintf(`<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="robots" content="noindex">
<title>gnoweb preview</title>
<style>
 :root{color-scheme:light dark}
 body{font-family:system-ui,sans-serif;max-width:52rem;margin:3rem auto;padding:0 1rem;line-height:1.5}
 code{font-family:ui-monospace,monospace}
 li{margin:.4em 0}
 a+a{font-size:.85em;margin-left:.5em;opacity:.7}
</style>
<h1>gnoweb preview</h1>
<p>%d realm(s) rendered with gnodev on a fresh chain. Static snapshot — transactions,
search and links outside the preview do not work.</p>
<ul>
%s</ul>
`, len(p.Realms), rows.String())
}
