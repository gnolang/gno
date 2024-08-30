package gnolotto

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
)

var (
	user1 = testutils.TestAddress("user1")
	user2 = testutils.TestAddress("user2")
)

func createMockLottery() *Lottery {
	drawTime := time.Now().Add(24 * time.Hour)
	prizePool := int64(1000)
	return NewLottery(drawTime, prizePool)
}

func TestNewLottery(t *testing.T) {
	drawTime := time.Now().Add(24 * time.Hour)
	prizePool := int64(1000)
	lottery := NewLottery(drawTime, prizePool)

	uassert.True(t, lottery.DrawTime.Equal(drawTime), "Draw time should be set correctly")
	uassert.Equal(t, lottery.PrizePool, prizePool, "Prize pool should be set correctly")
	uassert.Equal(t, len(lottery.Tickets), 0, "Lottery should start with no tickets")
}

func TestAddTicket(t *testing.T) {
	lottery := createMockLottery()
	numbers := []int{1, 2, 3, 4, 5}
	lottery.AddTicket(numbers, user1)

	uassert.Equal(t, len(lottery.Tickets), 1, "Expected 1 ticket")
	uassert.True(t, AreNumberMatching(lottery.Tickets[0].Numbers, numbers), "Ticket numbers should match")
	uassert.Equal(t, lottery.Tickets[0].Owner, user1, "Ticket owner should match")
}

func TestDraw(t *testing.T) {
	lottery := createMockLottery()
	lottery.Draw()

	uassert.Equal(t, len(lottery.WinningNumbers), MaxLottoNumbers, "Expected 5 winning numbers")

	uniqueNumbers := make(map[int]bool)
	for _, number := range lottery.WinningNumbers {
		uassert.True(t, number >= 1 && number <= 15, ufmt.Sprintf("Winning number out of range: %d", number))
		uassert.False(t, uniqueNumbers[number], ufmt.Sprintf("Duplicate winning number found: %d", number))
		uniqueNumbers[number] = true
	}
}

func TestCheckWinners(t *testing.T) {
	lottery := createMockLottery()

	lottery.AddTicket([]int{1, 2, 3, 4, 5}, user1)
	lottery.AddTicket([]int{6, 7, 8, 9, 10}, user2)
	lottery.WinningNumbers = []int{1, 2, 3, 4, 5}

	winners := lottery.CheckWinners()

	uassert.Equal(t, len(winners), 1, "Expected 1 winner")
	uassert.Equal(t, winners[0], user1, "Winner should be user1")
}

func TestStringToIntSlice(t *testing.T) {
	input := "1, 2, 3, 4, 5"
	expected := []int{1, 2, 3, 4, 5}
	result, err := StringToIntSlice(input)

	uassert.NoError(t, err, "Unexpected error")
	uassert.Equal(t, len(result), len(expected), "Slice length should match")

	for i := range expected {
		uassert.Equal(t, result[i], expected[i], ufmt.Sprintf("Expected %d at index %d, got %d", expected[i], i, result[i]))
	}
}

func TestAreNumberMatching(t *testing.T) {
	ticketNumbers := []int{1, 2, 3, 4, 5}
	winningNumbers := []int{1, 2, 3, 4, 5}
	uassert.True(t, AreNumberMatching(ticketNumbers, winningNumbers), "Expected numbers to match")

	winningNumbers = []int{1, 2, 3, 4, 6}
	uassert.False(t, AreNumberMatching(ticketNumbers, winningNumbers), "Expected numbers to not match")
}

func TestPayWinners(t *testing.T) {
	lottery := createMockLottery()
	realmAddress := std.GetOrigPkgAddr()
	std.TestSetOrigCaller(realmAddress)

	winners := []std.Address{user1, user2}
	initialPrizePool := lottery.PrizePool
	expectedReward := initialPrizePool / int64(len(winners))

	lottery.PayWinners(winners)

	uassert.Equal(t, lottery.PrizePool, int64(0), "The prize pool should be reset to 0 after distribution")

	coinsUser1 := std.GetBanker(std.BankerTypeRealmSend).GetCoins(user1)
	receivedRewardUser1 := coinsUser1.AmountOf("ugnot")
	uassert.Equal(t, receivedRewardUser1, expectedReward, ufmt.Sprintf("Winner %s should have received %d but received %d", user1.String(), expectedReward, receivedRewardUser1))

	coinsUser2 := std.GetBanker(std.BankerTypeRealmSend).GetCoins(user2)
	receivedRewardUser2 := coinsUser2.AmountOf("ugnot")
	uassert.Equal(t, receivedRewardUser2, expectedReward, ufmt.Sprintf("Winner %s should have received %d but received %d", user2.String(), expectedReward, receivedRewardUser2))
}
