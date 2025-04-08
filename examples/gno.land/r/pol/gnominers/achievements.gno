package gnominers

import "gno.land/r/pol/gnominers/pickaxes"

var achievements = map[string]iAchievement{
	"a gem":       &achMineAGem{},
	"1000rocks":   &achMine1000Rocks{},
	"alexandrite": &achAlexandrite{},
	"million":     &ach1Million{},
	"42k gems":    &ach42kGems{},
	"bday":        &achBDAY{},
}

type achievementDifficulty uint

const (
	diffEasy achievementDifficulty = iota
	diffMedium
	diffHard
	diffInsane
	diffImpossible
)

var achievementDifficultyNames = map[achievementDifficulty]string{
	diffEasy:       "Easy",
	diffMedium:     "Medium",
	diffHard:       "Hard",
	diffInsane:     "Insance",
	diffImpossible: "Impossible",
}

type iAchievement interface {
	Description() string
	Difficulty() achievementDifficulty
	Check(session *UserInfo)
}

func haveAchievement(u *UserInfo, a iAchievement) bool {
	for _, v := range u.hidden.achievements {
		if v.Description() == a.Description() {
			return true
		}
	}
	return false
}

// render

func renderAchievements(params map[string]string) string {
	// Assertions
	user, ok := params["user"]
	if !ok {
		return "user is not specified. [Go back to home](" + gnominersPath + ") to select a user."
	}
	session, ok := userSessions[user]
	if !ok {
		return "this user is not registered. [Go back to home](" + gnominersPath + ") to register."
	}

	// Content
	content := "# Achievements\n"
	content += "Go back to [DashBoard](" + gnominersDashboard + "?user=" + user + ").\n\n"
	for _, a := range session.hidden.achievements {
		content += "- `" + achievementDifficultyNames[a.Difficulty()] + "` - " + a.Description() + "\n"
	}
	return content
}

// achievements

// a gem

type achMineAGem struct{}

func (a achMineAGem) Description() string {
	return "Mine a gem."
}

func (a achMineAGem) Difficulty() achievementDifficulty {
	return diffEasy
}

func (it *achMineAGem) Check(session *UserInfo) {
	if session.hidden.overallMined >= 1 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// 1000 rocks

type achMine1000Rocks struct{}

func (a achMine1000Rocks) Description() string {
	return "Mine 1000 rocks."
}

func (a achMine1000Rocks) Difficulty() achievementDifficulty {
	return diffMedium
}

func (it *achMine1000Rocks) Check(session *UserInfo) {
	if session.hidden.rockMined >= 1000 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// alexandrite

type achAlexandrite struct{}

func (a achAlexandrite) Description() string {
	return "Mine an alexandrite."
}

func (a achAlexandrite) Difficulty() achievementDifficulty {
	return diffHard
}

func (it *achAlexandrite) Check(session *UserInfo) {
	if session.hidden.alexandrite && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// million

type ach1Million struct{}

func (a ach1Million) Description() string {
	return "Have 1 Million GnoCoins."
}

func (a ach1Million) Difficulty() achievementDifficulty {
	return diffInsane
}

func (it *ach1Million) Check(session *UserInfo) {
	if session.Money >= 1_000_000 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// 42k

type ach42kGems struct{}

func (a ach42kGems) Description() string {
	return "Mine 42k gems."
}

func (a ach42kGems) Difficulty() achievementDifficulty {
	return diffHard
}

func (it *ach42kGems) Check(session *UserInfo) {
	if session.hidden.overallMined >= 42_000 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// bday

type achBDAY struct{}

func (a achBDAY) Description() string {
	return "Mine 42k gems."
}

func (a achBDAY) Difficulty() achievementDifficulty {
	return diffHard
}

func (it *achBDAY) Check(session *UserInfo) {
	if haveAchievement(session, it) {
		return
	}
	pick := pickaxes.BirthdayPickaxe{}
	for _, p := range session.Pickaxes {
		if p.String() == pick.String() {
			session.hidden.achievements = append(session.hidden.achievements, it)
			return
		}
	}
}
