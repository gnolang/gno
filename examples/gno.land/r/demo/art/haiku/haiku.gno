package haiku

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
)

var (
	// URL for the realm
	pageURL = "/r/demo/art/haiku"
	// regex for sanitizer
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z ]+`)
	// definitions for the NFT
	nftName   = "TextNFT"
	nftSymbol = "NFTEXT"
	// storage for the fnts
	nfts *grc721.BasicNFT
	// storage for the syllable dictionary
	wordsWithSyllables avl.Tree
	// storage for the haikus
	nftDB avl.Tree // tokenID (string) -> haikuAmulet
)

// errors
var (
	ErrNotEnglish          error  = errors.New("non-english word detected")
	ErrIncorrectLineNumber error  = errors.New("incorrect number of lines")
	ErrIncorrectSyllables  error  = errors.New("incorrect number of syllables")
	StatusNotFound         string = "404"
)

// init initializes the NFT, the database, and the dictionary
// of syllables
func init() {
	nfts = grc721.NewBasicNFT(nftName, nftSymbol)
	nftDB = avl.Tree{}
	// load dicitionary of words
	wordsWithSyllables = avl.Tree{}
	for _, line := range strings.Split(rawwordlist, "\n") {
		line = strings.TrimSpace(line)
		word := strings.Replace(line, ";", "", -1)
		wordsWithSyllables.Set(word, strings.Count(line, ";")+1)
	}
}

// sanitizeString will remove puncuation and extra spaces that might
// be used to "game" the contract.
func sanitizeString(unclean string) string {
	cleaned := ""
	for _, line := range strings.Split(unclean, "\n") {
		line = strings.Join(strings.Fields(nonAlphanumericRegex.ReplaceAllString(line, "")), " ")
		if line != "" {
			cleaned += line + "\n"
		}
	}
	cleaned = strings.ToLower(cleaned)
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// countSyllables will count the syllables in a line
// and return an error if a non-english word is detected
func countSyllables(line string) (numSyllables int, err error) {
	for _, word := range strings.Fields(line) {
		v, ok := wordsWithSyllables.Get(word)
		if !ok {
			v, ok = wordsWithSyllables.Get(word + "s")
			if !ok {
				v, ok = wordsWithSyllables.Get(strings.TrimSuffix(word, "s"))
				if !ok {
					v, ok = wordsWithSyllables.Get(strings.TrimSuffix(word, "ed"))
					if !ok {
						err = errors.New(ufmt.Sprintf("'%s' is not an english word", word))
						return
					} else {
						numSyllables += v.(int)
					}
				} else {
					numSyllables += v.(int)
				}
			} else {
				numSyllables += v.(int)
			}
		} else {
			numSyllables += v.(int)
		}
	}
	return
}

// isHaiku determines whether the text contains
// three lines and syllables of each line is
// 5-7-5. the text is assumed to be sanitized.
func isHaiku(text string) (err error) {
	text = strings.Replace(text, `\n`, "\n", -1)
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) != 3 {
		return errors.New(ufmt.Sprintf("incorrect number of lines: %d != 3", len(lines)))
	}
	v, err := countSyllables(lines[0])
	if err != nil {
		return
	}
	if v != 5 {
		return errors.New(ufmt.Sprintf("incorrect number of syllables: %d != 5 for '%s'", v, lines[0]))
	}
	v, err = countSyllables(lines[1])
	if err != nil {
		return
	}
	if v != 7 {
		return errors.New(ufmt.Sprintf("incorrect number of syllables: %d != 7 for '%s'", v, lines[1]))
	}
	v, err = countSyllables(lines[2])
	if err != nil {
		return
	}
	if v != 5 {
		return errors.New(ufmt.Sprintf("incorrect number of syllables: %d != 5 for '%s'", v, lines[2]))
	}
	return
}

// calculateRarity creates an arbitrary definition of scarcity.
// The number of sequential 8s in the hash determines the rarity of the haiku:
//
//	8888: common
//	88888: uncommon
//	888888: rare
//	8888888: epic
//	88888888: legendary
//	888888888: mythic
//	8888888888+: ???
//
// https://archive.is/IADSf
func calculateRarity(s string) string {
	res := 0
	resBest := 0
	for _, v := range s {
		if string(v) == "8" {
			res++
			if res > resBest {
				resBest = res
			}
		} else {
			res = 0
		}
	}
	if resBest <= 4 {
		return "common"
	} else if resBest == 5 {
		return "uncommon"
	} else if resBest == 6 {
		return "rare"
	} else if resBest == 7 {
		return "epic"
	} else if resBest == 8 {
		return "legendary"
	} else {
		return "mythic"
	}
}

// haikuAmulet keeps track of the haiku and its metadata
type haikuAmulet struct {
	Text      string
	TokenID   string
	Rarity    string
	CreatedAt time.Time
	CreatedBy std.Address
}

// String renders the haiku to markdown
func (ha *haikuAmulet) String() string {
	tokenid := ha.TokenID
	for _, s := range []string{"888888888", "88888888", "8888888", "888888", "88888", "8888", "888", "88", "8"} {
		if strings.Contains(tokenid, s) {
			tokenid = strings.Replace(tokenid, s, "***"+s+"***", 1)
			break
		}
	}
	render := ufmt.Sprintf("### [%s](%s:%s)\n\n", tokenid, pageURL, ha.TokenID)
	render += "```\n" + ha.Text + "\n```\n\n"
	render += ufmt.Sprintf("[rarity](%s:rarity): **", pageURL) + ha.Rarity + "**\n\n"
	render += "*created on " + ha.CreatedAt.Format(time.RFC1123) + " "
	render += "by " + renderUserAddress1(ha.CreatedBy, "createdby") + ", "
	owner, err := nfts.OwnerOf(grc721.TokenID(ha.TokenID))
	if err == nil {
		render += "collected by " + renderUserAddress1(owner, "collectedby")
	}
	render += "*"
	render += "\n\n"
	return render
}

// Transfer allows the caller to transfer a haiku they own
// to a recipient.
func Transfer(to std.Address, tokenid string) string {
	caller := std.GetOrigCaller()
	value, found := nftDB.Get(tokenid)
	if !found {
		return ufmt.Sprintf("error: token ID '%s' not found", tokenid)
	}
	ha := value.(*haikuAmulet)

	// get the owner
	owner, err := nfts.OwnerOf(grc721.TokenID(ha.TokenID))
	if err != nil {
		return ufmt.Sprintf("error: owner for ID '%s' not found", tokenid)
	}

	// raise error if the owner is not the caller
	if caller != owner {
		return ufmt.Sprintf("error: caller '%s' is not owner '%s'", caller, owner)
	}

	// throws if the to address is empty
	if to == "" {
		return ufmt.Sprintf("error: to address cannot be empty", caller, owner)
	}

	// do the transfer
	nfts.TransferFrom(caller, to, grc721.TokenID(ha.TokenID))
	return ufmt.Sprintf("transfered '%s' from '%s' to '%s'", tokenid, caller.String(), to.String())
}

// Mint creates a Haiku, if it passes the basic
// criteria:
//   - only english words (dictionary in wordlist.gno)
//   - three lines
//   - each line has 5, 7, and 5 syllables respectively
//   - haiku is unique
func Mint(text string) string {
	// basic sanitiziation
	text = strings.Replace(text, `\n`, "\n", -1)
	text = strings.TrimSpace(text)
	// under-the-hood sanitation
	cleanedText := sanitizeString(text)
	if err := isHaiku(cleanedText); err != nil {
		return "error: " + err.Error()
	}

	var buf bytes.Buffer
	buf.Write([]byte(cleanedText))
	hash := sha256.Sum256(buf.Bytes())
	hashHex := hex.EncodeToString(hash[:])
	rarity := calculateRarity(hashHex)
	ha := haikuAmulet{
		Text:      text,
		TokenID:   hashHex,
		Rarity:    rarity,
		CreatedAt: time.Now(),
		CreatedBy: std.GetOrigCaller(),
	}

	err := nfts.Mint(std.GetOrigCaller(), grc721.TokenID(ha.TokenID))
	if err != nil {
		return "error: " + err.Error()
	}

	nftDB.Set(ha.TokenID, &ha)
	return ha.TokenID
}

// renderUserAddress1 will render a user address to point to the
// haikus that they own.
func renderUserAddress1(addr std.Address, prefix string) string {
	if u := users.GetUserByAddress(addr); u != nil {
		return ufmt.Sprintf("[%s](%s:%s:%s)", u.Name(), pageURL, prefix, addr.String())
	}
	return ufmt.Sprintf("[%s](%s:%s:%s)", addr.String(), pageURL, prefix, addr.String())
}

// renderUserAddress2 will render a user address to point to the
// user page generated from the users domain.
func renderUserAddress2(addr std.Address) string {
	if u := users.GetUserByAddress(addr); u != nil {
		return ufmt.Sprintf("[%s](%s:%s)", u.Name(), "/r/demo/users", addr.String())
	}
	return addr.String()
}

// renderAll renders all of the haikus
func renderAll() string {
	render := "# haikus\n\n"
	nftDB.Iterate("", "", func(key string, value interface{}) bool {
		ha := value.(*haikuAmulet)
		render += ha.String()
		return false
	})
	return render
}

// renderSingle renders a single haiku
func renderSingle(tokenid string) string {
	value, found := nftDB.Get(tokenid)
	if !found {
		return StatusNotFound
	}
	ha := value.(*haikuAmulet)
	return ha.String()
}

// renderCreatedBy renders all the haikus created by a specific user.
func renderCreatedBy(userid string) string {
	render := ""
	nftDB.Iterate("", "", func(key string, value interface{}) bool {
		ha := value.(*haikuAmulet)
		if ha.CreatedBy.String() == userid {
			if render == "" {
				render = "# " + renderUserAddress2(ha.CreatedBy) + " creations\n\n"
			}
			render += ha.String()
		}
		return false
	})
	return render
}

// renderCreatedBy renders all the haikus collected by a specific user.
func renderCollectedBy(userid string) string {
	render := ""
	nftDB.Iterate("", "", func(key string, value interface{}) bool {
		ha := value.(*haikuAmulet)
		owner, err := nfts.OwnerOf(grc721.TokenID(ha.TokenID))
		if err != nil {
			return false
		}
		if owner.String() == userid {
			if render == "" {
				render = "# " + renderUserAddress2(owner) + " collection\n\n"
			}
			render += ha.String()
		}
		return false
	})
	return render
}

// Render is the main screen render
// - / -> renders all
// - /rarity -> renders rarity information
// - /createdby:userid -> renders all haikus created by a user
// - /collectedby:userid -> renders all haikus collected by a user
// - /tokenid -> renders a specific haiku
func Render(path string) string {
	if path == "" {
		return renderAll()
	} else if path == "rarity" {
		return `## about "rarity"

The number of sequential 8s in the SHA-256 hash of the sanitized text determines the rarity of the haiku:

	8888: common
	88888: uncommon
	888888: rare
	8888888: epic
	88888888: legendary
	888888888+: mythic

Please note that a haiku should be judged by its total effect and not simply by its adherence to this arbitrary definition of rarity.
	`
	} else if strings.HasPrefix(path, "createdby:") {
		return renderCreatedBy(strings.TrimPrefix(path, "createdby:"))
	} else if strings.HasPrefix(path, "collectedby:") {
		return renderCollectedBy(strings.TrimPrefix(path, "collectedby:"))
	} else {
		return renderSingle(path)
	}
}
