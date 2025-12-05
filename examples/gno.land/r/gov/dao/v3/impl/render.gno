package impl

import (
	"chain/runtime"
	"sort"
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
		rw.Write(ren.renderActivityPage(d, 10))
	})

	r.HandleFunc("activity/{count}", func(rw *mux.ResponseWriter, req *mux.Request) {
		count, err := strconv.Atoi(req.GetVar("count"))
		if err != nil || count <= 0 {
			count = 10
		}
		if count > 20 {
			count = 20
		}
		rw.Write(ren.renderActivityPage(d, count))
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
	addr       address
	tier       string
	activity   uint32
	available  uint32
	percentage float64
}

// memberActivitySorter implements sort.Interface for sorting member activities by participation percentage
type memberActivitySorter []memberActivity

func (mas memberActivitySorter) Len() int {
	return len(mas)
}

func (mas memberActivitySorter) Less(i, j int) bool {
	if mas[i].percentage != mas[j].percentage {
		return mas[i].percentage > mas[j].percentage
	}
	return mas[i].activity > mas[j].activity
}

func (mas memberActivitySorter) Swap(i, j int) {
	mas[i], mas[j] = mas[j], mas[i]
}

// getRecentProposalIDs returns the most recent proposal IDs, up to maxCount
func (d *GovDAO) getRecentProposalIDs(maxCount int) []dao.ProposalID {
	var proposalIDs []dao.ProposalID

	// Iterate in reverse order to get most recent first
	d.pss.ReverseIterate("", "", func(id string, _ interface{}) bool {
		seqpid, err := seqid.FromString(id)
		if err != nil {
			return false
		}
		proposalIDs = append(proposalIDs, dao.ProposalID(seqpid))

		if len(proposalIDs) >= maxCount {
			return true
		}
		return false
	})

	return proposalIDs
}

func (ren *render) renderActivityPage(d *GovDAO, recentProposalsCount int) string {
	out := "# GovDAO Activity\n\n"

	out += "### Configuration\n"
	out += "**Proposals to analyze:** [5](" + ren.relativeRealmPath + ":activity/5) | "
	out += "[10](" + ren.relativeRealmPath + ":activity/10) | "
	out += "[20](" + ren.relativeRealmPath + ":activity/20)\n\n"
	out += "---\n\n"
	out += ufmt.Sprintf("## Members Activity (Last %d Proposals)\n\n", recentProposalsCount)

	tiers := memberstore.Get()
	if tiers.Size() == 0 {
		out += "### No members available\n"
		return out
	}

	recentProposalIDs := d.getRecentProposalIDs(recentProposalsCount)
	if len(recentProposalIDs) == 0 {
		out += "### No proposals available\n"
		return out
	}

	memberVotes := make(map[string]uint32) // address -> vote count

	for _, pid := range recentProposalIDs {
		ps := d.pss.GetStatus(pid)
		if ps == nil {
			continue
		}
		ps.AllVotes.Iterate("", "", func(tierName string, tierValue any) bool {
			memberTree, ok := tierValue.(*avl.Tree)
			if !ok {
				return false
			}

			memberTree.Iterate("", "", func(addrKey string, _ any) bool {
				memberVotes[addrKey]++
				return false
			})
			return false
		})
	}

	var activities []memberActivity
	tiers.Iterate("", "", func(tierName string, tierValue any) bool {
		memberTree, ok := tierValue.(*avl.Tree)
		if !ok {
			return false
		}

		memberTree.Iterate("", "", func(addrKey string, _ any) bool {
			memberAddr := address(addrKey)
			member, _ := tiers.GetMember(memberAddr)
			if member == nil {
				return false
			}

			var available uint32
			for _, pid := range recentProposalIDs {
				ps := d.pss.GetStatus(pid)
				if ps != nil && !member.JoinTime.After(ps.CreatedAt) {
					available++
				}
			}

			if available == 0 {
				return false
			}

			activity := memberVotes[addrKey]
			var percentage float64
			if available > 0 {
				percentage = float64(activity) / float64(available) * 100.0
			}

			activities = append(activities, memberActivity{
				addr:       memberAddr,
				tier:       tierName,
				activity:   activity,
				available:  available,
				percentage: percentage,
			})

			return false
		})
		return false
	})

	sort.Sort(memberActivitySorter(activities))

	for _, ma := range activities {
		if ma.available > 0 {
			out += ufmt.Sprintf("**%s** (%s): %d/%d (%.1f%%)\n\n",
				tryResolveAddr(ma.addr), ma.tier, ma.activity, ma.available, ma.percentage)
		} else {
			out += ufmt.Sprintf("**%s** (%s): %d/%d\n\n",
				tryResolveAddr(ma.addr), ma.tier, ma.activity, ma.available)
		}
	}

	out += ufmt.Sprintf("_Total: %d members_\n\n", len(activities))

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
