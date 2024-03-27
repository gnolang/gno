package gnolotto_factory

import (
	"bytes"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/gnolotto"
	"gno.land/p/demo/ufmt"
)

var lotteries *avl.Tree

// Replace this address with your address
var admin std.Address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

// Initializes the lottery AVL tree
func init() {
	lotteries = avl.NewTree()
}

// Creates a new lottery, only callable by admin.
func CreateLottery(drawTime int64, prizePool int64) (int, string) {
	sentCoins := std.GetOrigSend()
	amount := sentCoins.AmountOf("ugnot")
	banker := std.GetBanker(std.BankerTypeRealmSend)
	send := std.Coins{{"ugnot", int64(amount)}}

	if prizePool != amount {
		banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
		return -1, "Prize pool must match the transaction value"
	}

	if drawTime < time.Now().Unix() {
		banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
		return -1, "Invalid draw time"
	}

	if std.GetOrigCaller() != admin {
		banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
		return -1, "Only the admin can create a lottery"
	}

	lotteryID := lotteries.Size()
	lottery := gnolotto.NewLottery(time.Unix(drawTime, 0), prizePool)

	lotteries.Set(ufmt.Sprintf("%d", lotteryID), lottery)
	return lotteryID, "Lottery created successfully"
}

// Buy ticket for a specific lottery.
func BuyTicket(lotteryID int, numbersStr string) (int, string) {
	sentCoins := std.GetOrigSend()
	amount := sentCoins.AmountOf("ugnot")
	banker := std.GetBanker(std.BankerTypeRealmSend)
	send := std.Coins{{"ugnot", int64(amount)}}

	id := ufmt.Sprintf("%d", lotteryID)
	lotteryRaw, exists := lotteries.Get(id)
	if !exists {
		banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
		return -1, "Lottery not found"
	}

	// Convert string to slice of integers.
	numbersSlice := strings.Split(numbersStr, ",")
	numbers := make([]int, len(numbersSlice))
	for i, numStr := range numbersSlice {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
			panic("Invalid number: " + err.Error())
		}
		numbers[i] = num
	}

	//Verify if the amount sent is equal to the ticket price.
	if amount != 10 {
		banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
		return -1, "Ticket price must be 10 UGNOT"
	}

	// Verify if the numbers are unique.
	uniqueNumbers := make(map[int]bool)
	for _, num := range numbers {
		if uniqueNumbers[num] {
			banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
			return -1, "Numbers must be unique"
		}
		uniqueNumbers[num] = true
	}

	l, _ := lotteryRaw.(*gnolotto.Lottery)

	if time.Now().Unix() > l.DrawTime.Unix() {
		banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
		return -1, "This lottery has already ended"
	}

	if len(numbers) > 5 || len(numbers) < 5 {
		banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
		return -1, "You must select exactly 5 numbers"
	}

	for _, num := range numbers {
		if num > 15 || num < 1 {
			banker.SendCoins(std.GetOrigPkgAddr(), std.GetOrigCaller(), send)
			return -1, "Invalid number, select number range from 1 to 15"
		}
	}

	caller := std.GetOrigCaller()
	l.AddTicket(numbers, caller)

	return 1, "Ticket purchased successfully"
}

// Draws the winning numbers for a specific lottery, only callable by admin the draw time has passed.
func Draw(lotteryID int) (int, string) {
	id := ufmt.Sprintf("%d", lotteryID)
	if std.GetOrigCaller() != admin {
		return -1, "Only the admin can draw the winning numbers"
	}
	lotteryRaw, exists := lotteries.Get(id)
	if !exists {
		return -1, "Lottery not found"
	}

	l, _ := lotteryRaw.(*gnolotto.Lottery)

	if time.Now().Unix() < l.DrawTime.Unix() {
		return -1, "Draw time has not passed yet"
	}

	l.Draw()
	return 1, "Winning numbers drawn successfully"
}

func Render(path string) string {
	if path == "" {
		return renderHomepage()
	}

	return "unknown page"
}

func renderHomepage() string {
	var b bytes.Buffer

	b.WriteString("# Welcome to GnoLotto\n\n")

	if lotteries.Size() == 0 {
		b.WriteString("### *No lotteries available currently!*\n")
		return b.String()
	}

	lotteries.Iterate("", "", func(key string, value interface{}) bool {
		l := value.(*gnolotto.Lottery)

		b.WriteString(
			ufmt.Sprintf(
				"## Lottery ID: *%s*\n",
				key,
			),
		)

		b.WriteString(
			ufmt.Sprintf(
				"Draw Time: *%s*\n",
				l.DrawTime.Format("Mon Jan _2 15:04:05 2006"),
			),
		)

		b.WriteString(
			ufmt.Sprintf(
				"Prize Pool: *%d* UGNOT\n\n",
				l.PrizePool,
			),
		)

		if time.Now().Unix() > l.DrawTime.Unix() {
			// If the lottery has ended, display the winners.
			var numbersStr string
			for i, number := range l.WinningNumbers {
				if i > 0 {
					numbersStr += ", "
				}
				numbersStr += ufmt.Sprintf("%d", number)
			}
			b.WriteString(ufmt.Sprintf("- Winning numbers [%s]\n\n", numbersStr))

			winners := l.CheckWinners()

			l.PayWinners(winners)

			if len(winners) > 0 {
				b.WriteString("Winners:\n\n")
				for _, winner := range winners {
					b.WriteString(ufmt.Sprintf("*%s*\n\n", winner.String()))
				}
			} else {
				b.WriteString("*No winners for this lottery.*\n")
			}
		} else {
			// If the lottery is still ongoing, display the participants.
			if len(l.Tickets) > 0 {
				b.WriteString("Participants:\n")
				for _, ticket := range l.Tickets {
					// Initialise string for displaying numbers
					var numbersStr string
					for i, number := range ticket.Numbers {
						if i > 0 {
							numbersStr += ", "
						}
						numbersStr += ufmt.Sprintf("%d", number)
					}
					b.WriteString(ufmt.Sprintf("- *%s* with numbers [%s]\n", ticket.Owner.String(), numbersStr))
				}
			} else {
				b.WriteString("*No participants yet.*\n")
			}
		}

		b.WriteString("\n")
		return false
	})

	banker := std.GetBanker(std.BankerTypeReadonly)
	contractAddress := std.GetOrigPkgAddr()
	coins := banker.GetCoins(contractAddress)

	b.WriteString("## Contract Balance:\n")
	b.WriteString(coins.String() + "\n\n")

	return b.String()
}
