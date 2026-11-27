package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// CommentMarker identifies the sticky PR comment so the publishing job updates
// it in place instead of adding one per push.
const CommentMarker = "<!-- gnoweb-pr-preview -->"

// safePkgPath is the only shape a realm or package path read back from
// preview.json may take. That file is written by the pull request's own code,
// so anything else is refused rather than escaped: no backtick, bracket, angle
// bracket, parenthesis or space ever reaches the comment's markdown.
var safePkgPath = regexp.MustCompile(`^gno\.land/[pr](/[A-Za-z0-9_][A-Za-z0-9_.-]*)+$`)

// TrustedPlan rebuilds the plan a comment is written from out of an untrusted
// preview.json and the snapshot directory it came with. The publishing job
// runs this from the default branch, so the comment it posts as the bot is
// built by trusted code: package paths are validated, and every file path
// and caption is derived again from the realm name and this binary's own
// tables instead of being taken from the file. A screenshot the snapshot
// does not hold is dropped rather than embedded as a broken image.
func TrustedPlan(in *Plan, snapshot string) (*Plan, error) {
	if in.Dropped < 0 {
		return nil, fmt.Errorf("dropped: refusing %d", in.Dropped)
	}
	p := &Plan{Gnoweb: in.Gnoweb, Dropped: in.Dropped}
	for _, l := range []struct {
		name string
		in   []string
		out  *[]string
	}{
		{"realms", in.Realms, &p.Realms},
		{"changed_realms", in.ChangedRealms, &p.ChangedRealms},
		{"changed_pkgs", in.ChangedPkgs, &p.ChangedPkgs},
		{"missed", in.Missed, &p.Missed},
	} {
		for _, v := range l.in {
			if !safePkgPath.MatchString(v) {
				return nil, fmt.Errorf("%s: refusing %q", l.name, v)
			}
			*l.out = append(*l.out, v)
		}
	}
	exists := func(rel string) bool {
		fi, err := os.Lstat(filepath.Join(snapshot, filepath.FromSlash(rel)))
		return err == nil && fi.Mode().IsRegular()
	}
	for _, sp := range shotPlan {
		f := path.Join(shotsDir, sp.name+".png")
		for _, s := range in.Shots {
			if s.File == f && exists(f) {
				p.Shots = append(p.Shots, Shot{File: f, Label: sp.label})
				break
			}
		}
	}
	for _, u := range in.Pairs {
		if len(p.Pairs) >= maxPairs {
			break
		}
		if !contains(p.ChangedRealms, u.Realm) || slices.ContainsFunc(p.Pairs, func(q ShotPair) bool { return q.Realm == u.Realm }) {
			continue
		}
		name := slug(strings.TrimPrefix(urlOf(u.Realm), "/"))
		pair := ShotPair{
			Realm: u.Realm,
			URL:   path.Dir(urlToFile(urlOf(u.Realm))) + "/",
			After: path.Join(shotsDir, name+"-after.png"),
			New:   u.New,
		}
		if !exists(pair.After) {
			continue
		}
		if before := path.Join(shotsDir, name+"-before.png"); u.Before != "" && exists(before) {
			pair.Before = before
		}
		p.Pairs = append(p.Pairs, pair)
	}
	return p, nil
}

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
	// A realm with no page in the snapshot is named without a link: the
	// alternative is a bullet pointing at a 404. Either the crawl never
	// captured it, or the realm cap cut it before the crawl started, which
	// past 25 changed realms includes changed ones.
	missed := map[string]bool{}
	for _, r := range p.Missed {
		missed[r] = true
	}
	bullet := func(r string) string {
		switch {
		case missed[r]:
			return "`" + r + "` (not rendered: the page did not load)"
		case !contains(p.Realms, r):
			return "`" + r + "` (not rendered: realm cap reached)"
		}
		return link(urlOf(r), r) + tabs(base, r)
	}

	switch p.Mode() {
	case "gnoweb":
		b.WriteString("This PR changes **gnoweb itself**, so the preview is a sample of pages rendered with it:\n\n")
		if base != "" {
			b.WriteString(fmt.Sprintf("**[Open the preview homepage](%s/)**\n\n", base))
		}
		b.WriteString(shotGrid(p.Shots, base))
		for _, r := range p.Realms {
			b.WriteString("- " + bullet(r) + "\n")
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
		b.WriteString(pairGrid(p.Pairs, base))
		direct := map[string]bool{}
		for _, r := range p.ChangedRealms {
			direct[r] = true
		}
		if len(p.ChangedRealms) > 0 {
			b.WriteString(fmt.Sprintf("**Changed realms (%d)**\n\n", len(p.ChangedRealms)))
			for _, r := range p.ChangedRealms {
				b.WriteString("- " + bullet(r) + "\n")
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
				b.WriteString("- " + bullet(r) + "\n")
			}
		}
	}

	if p.Dropped > 0 {
		b.WriteString(fmt.Sprintf("\n⚠️ %d affected realm(s) were **not** rendered (cap reached). Changed realms are rendered first.\n", p.Dropped))
	}
	b.WriteString("\n<sub>Static snapshot of a fresh chain: realms render their post-`init` state, transactions and search do not work, and links out of the preview go to the live site.")
	if pr != "" {
		b.WriteString(fmt.Sprintf(" Rebuilt on every push to this PR; removed when PR #%s closes.", pr))
	}
	b.WriteString("</sub>\n")
	return b.String()
}

// pairGrid shows each changed realm as the base branch renders it and as this
// branch renders it, side by side. Both columns come from the same gnoweb, so
// the difference is the realm change.
func pairGrid(pairs []ShotPair, base string) string {
	if len(pairs) == 0 || base == "" {
		return ""
	}
	var b strings.Builder
	for _, p := range pairs {
		b.WriteString(fmt.Sprintf("**`%s`**\n\n", p.Realm))
		if p.Before == "" {
			note := ""
			if p.New {
				note = "\n\n<sub>New in this PR — nothing to compare against.</sub>"
			}
			b.WriteString(fmt.Sprintf(
				`<a href="%s/%s"><img src="%s/%s" width="600" alt="%s"></a>`+"%s\n\n",
				base, p.URL, base, p.After, p.Realm, note))
			continue
		}
		b.WriteString("<table><tr>")
		b.WriteString(fmt.Sprintf(
			`<td width="50%%"><img src="%s/%s" width="100%%" alt="%s before"><br><sub>before — base branch</sub></td>`,
			base, p.Before, p.Realm))
		b.WriteString(fmt.Sprintf(
			`<td width="50%%"><a href="%s/%s"><img src="%s/%s" width="100%%" alt="%s after"></a><br><sub><b>after — this PR</b></sub></td>`,
			base, p.URL, base, p.After, p.Realm))
		b.WriteString("</tr></table>\n\n")
	}
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
	n := 0
	for _, r := range p.Realms {
		u := urlOf(r)
		if _, ok := c.pages[u]; !ok {
			continue
		}
		n++
		rows.WriteString(fmt.Sprintf(
			`    <li><a href="%s/"><code>gno.land%s</code></a> <a href="%s/_t/source/">source</a> <a href="%s/_t/help/">help</a></li>`+"\n",
			strings.TrimPrefix(path.Dir(urlToFile(u)), "/"), u,
			strings.TrimPrefix(path.Dir(urlToFile(u)), "/"), strings.TrimPrefix(path.Dir(urlToFile(u)), "/")))
	}
	return fmt.Sprintf(`<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="robots" content="noindex, nofollow">
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
`, n, rows.String())
}
