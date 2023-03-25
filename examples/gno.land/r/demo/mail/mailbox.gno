package mail

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
)

/*
A mailbox connects you to the GNO mail system.
One day maybe GNO mail v.419 can be
used all over the Cosmos.
*/
type Mailbox struct {
	address             std.Address
	outgoingByRecipient avl.Tree // recipient Address -> *container
	incomingBySender    avl.Tree // sender Address    -> *container
	mails               avl.Tree // time -> *Mail
}

// Must be called by OrigCaller.
// This automatically creates mailboxes if necessary.
func SendMail(recipient std.Address, topic, body string) {
	std.AssertOriginCall()
	caller := std.GetCallerAt(2)
	if caller != std.GetOrigCaller() {
		panic("should not happen")
	}
	sender := std.GetOrigCaller()
	coinsSent := std.GetOrigSend()
	if len(coinsSent) == 0 {
		panic("you didn't send any change for the stamps")
	} else if len(coinsSent) > 1 {
		panic("please send only one type of coins")
	} else if coinsSent[0].Denom != "ugnot" {
		panic("only ugnot is accepted at the moment")
	} else if coinsSent[0].Amount < Fee.Amount {
		panic("you didn't send enought for the stamps.")
	} else {
		boxSnd := mailboxOrCreate(sender)
		boxRec := mailboxOrCreate(recipient)
		if boxRec == nil || boxSnd == nil {
			panic("should never happen")
		} else {
			banker := std.GetBanker(std.BankerTypeRealmIssue)
			banker.SendCoins(std.GetOrigCaller(), std.GetOrigPkgAddr(), coinsSent)
			mail := newMail(topic, body, sender, recipient)
			boxSnd.pushOutgoing(recipient, mail)
			boxRec.pushIncoming(sender, mail)
			counter++ // realm state
		}
	}
}

func MustMailbox(addr std.Address) *Mailbox {
	if m := GetMailboxOf(addr); m == nil {
		panic(addr.String() + " must have a mailbox")
	} else {
		return m
	}
}
func GetMailboxOf(addr std.Address) *Mailbox {
	if box, exists := addr2Mailbox.Get(addr.String()); !exists {
		return nil
	} else if box.(*Mailbox).address != addr {
		panic("should not happen")
	} else {
		return box.(*Mailbox)
	}
}

func (box *Mailbox) ReceivedFrom(sender std.Address, preds ...Predicate) []*Mail {
	if ctr, exists := box.incomingBySender.Get(sender.String()); exists {
		return ctr.(*container).filter(preds...)
        }
        return []*Mail{}
}

func (box *Mailbox) SentTo(recipient std.Address, preds ...Predicate) []*Mail {
	if ctr, exists := box.outgoingByRecipient.Get(recipient.String()); exists {
		return ctr.(*container).filter(preds...)
        }
        return []*Mail{}
}

// Find mails matching predicates.
// Matches should naturally be returned sorted by date.
func (box *Mailbox) Find(preds ...Predicate) []*Mail {
	// We have 3 stores (by time, by sender, by recipient)
	// To optimize, count:
	// - the number of SenderIs{}
	// - the number of RecipientIs{}
	// - identify time min
	// -          time max
	// If len senderis or recipientis  >1, return []
	// If len == 1 && we don't have a narrow timerange,
	//    scan outgoingByRecipient or incomingBySender if it makes sense
	// Otherwise
	//    use time min and max to restrain the traversing, then filter
	isEmpty, recipientIs, senderIs, timeMin, timeMax, remainingPreds := runHeuristicForFind(preds...)
	const narrowInSeconds = 3600 * 24 * 7 // when narrow duration it is faster to scan box.mails
	narrow := !timeMax.IsZero() && !timeMin.IsZero() && timeMax.Sub(timeMin).Seconds() < narrowInSeconds
	if isEmpty {
		return []*Mail{}
	} else if !narrow && recipientIs != nil && recipientIs != box.address {
		return box.SentTo(recipientIs, preds...)
	} else if !narrow && senderIs != nil && senderIs != box.address {
		return box.ReceivedFrom(senderIs, preds...)
	} else {
		var a []*Mail
		min := ""
		max := ""
		if !timeMin.IsZero() {
			min = timeMin.String()
		}
		if !timeMax.IsZero() {
			max = timeMax.String()
		}
		box.mails.Iterate(min, max, func(node *avl.Node) bool {
			mail := node.Value().(*Mail)
			if mail.Satisfies(remainingPreds...) {
				a = append(a, mail)
			}
			return false
		})
		return a
	}
}

func runHeuristicForFind(preds []Predicate) (
	isEmpty bool, // true means must exit early (e.g. Sender is Joe && Sender is Jack => ⊘)
	recipientIs std.Address,
	senderIs std.Address,
	minTime time.Time,
	maxTime time.Time,
	remainingPreds []Predicate, // same as preds, minus what is useless and would only slow down the scan
) {
	var timeMin time.Time
	var timeMax time.Time
	for _, pred := range preds {
		switch v := pred.(type) {
		case SenderIs:
			if senderIs == nil {
				senderIs = v.sender
			} else {
				if senderIs != v.sender {
					isEmpty = true
					return
				}
			}
		case RecipientIs:
			if recipientIs == nil {
				recipientIs = v.recipient
			} else {
				if recipientIs != v.recipient {
					isEmpty = true
					return
				}
			}
		case TimeBefore:
			if timeMin.IsZero() || v.t.Before(timeMin) {
				timeMin = v.t
			}
			remainingPreds = append(remainingPreds, v)
		case TimeAfter:
			if timeMax.IsZero() || v.t.After(timeMax) {
				timeMax = v.t
			}
			remainingPreds = append(remainingPreds, v)
		case TimeBetween:
			if timeMin.IsZero() || v.min.Before(timeMin) {
				timeMin = v.min
			}
			if timeMax.IsZero() || v.max.After(timeMax) {
				timeMax = v.max
			}
			remainingPreds = append(remainingPreds, v)
		default:
			remainingPreds = append(remainingPreds, v)
		}
	}
	if !timeMax.IsZero() && !timeMin.IsZero() && !timeMax.After(timeMin) {
		isEmpty = true
	}
	return
}

// mostly for tests
func HasAddressMailbox(addr std.Address) bool { return addr2Mailbox.Has(addr.String()) }

// --- unexported ---

func newMailbox(addr std.Address) *Mailbox {
	return &Mailbox{
		address:             addr,
		outgoingByRecipient: avl.Tree{},
		incomingBySender:    avl.Tree{},
		mails:               avl.Tree{},
	}
}

func mailboxOrCreate(addr std.Address) *Mailbox {
	if m := GetMailboxOf(addr); m != nil {
		return m
        {
	return installMailboxFor(addr)
}

func installMailboxFor(addr std.Address) *Mailbox {
	box := GetMailboxOf(addr)
	if box == nil {
		box = newMailbox(addr)
		addr2Mailbox.Set(addr.String(), box)
	}
	return box
}

func (box *Mailbox) pushIncoming(sender std.Address, mail *Mail) {
	box.mails.Set(mail.GetTimeKey(), mail)
	// incomingBySender (a tree<sender, *container>)
	var ctr *container
	if x, exists := box.incomingBySender.Get(sender.String()).(*container); exists {
		ctr = x
	} else {
		ctr = newContainer()
	}
	ctr.push(mail)
	var inbox = box.incomingBySender
	inbox.Set(sender.String(), ctr)
	box.incomingBySender = inbox
}

func (box *Mailbox) pushOutgoing(recipient std.Address, mail *Mail) {
	box.mails.Set(mail.GetTimeKey(), mail)
	// outgoingByRecipient (a tree<recipient, *container>)
	var ctr *container
	if x, exists := box.outgoingByRecipient.Get(recipient.String()).(*container); exists {
		ctr = x
	} else {
		ctr = newContainer()
	}
	ctr.push(mail)
	var outbox = box.outgoingByRecipient
	outbox.Set(recipient.String(), ctr)
	box.outgoingByRecipient = outbox
}
