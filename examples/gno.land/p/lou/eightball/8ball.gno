package eightball

import (
	"crypto/sha256"
)

var (
	decisions = []string{
		"It is certain.",
		"It is decidedly so.",
		"Without a doubt.",
		"Yes definitely.",
		"You may rely on it.",
		"As I see it, yes.",
		"Most likely.",
		"Outlook looks good.",
		"Yes.",
		"Signs point to yes.",
		"Reply hazy, try again.",
		"Ask again later.",
		"Better not tell you now.",
		"Cannot predict now.",
		"Concentrate and ask again.",
		"Don't count on it.",
		"My reply is no.",
		"My sources say no.",
		"Outlook not so good.",
		"Very doubtful.",
	}
)

func GetResponse(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	index := int(hash[0]) % len(decisions)
	return decisions[index]
}
