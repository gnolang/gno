package memberstore

import (
	"chain/runtime"
	"strings"

	"gno.land/p/demo/svg"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
	"gno.land/r/gov/dao"
)

var (
	members MembersByTier
	tiers   TiersByName // private to prevent external modification
	router  *mux.Router
)

const (
	T1 = "T1"
	T2 = "T2"
	T3 = "T3"
)

func init() {
	members = NewMembersByTier()

	tiers = TiersByName{avl.NewTree()}
	tiers.Set(T1, Tier{
		InvitationPoints: 3,
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 70
		},
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		BasePower: 3,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			return 3
		},
	})

	tiers.Set(T2, Tier{
		InvitationPoints: 2,
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return membersByTier.GetTierSize(T1) * 2
		},
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return membersByTier.GetTierSize(T1) / 4
		},
		BasePower: 2,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			t1ms := float64(membersByTier.GetTierSize(T1))
			t1, _ := tiersByName.GetTier(T1)
			t2ms := float64(membersByTier.GetTierSize(T2))
			t2, _ := tiersByName.GetTier(T2)

			t1p := t1.BasePower * t1ms
			t2p := t2.BasePower * t2ms

			// capped to 2/3 of tier 1
			t1ptreshold := t1p * (2.0 / 3.0)
			if t2p > t1ptreshold {
				return t1ptreshold / t2ms
			}

			return t2.BasePower
		},
	})

	tiers.Set(T3, Tier{
		InvitationPoints: 1,
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		BasePower: 1,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			t1ms := float64(membersByTier.GetTierSize(T1))
			t1, _ := tiersByName.GetTier(T1)
			t3ms := float64(membersByTier.GetTierSize(T3))
			t3, _ := tiersByName.GetTier(T3)

			t1p := t1.BasePower * t1ms
			t3p := t3.BasePower * t3ms

			// capped to 1/3 of tier 1
			t1ptreshold := t1p * (1.0 / 3.0)
			if t3p > t1ptreshold {
				return t1ptreshold / t3ms
			}

			return t3.BasePower
		},
	})

	initRouter()
}

// initRouter initializes the router for the memberstore.
func initRouter() {
	router = mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("members", renderMembers)
	router.NotFoundHandler = renderNotFound
}

// renderHome displays the tiers data (Number of members and powers) and tiers charts.
func renderHome(res *mux.ResponseWriter, req *mux.Request) {
	var sb strings.Builder

	// Navigation
	sb.WriteString("[View Members List](/r/gov/dao/v3/memberstore:members) | ")
	sb.WriteString("[Back to Proposals](/r/gov/dao/v3/impl)\n\n")

	// Tier summary section
	sb.WriteString("## Tier Summary\n\n")

	sb.WriteString("| Tier | Members | Power |\n")
	sb.WriteString("|------|---------|-------|\n")

	members.Iterate("", "", func(tn string, ti interface{}) bool {
		tree, ok := ti.(*avl.Tree)
		if !ok {
			return false
		}

		tier, ok := tiers.GetTier(tn)
		if !ok {
			return false
		}

		tp := (tier.PowerHandler(members, tiers) * float64(members.GetTierSize(tn)))

		sb.WriteString(ufmt.Sprintf("| %s | %v | %.1f |\n", tn, tree.Size(), tp))

		return false
	})

	sb.WriteString("\n")

	// Charts section
	sb.WriteString(RenderCharts(members))

	res.Write(sb.String())
}

// renderMembers displays the members list.
func renderMembers(res *mux.ResponseWriter, req *mux.Request) {
	path := strings.Replace(req.RawPath, "members", "", 1) // We have to clean the path
	res.Write(RenderMembers(path, members))
}

func renderNotFound(res *mux.ResponseWriter, req *mux.Request) {
	res.Write("# 404\n\nThat page was not found. Would you like to [**go home**?](/r/gov/dao/v3/memberstore)")
}

func tierColor(tn string) string {
	switch tn {
	case T1:
		return "#329175"
	case T2:
		return "#21577A"
	case T3:
		return "#F3D3BC"
	default:
		return "#FFF"
	}
}

// tierColoredChip returns a colored chip svg for the given tier name.
func tierColoredChip(tn string) string {
	canvas := svg.NewCanvas(16, 16)
	canvas.Append(svg.NewRectangle(0, 0, 16, 16, tierColor(tn)))
	return canvas.Render(tn + " colored chip")
}

func Render(path string) string {
	var sb strings.Builder
	sb.WriteString(md.H1("GovDAO Members"))
	sb.WriteString(router.Render(path))
	return sb.String()
}

// Get gets the Members store
func Get() MembersByTier {
	currealm := runtime.CurrentRealm().PkgPath()
	if !dao.InAllowedDAOs(currealm) {
		panic("this Realm is not allowed to get the Members data: " + currealm)
	}

	return members
}

// GetTier returns a tier by name. This is a read-only accessor.
func GetTier(name string) (Tier, bool) {
	return tiers.GetTier(name)
}

// IterateTiers iterates over all tiers in order. This is a read-only accessor.
// The callback receives the tier name and tier data.
// Return true from the callback to stop iteration.
func IterateTiers(fn func(name string, tier Tier) bool) {
	tiers.Iterate("", "", func(name string, value interface{}) bool {
		tier, ok := value.(Tier)
		if !ok {
			return false
		}
		return fn(name, tier)
	})
}

// setTiers replaces the tiers configuration.
// This is internal and should only be called via governance proposal execution.
func setTiers(newTiers TiersByName) {
	tiers = newTiers
}

// GetTierPower calculates the effective voting power for a tier given the current members.
// This is a safe accessor that uses the internal tiers configuration.
func GetTierPower(tierName string, members MembersByTier) float64 {
	tier, ok := tiers.GetTier(tierName)
	if !ok {
		return 0
	}
	return tier.PowerHandler(members, tiers)
}
