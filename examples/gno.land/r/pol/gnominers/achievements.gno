package gnominers

var achievements = map[string]IAchievement{
	"a gem": &AchMineAGem{},
	"1000rocks": &AchMine1000Rocks{},
	"alexandrite": &AchAlexandrite{},
	"million": &Ach1Million{},
	"42k gems": &Ach42kGems{},
	"bday": &AchBDAY{},
}

type achievementDifficulty uint
const (
	diffEasy achievementDifficulty = iota
	diffMedium
	diffHard
	diffInsane
	diffImpossible
)

var achievementDifficultyNames = map[achievementDifficulty]string {
	diffEasy: 		"Easy",
	diffMedium: 	"Medium",
	diffHard: 		"Hard",
	diffInsane: 	"Insance",
	diffImpossible: "Impossible",
}

type IAchievement interface {
	Description() string
	Difficulty() achievementDifficulty
	Check(session *UserInfo)
}

func CheckAllAchievements(u *UserInfo) {
	for _, v := range achievements {
		v.Check(u)
	}
}

func haveAchievement(u *UserInfo, a IAchievement) bool {
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

type AchMineAGem struct {}

func (a AchMineAGem) Description() string {
	return "Mine a gem."
}

func (a AchMineAGem) Difficulty() achievementDifficulty {
	return diffEasy
}

func (it *AchMineAGem) Check(session *UserInfo) {
	if session.hidden.overallMined >= 1 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// 1000 rocks

type AchMine1000Rocks struct {}

func (a AchMine1000Rocks) Description() string {
	return "Mine 1000 rocks."
}

func (a AchMine1000Rocks) Difficulty() achievementDifficulty {
	return diffMedium
}

func (it *AchMine1000Rocks) Check(session *UserInfo) {
	if session.hidden.rockMined >= 1000 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// alexandrite

type AchAlexandrite struct {}

func (a AchAlexandrite) Description() string {
	return "Mine an alexandrite."
}

func (a AchAlexandrite) Difficulty() achievementDifficulty {
	return diffHard
}

func (it *AchAlexandrite) Check(session *UserInfo) {
	if session.hidden.alexandrite && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// million

type Ach1Million struct {}

func (a Ach1Million) Description() string {
	return "Have 1 Million GnoCoins."
}

func (a Ach1Million) Difficulty() achievementDifficulty {
	return diffInsane
}

func (it *Ach1Million) Check(session *UserInfo) {
	if session.Money >= 1_000_000 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// 42k

type Ach42kGems struct {}

func (a Ach42kGems) Description() string {
	return "Mine 42k gems."
}

func (a Ach42kGems) Difficulty() achievementDifficulty {
	return diffHard
}

func (it *Ach42kGems) Check(session *UserInfo) {
	if session.hidden.overallMined >= 42_000 && !haveAchievement(session, it) {
		session.hidden.achievements = append(session.hidden.achievements, it)
	}
}

// bday

type AchBDAY struct {}

func (a AchBDAY) Description() string {
	return "Mine 42k gems."
}

func (a AchBDAY) Difficulty() achievementDifficulty {
	return diffHard
}

func (it *AchBDAY) Check(session *UserInfo) {
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