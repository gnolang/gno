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
	price       uint64 // in gnot
	ticketsSold int
}

var (
	events       = avl.NewTree()
	eventCounter uint64 = 0
	tickets      = grc721.NewBasicNFT("Event Ticket", "EVTIX")
)

func CreateEvent(name, description string, dateStr string, maxTickets int, price uint64) uint64 {

	date, err := time.Parse("2006-01-02T15:04:05Z", dateStr)
	if err != nil {
		panic("Invalid date format. Use: YYYY-MM-DDThh:mm:ssZ")
	}

	newID := eventCounter + 1
	event := Event{
		name:        name,
		description: description,
		date:        date,
		maxTickets:  maxTickets,
		price:       price,
		ticketsSold: 0,
	}
	events.Set(strconv.FormatUint(newID, 10), event)
	eventCounter = newID
	return newID
}

func getEvent(eventId uint64) (Event, bool) {
	value, exists := events.Get(strconv.FormatUint(eventId, 10))
	if !exists {
		return Event{}, false
	}
	return value.(Event), true
}

func setEvent(eventId uint64, event Event) {
	events.Set(strconv.FormatUint(eventId, 10), event)
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
	amount := sent.AmountOf("ugnot")
	if amount != int64(event.price) {
		panic(ufmt.Sprintf("Please send exactly %d ugnot", event.price))
	}

	caller := std.PrevRealm().Addr()
	tokenId := grc721.TokenID(strconv.FormatUint(eventId, 10) + "-" + strconv.Itoa(event.ticketsSold+1))
	tickets.Mint(caller, tokenId)

	event.ticketsSold++
	setEvent(eventId, event)
	// Emit a "TicketPurchased" event with details about the purchase.
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
		output += ufmt.Sprintf("Date: %s\n", event.date.Format("2006-01-02 15:04:05"))
		output += ufmt.Sprintf("Tickets: %d/%d\n", event.ticketsSold, event.maxTickets)
		output += ufmt.Sprintf("Price: %d ugnot\n\n", event.price)

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
