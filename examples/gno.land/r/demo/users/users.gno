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
	admin std.Address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul

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
	// 기존 validation 코드는 그대로 유지...
	std.AssertOriginCall()
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen")
	}

	// Fee 체크 로직 유지...
	sentCoins := std.GetOrigSend()
	minCoin := std.NewCoin("ugnot", minFee)

	if inviter == "" {
		if len(sentCoins) == 1 && sentCoins[0].IsGTE(minCoin) {
			if sentCoins[0].Amount > minFee*maxFeeMult {
				panic("payment must not be greater than " + strconv.Itoa(int(minFee*maxFeeMult)))
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

	// 기존 중복 체크 로직 유지...
	_, ok := name2User.Get(name)
	if ok {
		panic("name already registered: " + name)
	}
	_, ok = addr2User.Get(caller.String())
	if ok {
		panic("address already registered: " + caller.String())
	}

	// invites 계산 로직...
	inviteCount := int(0)
	if len(sentCoins) == 1 {
		if sentCoins[0].Denom == "ugnot" && sentCoins[0].Amount >= minFee {
			inviteCount = int(sentCoins[0].Amount / minFee)
			if inviter == "" && inviteCount > 0 {
				inviteCount -= 1
			}
		}
	}

	// 먼저 Profile 생성
	profile1 := users.Profile{ // & 제거
		Header:  name,    // 대문자로 시작 (users 패키지에서 정의된 대로)
		Profile: profile, // 대문자로 시작
		Image:   "",
		Blog:    avl.NewTree(), // 대문자로 시작
		Post:    avl.NewTree(), // 대문자로 시작
	}

	// counter 증가
	counter++

	// Profile을 포함한 User 생성
	user := &users.User{
		Address: caller,
		Name:    name,
		Invites: inviteCount,
		Inviter: inviter,
		Profile: profile1, // Profile 필드도 대문자로 시작
		Number:  counter,
	}
	// 저장
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

// Update profile and blog
// 프로필 조회 함수
func GetUserProfile(addr std.Address) (*users.Profile, error) {
	user := GetUserByAddress(addr)
	if user == nil {
		return nil, std.Error("user not found")
	}
	return &user.Profile, nil
}

// 프로필 수정 함수
func UpdateUserProfile(header string, profile string, image string) {
	// 호출자 검증
	std.AssertOriginCall()
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen")
	}

	// 유저 존재 여부 확인
	user := GetUserByAddress(caller)
	if user == nil {
		panic("user not found")
	}

	// 헤더 검증
	if len(header) > 100 {
		panic("header too long: max 100 characters")
	}
	if len(header) == 0 {
		header = user.Name // 비어있으면 유저 이름으로 설정
	}

	// 프로필 설명 검증
	if len(profile) > 500 {
		panic("profile too long: max 500 characters")
	}

	// 이미지 검증 (base64 형식)
	if len(image) > 0 && !strings.HasPrefix(image, "data:image/") {
		panic("invalid image format: must be base64 encoded image")
	}

	// 프로필 업데이트
	user.Profile.Header = header
	user.Profile.Profile = profile
	if len(image) > 0 {
		user.Profile.Image = image
	}

	// addr2User 업데이트는 필요 없음 (포인터라서 자동 반영)
	// name2User는 이름 변경이 없으므로 업데이트 불필요
}

// 블로그 포스트 추가
func AddBlogPost(title string, content string) {
	std.AssertOriginCall()
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen")
	}

	user := GetUserByAddress(caller)
	if user == nil {
		panic("user not found")
	}

	if len(title) > 200 {
		panic("title too long: max 200 characters")
	}

	// 블로그 포스트 번호 생성
	postNum := user.Profile.Blog.Size() + 1

	// 블로그 포스트 저장
	user.Profile.Blog.Set(strconv.Itoa(postNum), title+"\n"+content)
}

// 포스트(짧은 메시지) 추가
func AddPost(content string) {
	std.AssertOriginCall()
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen")
	}

	user := GetUserByAddress(caller)
	if user == nil {
		panic("user not found")
	}

	if len(content) > 280 {
		panic("post too long: max 280 characters")
	}

	// 포스트 번호 생성
	postNum := user.Profile.Post.Size() + 1

	// 포스트 저장
	user.Profile.Post.Set(strconv.Itoa(postNum), content)
}

// 블로그 포스트 삭제
func DeleteBlogPost(postNum string) {
	std.AssertOriginCall()
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen")
	}

	user := GetUserByAddress(caller)
	if user == nil {
		panic("user not found")
	}

	if !user.Profile.Blog.Has(postNum) {
		panic("blog post not found")
	}

	user.Profile.Blog.Remove(postNum)
}

// 포스트 삭제
func DeletePost(postNum string) {
	std.AssertOriginCall()
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen")
	}

	user := GetUserByAddress(caller)
	if user == nil {
		panic("user not found")
	}

	if !user.Profile.Post.Has(postNum) {
		panic("post not found")
	}

	user.Profile.Post.Remove(postNum)
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

	page := pager.NewPager(&name2User, 50).MustGetPageByPath(path)

	for _, item := range page.Items {
		user := item.Value.(*users.User)
		doc += " * [" + user.Name + "](/r/demo/users:" + user.Name + ")\n"
	}
	doc += "\n"
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
