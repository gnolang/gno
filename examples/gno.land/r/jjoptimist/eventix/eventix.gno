package eventix

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"strconv"
	"time"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/ufmt/v0"
)

type Event struct {
	name         string
	description  string
	date         time.Time
	maxTickets   int
	price        uint64
	paymentToken any
	ticketsSold  int
}

var (
	events              = avl.NewTree()
	eventCounter uint64 = 0
	tickets             = grc721.NewBasicNFT("Event Ticket", "EVTIX")
)

func CreateEvent(_ realm, name, description, dateStr string, paymentToken any, maxTickets int, price uint64) uint64 {
	// validate inputs
	if maxTickets <= 0 {
		panic("Maximum tickets must be greater than 0")
	}

	date, err := time.Parse("2006-01-02T15:04:05Z", dateStr)
	if err != nil {
		panic("Invalid date format. Use: YYYY-MM-DDThh:mm:ssZ")
	}

	switch pt := paymentToken.(type) {
	case string:
		if pt != "ugnot" {
			panic("Unsupported native token")
		}
	case *grc20.Token:
		if pt == nil {
			panic("Invalid GRC20 token")
		}
	default:
		panic("Unsupported payment token type")
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
	chain.Emit("EventCreated", "id", strconv.FormatUint(newID, 10), "name", name)
	return newID
}

func BuyTicket(_ realm, eventId uint64) {
	event, exists := getEvent(eventId)
	if !exists {
		panic("Event does not exist")
	}

	if event.ticketsSold >= event.maxTickets {
		panic("Event is sold out")
	}

	buyer := runtime.PreviousRealm().Address()

	switch pt := event.paymentToken.(type) {
	case string:
		if pt != "ugnot" {
			panic("Unsupported native token")
		}
		if banker.OriginSend().AmountOf("ugnot") != int64(event.price) {
			panic(ufmt.Sprintf("Invalid payment amount: needs to be %dugnot", event.price))
		}
	case *grc20.Token:
		if err := pt.CallerTeller().Transfer(runtime.CurrentRealm().Address(), int64(event.price)); err != nil {
			panic("GRC20 transfer error: " + err.Error())
		}
	default:
		panic("Unsupported payment token type")
	}

	// Mint NFT ticket
	tokenId := grc721.TokenID(ufmt.Sprintf("event_%d_ticket_%d", eventId, event.ticketsSold+1))
	tickets.Mint(buyer, tokenId)

	event.ticketsSold++
	events.Set(strconv.Itoa(int(eventId)), event)
}

func Render(path string) string {
	output := "# Event Ticketing System\n\n"

	pg := pager.NewPager(events, 10, true).MustGetPageByPath(path)

	for _, item := range pg.Items {
		id, _ := strconv.ParseUint(item.Key, 10, 64)
		event := item.Value.(Event)

		output += ufmt.Sprintf("## Event #%d: %s\n", id, event.name)
		output += ufmt.Sprintf("Description: %s\n", event.description)
		output += ufmt.Sprintf("Date: %s\n", event.date.Format(time.RFC3339))
		output += ufmt.Sprintf("Tickets: %d/%d\n", event.ticketsSold, event.maxTickets)
		output += ufmt.Sprintf("Price: %d %v\n\n", event.price, event.paymentToken)

		if event.ticketsSold < event.maxTickets {
			output += ufmt.Sprintf("[Buy Ticket](/r/jjoptimist/eventix/BuyTicket?eventId=%d)\n", id)
		} else {
			output += "**SOLD OUT**\n"
		}
		output += "---\n\n"
	}

	output += pg.Picker(path)

	return output
}

func getEvent(eventId uint64) (Event, bool) {
	value, exists := events.Get(strconv.Itoa(int(eventId)))
	if !exists {
		return Event{}, false
	}
	return value.(Event), true
}
