package gnolotto

import (
	"std"
	"strconv"
	"strings"
	"time"
)

type Ticket struct {
	Numbers []int       // Holds the selected numbers for the lottery ticket
	Owner   std.Address // Address of the ticket owner
}

type Lottery struct {
	Tickets        []Ticket  // All tickets in the lottery
	WinningNumbers []int     // Winning numbers after the draw
	DrawTime       time.Time // Time of the draw
	PrizePool      int64     // Total prize pool amount
}

// Intializes a new lottery instance with a specified draw time and prize pool
func NewLottery(drawTime time.Time, prizePool int64) *Lottery {
	return &Lottery{
		DrawTime:  drawTime,
		PrizePool: prizePool,
		Tickets:   make([]Ticket, 0),
	}
}

const MaxLottoNumbers = 5

// Adds a new ticket to the lottery
func (l *Lottery) AddTicket(numbers []int, owner std.Address) {
	l.Tickets = append(l.Tickets, Ticket{Numbers: numbers, Owner: owner})
}

// Conducts the draw by generating 5 pseudo-random numbers between 1 and 15 inclusive
func (l *Lottery) Draw() {
	var blockHeight int64 = std.GetHeight()

	l.WinningNumbers = nil
	numbersMap := make(map[int]bool)

	// Add variability to the pseudo-random number generation
	var variabilityFactor int64 = 1

	for len(l.WinningNumbers) < MaxLottoNumbers {
		simpleSeed := (blockHeight + variabilityFactor*251) % 233280
		number := int(simpleSeed%15) + 1 // Ensure number is between 1 and 15

		if !numbersMap[number] {
			l.WinningNumbers = append(l.WinningNumbers, number)
			numbersMap[number] = true
		}

		variabilityFactor += 13 // Adjusts for increased variability
	}
}

func (l *Lottery) CheckWinners() []std.Address {
	var winners []std.Address

	for _, ticket := range l.Tickets {
		if AreNumberMatching(ticket.Numbers, l.WinningNumbers) {
			winners = append(winners, ticket.Owner)
		}
	}

	return winners
}

// Distributes the prize pool equally among the winning ticket owners
func (l *Lottery) PayWinners(winners []std.Address) {
	if len(winners) == 0 {
		return
	}

	var reward int64 = l.PrizePool / int64(len(winners))

	banker := std.GetBanker(std.BankerTypeRealmSend)

	for _, owner := range winners {
		send := std.Coins{{"ugnot", reward}}
		banker.SendCoins(std.GetOrigPkgAddr(), owner, send)
	}

	l.PrizePool = 0
}

func StringToIntSlice(numbersStr string) ([]int, error) {
	numbersSlice := strings.Split(numbersStr, ",")
	numbers := make([]int, len(numbersSlice))
	for i, numStr := range numbersSlice {
		num, err := strconv.Atoi(strings.TrimSpace(numStr))
		if err != nil {
			return nil, err
		}
		numbers[i] = num
	}

	return numbers, nil
}

func AreNumberMatching(ticketNumbers, winningNumbers []int) bool {
	for i := 0; i < MaxLottoNumbers; i++ {
		if ticketNumbers[i] != winningNumbers[i] {
			return false
		}
	}

	return true
}
