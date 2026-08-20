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
	// panics at ps.String below. renderProposalListItem needs no such guard:
	// its pids are iterated out of pss itself, so a status always exists.
	ps := d.pss.GetStatus(dao.ProposalID(pid))
	if ps == nil {
		return ufmt.Sprintf("# Proposal #%v not available\n\nThis proposal was not created by the current govDAO implementation, so it has no voting status here.", pid)
	}

	out := ufmt.Sprintf("## Prop #%v - %v\n", pid, md.EscapeText(clampField(p.Title(), maxRenderedTitle)))
	out += "Author: " + tryResolveAddr(p.Author()) + "\n\n"

	out += p.Description()
	out += "\n\n"

	// Add executor metadata if available
	if p.ExecutorString() != "" {
		out += ufmt.Sprintf(`This proposal contains the following metadata:

%s

`, p.ExecutorString())
	}

	// Disclosed independently of the description: this names the realm whose
	// code runs if the proposal passes. Sharing the gate above hid it for every
	// proposal built with an empty executor description — 16 call sites across
	// 7 production realms, nearly all r/sys/* governance actions.
	//
	// Escaped, not trusted: CreationRealm() is dispatched through the public
	// dao.Executor interface, so only SimpleExecutor's value is VM-captured;
	// a third-party executor returns any string it likes.
	// Guarded on the sanitized value, not the raw one: InlineCode returns ""
	// for input that strips to nothing (bidi/zero-width only), and TrimSpace
	// catches plain whitespace, so neither prints a label with nothing after
	// it. This is tidiness, not a boundary: a zero-width/space mixture still
	// renders an empty-looking span, and a hostile executor can always return
	// a plausible-looking lie. The security property is the escaping below.
	if cr := sanitize.InlineCode(strings.TrimSpace(clampField(p.ExecutorCreationRealm(), maxRenderedRealm))); cr != "" {
		out += ufmt.Sprintf("Executor created in: %s\n", cr)
		out += "\n\n"
	}

	out += "\n\n---\n\n"
	out += ps.String(0, rlm)
	out += "\n"
	out += ufmt.Sprintf("[Detailed voting list](%v:%v/votes)", ren.relativeRealmPath, pid)
	out += "\n\n---\n\n"

	out += renderActionBar(ufmt.Sprintf("%v", pid))

	return out
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
