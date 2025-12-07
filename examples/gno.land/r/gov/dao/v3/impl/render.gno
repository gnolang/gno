package impl

import (
	"chain/runtime"
	"strconv"
	"strings"

	"gno.land/p/moul/helplink"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
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
		pssPager: pager.NewPager(d.pss.Tree, 5, true),
	}

	r := mux.NewRouter()

	r.HandleFunc("", func(rw *mux.ResponseWriter, req *mux.Request) {
		rw.Write(ren.renderActiveProposals(req.RawPath, d))
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
	out := "# 🏛️ Gno.land Governance DAO\n\n"

	// Quick stats
	page := ren.pssPager.MustGetPageByPath(url)
	totalProposals := d.pss.Size()
	activeCount := countActiveProposals(d)

	out += ufmt.Sprintf("**%d** proposals total • **%d** active • **%d** completed", totalProposals, activeCount, totalProposals-activeCount)
	out += " | [👥 Members](/r/gov/dao/v3/memberstore)\n\n"

	out += "## 📜 Proposals\n\n---\n\n"

	if len(page.Items) == 0 {
		out += "_No proposals to display._\n\n"
	} else {
		for _, item := range page.Items {
			seqpid, err := seqid.FromString(item.Key)
			if err != nil {
				panic(err.Error())
			}
			out += ren.renderProposalListItem(ufmt.Sprintf("%v", int64(seqpid)), d)
			out += "\n"
		}
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

	statusIcon := getStatusIcon(ps)
	statusBadge := getStatusBadge(ps)

	// Header
	out := ufmt.Sprintf("# %s Proposal #%v\n\n", statusIcon, pid)
	out += ufmt.Sprintf("## %s\n\n", p.Title())
	out += ufmt.Sprintf("> %s\n\n", statusBadge)

	out += "---\n\n"

	// Metadata in a clean layout
	out += "| | |\n"
	out += "|---|---|\n"
	out += "| **👤 Author** | " + tryResolveAddr(p.Author()) + " |\n"
	out += "| **📊 Status** | " + ufmt.Sprintf("%s %s", statusIcon, getPropStatus(ps))
	if ps.DeniedReason != "" {
		out += " _(Reason: " + ps.DeniedReason + ")_"
	}
	out += " |\n"
	out += "| **🎯 Eligible Tiers** | " + strings.Join(ps.TiersAllowedToVote, ", ") + " |\n"

	// Add voting percentages for active proposals
	if !ps.Accepted && !ps.Denied {
		out += ufmt.Sprintf("| **✅ YES** | %.2f%% |\n", ps.YesPercent())
		out += ufmt.Sprintf("| **❌ NO** | %.2f%% |\n", ps.NoPercent())
	}

	out += "\n"

	out += "---\n\n"

	out += "### 📋 Description\n\n"
	out += p.Description()
	out += "\n\n"

	// Show executor metadata if exists
	executorMeta := p.ExecutorString()
	if executorMeta != "" {
		out += "---\n\n"
		out += "### ⚙️ Execution Details\n\n"
		out += "> This proposal contains executable actions that will be performed if accepted.\n\n"
		out += "```\n"
		out += executorMeta
		out += "\n```\n\n"
	}

	out += "---\n\n"

	out += "### 📊 Voting Statistics\n\n"
	out += ps.String()

	out += "\n"
	out += ufmt.Sprintf("� [View All Votes](%v:%v/votes)\n\n", ren.relativeRealmPath, pid)

	out += "---\n\n"

	out += renderActionBar(ufmt.Sprintf("%v", pid), ps)

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

	statusIcon := getStatusIcon(ps)
	statusBadge := getStatusBadge(ps)

	// Card-style layout
	out := "### " + statusIcon + " [" + p.Title() + "](" + ufmt.Sprintf("%v:%v", ren.relativeRealmPath, pid) + ")\n\n"
	out += ufmt.Sprintf("> **Proposal #%d** • %s • by %s\n\n", pid, statusBadge, tryResolveAddr(p.Author()))

	// Voting stats for active proposals - compact format
	if !ps.Accepted && !ps.Denied {
		yesPercent := ps.YesPercent()
		noPercent := ps.NoPercent()
		yesVotes := ps.YesVotes.Size()
		noVotes := ps.NoVotes.Size()

		out += ufmt.Sprintf("✅ **YES:** %d (%.1f%%) • ❌ **NO:** %d (%.1f%%) • _Eligible: %s_\n\n",
			yesVotes, yesPercent, noVotes, noPercent, strings.Join(ps.TiersAllowedToVote, ", "))
	} else if ps.DeniedReason != "" {
		out += ufmt.Sprintf("> 📝 %s\n\n", ps.DeniedReason)
	}

	out += "---\n\n"
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

	p, err := dao.GetProposal(cross, dao.ProposalID(pid))
	if err != nil {
		return ufmt.Sprintf("# Proposal not found\n\n%s", err.Error())
	}

	statusIcon := getStatusIcon(ps)
	statusBadge := getStatusBadge(ps)

	out := ufmt.Sprintf("# %s Proposal #%v - Voting Details\n\n", statusIcon, pid)

	out += ufmt.Sprintf("[⬅️ Back to Proposal](%v:%v)\n\n", ren.relativeRealmPath, pid)

	out += "---\n\n"

	out += "**" + p.Title() + "**\n\n"
	out += ufmt.Sprintf("> %s\n\n", statusBadge)

	out += "---\n\n"

	out += "## 📊 Vote Summary\n\n"

	yesPercent := ps.YesPercent()
	noPercent := ps.NoPercent()
	yesVotes := ps.YesVotes.Size()
	noVotes := ps.NoVotes.Size()
	totalVotes := yesVotes + noVotes

	out += "| Vote | Count | Percentage |\n"
	out += "|------|-------|------------|\n"
	out += ufmt.Sprintf("| ✅ **YES** | %d | %.2f%% |\n", yesVotes, yesPercent)
	out += ufmt.Sprintf("| ❌ **NO** | %d | %.2f%% |\n", noVotes, noPercent)
	out += ufmt.Sprintf("| **TOTAL** | **%d** | |\n\n", totalVotes)

	out += "---\n\n"

	out += "## 🗳️ Individual Votes\n\n"

	votes := StringifyVotes(ps)
	if votes == "" || votes == "No one voted yet." {
		out += "_No votes cast yet._\n\n"
	} else {
		out += votes
	}

	return out
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

func getStatusIcon(ps *proposalStatus) string {
	if ps.Accepted {
		return "✅"
	} else if ps.Denied {
		return "❌"
	}
	return "🔄"
}

func getStatusBadge(ps *proposalStatus) string {
	if ps.Accepted {
		return "✅ **ACCEPTED** - This proposal has been approved and executed"
	} else if ps.Denied {
		return "❌ **REJECTED** - This proposal has been denied"
	}
	return "🔄 **ACTIVE** - Voting in progress"
}

func countActiveProposals(d *GovDAO) int {
	count := 0
	d.pss.Iterate("", "", func(key string, value interface{}) bool {
		ps := value.(*proposalStatus)
		if !ps.Accepted && !ps.Denied {
			count++
		}
		return false
	})
	return count
}

func renderActionBar(sPid string, ps *proposalStatus) string {
	out := "### ⚡ Actions\n\n"

	// Show different actions based on proposal status
	if ps.Accepted {
		out += "> ✅ This proposal has been accepted and executed. No further actions needed.\n"
		return out
	} else if ps.Denied {
		out += "> ❌ This proposal has been rejected. No further actions available.\n"
		return out
	}

	// For active proposals, show voting options
	proxy := helplink.Realm("gno.land/r/gov/dao")
	out += "**Cast your vote:**\n\n"
	out += proxy.Func("✅ Vote YES", "MustVoteOnProposalSimple", "pid", sPid, "option", "YES") + " | "
	out += proxy.Func("❌ Vote NO", "MustVoteOnProposalSimple", "pid", sPid, "option", "NO") + " | "
	out += proxy.Func("⚪ Vote ABSTAIN", "MustVoteOnProposalSimple", "pid", sPid, "option", "ABSTAIN")

	out += "\n\n"
	out += "> **Important:** Please verify all transaction details before submitting your vote.\n\n"
	out += "> **Requirement:** You must have a registered namespace to cast a vote."
	return out
}

func tryResolveAddr(addr address) string {
	userData := users.ResolveAddress(addr)
	if userData == nil {
		return addr.String()
	}
	return userData.RenderLink("")
}
