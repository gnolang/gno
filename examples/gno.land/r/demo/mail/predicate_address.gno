package mail

import "std"

type SenderIs struct{ sender std.Address }
type RecipientIs struct{ recipient std.Address }

func (x SenderIs) Satisfy(mail *Mail) bool    { return mail.sender == x.sender }
func (x RecipientIs) Satisfy(mail *Mail) bool { return mail.recipient == x.recipient }
