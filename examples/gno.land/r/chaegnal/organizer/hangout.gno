package organizer

// HangoutInfo is a struct for the hangout set by each friend group
type HangoutInfo struct {
	Id          string
	Date        string
	Time        string
	Place       string
	Description string
	Active      bool
}

// SetHangoutTimeAndPlace creates a hangout after validation from the dao
// Set the new hangout into a dedicated AVL Tree
func (g *FriendGroup) SetHangoutTimeAndPlace(date, time, place, desc string) error {
	id := g.nextHangoutID.Next()

	hangout := HangoutInfo{
		Id:          id.String(),
		Date:        date,
		Time:        time,
		Place:       place,
		Description: desc,
		Active:      true,
	}

	g.Hangout.Set(id.String(), hangout)
	return nil
}
