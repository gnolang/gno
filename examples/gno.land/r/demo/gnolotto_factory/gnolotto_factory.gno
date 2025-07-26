package gnolotto_factory

import (
	"bytes"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/gnolotto"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

var (
	lotteries   *avl.Tree
	lotteriesID seqid.ID
)

// Replace this address with your address
var admin std.Address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

const TICKET_PRICE = 10

// Initializes the lottery AVL tree
func init() {
	lotteries = avl.NewTree()
}

// Creates a new lottery, only callable by admin.
func CreateLottery(drawTime int64, prizePool int64) string {
	sentCoins := std.GetOrigSend()
	amount := sentCoins.AmountOf("ugnot")

	if prizePool != amount {
		panic("Prize pool must match the transaction value")
	}

	if drawTime < time.Now().Unix() {
		panic("Invalid draw time")
	}

	if std.PrevRealm().Addr() != admin {
		panic("Only the admin can create a lottery")
	}

	lotteryID := lotteriesID.Next()

	lottery := gnolotto.NewLottery(time.Unix(drawTime, 0), prizePool)

	lotteries.Set(lotteryID.Binary(), lottery)

	return ufmt.Sprintf("Lottery with ID: %s created successfully", lotteryID.String())
}

// Buy ticket for a specific lottery.
func BuyTicket(lotteryIDStr string, numbersStr string) string {
	sentCoins := std.GetOrigSend()
	amount := sentCoins.AmountOf("ugnot")

	lotteryID, _ := seqid.FromString(lotteryIDStr)
	id := lotteryID.Binary()

	lotteryRaw, exists := lotteries.Get(id)
	if !exists {
		panic("Lottery not found")
	}

	// Convert string to slice of integers.
	numbers, err := gnolotto.StringToIntSlice(numbersStr)
	if err != nil {
		panic("Invalid number: " + err.Error()) // Error handling if conversion fails.
	}

	str := ufmt.Sprintf("Ticket costs %d ugnot", TICKET_PRICE)
	// Verify if the amount sent is equal to the ticket price.
	if amount != TICKET_PRICE {
		panic(str)
	}

	l, _ := lotteryRaw.(*gnolotto.Lottery)

	if time.Now().Unix() > l.DrawTime.Unix() {
		panic("This lottery has already ended")
	}

	if len(numbers) != gnolotto.MaxLottoNumbers {
		panic("You must select exactly 5 numbers")
	}

	for _, num := range numbers {
		if num > 15 || num < 1 {
			panic("Invalid number, select number range from 1 to 15")
		}
	}

	// Verify if the numbers are unique.
	uniqueNumbers := make(map[int]bool)
	for _, num := range numbers {
		if uniqueNumbers[num] {
			panic("Ticket numbers must be unique")
		}
		uniqueNumbers[num] = true
	}

	l.AddTicket(numbers, std.PrevRealm().Addr())

	return "Ticket purchased successfully"
}

// Draws the winning numbers for a specific lottery, only callable by admin the draw time has passed.
func Draw(lotteryIDStr string) string {
	lotteryID, _ := seqid.FromString(lotteryIDStr)
	id := lotteryID.Binary()

	if std.PrevRealm().Addr() != admin {
		panic("Only the admin can draw the winning numbers")
	}

	lotteryRaw, exists := lotteries.Get(id)

	if !exists {
		panic("Lottery not found")
	}

	l, _ := lotteryRaw.(*gnolotto.Lottery)

	if time.Now().Unix() < l.DrawTime.Unix() {
		panic("Draw time has not passed yet")
	}

	l.Draw()

	return "Winning numbers drawn successfully"
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

		lotteryID, _ := seqid.FromBinary(key)
		readableID := lotteryID.String()

		b.WriteString(
			ufmt.Sprintf(
				"## Lottery ID: *%s*\n",
				readableID,
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
