package eventix

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
)

type Event struct {
	name        string
	description string
	date        time.Time
	maxTickets  int
	price       uint64 // price per ticket (in tokens)
	paymentToken string  // payment token symbol (e.g., ugnot or any GRC20 token)
	ticketsSold int
}

var (
	events       = avl.NewTree()
	eventCounter uint64 = 0
	tickets      = grc721.NewBasicNFT("Event Ticket", "EVTIX")
)

func CreateEvent(name, description string, dateStr string, paymentToken string, maxTickets int, price uint64) uint64 {

	date, err := time.Parse("2006-01-02T15:04:05Z", dateStr)
	if err != nil {
		panic("Invalid date format. Use: YYYY-MM-DDThh:mm:ssZ")
	}

	newID := eventCounter + 1
	event := Event{
		name:         name,
		description:  description,
		date:         date,
		maxTickets:   maxTickets,
		price:        price,
		ticketsSold:  0,
		paymentToken: paymentToken,
	}
	events.Set(strconv.Itoa(int(newID)), event)
	eventCounter = newID
	std.Emit("EventCreated", ufmt.Sprintf("Event %d: %s created", newID, name))
	return newID
}

func getEvent(eventId uint64) (Event, bool) {
	value, exists := events.Get(strconv.Itoa(int(eventId)))
	if !exists {
		return Event{}, false
	}
	return value.(Event), true
}

func BuyTicket(eventId uint64) {
	event, exists := getEvent(eventId)
	if !exists {
		panic("Event does not exist")
	}

	if event.ticketsSold >= event.maxTickets {
		panic("Event is sold out")
	}

	sent := std.GetOrigSend()
	amount := sent.AmountOf(event.paymentToken)
	if amount != int64(event.price) {
		panic(ufmt.Sprintf("Please send exactly %d %s", event.price, event.paymentToken))
	}

	caller := std.PrevRealm().Addr()
	tokenId := grc721.TokenID(strconv.Itoa(int(eventId)) + "-" + strconv.Itoa(event.ticketsSold+1))
	tickets.Mint(caller, tokenId)

	event.ticketsSold++
	events.Set(strconv.Itoa(int(eventId)), event)
	std.Emit("TicketPurchased", ufmt.Sprintf("Ticket %s purchased for Event %d by %s", tokenId, eventId, caller))
}

// all events and ticket info
func Render(path string) string {
	output := "# Event Ticketing System\n\n"

	events.Iterate("", "", func(key string, value interface{}) bool {
		id, _ := strconv.ParseUint(key, 10, 64)
		event := value.(Event)

		output += ufmt.Sprintf("## Event #%d: %s\n", id, event.name)
		output += ufmt.Sprintf("Description: %s\n", event.description)
		output += ufmt.Sprintf("Date: %s\n", event.date.Format(time.RFC3339))
		output += ufmt.Sprintf("Tickets: %d/%d\n", event.ticketsSold, event.maxTickets)
		output += ufmt.Sprintf("Price: %d %s\n\n", event.price, event.paymentToken)

		if event.ticketsSold < event.maxTickets {
			output += ufmt.Sprintf("[Buy Ticket](/r/jjoptimist/eventix/BuyTicket?eventId=%d)\n", id)
		} else {
			output += "**SOLD OUT**\n"
		}
		output += "---\n\n"
		return false
	})

	return output
}
