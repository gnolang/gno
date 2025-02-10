package poap

import (
	"strconv"
	"time"
)

// POAPListingPage generates a listing page content for all POAP events.
func POAPListingPage() string {
	events := ListPOAPs()
	if len(events) == 0 {
		return "No POAP events available."
	}
	
	page := "=== POAP Listing Page ===\n\n"
	for _, event := range events {
		page += "ID: " + strconv.FormatUint(event.ID, 10) + "\n"
		page += "Name: " + event.Name + "\n"
		page += "Description: " + event.Description + "\n"
		page += "Date: " + event.Date.Format(time.RFC3339) + "\n"
		page += "Mint Count: " + strconv.Itoa(event.MintCount) + " / " + strconv.Itoa(event.MaxMints) + "\n"
		page += "---------------------------------\n"
	}
	return page
}

func Render(path string) string {
	return POAPListingPage()
} 
