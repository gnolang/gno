package poap

import (
	"strconv"
	"time"

	"gno.land/p/moul/md"
)

const totalWidth int = 62
const interiorWidth int = totalWidth - 2 // 60 characters available

func POAPListingPage() string {
	events := ListPOAPs()

	globalHeader := centerText("🎟️ POAP Minting Dashboard", totalWidth) + "\n" +
		centerText("Explore, view, and mint your exclusive POAP events", totalWidth) + "\n\n"
	
	querySection := md.H4(md.Link("Query POAPs Owned by a Specific Address", "https://gno.studio/connect/view/gno.land/r/jjoptimist/poap/GetPOAPsByOwner?args=g16vfw3r7zuz43fhky3xfsuc2hdv9tnhvlkyn0nj")) + "\n" +
	                md.H4(md.Link("Retrieve All NFT Tokens for a Specific POAP Event", "https://gno.studio/connect/view/gno.land/r/jjoptimist/poap/GetPOAPTokens?args=1")) + "\n\n"
	
	globalHeader = globalHeader + querySection

	globalFooter := "\n" + centerText("🏅 Proof of Attendance Protocol 🏅", totalWidth) + "\n"

	eventContent := ""
	for _, event := range events {
		eventTopBorder := "╔" + repeat("═", interiorWidth) + "╗\n"
		eventBottomBorder := "╚" + repeat("═", interiorWidth) + "╝\n"

		line1 := padRight("🆔 POAP: " + strconv.FormatUint(event.ID, 10), interiorWidth) + "\n"
		line2 := padRight("🎫 Name: " + event.Name, interiorWidth) + "\n"
		line3 := padRight("📝 Desc: " + event.Description, interiorWidth) + "\n"
		line4 := padRight("🗓 Date: " + event.Date.Format(time.RFC3339), interiorWidth) + "\n"
		line5 := padRight("🖨 Mint Count: " + strconv.Itoa(event.MintCount) + " / " + strconv.Itoa(event.MaxMints), interiorWidth) + "\n"

		var actionContent string
		if event.MintCount < event.MaxMints {
			actionContent =  md.Link("[Mint Now 🛒]", "https://gno.studio/connect/view/gno.land/r/jjoptimist/poap/MintPOAP?poapId=" + strconv.FormatUint(event.ID, 10))
		} else {
			actionContent = "🚫 Sold Out"
		}
		actionLine := centerText(actionContent, interiorWidth) + "\n"

		eventBlock := eventTopBorder + line1 + line2 + line3 + line4 + line5 + actionLine + "\n" + eventBottomBorder + "\n"
		eventContent += eventBlock
	}

	content := globalHeader + eventContent + globalFooter

	return content
}

func Render(path string) string {
	return POAPListingPage()
} 
