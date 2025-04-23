package card

import (
	"strings"

	"gno.land/p/lou/ascii"
	"gno.land/p/moul/md"
)

type Card struct {
	header          string
	hWidthAlignment ascii.WAlignment

	content          []string
	cWidthAlignment  ascii.WAlignment
	cHeightAlignment ascii.HAlignment

	width  uint64
	height uint64

	footer *string
	link   *string
}

func (card Card) Header() string {
	return card.header
}

func (card Card) Content() []string {
	return card.content
}

func (card *Card) SetHeader(header string) {
	card.header = header
}

func (card *Card) SetContent(content []string) {
	card.content = content
}

func (card *Card) AppendContent(line string) {
	card.content = append(card.content, line)
}

func (card *Card) SetHeight(height uint64) {
	card.height = height
}

func (card *Card) SetWidth(width uint64) {
	maxW := uint64(len(card.header))
	if maxW > width && width != 0 {
		width = maxW
	}

	card.width = width
}

func (card *Card) SetHeaderAlignmentW(
	width ascii.WAlignment,
) {
	card.hWidthAlignment = width
}

func (card *Card) SetContentAlignmentW(
	width ascii.WAlignment,
) {
	card.cWidthAlignment = width
}

func (card *Card) SetContentAlignmentH(
	height ascii.HAlignment,
) {
	card.cHeightAlignment = height
}

// word-based
func contentRender(
	lines []string,
	wAlign ascii.WAlignment,
	width, height int,
	hAlign ascii.HAlignment,
) string {
	var wrapped []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			wrapped = append(wrapped, "")
			continue
		}

		words := strings.Fields(line)
		current := ""

		for j, word := range words {
			if len(current)+len(word)+1 > width {
				wrapped = append(wrapped, current)
				current = word
			} else {
				if current == "" {
					current = word
				} else {
					current += " " + word
				}
			}
			if j == len(words)-1 && current != "" {
				wrapped = append(wrapped, current)
			}
		}
	}

	wrapped = ascii.PadHeight(wrapped, height, hAlign)

	body := ""
	for _, line := range wrapped {
		body += ascii.PadLine(line, width, wAlign, "⠀") + "\n"
	}

	return body
}

func (card Card) Render() string {
	header := ascii.PadLine(card.header, int(card.width)/3, card.hWidthAlignment, "⠀")
	formattedContent := contentRender(card.content, card.cWidthAlignment, int(card.width), int(card.height), card.cHeightAlignment)

	out := md.H1(header) + "\n\n" + formattedContent + "\n\n"
	return out
}

func NewCard(
	header string,
	content []string,
	width,
	height uint64,
) *Card {
	maxW := uint64(len(header))
	if maxW > width && width != 0 {
		width = maxW
	}

	return &Card{
		header:          header,
		hWidthAlignment: ascii.AlignLeft,

		content:          content,
		cWidthAlignment:  ascii.AlignLeft,
		cHeightAlignment: ascii.AlignTop,

		width:  width,
		height: height,
	}
}
