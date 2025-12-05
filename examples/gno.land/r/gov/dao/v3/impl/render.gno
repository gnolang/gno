package impl

import (
	"chain/runtime"
	"strconv"
	"strings"

	"gno.land/p/moul/helplink"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	"gno.land/r/sys/users"
)

type render struct {
	relativeRealmPath string
	router            *mux.Router
	pssPager          *pager.Pager
}

func NewRender(d *GovDAO) *render {
	ren := &render{
		pssPager: pager.NewPager(d.pss.Tree, 5, true),
	}

	r := mux.NewRouter()

	r.HandleFunc("", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderActiveProposals(req.RawPath, d))
	})

	r.HandleFunc("activity", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderActivityPage(req.RawPath, d))
	})

	r.HandleFunc("{pid}", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderProposalPage(req.GetVar("pid"), d))
	})

	r.HandleFunc("{pid}/votes", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderVotesForProposal(req.GetVar("pid"), d))
	})

	ren.router = r

	return ren
}

func (ren *render) Render(pkgPath string, path string) string {
	relativePath, found := strings.CutPrefix(pkgPath, runtime.ChainDomain())
	if !found {
		panic(ufmt.Sprintf(
			"realm package with unexpected name found: %v in chain domain %v",
			pkgPath, runtime.ChainDomain()))
	}
	ren.relativeRealmPath = relativePath
	return ren.router.Render(path)
}

func (ren *render) renderActiveProposals(url string, d *GovDAO) string {
	out := "# GovDAO\n"
	out += "[> View Activity <](" + ren.relativeRealmPath + ":activity)\n\n"
	out += "## Members\n"
	out += "[> Go to Memberstore <](/r/gov/dao/v3/memberstore)\n"
	out += "## Proposals\n"
	page := ren.pssPager.MustGetPageByPath(url)
	if len(page.Items) == 0 {
		out += "\nNo proposals yet.\n\n"
		return out
	}

	for _, item := range page.Items {
		seqpid, err := seqid.FromString(item.Key)
		if err != nil {
			continue
		}
		out += ren.renderProposalListItem(ufmt.Sprintf("%v", int64(seqpid)), d)
		out += "---\n\n"
	}

	out += page.Picker("")

	return out
}

func (ren *render) renderProposalPage(sPid string, d *GovDAO) string {
	pid, err := strconv.ParseInt(sPid, 10, 64)
	if err != nil {
		return ufmt.Sprintf("# Error: Invalid proposal ID format.\n\n\n%s\n\n", err.Error())
	}

	p, err := dao.GetProposal(cross, dao.ProposalID(pid))
	if err != nil {
		return ufmt.Sprintf("# Proposal not found\n\n%s", err.Error())
	}

	ps := d.pss.GetStatus(dao.ProposalID(pid))
	out := ufmt.Sprintf("## Prop #%v - %v\n", pid, p.Title())
	out += "Author: " + tryResolveAddr(p.Author()) + "\n\n"

	out += p.Description()
	out += "\n\n"

	// Add executor metadata if available
	if p.ExecutorString() != "" {
		out += ufmt.Sprintf(`This proposal contains the following metadata:

%s

Executor created in: %s
`, p.ExecutorString(), p.ExecutorCreationRealm())
		out += "\n\n"
	}

	out += "\n\n---\n\n"
	out += ps.String()
	out += "\n"
	out += ufmt.Sprintf("[Detailed voting list](%v:%v/votes)", ren.relativeRealmPath, pid)
	out += "\n\n---\n\n"

	out += renderActionBar(ufmt.Sprintf("%v", pid))

	return out
}

type memberActivity struct {
	addr      address
	tier      string
	available uint32
	votes     uint32
}

func (ren *render) renderActivityPage(path string, d *GovDAO) string {
	out := "# GovDAO Activity\n\n"
	out += "## Members Activity\n\n"

	tiers := memberstore.Get()
	if tiers.Size() == 0 {
		out += "### No members available\n"
		return out
	}

	var activities []memberActivity
	tiers.Iterate("", "", func(tierName string, tierValue any) bool {
		memberTree, ok := tierValue.(*avl.Tree)
		if !ok {
			return false
		}

		memberTree.Iterate("", "", func(addrKey string, _ any) bool {
			memberAddr := address(addrKey)

			activity := GetMemberActivity(memberAddr)

			activities = append(activities, memberActivity{
				addr:      memberAddr,
				tier:      tierName,
				available: activity.Available,
				votes:     activity.Votes,
			})

			return false
		})
		return false
	})

	activityTree := avl.NewTree()
	var index int
	for _, ma := range activities {
		activityTree.Set(ufmt.Sprintf("%05d", index), ma)
		index++
	}

	activityPager := pager.NewPager(activityTree, 15, false)
	page := activityPager.MustGetPageByPath(path)

	for _, item := range page.Items {
		ma, ok := item.Value.(memberActivity)
		if !ok {
			continue
		}

		percentage := float64(0)
		if ma.available > 0 {
			percentage = (float64(ma.votes) / float64(ma.available)) * 100.0
		}

		out += ufmt.Sprintf("**%s** (%s): %d/%d votes (%.2f%%)\n\n",
			tryResolveAddr(ma.addr), ma.tier, ma.votes, ma.available, percentage)
	}

	out += "\n\n" + page.Picker("") + "\n\n"

	startIdx := (page.PageNumber-1)*page.PageSize + 1
	endIdx := startIdx + len(page.Items) - 1
	out += ufmt.Sprintf("_Showing %d-%d of %d members_\n\n",
		startIdx, endIdx, len(activities))

	return out
}

func (ren *render) renderProposalListItem(sPid string, d *GovDAO) string {
	pid, err := strconv.ParseInt(sPid, 10, 64)
	if err != nil {
		return ufmt.Sprintf("# Error: Invalid proposal ID format.\n\n\n%s\n\n", err.Error())
	}

	p, err := dao.GetProposal(cross, dao.ProposalID(pid))
	if err != nil {
		return ufmt.Sprintf("# Proposal not found\n\n%s\n\n", err.Error())
	}

	ps := d.pss.GetStatus(dao.ProposalID(pid))
	out := ufmt.Sprintf("### [Prop #%v - %v](%v:%v)\n", pid, p.Title(), ren.relativeRealmPath, pid)
	out += ufmt.Sprintf("Author: %s\n\n", tryResolveAddr(p.Author()))

	out += "Status: " + getPropStatus(ps)
	out += "\n\n"

	out += "Tiers eligible to vote: "
	out += strings.Join(ps.TiersAllowedToVote, ", ")

	out += "\n\n"
	return out
}

func (ren *render) renderVotesForProposal(sPid string, d *GovDAO) string {
	pid, err := strconv.ParseInt(sPid, 10, 64)
	if err != nil {
		return ufmt.Sprintf("# Error: Invalid proposal ID format.\n\n\n%s\n\n", err.Error())
	}

	ps := d.pss.GetStatus(dao.ProposalID(pid))
	if ps == nil {
		return ufmt.Sprintf("# Proposal not found\n\nProposal %v does not exist.", pid)
	}

	out := ""
	out += ufmt.Sprintf("# Proposal #%v - Vote List\n\n", pid)
	out += StringifyVotes(ps)

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

func tryResolveAddr(addr address) string {
	userData := users.ResolveAddress(addr)
	if userData == nil {
		return addr.String()
	}
	return userData.RenderLink("")
}
