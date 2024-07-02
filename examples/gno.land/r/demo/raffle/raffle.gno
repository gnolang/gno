package raffle

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

type EntryData struct {
	txorigin   std.Address
	caller     std.Realm
	raffleCode string
	codeHash   string
	ghUsername string
}

// Top-level variables are automatically persisted to storage
var (
	o                *ownable.Ownable    // admin of realm
	partialEntries   []*EntryData        // keeps registered partialEntries
	completeEntries  []*EntryData        // keeps complete registrations: valid code + gh username
	CodeHashes       []string            // valid code hashes
	registeredHashes map[string]struct{} // tracks if a code has been registered before
	winner1, winner2 *EntryData
	numReg           int
)

const (
	codeLength = 10
	amtOfCodes = 300
)

func init() {
	o = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

	partialEntries = make([]*EntryData, 0)
	completeEntries = make([]*EntryData, 0)
	registeredHashes = make(map[string]struct{})
	CodeHashes = make([]string, 300)
}

// RegisterCode enters the raffle for a specific raffle code, and caller
// Can only be called from within other code; figure out a way to do it!
func RegisterCode(code string) string {
	if code == "" && len(code) != codeLength {
		panic("invalid code: " + code)
	}

	caller := std.PrevRealm()     // save realm used to call
	origin := std.GetOrigCaller() // save deployer of realm

	// Deny non-code entries
	if caller.IsUser() {
		panic("denied; can only be called from within code")
	}

	// Get hash of code
	hash := sha256.Sum256([]byte(code))
	hashString := hex.EncodeToString(hash[:])

	// Check if code has already been registered
	if _, ok := registeredHashes[hashString]; ok {
		panic("code already registered: " + code)
	}

	// Try to find the hash in the official hash list
	var found bool
	for _, ch := range CodeHashes {
		if ch == hashString {
			found = true
			break
		}
	}

	if !found {
		panic("specified code is not a valid raffle code: " + code)
	}

	entry := &EntryData{
		txorigin:   origin,
		caller:     caller,
		raffleCode: code,
		codeHash:   hashString,
		ghUsername: "No username yet!",
	}

	// Save to hash tracker
	registeredHashes[hashString] = struct{}{}

	// Save raffle entry
	partialEntries = append(partialEntries, entry)

	return ufmt.Sprintf("Successfully registered raffle code!\n%s\nRegister your username to complete your raffle entry.", entry.String())
}

// RegisterUsername registers a GitHub username to an already existing partialEntries entry
func RegisterUsername(username string) string {
	if username == "" {
		panic("invalid username: " + username)
	}

	origin := std.GetOrigCaller()

	for _, entry := range partialEntries {
		if entry.txorigin == origin && entry.ghUsername == "No username yet!" {
			entry.ghUsername = username
			completeEntries = append(completeEntries, entry)
			numReg += 1
			return ufmt.Sprintf("successfully registered \"%s\" for address %s", username, entry.txorigin)
		}
	}

	panic("could not find partialEntries entry for caller address; did you register your raffle code yet?")
}

func PickWinner1() string {
	o.AssertCallerIsOwner()
	winner1 = pickWinner()

	return winner1.ghUsername
}

func PickWinner2() string {
	o.AssertCallerIsOwner()
	winner2 = pickWinner()

	return winner2.ghUsername
}

func UploadCodes(delimCodes string) {
	o.AssertCallerIsOwner()

	tokens := strings.Split(delimCodes, ",")

	if len(tokens) != amtOfCodes {
		panic(ufmt.Sprintf("invalid amount of codes; wanted %d got %d", amtOfCodes, len(tokens)))
	}

	copy(CodeHashes, tokens)
}

func Render(_ string) string {
	output := "# Gno.land Raffle - GopherCon US 2024\n\n"

	output += ufmt.Sprintf("### Currently signed up users: %d\n\n", len(completeEntries))

	chanceOfWinning := 0

	if numReg > 0 {
		chanceOfWinning = 100 / numReg
		output += ufmt.Sprintf("### Current chance to win a prize: %d%%\n\n", chanceOfWinning)
	}

	if winner1 != nil {
		output += ufmt.Sprintf("### Winner 1: @%s!\n\n", winner1.ghUsername)
	}

	if winner2 != nil {
		output += ufmt.Sprintf("### Winner 2: @%s!\n\n", winner2.ghUsername)
	}

	return output
}

func (entry *EntryData) String() string {
	return ufmt.Sprintf("Address: %s\nRealm Path: %s\nCode: %s\nHash: %s\nGitHub username: %s\n", entry.txorigin.String(), entry.caller.PkgPath(), entry.raffleCode, entry.codeHash, entry.ghUsername)
}

func pickWinner() *EntryData {
	if len(completeEntries) == 0 {
		panic("No complete entries yet!")
	}

	var randSource = rand.New(rand.NewPCG(uint64(time.Now().Unix()), uint64(std.GetHeight())))

	r := rand.New(randSource)
	winnerIndex := r.IntN(len(completeEntries))
	winner1 = partialEntries[winnerIndex]

	// remove winner from entry list
	completeEntries = append(completeEntries[:winnerIndex], completeEntries[winnerIndex+1:]...)

	return winner1
}
