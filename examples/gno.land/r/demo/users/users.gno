package users

import (
	"regexp"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/avlhelpers"
	"gno.land/p/demo/users"
)

//----------------------------------------
// State

var (
	admin std.Address = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul

	restricted avl.Tree                  // Name -> true - restricted name
	name2User  avl.Tree                  // Name -> *users.User
	addr2User  avl.Tree                  // std.Address -> *users.User
	invites    avl.Tree                  // string(inviter+":"+invited) -> true
	counter    int                       // user id counter
	minFee     int64    = 20 * 1_000_000 // minimum gnot must be paid to register.
	maxFeeMult int64    = 10             // maximum multiples of minFee accepted.
)

//----------------------------------------
// Top-level functions

func Register(inviter std.Address, name string, profile string) {
	// assert CallTx call.
	std.AssertOriginCall()
	// assert invited or paid.
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen") // because std.AssertOrigCall().
	}

	sentCoins := std.GetOrigSend()
	minCoin := std.NewCoin("ugnot", minFee)

	if inviter == "" {
		// banker := std.GetBanker(std.BankerTypeOrigSend)
		if len(sentCoins) == 1 && sentCoins[0].IsGTE(minCoin) {
			if sentCoins[0].Amount > minFee*maxFeeMult {
				panic("payment must not be greater than " + strconv.Itoa(int(minFee*maxFeeMult)))
			} else {
				// ok
			}
		} else {
			panic("payment must not be less than " + strconv.Itoa(int(minFee)))
		}
	} else {
		invitekey := inviter.String() + ":" + caller.String()
		_, ok := invites.Get(invitekey)
		if !ok {
			panic("invalid invitation")
		}
		invites.Remove(invitekey)
	}

	// assert not already registered.
	_, ok := name2User.Get(name)
	if ok {
		panic("name already registered: " + name)
	}
	_, ok = addr2User.Get(caller.String())
	if ok {
		panic("address already registered: " + caller.String())
	}

	isInviterAdmin := inviter == admin

	// check for restricted name
	if _, isRestricted := restricted.Get(name); isRestricted {
		// only address invite by the admin can register restricted name
		if !isInviterAdmin {
			panic("restricted name: " + name)
		}

		restricted.Remove(name)
	}

	// assert name is valid.
	// admin inviter can bypass name restriction
	if !isInviterAdmin && !reName.MatchString(name) {
		panic("invalid name: " + name + " (must be at least 6 characters, lowercase alphanumeric with underscore)")
	}

	// remainder of fees go toward invites.
	invites := int(0)
	if len(sentCoins) == 1 {
		if sentCoins[0].Denom == "ugnot" && sentCoins[0].Amount >= minFee {
			invites = int(sentCoins[0].Amount / minFee)
			if inviter == "" && invites > 0 {
				invites -= 1
			}
		}
	}
	// register.
	counter++
	user := &users.User{
		Address: caller,
		Name:    name,
		Profile: profile,
		Number:  counter,
		Invites: invites,
		Inviter: inviter,
	}
	name2User.Set(name, user)
	addr2User.Set(caller.String(), user)
}

func Invite(invitee string) {
	// assert CallTx call.
	std.AssertOriginCall()
	// get caller/inviter.
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen") // because std.AssertOrigCall().
	}
	lines := strings.Split(invitee, "\n")
	if caller == admin {
		// nothing to do, all good
	} else {
		// ensure has invites.
		userI, ok := addr2User.Get(caller.String())
		if !ok {
			panic("user unknown")
		}
		user := userI.(*users.User)
		if user.Invites <= 0 {
			panic("user has no invite tokens")
		}
		user.Invites -= len(lines)
		if user.Invites < 0 {
			panic("user has insufficient invite tokens")
		}
	}
	// for each line...
	for _, line := range lines {
		if line == "" {
			continue // file bodies have a trailing newline.
		} else if strings.HasPrefix(line, `//`) {
			continue // comment
		}
		// record invite.
		invitekey := string(caller) + ":" + string(line)
		invites.Set(invitekey, true)
	}
}

func GrantInvites(invites string) {
	// assert CallTx call.
	std.AssertOriginCall()
	// assert admin.
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen") // because std.AssertOrigCall().
	}
	if caller != admin {
		panic("unauthorized")
	}
	// for each line...
	lines := strings.Split(invites, "\n")
	for _, line := range lines {
		if line == "" {
			continue // file bodies have a trailing newline.
		} else if strings.HasPrefix(line, `//`) {
			continue // comment
		}
		// parse name and invites.
		var name string
		var invites int
		parts := strings.Split(line, ":")
		if len(parts) == 1 { // short for :1.
			name = parts[0]
			invites = 1
		} else if len(parts) == 2 {
			name = parts[0]
			invites_, err := strconv.Atoi(parts[1])
			if err != nil {
				panic(err)
			}
			invites = int(invites_)
		} else {
			panic("should not happen")
		}
		// give invites.
		userI, ok := name2User.Get(name)
		if !ok {
			// maybe address.
			userI, ok = addr2User.Get(name)
			if !ok {
				panic("invalid user " + name)
			}
		}
		user := userI.(*users.User)
		user.Invites += invites
	}
}

// Any leftover fees go toward invitations.
func SetMinFee(newMinFee int64) {
	// assert CallTx call.
	std.AssertOriginCall()
	// assert admin caller.
	caller := std.GetCallerAt(2)
	if caller != admin {
		panic("unauthorized")
	}
	// update global variables.
	minFee = newMinFee
}

// This helps prevent fat finger accidents.
func SetMaxFeeMultiple(newMaxFeeMult int64) {
	// assert CallTx call.
	std.AssertOriginCall()
	// assert admin caller.
	caller := std.GetCallerAt(2)
	if caller != admin {
		panic("unauthorized")
	}
	// update global variables.
	maxFeeMult = newMaxFeeMult
}

//----------------------------------------
// Exposed public functions

func GetUserByName(name string) *users.User {
	userI, ok := name2User.Get(name)
	if !ok {
		return nil
	}
	return userI.(*users.User)
}

func GetUserByAddress(addr std.Address) *users.User {
	userI, ok := addr2User.Get(addr.String())
	if !ok {
		return nil
	}
	return userI.(*users.User)
}

// unlike GetUserByName, input must be "@" prefixed for names.
func GetUserByAddressOrName(input users.AddressOrName) *users.User {
	name, isName := input.GetName()
	if isName {
		return GetUserByName(name)
	}
	return GetUserByAddress(std.Address(input))
}

// Get a list of user names starting from the given prefix. Limit the
// number of results to maxResults. (This can be used for a name search tool.)
func ListUsersByPrefix(prefix string, maxResults int) []string {
	return avlhelpers.ListByteStringKeysByPrefix(name2User, prefix, maxResults)
}

func Resolve(input users.AddressOrName) std.Address {
	name, isName := input.GetName()
	if !isName {
		return std.Address(input) // TODO check validity
	}

	user := GetUserByName(name)
	return user.Address
}

// Add restricted name to the list
func AdminAddRestrictedName(name string) {
	// assert CallTx call.
	std.AssertOriginCall()
	// get caller
	caller := std.GetOrigCaller()
	// assert admin
	if caller != admin {
		panic("unauthorized")
	}

	if user := GetUserByName(name); user != nil {
		panic("already registered name")
	}

	// register restricted name

	restricted.Set(name, true)
}

//----------------------------------------
// Constants

// NOTE: name length must be clearly distinguishable from a bech32 address.
var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{5,16}$`)

//----------------------------------------
// Render main page

func Render(fullPath string) string {
	path, _ := splitPathAndQuery(fullPath)
	if path == "" {
		return renderHome(fullPath)
	} else if len(path) >= 38 { // 39? 40?
		if path[:2] != "g1" {
			return "invalid address " + path
		}
		user := GetUserByAddress(std.Address(path))
		if user == nil {
			// TODO: display basic information about account.
			return "unknown address " + path
		}
		return user.Render()
	} else {
		user := GetUserByName(path)
		if user == nil {
			return "unknown username " + path
		}
		return user.Render()
	}
}

func renderHome(path string) string {
	doc := ""

	page := pager.NewPager(&name2User).MustGetPageByPath(path)

	for _, item := range page.Items {
		user := item.Value.(*users.User)
		doc += " * [" + user.Name + "](/r/demo/users:" + user.Name + ")\n"
	}

	doc += page.Selector()
	return doc
}

func splitPathAndQuery(fullPath string) (string, string) {
	parts := strings.SplitN(fullPath, "?", 2)
	path := parts[0]
	queryString := ""
	if len(parts) > 1 {
		queryString = "?" + parts[1]
	}
	return path, queryString
}
