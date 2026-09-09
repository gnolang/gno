package impl

import (
	"chain/runtime"
	"strconv"
	"strings"

	"gno.land/p/moul/helplink"
	"gno.land/p/moul/md"
	"gno.land/p/nt/bptree/v0/pager"
	"gno.land/p/nt/markdown/sanitize/v0"
	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/sys/users"
)

type render struct {
	relativeRealmPath string
	router            *mux.Router
	pssPager          *pager.Pager
}

func NewRender(d *GovDAO) *render {
	ren := &render{
		pssPager: pager.NewPager(d.pss.BPTree, 5, true),
	}

	r := mux.NewRouter()

	// Handlers use mux's rlm-aware shape: rlm is supplied at RenderRlm
	// dispatch time rather than captured at NewRender time. This lets
	// downstream crossing reads (dao.GetProposal etc.) use cross(rlm)
	// without relying on bare cross or restructuring the router.
	r.HandleFuncRlm("", func(_ int, rlm realm, rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderActiveProposals(0, rlm, req.RawPath, d))
	})

	r.HandleFuncRlm("{pid}", func(_ int, rlm realm, rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderProposalPage(0, rlm, req.GetVar("pid"), d))
	})

	r.HandleFuncRlm("{pid}/votes", func(_ int, rlm realm, rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderVotesForProposal(0, rlm, req.GetVar("pid"), d))
	})

	ren.router = r

	return ren
}

func (ren *render) Render(_ int, rlm realm, pkgPath string, path string) string {
	relativePath, found := strings.CutPrefix(pkgPath, runtime.ChainDomain())
	if !found {
		panic(ufmt.Sprintf(
			"realm package with unexpected name found: %v in chain domain %v",
			pkgPath, runtime.ChainDomain()))
	}
	ren.relativeRealmPath = relativePath
	return ren.router.RenderRlm(0, rlm, path)
}

func (ren *render) renderActiveProposals(_ int, rlm realm, url string, d *GovDAO) string {
	out := "# GovDAO\n"
	out += "## Members\n"
	out += "[> Go to Memberstore <](/r/gov/dao/v3/memberstore)\n"
	out += "## Proposals\n"
	page, perr := ren.pssPager.GetPageByPath(url)
	if perr != nil {
		// A query url.Parse rejects (e.g. a control byte) is the caller's own
		// malformed input; render the default first page instead of aborting
		// this read-only render. ParseQuery already tolerates bad page/size
		// values, so every currently-working input is unaffected.
		page = ren.pssPager.GetPage(1)
	}
	if len(page.Items) == 0 {
		out += "\nNo proposals yet.\n\n"
		return out
	}

	for _, item := range page.Items {
		seqpid, err := seqid.FromString(item.Key)
		if err != nil {
			continue
		}
		out += ren.renderProposalListItem(0, rlm, ufmt.Sprintf("%v", int64(seqpid)), d)
		out += "---\n\n"
	}

	out += page.Picker("")

	return out
}

func (ren *render) renderProposalPage(_ int, rlm realm, sPid string, d *GovDAO) string {
	pid, err := strconv.ParseInt(sPid, 10, 64)
	if err != nil {
		// err echoes the caller's raw pid segment (strconv quotes it but leaves
		// markdown/HTML metachars); escape+clamp it like every other untrusted
		// slot on this page, since Render is reachable unauthenticated.
		return ufmt.Sprintf("# Error: Invalid proposal ID format.\n\n\n%s\n\n", md.EscapeText(clampField(err.Error(), maxRenderedError)))
	}

	p, err := dao.GetProposal(dao.ProposalID(pid))
	if err != nil {
		return ufmt.Sprintf("# Proposal not found\n\n%s", err.Error())
	}

	// pid is user-supplied and the proposal itself lives on the proxy, so
	// GetProposal can succeed for a proposal this implementation has no status
	// for — the same post-upgrade case PreExecuteProposal guards (statuses live
	// on the instance, so a fresh instance has none). Without this the page
	// panics at ps.statusLine below. renderProposalListItem needs no such guard:
	// its pids are iterated out of pss itself, so a status always exists.
	ps := d.pss.GetStatus(dao.ProposalID(pid))
	if ps == nil {
		return ufmt.Sprintf("# Proposal #%v not available\n\nThis proposal was not created by the current govDAO implementation, so it has no voting status here.", pid)
	}

	// The trailing backslashes are markdown hard breaks: gnoweb folds a plain
	// newline into a space, and the block would render as one paragraph.
	out := ufmt.Sprintf("## Prop #%v - %v\n", pid, md.EscapeText(clampField(p.Title(), maxRenderedTitle)))
	out += "Author: " + tryResolveAddr(p.Author()) + "\\\n"
	out += "Executes: " + executesLine(p) + "\\\n"
	out += ps.statusLine() + "\\\n"
	out += "Eligible: " + strings.Join(ps.TiersAllowedToVote, ", ") + "\\\n"
	out += ps.tally(0, rlm)
	out += ufmt.Sprintf(" ([detail](%v:%v/votes))\n\n", ren.relativeRealmPath, pid)

	// The page draws every rule on it, and escapeThematicBreaks takes the mark
	// out of the description, so an imitation can only render below this one.
	out += renderExecution(p)
	out += "---\n\n"

	if description := escapeThematicBreaks(p.Description()); description != "" {
		out += "### Description\n\n"
		out += description
		out += "\n\n---\n\n"
	}

	out += renderActionBar(ufmt.Sprintf("%v", pid))

	return out
}

// escapeThematicBreaks backslashes every line of the description that would
// render as a horizontal rule, leaving the rule to the page. Fenced blocks are
// skipped, where a rule is already literal text.
func escapeThematicBreaks(description string) string {
	if !strings.ContainsAny(description, "-_*") {
		return description
	}

	lines := strings.Split(description, "\n")
	fenced := false
	for i, line := range lines {
		// A CRLF description leaves a trailing carriage return on every line,
		// which gnoweb drops before it reads the line.
		trimmed := strings.TrimSuffix(line, "\r")
		if isCodeFence(trimmed) {
			fenced = !fenced
			continue
		}
		if !fenced && isThematicBreak(trimmed) {
			lines[i] = "\\" + line
		}
	}

	return strings.Join(lines, "\n")
}

// isCodeFence reports whether the line opens or closes a fenced code block.
func isCodeFence(line string) bool {
	line = trimIndent(line)
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

// isThematicBreak reports whether the line renders as a horizontal rule, by
// the CommonMark rule. A setext underline is the same run of dashes.
func isThematicBreak(line string) bool {
	line = trimIndent(line)
	if line == "" {
		return false
	}

	mark := line[0]
	if mark != '-' && mark != '_' && mark != '*' {
		return false
	}

	count := 0
	for i := 0; i < len(line); i++ {
		switch c := line[i]; {
		case c == mark:
			count++
		case c == ' ' || c == '\t':
		default:
			return false
		}
	}

	return count >= 3
}

// trimIndent drops the up-to-three leading spaces a block marker may carry. A
// fourth space makes the line indented code, which renders literally.
func trimIndent(line string) string {
	for i := 0; i < 3 && i < len(line); i++ {
		if line[i] != ' ' {
			return line[i:]
		}
	}
	if len(line) >= 3 {
		return line[3:]
	}

	return ""
}

// executesLine names the realm that runs if the proposal passes. The line is
// never dropped: silence reads as a shorter page, not as a missing disclosure.
func executesLine(p *dao.Proposal) string {
	if creationRealm := renderCreationRealm(p.ExecutorCreationRealm()); creationRealm != "" {
		return creationRealm
	}

	return "none reported"
}

// renderExecution returns the executor's own account of what runs, under a
// heading of its own, and "" for an executor with nothing to say.
func renderExecution(p *dao.Proposal) string {
	description := p.ExecutorString()
	if description == "" {
		return ""
	}

	return "### Execution\n\n" + description + "\n\n"
}

// renderCreationRealm renders the realm an executor reports it was created in:
// the escaped code span it has always rendered as, linked to that realm's
// source when the value parses as a package path this chain serves. Any
// executor returns any string it likes, so the escaping is the guard here.
func renderCreationRealm(creationRealm string) string {
	// Clamp before escaping, per clamp.gno: an escaped value cut afterwards can
	// lose its closing fence.
	pkgPath := strings.TrimSpace(clampField(creationRealm, maxRenderedRealm))
	code := sanitize.InlineCode(pkgPath)
	if code == "" {
		return ""
	}

	source, ok := realmSourcePath(pkgPath)
	if !ok {
		return code
	}

	// The destination is a value realmSourcePath accepted, which carries no
	// backtick, bracket, parenthesis or colon to break out of the link.
	return "[" + code + "](" + source + ")"
}

// realmSourcePath maps a package path this chain serves to its gnoweb source
// path. Anything else renders unlinked rather than pointing at a 404.
func realmSourcePath(pkgPath string) (string, bool) {
	rel, found := strings.CutPrefix(pkgPath, runtime.ChainDomain())
	if !found {
		return "", false
	}

	// "/r/gov/dao/v3/impl": the letter directory, then one or more segments.
	if len(rel) < len("/r/x") || rel[0] != '/' || rel[2] != '/' ||
		(rel[1] != 'r' && rel[1] != 'p') {
		return "", false
	}
	for _, segment := range strings.Split(rel[3:], "/") {
		if !isPkgPathSegment(segment) {
			return "", false
		}
	}

	// The overview page, which is $source with no file selected.
	return rel + "$source", true
}

// isPkgPathSegment reports whether s is one segment of a package path, by the
// grammar the VM enforces on a deployed path: Re_name in
// gnovm/pkg/gnolang/mempackage.go.
func isPkgPathSegment(s string) bool {
	if s == "" || s[0] < 'a' || s[0] > 'z' {
		return false
	}

	separator := false
	for i := 1; i < len(s); i++ {
		switch c := s[i]; {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			separator = false
		case c == '_' || c == '-':
			if separator {
				return false
			}
			separator = true
		default:
			return false
		}
	}

	return !separator
}

func (ren *render) renderProposalListItem(_ int, rlm realm, sPid string, d *GovDAO) string {
	pid, err := strconv.ParseInt(sPid, 10, 64)
	if err != nil {
		// err echoes the caller's raw pid segment (strconv quotes it but leaves
		// markdown/HTML metachars); escape+clamp it like every other untrusted
		// slot on this page, since Render is reachable unauthenticated.
		return ufmt.Sprintf("# Error: Invalid proposal ID format.\n\n\n%s\n\n", md.EscapeText(clampField(err.Error(), maxRenderedError)))
	}

	p, err := dao.GetProposal(dao.ProposalID(pid))
	if err != nil {
		return ufmt.Sprintf("# Proposal not found\n\n%s\n\n", err.Error())
	}

	ps := d.pss.GetStatus(dao.ProposalID(pid))
	out := ufmt.Sprintf("### [Prop #%v - %v](%v:%v)\n", pid, md.EscapeText(clampField(p.Title(), maxRenderedTitle)), ren.relativeRealmPath, pid)
	out += ufmt.Sprintf("Author: %s\n\n", tryResolveAddr(p.Author()))

	out += "Status: " + getPropStatus(ps)
	out += "\n\n"

	out += "Tiers eligible to vote: "
	out += strings.Join(ps.TiersAllowedToVote, ", ")

	out += "\n\n"
	return out
}

func (ren *render) renderVotesForProposal(_ int, rlm realm, sPid string, d *GovDAO) string {
	pid, err := strconv.ParseInt(sPid, 10, 64)
	if err != nil {
		// err echoes the caller's raw pid segment (strconv quotes it but leaves
		// markdown/HTML metachars); escape+clamp it like every other untrusted
		// slot on this page, since Render is reachable unauthenticated.
		return ufmt.Sprintf("# Error: Invalid proposal ID format.\n\n\n%s\n\n", md.EscapeText(clampField(err.Error(), maxRenderedError)))
	}

	ps := d.pss.GetStatus(dao.ProposalID(pid))
	if ps == nil {
		return ufmt.Sprintf("# Proposal not found\n\nProposal %v does not exist.", pid)
	}

	out := ""
	out += ufmt.Sprintf("# Proposal #%v - Vote List\n\n", pid)
	out += StringifyVotes(0, rlm, ps)

	return out
}

func isPropActive(ps *proposalStatus) bool {
	return !ps.Accepted && !ps.Denied
}

func getPropStatus(ps *proposalStatus) string {
	if ps == nil {
		return "UNKNOWN"
	}
	if ps.Accepted {
		return "ACCEPTED"
	} else if ps.Denied {
		return "REJECTED"
	}
	return "ACTIVE"
}

func renderActionBar(sPid string) string {
	out := "### Actions\n"

	proxy := helplink.Realm("gno.land/r/gov/dao")
	out += proxy.Func("Vote YES", "MustVoteOnProposalSimple", "pid", sPid, "option", "YES") + " | "
	out += proxy.Func("Vote NO", "MustVoteOnProposalSimple", "pid", sPid, "option", "NO") + " | "
	out += proxy.Func("Vote ABSTAIN", "MustVoteOnProposalSimple", "pid", sPid, "option", "ABSTAIN")

	out += "\n\n"
	out += "WARNING: Please double check transaction data before voting."
	return out
}

// tryResolveAddr renders the author/voter as a username link. RenderLink
// interpolates the username raw into "[@name](/u/name)" (r/sys/users), so this
// is markdown-safe ONLY because r/sys/users validateName restricts names to
// ^[a-z][a-z0-9]*([_-][a-z0-9]+)*$ (max 64) — no markdown/HTML metachar. If that
// charset ever loosens, this line and writeVotes in types.gno need escaping.
func tryResolveAddr(addr address) string {
	userData := users.ResolveAddress(addr)
	if userData == nil {
		return addr.String()
	}
	return userData.RenderLink("")
}
