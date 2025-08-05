package memberstore

import (
	"net/url"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
)

const (
	T1 = "T1"
	T2 = "T2"
	T3 = "T3"
)

var (
	members  = NewMembersByTier()
	Tiers    = TiersByName{avl.NewTree()}
	AllTiers = []string{T1, T2, T3} // Common tiers list used throughout
)

// Initializes the tiers with their properties and power handlers
func init() {
	Tiers.Set(T1, Tier{
		InvitationPoints: 3,
		MinSize:          func(mbt MembersByTier, tbn TiersByName) int { return 70 },
		MaxSize:          func(mbt MembersByTier, tbn TiersByName) int { return 0 },
		BasePower:        3,
		PowerHandler:     func(mbt MembersByTier, tbn TiersByName) float64 { return 3 },
	})

	Tiers.Set(T2, Tier{
		InvitationPoints: 2,
		MinSize: func(mbt MembersByTier, tbn TiersByName) int {
			return mbt.GetTierSize(T1) / 4
		},
		MaxSize: func(mbt MembersByTier, tbn TiersByName) int {
			return mbt.GetTierSize(T1) * 2
		},
		BasePower: 2,
		PowerHandler: func(mbt MembersByTier, tbn TiersByName) float64 {
			t1ms := float64(mbt.GetTierSize(T1))
			t1, _ := tbn.GetTier(T1)
			t2ms := float64(mbt.GetTierSize(T2))
			t2, _ := tbn.GetTier(T2)

			t1p := t1.BasePower * t1ms
			t2p := t2.BasePower * t2ms

			// capped to 2/3 of tier 1
			t1pthreshold := t1p * (2.0 / 3.0)
			if t2p > t1pthreshold {
				return t1pthreshold / t2ms
			}

			return t2.BasePower
		},
	})

	Tiers.Set(T3, Tier{
		InvitationPoints: 1,
		MinSize:          func(mbt MembersByTier, tbn TiersByName) int { return 0 },
		MaxSize:          func(mbt MembersByTier, tbn TiersByName) int { return 0 },
		BasePower:        1,
		PowerHandler: func(mbt MembersByTier, tbn TiersByName) float64 {
			t1ms := float64(mbt.GetTierSize(T1))
			t1, _ := tbn.GetTier(T1)
			t3ms := float64(mbt.GetTierSize(T3))
			t3, _ := tbn.GetTier(T3)

			t1p := t1.BasePower * t1ms
			t3p := t3.BasePower * t3ms

			//capped to 1/3 of tier 1
			t1pthreshold := t1p * (1.0 / 3.0)
			if t3p > t1pthreshold {
				return t1pthreshold / t3ms
			}

			return t3.BasePower
		},
	})
}

// Returns the member store for the current realm
func Get() MembersByTier {
	currealm := std.CurrentRealm().PkgPath()
	if !dao.InAllowedDAOs(currealm) {
		panic("this Realm is not allowed to get the Members data: " + currealm)
	}
	return members
}

// Displays the full memberstore with tier filters and pagination
func Render(path string) string {
	u, _ := url.Parse(path)
	query := u.Query()
	filter := parseFilterMap(query.Get("filter"))
	pageNum := parsePageNum(query.Get("page"))
	pageSize := 14
	allMembers := avl.NewTree()
	sb := strings.Builder{}

	sb.WriteString("# Memberstore Govdao v3\n## Tiers:\n")
	collectMembersAndRenderTiers(&sb, filter, allMembers)

	sb.WriteString("## Members:\n")
	renderTierFilters(&sb, u, query, filter)

	pager := pager.NewPager(allMembers, pageSize, false)
	page := pager.GetPageWithSize(pageNum, pageSize)
	renderMembersPage(&sb, page, allMembers)
	renderPagination(&sb, u, query, page, "page")

	return sb.String()
}

// Parses the page number from the query string, defaults to 1 if invalid
func parsePageNum(s string) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return 1
}

// Parses filter query into a map for quick lookup
func parseFilterMap(filter string) map[string]bool {
	m := map[string]bool{}
	for _, tn := range strings.Split(filter, ",") {
		trimTn := strings.TrimSpace(tn)
		if trimTn != "" {
			m[trimTn] = true
		}
	}
	return m
}

// Renders the tiers and their summaries into the provided string builder
func collectMembersAndRenderTiers(sb *strings.Builder, filter map[string]bool, allMembers *avl.Tree) {
	members.Iterate("", "", func(tn string, ti interface{}) bool {
		tree, ok := ti.(*avl.Tree)
		tier, ok2 := Tiers.GetTier(tn)
		if !ok || !ok2 {
			return false
		}
		tree.Iterate("", "", func(addr string, _ interface{}) bool {
			if len(filter) == 0 || filter[tn] {
				allMembers.Set(addr, tn)
			}
			return false
		})

		tp := tier.PowerHandler(members, Tiers) * float64(members.GetTierSize(tn))
		sb.WriteString(ufmt.Sprintf("- Tier %v contains %v members with power: %v\n", tn, tree.Size(), tp))

		return false
	})
}

// Renders the tier filter hyperlinks into the provided string builder
func renderTierFilters(sb *strings.Builder, u *url.URL, query url.Values, filter map[string]bool) {
	sb.WriteString("Filter members by tier: \n")
	for _, tn := range AllTiers {
		newFilters := toggleFilter(filter, tn)

		newQuery := copyQueryWithoutKey(query, "filter")
		if len(newFilters) > 0 {
			newQuery.Set("filter", strings.Join(newFilters, ","))
		}

		urlStr := buildURL(u.Path, newQuery)
		label := "**" + tn + "**"
		if !filter[tn] {
			label = "~~" + tn + "~~"
		}
		sb.WriteString(ufmt.Sprintf(" | [%v](%v) ", label, urlStr))
	}
	sb.WriteString("\n")
}

// Toggles the filter for a specific tier and returns the updated filter list
func toggleFilter(filter map[string]bool, toggleTn string) []string {
	result := []string{}
	for _, tn := range AllTiers {
		if tn == toggleTn {
			if !filter[tn] {
				result = append(result, tn)
			}
		} else if filter[tn] {
			result = append(result, tn)
		}
	}
	return result
}

// Returns the query parameters, excluding the specified key
func copyQueryWithoutKey(query url.Values, key string) url.Values {
	newQuery := url.Values{}
	for k, v := range query {
		if k != key {
			for _, vv := range v {
				newQuery.Add(k, vv)
			}
		}
	}
	return newQuery
}

// Builds a URL with the given path and query parameters
func buildURL(path string, query url.Values) string {
	if enc := query.Encode(); enc != "" {
		return path + "?" + enc
	}
	return path
}

// Writes the members of the current page into the provided string builder
func renderMembersPage(sb *strings.Builder, page *pager.Page, allMembers *avl.Tree) {
	for _, item := range page.Items {
		addr := item.Key
		tn, _ := allMembers.Get(addr)
		sb.WriteString(ufmt.Sprintf("- %v · %v\n", tn, addr))
	}
}

// Renders the pagination UI for the current page
func renderPagination(sb *strings.Builder, u *url.URL, query url.Values, page *pager.Page, pageKey string) {
	sb.WriteString("\n")
	baseQuery := copyQueryWithoutKey(query, pageKey)
	basePath := buildURL(u.Path, baseQuery)
	sb.WriteString(page.Picker(basePath))
	sb.WriteString("\n\n")
}
