package gnolotto_factory

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var nonAdmin = testutils.TestAddress("nonAdmin")

func createLottery(t *testing.T, admin std.Address, drawTime int64, prizePool int64) seqid.ID {
	std.TestSetOrigCaller(admin)
	std.TestSetOrigSend(std.Coins{{Denom: "ugnot", Amount: prizePool}}, nil)

	result := CreateLottery(drawTime, prizePool)

	lotteryIDStr := strings.Split(result, " ")[3]

	lotteryID, err := seqid.FromString(lotteryIDStr)
	if err != nil {
		t.Fatalf("Failed to parse lottery ID: %s", lotteryIDStr)
	}

	return lotteryID
}

func TestCreateLottery(t *testing.T) {
	drawTime := time.Now().Unix() + 60
	prizePool := int64(1000)

	lotteryID := createLottery(t, admin, drawTime, prizePool)
	uassert.Equal(t, "0000001", lotteryID.String())

	std.TestSetOrigCaller(nonAdmin)
	uassert.PanicsWithMessage(t, "Only the admin can create a lottery", func() {
		CreateLottery(drawTime, prizePool)
	})
}

func TestBuyTicket(t *testing.T) {
	drawTime := time.Now().Unix() + 60
	prizePool := int64(1000)

	lotteryID := createLottery(t, admin, drawTime, prizePool)
	numbers := "1,2,3,4,5"

	std.TestSetOrigSend(std.Coins{{Denom: "ugnot", Amount: 10}}, nil)

	result := BuyTicket(lotteryID.String(), numbers)
	uassert.Equal(t, "Ticket purchased successfully", result)

	std.TestSetOrigSend(std.Coins{{Denom: "ugnot", Amount: 5}}, nil)
	uassert.PanicsWithMessage(t, "Ticket costs 10 ugnot", func() {
		BuyTicket(lotteryID.String(), numbers)
	})
}

func TestDraw(t *testing.T) {
	drawTime := time.Now().Unix()
	prizePool := int64(1000)
	lotteryID := createLottery(t, admin, drawTime, prizePool)

	result := Draw(lotteryID.String())
	uassert.Equal(t, "Winning numbers drawn successfully", result)
}
