package mail

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/ufmt"
)

type Mail struct {
	id        int
	topic     string
	body      string
	time      time.Time
	sender    std.Address
	recipient std.Address
}

func newMail(topic, body string, sender std.Address, recipient std.Address) *Mail {
	return &Mail{counter + 1, topic, body, time.Now(), sender, recipient}
}

func (mail *Mail) Satisfies(preds []Predicate) bool {
	for _, pred := range preds {
		if !pred.Satisfy(mail) {
			return false
		}
	}
	return true
}

func (mail *Mail) Contains(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(strings.ToLower(mail.topic), s) || strings.Contains(strings.ToLower(mail.body), s)
}

// time key for avl, like 1234567890_<counter>
func (mail *Mail) GetTimeKey() string {
	return ufmt.Sprintf("%d_%d", time.Now().Unix(), counter)
}
