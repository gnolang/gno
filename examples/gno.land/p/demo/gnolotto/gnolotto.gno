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

// Adds a new ticket to the lottery
func (l *Lottery) AddTicket(numbers []int, owner std.Address) {
	l.Tickets = append(l.Tickets, Ticket{Numbers: numbers, Owner: owner})
}

// Conducts the draw by generating 5 random numbers between 1 and 15
func (l *Lottery) Draw() {
	var blockHeight int64 = std.GetHeight()

	l.WinningNumbers = nil
	numbersMap := make(map[int]bool)

	// Add variability to the pseudo-random number generation
	var variabilityFactor int64 = 1

	for len(l.WinningNumbers) < 5 {
		simpleSeed := (blockHeight + variabilityFactor*251) % 233280
		number := int(simpleSeed%15) + 1 // Ensure number is between 1 and 15

		if !numbersMap[number] {
			l.WinningNumbers = append(l.WinningNumbers, number)
			numbersMap[number] = true
		}

		variabilityFactor += 13 // Adjusts for increased variability
	}
}

// Itterate over all tickets to identify and return the addresses of the winners
func (l *Lottery) CheckWinners() []std.Address {
	var winningOwners []std.Address

	for _, ticket := range l.Tickets {
		matchCount := 0

		for _, tNum := range ticket.Numbers {
			for _, wNum := range l.WinningNumbers {
				if tNum == wNum {
					matchCount++
					break
				}
			}
		}

		if matchCount == len(l.WinningNumbers) {
			winningOwners = append(winningOwners, ticket.Owner)
		}
	}

	return winningOwners
}

// Distributes the prize pool equally among the winning ticket owners
func (l *Lottery) PayWinners(winningOwners []std.Address) {
	if len(winningOwners) == 0 {
		return
	} else {
		// Calculate reward per winner
		var reward int64 = l.PrizePool / int64(len(winningOwners))

		banker := std.GetBanker(std.BankerTypeRealmSend)

		for _, owner := range winningOwners {
			send := std.Coins{{"ugnot", reward}}
			banker.SendCoins(std.GetOrigPkgAddr(), owner, send) // Send reward to each winner
		}

		l.PrizePool = 0 // Reset the prize pool after distribution
	}
}
