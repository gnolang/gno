package organizer

import "strconv"

// HangoutInfo is a struct for the hangout set by each friend group
type HangoutInfo struct {
	Id          int
	Date        string
	Time        string
	Place       string
	Description string
	Active      bool
}

// SetHangoutTimeAndPlace creates a hangout after validation from the dao
// Set the new hangout into a dedicated AVL Tree
func (g *FriendGroup) SetHangoutTimeAndPlace(date, time, place, desc string) error {
	id := g.Hangout.Size() + 1
	hangout := HangoutInfo{
		Id:          id,
		Date:        date,
		Time:        time,
		Place:       place,
		Description: desc,
		Active:      true,
	}
	g.Hangout.Set(strconv.Itoa(id), hangout)
	return nil
}
