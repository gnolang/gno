package gnominers

var achievements = map[string]iAchievement{
	"a gem":       &achMineAGem{},
	"dao":         &achDAO{},
	"1000rocks":   &achMine1000Rocks{},
	"alexandrite": &achAlexandrite{},
	"million":     &ach1Million{},
	"42k gems":    &ach42kGems{},
	"bday":        &achBDAY{},
	"diamonds":    &achDiamonds{},
	"billion":    &ach1Billion{},
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
	diffInsane:     "Insane",
	diffImpossible: "Impossible",
}

type iAchievement interface {
	Description() string
	Difficulty() achievementDifficulty
	Check(session *UserSession)
}

func haveAchievement(u *UserSession, a iAchievement) bool {
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

func (it *achMineAGem) Check(session *UserSession) {
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

func (it *achMine1000Rocks) Check(session *UserSession) {
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

func (it *achAlexandrite) Check(session *UserSession) {
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

func (it *ach1Million) Check(session *UserSession) {
	if session.Money >= 1_000_000 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// billion

type ach1Billion struct{}

func (a ach1Billion) Description() string {
	return "Have 1 Million GnoCoins."
}

func (a ach1Billion) Difficulty() achievementDifficulty {
	return diffImpossible
}

func (it *ach1Billion) Check(session *UserSession) {
	if session.Money >= 1_000_000 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// dao

type achDAO struct{}

func (a achDAO) Description() string {
	return "Open your own DAO."
}

func (a achDAO) Difficulty() achievementDifficulty {
	return diffEasy
}

func (it *achDAO) Check(session *UserSession) {
	if session.DAO != nil && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// 42k

type ach42kGems struct{}

func (a ach42kGems) Description() string {
	return "Mine 4200 gems."
}

func (a ach42kGems) Difficulty() achievementDifficulty {
	return diffHard
}

func (it *ach42kGems) Check(session *UserSession) {
	if session.hidden.overallMined >= 4_200 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// diamonds

type achDiamonds struct{}

func (a achDiamonds) Description() string {
	return "Mine your first diamond."
}

func (a achDiamonds) Difficulty() achievementDifficulty {
	return diffMedium
}

func (it *achDiamonds) Check(session *UserSession) {
	if session.GemsCount[gemDiamond] > 0 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// bday

type achBDAY struct{}

func (a achBDAY) Description() string {
	return "Get the birthday pickaxe."
}

func (a achBDAY) Difficulty() achievementDifficulty {
	return diffImpossible
}

func (it *achBDAY) Check(session *UserSession) {
	if haveAchievement(session, it) {
		return
	}
	pick := BirthdayPickaxe{}
	for _, p := range session.Pickaxes {
		if p.String() == pick.String() {
			session.hidden.achievements = append(session.hidden.achievements, it)
			return
		}
	}
}
