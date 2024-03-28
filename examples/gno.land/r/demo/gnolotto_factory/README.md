# Write a simple Lottery on Gno.land

## Overview

This guide will demonstrate how to write a simple lottery on GnoLand. We'll cover adding funds to the realm, buying tickets for participation, and finally, distributing the winnings once the winning numbers are drawn. Each step is designed to ensure a smooth, transparent process from the lottery's inception to the awarding the prizepool.

## Lottery functionality

- **Lottery Creation**: Admin can create a lottery specifying the draw time and the prize pool. The amount sent with the transaction must match the prize pool specified.
- **Buying Tickets**: Users can buy tickets by specifying the lottery they want to enter and their chosen numbers. Each ticket costs a fixed amount at 10ugnot, and users can only buy tickets before the draw time.
- **Drawing Winners**: Once the draw time has passed, the admin can draw the winning numbers. This process is handled by the `Draw` function, which selects pseudo-random numbers as winners.
- **Rendering Results**: The `Render` function generates a readable output for the homepage, showing available lotteries, their details, and results if available.

## Package 

```go
package gnolotto

import (
    "std"
    "time"
    "strings"
    "strconv"
)

type Ticket struct {
    Numbers []int // Holds the selected numbers for the lottery ticket
    Owner std.Address // Address of the ticket owner
}

type Lottery struct {
    Tickets []Ticket // All tickets in the lottery
    WinningNumbers []int // Winning numbers after the draw
    DrawTime time.Time // Time of the draw
    PrizePool int64 // Total prize pool amount
}

// Intializes a new lottery instance with a specified draw time and prize pool
func NewLottery(drawTime time.Time, prizePool int64) *Lottery {
    return &Lottery{
        DrawTime: drawTime,
        PrizePool: prizePool,
        Tickets: make([]Ticket, 0),
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
        number := int(simpleSeed % 15) + 1 // Ensure number is between 1 and 15

        if !numbersMap[number] {
            l.WinningNumbers = append(l.WinningNumbers, number)
            numbersMap[number] = true
        }

        variabilityFactor += 13 // Adjusts for increased variability
    }
}

// Itterate over all tickets to identify and return the addresses of the winners
func(l *Lottery) CheckWinners() []std.Address {
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
            banker.SendCoins(std.GetOrigPkgAddr(), owner, send)
        }

        l.PrizePool = 0 // Reset the prize pool after distribution
    }
}
```

A few remarks : 

- In the blockchain world, it's difficult to generate random numbers without using an oracle. Since Gno.land doesn't yet offer an oracle, the `Draw()` function generates random numbers based on the height of the block. This solution is not viable in real-life conditions, but is sufficient for this tutorial.
- In the `PayWinners()` function, we use the `std` package to manipulate the funds available in the realm.

## Realm

```go
package gnolotto_factory

import (
    "bytes"
    "time"
    "strconv"
    "strings"
    "std"

    "gno.land/p/demo/avl"
    "gno.land/p/demo/ufmt"
    "gno.land/p/demo/gnolotto"
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
```

A few remarks :

- The `Draw()` function generates 5 winning numbers. A ticket purchase must be accompanied by a selection of 5 numbers in order to participate. 
- In the `BuyTicket()` function, we take as arguments the winning numbers in string type, as it's not possible to pass a slice as an argument in  `gnokey`. We therefore retrieve the winning numbers in string type, split them and convert them to slice to add them to our `Ticket` struct in our package
- When we make a function call using `gnokey` and add an amount in `-send`, this amount will be sent to the realm even if a condition does not allow the action in our code. This is why, in the `CreateLottery()` and `BuyTicket()` functions, we use the `std` package to refund the wallet that sent the funds in the event that a condition is not met.
- For this lottery, we have chosen to set the price of a ticket at 10ugnot. If the user buys a ticket and sends + or - 10ugnot, he will be refunded the amount sent. At the end of the lottery creation process, we check that the amount sent to the realm is equal to the amount defined in the prize pool. Sending the amount to the realm when the lottery is created allows us to distribute the winnings to the winner(s) automatically after the draw.

## Render

And finally, our Render() function, which displays our lottery.

```go
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
```

Congratulations, your lottery has been successfully created 🥳 ! Below you'll find the commands for using this lottery with `gnokey`.

**Create a new Lottery (Admin) :**
```
gnokey maketx call -pkgpath "gno.land/r/demo/gnolotto_factory" -func "CreateLottery" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "10000ugnot" -broadcast -chainid "dev" -args "1711487446" -args "10000" -remote "tcp://127.0.0.1:36657" test1
```
*The first argument corresponds to the date and time of the draw run, in unix format*
*The second is the prize pool amount, so don't forget to put the same amount in `-send`.*

**Buy a ticket :**
```
gnokey maketx call -pkgpath "gno.land/r/demo/gnolotto_factory" -func "BuyTicket" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "10ugnot" -broadcast -chainid "dev" -args "0" -args "1,2,3,4,5" -remote "tcp://127.0.0.1:36657" test1
```
*The first argument corresponds to the ID of Lottery*
*The second arguments corresponds to the lottery participation numbers*
*Don't forget to add 10ugnot to `-send`, which corresponds to the price of a ticket.*

**Drawing (Admin) :**

```
gnokey maketx call -pkgpath "gno.land/r/demo/gnolotto_factory" -func "Draw" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "0" -remote "tcp://127.0.0.1:36657" test1
```

*The argument corresponds to the ID of the lottery for which you wish to perform the draw. (Don't forget that you can't make a draw until the date defined at creation has passed.)*
