package basedao

import (
	"strconv"
	"strings"

	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/samcrew/daokit"
)

// WELL KNOWN PATHS
const (
	HOME_PATH             = ""
	HOME_NO_PROFILE_PATH  = "noprofile"
	CONFIG_PATH           = "config"
	PROPOSAL_HISTORY_PATH = "history"
	MEMBER_DETAIL_PATH    = "member/{address}"
	PROPOSAL_DETAIL_PATH  = "proposal/{id}"
	FALLBACK_DISPLAY_NAME = "Anon"
)

func (d *DAOPrivate) initRenderingRouter() {
	if d.RenderRouter == nil {
		d.RenderRouter = mux.NewRouter()
	}
}

func (d *DAOPrivate) InitDefaultRendering() {
	d.RenderRouter.HandleFunc(HOME_PATH, d.MuxHomePage)
	d.RenderRouter.HandleFunc(HOME_NO_PROFILE_PATH, d.MuxHomePage)
	d.RenderRouter.HandleFunc(CONFIG_PATH, d.MuxConfigPage)
	d.RenderRouter.HandleFunc(PROPOSAL_HISTORY_PATH, d.MuxProposalHistoryPage)
	d.RenderRouter.HandleFunc(MEMBER_DETAIL_PATH, d.MuxMemberDetailPage)
	d.RenderRouter.HandleFunc(PROPOSAL_DETAIL_PATH, d.MuxProposalDetailPage)
}

func (d *DAOPrivate) Render(path string) string {
	return d.RenderRouter.Render(path)
}

// DEFAULT_HANDLERS

func (d *DAOPrivate) MuxHomePage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.HomePageView())
}

func (d *DAOPrivate) MuxConfigPage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.ConfigPageView())
}

func (d *DAOPrivate) MuxProposalHistoryPage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.ProposalHistoryPageView())
}

func (d *DAOPrivate) MuxMemberDetailPage(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(d.MemberDetailPageView(req.GetVar("address")))
}

func (d *DAOPrivate) MuxProposalDetailPage(res *mux.ResponseWriter, req *mux.Request) {
	id, err := strconv.ParseUint(req.GetVar("id"), 10, 64)
	if err != nil {
		panic(err)
	}
	res.Write(d.ProposalDetailPageView(id))
}

// PUBLIC PAGES

func (d *DAOPrivate) HomePageView() string {
	return d.HomeProfileHeaderView(false) +
		d.HomeMembersView() +
		d.HomeProposalsView()
}

func (d *DAOPrivate) ConfigPageView() string {
	return d.ConfigHeaderView() +
		d.ConfigRolesView() +
		d.ConfigResourcesView()
}

func (d *DAOPrivate) ProposalHistoryPageView() string {
	return d.ProposalHistoryHeaderView() +
		d.ProposalHistoryView()
}

func (d *DAOPrivate) MemberDetailPageView(address string) string {
	return d.MemberDetailHeaderView() +
		d.MemberDetailView(address)
}

func (d *DAOPrivate) ProposalDetailPageView(idu uint64) string {
	return d.ProposalDetailHeaderView() +
		d.ProposalDetailView(idu)
}

// --- PUBLIC COMPONENT VIEWS ---

// HOME PAGE COMPONENTS VIEWS

func (d *DAOPrivate) HomeProfileHeaderView(noprofile bool) string {
	s := ""
	name := d.GetProfileString(d.RealmAddress, "DisplayName", "DAO")
	description := d.GetProfileString(d.RealmAddress, "Bio", "Created with daokit")
	pkgPath := d.RealmPkgPath
	linkPath := getLinkPath(pkgPath)

	if !noprofile {
		s += ufmt.Sprintf("# %s\n\n", name)
		imageURI := d.GetProfileString(d.RealmAddress, "Avatar", "")
		if imageURI != "" {
			s += ufmt.Sprintf("![Image](%s)\n\n", imageURI)
		}
		s += ufmt.Sprintf("%s\n\n", description)
	}

	s += ufmt.Sprintf("> Realm address: %s\n\n", d.RealmAddress)
	s += ufmt.Sprintf("Discover more about this DAO on the [configuration page ⚙️](%s:%s)\n\n", linkPath, CONFIG_PATH)
	s += ufmt.Sprintf("\n--------------------------------\n")

	return s
}

func (d *DAOPrivate) HomeMembersView() string {
	pkgPath := d.RealmPkgPath
	linkPath := getLinkPath(pkgPath)
	s := ""
	s += ufmt.Sprintf("## Members 👤 \n\n")
	i := 1
	d.Members.Members.Iterate("", "", func(key string, value interface{}) bool {
		s += ufmt.Sprintf("- **Member %d: [%s](%s:%s/%s)**\n\n", i, key, linkPath, "member", key)
		s += ufmt.Sprintf("	- **Profile:** %s\n", d.GetProfileString(address(key), "DisplayName", FALLBACK_DISPLAY_NAME))
		s += ufmt.Sprintf(" 	- **Roles:** %s\n\n", strings.Join(d.Members.GetMemberRoles(key), ", "))
		i += 1
		return false
	})
	s += ufmt.Sprintf("> You can find more information about a member by clicking on their address\n\n")
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) HomeProposalsView() string {
	pkgPath := d.RealmPkgPath
	linkPath := getLinkPath(pkgPath)
	s := ""
	s += ufmt.Sprintf("## Proposals 🗳️ \n\n")
	i := 0
	d.Core.Proposals.Tree.Iterate("", "", func(key string, value interface{}) bool {
		proposal := value.(*daokit.Proposal)
		if proposal.Status != daokit.ProposalStatusOpen {
			return false
		}
		id, err := seqid.FromString(key)
		if err != nil {
			panic(err)
		}
		s += ufmt.Sprintf("- **Proposal %d: [%s](%s:%s/%d)**\n\n", uint64(id), proposal.Title, linkPath, "proposal", uint64(id))
		i += 1
		return false
	})
	if i == 0 {
		s += ufmt.Sprintf("\t⚠️ There are no running proposals at the moment\n\n")
	}
	s += ufmt.Sprintf("> See the [proposal history 📜](%s:%s) for more information\n\n", linkPath, PROPOSAL_HISTORY_PATH)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

// CONFIG PAGE COMPONENTS VIEWS

func (d *DAOPrivate) ConfigHeaderView() string {
	name := d.GetProfileString(d.RealmAddress, "DisplayName", "DAO")
	s := ""
	s += ufmt.Sprintf("# %s - Config ⚙️\n\n", name)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) ConfigRolesView() string {
	roles := d.Members.GetRoles()
	s := ""
	s += ufmt.Sprintf("## Roles 🏷️\n\n")
	for _, role := range roles {
		s += ufmt.Sprintf("- %s\n\n", role)
		info := d.Members.RoleInfo(role)
		s += ufmt.Sprintf("  %s\n\n", info.Description)
	}
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) ConfigResourcesView() string {
	s := ""
	s += ufmt.Sprintf("## Resources 📦\n\n")
	i := 1
	d.Core.Resources.Tree.Iterate("", "", func(key string, value interface{}) bool {
		resource := value.(*daokit.Resource)
		s += ufmt.Sprintf("- **Resource #%d: %s**\n\n", i, key)
		// TODO: add doc to handler and print here
		s += ufmt.Sprintf("  - **Name:** %s\n", resource.DisplayName)
		s += ufmt.Sprintf("  - **Description:** %s\n", resource.Description)
		s += ufmt.Sprintf("  - **Condition:** %s\n\n", resource.Condition.Render())
		i += 1
		return false
	})
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

// HISTORY PAGE COMPONENTS VIEWS

func (d *DAOPrivate) ProposalHistoryHeaderView() string {
	name := d.GetProfileString(d.RealmAddress, "DisplayName", "DAO")
	s := ""
	s += ufmt.Sprintf("# %s - Proposal History\n\n", name)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) ProposalHistoryView() string {
	pkgPath := d.RealmPkgPath
	linkPath := getLinkPath(pkgPath)
	s := ""
	s += ufmt.Sprintf("## Proposals 🗳️\n\n")
	i := 1
	d.Core.Proposals.Tree.Iterate("", "", func(key string, value interface{}) bool {
		proposal, ok := value.(*daokit.Proposal)
		if !ok {
			panic("unexpected invalid proposal type")
		}
		id, err := seqid.FromString(key)
		if err != nil {
			panic(err)
		}
		s += ufmt.Sprintf("- **Proposal %d: [%s](%s:%s/%d) - %s**\n\n", uint64(id), proposal.Title, linkPath, "proposal", uint64(id), proposal.Status)
		i += 1
		return false
	})
	s += ufmt.Sprintf("[Add a new proposal 🗳️](%s$help)\n\n", linkPath)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

// MEMBER DETAIL PAGE COMPONENTS VIEWS

func (d *DAOPrivate) MemberDetailHeaderView() string {
	name := d.GetProfileString(d.RealmAddress, "DisplayName", "DAO")
	s := ""
	s += ufmt.Sprintf("# %s - Member Detail\n\n", name)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) MemberDetailView(addr string) string {
	pkgPath := d.RealmPkgPath
	linkPath := getLinkPath(pkgPath)
	roles := d.Members.GetMemberRoles(addr)
	displayName := d.GetProfileString(address(addr), "DisplayName", FALLBACK_DISPLAY_NAME)
	bio := d.GetProfileString(address(addr), "Bio", "No bio")
	pp := d.GetProfileString(address(addr), "Avatar", "")
	s := ""
	s += ufmt.Sprintf("## Profile 👤\n\n")
	s += ufmt.Sprintf("- **Display Name:** %s\n\n", displayName)
	s += ufmt.Sprintf("- **Bio:** %s\n\n", bio)
	if pp != "" {
		s += ufmt.Sprintf("![Avatar](%s)\n\n", pp)
	}
	s += ufmt.Sprintf("## Roles 🏷️\n\n")
	for _, role := range roles {
		s += ufmt.Sprintf("- %s\n\n", role)
	}
	s += ufmt.Sprintf("> Learn more about the roles on the [configuration page ⚙️](%s:%s)\n\n", linkPath, CONFIG_PATH)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

// PROPOSAL DETAIL PAGE COMPONENTS VIEWS

func (d *DAOPrivate) ProposalDetailHeaderView() string {
	name := d.GetProfileString(d.RealmAddress, "DisplayName", "DAO")
	s := ""
	s += ufmt.Sprintf("# %s - Proposal Detail\n\n", name)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) ProposalDetailView(idu uint64) string {
	pkgPath := d.RealmPkgPath
	linkPath := getLinkPath(pkgPath)
	id := seqid.ID(idu)
	proposal := d.Core.Proposals.GetProposal(uint64(id))
	s := ""
	s += ufmt.Sprintf("## Title - %s 📜\n\n", proposal.Title)
	s += ufmt.Sprintf("## Description 📝\n\n%s\n\n", proposal.Description)
	s += ufmt.Sprintf("## Resource - %s 📦\n\n", proposal.Action.Type())
	resource := d.Core.Resources.Get(proposal.Action.Type())
	s += ufmt.Sprintf("  - **Name:** %s\n", resource.DisplayName)
	s += ufmt.Sprintf("  - **Description:** %s\n", resource.Description)
	s += ufmt.Sprintf("  - **Condition:** %s\n\n", resource.Condition.Render())
	s += ("---\n\n")
	s += proposal.Action.String() + "\n\n"
	s += ("---\n\n")
	if proposal.Status == daokit.ProposalStatusOpen {
		s += ufmt.Sprintf("## Status - Open 🟡\n\n")
		s += ufmt.Sprintf("[Vote on this proposal 🗳️](%s$help)\n\n", linkPath)
	} else if proposal.Status == daokit.ProposalStatusPassed {
		s += ufmt.Sprintf("## Status - Passed 🟢\n\n")
		s += ufmt.Sprintf("[Execute this proposal 🗳️](%s$help)\n\n", linkPath)
	} else if proposal.Status == daokit.ProposalStatusExecuted {
		s += ufmt.Sprintf("## Status - Executed ✅\n\n")
	} else {
		s += ufmt.Sprintf("## Status - Closed 🔴\n\n")
	}
	s += ufmt.Sprintf("> proposed by %s 👤\n\n", proposal.ProposerID)
	s += ufmt.Sprintf("\n--------------------------------\n")
	s += ufmt.Sprintf("## Votes 🗳️\n\n%s\n\n", proposal.Condition.RenderWithVotes(proposal.Ballot))
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func getLinkPath(pkgPath string) string {
	slashIdx := strings.IndexRune(pkgPath, '/')
	if slashIdx != 1 {
		return pkgPath[slashIdx:]
	}
	return ""
}
