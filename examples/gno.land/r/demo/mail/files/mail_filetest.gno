// PKGPATH: gno.land/r/demo/mail_test
package mail_test

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/r/demo/mail"
)

// Send mail, a great movie - SYNOPSIS
// Starring "dude" (the OrigCaller) and "sanders"
// ------------------------------------------------------------------
// Let there be some dude who wants to send mail:
// .dude gets some money, so he can pay for stamps
// .dude sent a mail to sanders
// .now he has a mailbox, so does sanders
// .mail got received by sanders
// .mail also gets saved in the outbox
// .the realm got money from the stamps
// .sanders still has 0gno

// TODO send more emails.
// TODO mail.Id is not checked

func init() {}

func main() {
	var (
		dude    = testutils.TestAddress("dude")
		sanders = testutils.TestAddress("sanders")
		banker  = std.GetBanker(std.BankerTypeReadonly)
	)
	std.TestSetOrigCaller(dude)
	std.TestSetOrigSend(std.Coins{{mail.Fee.Denom, mail.Fee.Amount}}, nil)
	coinsBefore := banker.GetCoins(std.GetOrigPkgAddr()).AmountOf(mail.Fee.Denom)
	coins := std.Coins{{mail.Fee.Denom, 10 * mail.Fee.Amount}}
	std.TestIssueCoins(dude, coins)
	if banker.GetCoins(dude).AmountOf(mail.Fee.Denom) == 0 {
		panic(".dude luckily gets some money, so he can pay for stamps")
	}
	// ------------------------------------------
	mail.SendMail(sanders, "hi", "this is the dude")
	// ------------------------------------------
	if !mail.HasAddressMailbox(dude) {
		panic(".now he has a mailbox")
	}
	if !mail.HasAddressMailbox(sanders) {
		panic(".and sanders also has a mailbox")
	}
	dudebox := mail.MustMailbox(dude)
	sandersbox := mail.MustMailbox(sanders)
	if len(sandersbox.Find(mail.RecipientIs{sanders})) < 1 {
		panic(".mail got received by sanders")
	}
	if len(sandersbox.Find(mail.RecipientIs{sanders}, mail.SenderIs{dude})) < 1 {
		panic(".mail got received by sanders")
	}
	if len(dudebox.Find(mail.SenderIs{dude}, mail.RecipientIs{sanders})) < 1 {
		panic(".mail to sanders is in large outbox, by date")
	}
	if len(sandersbox.ReceivedFrom(dude)) < 1 {
		panic(".mail got received by sanders")
	}
	if len(dudebox.SentTo(sanders)) < 1 {
		panic(".mail also saved in outbox")
	}
	coinsAfter := banker.GetCoins(std.GetOrigPkgAddr()).AmountOf(mail.Fee.Denom)
	if coinsAfter-coinsBefore != mail.Fee.Amount {
		panic(".realm received the expected amount from the stamps: recvd=" + strconv.Itoa(int(coinsAfter-coinsBefore)))
	}
	if banker.GetCoins(sanders).AmountOf(mail.Fee.Denom) != 0 {
		panic(".sanders still has 0gno")
	}
}

// Output:

// Error:
