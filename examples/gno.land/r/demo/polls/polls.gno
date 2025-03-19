package pollrealm

import (
	"bytes"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/poll"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

//----------------------------------------
// Global State

var (
	polls         *avl.Tree // TODO: This would be better as a btree; no need to key off a string, then.
	pollIDCounter seqid.ID
	pageSize      = 5
)

func init() {
	polls = avl.NewTree()
}

//----------------------------------------
// Poll Management Functions

// CreatePoll creates a new poll with no options initially.
// The deadline parameter should be a date/time string in RFC3339 format,
// e.g. "2025-12-31T23:59:59Z".
func CreatePoll(
	title string,
	description string,
	maxChoices int,
	deadlineStr string,
) string {
	caller := std.OriginCaller()

	deadline, err := time.Parse(time.RFC3339, deadlineStr)
	if err != nil {
		panic("invalid deadline format; please use RFC3339 format (e.g., 2025-12-31T23:59:59Z)")
	}

	id := pollIDCounter.Next().String()
	// Create with an empty slice of options
	newP := poll.NewPoll(title, description, []string{}, maxChoices, deadline, caller)
	polls.Set(id, newP)

	return ufmt.Sprintf("Successfully created poll #%s!", id)
}

// EditPoll modifies basic poll metadata (title, description, deadline, maxChoices)
// but does NOT change the options. Only the poll creator can edit.
func EditPoll(
	id string,
	newTitle string,
	newDescription string,
	newDeadlineStr string,
	newMaxChoices int,
) string {
	newDeadline, err := time.Parse(time.RFC3339, newDeadlineStr)
	if err != nil {
		panic("invalid deadline format; please use RFC3339 format (e.g., 2025-12-31T23:59:59Z)")
	}
	p := getPoll(id)
	p.Edit(std.OriginCaller(), newTitle, newDescription, newDeadline, newMaxChoices)
	polls.Set(id, p)
	return ufmt.Sprintf("Poll #%s edited successfully!", id)
}

// AddOption adds a single option string to a poll (before any votes).
func AddOption(id string, optionText string) string {
	p := getPoll(id)
	// Only the creator can add options
	if p.Creator != std.OriginCaller() {
		panic("only the poll creator can add options")
	}
	// Disallow if any votes have been cast
	if p.VoterCount() > 0 {
		panic("cannot add options after the first vote has been cast")
	}
	p.AddOption(optionText)
	polls.Set(id, p)
	return ufmt.Sprintf("Added a new option to poll #%s", id)
}

// EditOption changes the text of an existing option in a poll.
// Allowed even if votes have already been cast, because it does not
// change the *number* of options.
func EditOption(id string, optionIndex int, newText string) string {
	p := getPoll(id)
	// Only the creator can edit
	if p.Creator != std.OriginCaller() {
		panic("only the poll creator can edit options")
	}
	p.EditOption(optionIndex, newText)
	polls.Set(id, p)
	return ufmt.Sprintf("Edited option #%d in poll %s", optionIndex, id)
}

// DeleteOption removes an option by index (before any votes).
func DeleteOption(id string, optionIndex int) string {
	p := getPoll(id)
	// Only the creator can delete
	if p.Creator != std.OriginCaller() {
		panic("only the poll creator can delete options")
	}
	// Disallow deleting if any votes have been cast
	if p.VoterCount() > 0 {
		panic("cannot delete options after the first vote has been cast")
	}
	p.DeleteOption(optionIndex)
	polls.Set(id, p)
	return ufmt.Sprintf("Deleted option #%d from poll %s", optionIndex, id)
}

// Vote: Cast a vote for the poll with given ID.
// `choices` is a list of option indices the user picks (a comma-separated string,
// e.g. "1,2,4"). The string is parsed into a slice of ints.
func Vote(id string, choices string) string {
	choiceIndices := parseChoices(choices)

	p := getPoll(id)
	p.Vote(std.OriginCaller(), choiceIndices)
	polls.Set(id, p)

	return ufmt.Sprintf("Successfully voted on poll #%s!", id)
}

//----------------------------------------
// Rendering

func Render(path string) string {
	var b bytes.Buffer

	path = strings.TrimSpace(path)
	if path == "" || strings.HasPrefix(path, "page=") {
		return renderPollList(&b, path)
	}

	val, exists := polls.Get(path)
	if !exists {
		return renderPollList(&b, "")
	}
	p := val.(*poll.Poll)
	renderSinglePoll(&b, path, p)
	return b.String()
}

// renderSinglePoll shows the details of a single poll.
func renderSinglePoll(b *bytes.Buffer, id string, p *poll.Poll) {
	b.WriteString(ufmt.Sprintf("# Poll %s: %s\n\n", id, p.Title))
	b.WriteString(ufmt.Sprintf("**Description:** %s\n\n", p.Description))
	b.WriteString(ufmt.Sprintf("**Deadline:** %s\n", p.Deadline.Format(time.RFC1123)))
	b.WriteString(ufmt.Sprintf("**Creator:** %s\n", p.Creator))
	b.WriteString(ufmt.Sprintf("**MaxChoices:** %d\n", p.MaxChoices))

	tally := p.Tally()
	b.WriteString("\n## Options:\n")
	for i, opt := range p.Options {
		b.WriteString(ufmt.Sprintf("  - (%d) %s — Votes: %d -- %f%%\n", i, opt, tally[i], 100*float64(tally[i])/float64(p.VoteTotal)))
	}

	b.WriteString(ufmt.Sprintf("\n**Distinct Voters:** %d\n", p.VoterCount()))
	sumVotes := 0
	for _, count := range tally {
		sumVotes += count
	}
	if sumVotes > p.VoterCount() && p.MaxChoices > 1 {
		b.WriteString(
			ufmt.Sprintf("*Multiple selections (%d) are allowed; total votes can exceed distinct voters.*\n\n", p.MaxChoices))
	}

	b.WriteString("\n## Detailed Votes:\n")
	p.Votes.Iterate("", "", func(addr string, val interface{}) bool {
		indexes := val.([]int)
		strIndexes := make([]string, len(indexes))
		for i, num := range indexes {
			strIndexes[i] = strconv.Itoa(num)
		}
		b.WriteString(ufmt.Sprintf("- %s => %s\n", addr, strings.Join(strIndexes, ", ")))
		return false
	})
}

// renderPollList displays a paginated list of all polls.
func renderPollList(b *bytes.Buffer, path string) string {
	realmAddr := strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land")
	b.WriteString("# All Polls\n\n")

	if polls.Size() == 0 {
		b.WriteString("No polls have been created yet.\n")
		return b.String()
	}

	page := 1
	if strings.HasPrefix(path, "page=") {
		pStr := strings.TrimPrefix(path, "page=")
		pInt, err := strconv.Atoi(pStr)
		if err == nil && pInt > 0 {
			page = pInt
		}
	}
	start := (page - 1) * pageSize
	end := start + pageSize

	var ids []string
	polls.Iterate("", "", func(k string, _ interface{}) bool {
		ids = append(ids, k)
		return false
	})

	if start >= len(ids) {
		b.WriteString("No polls on this page.\n")
		return b.String()
	}
	if end > len(ids) {
		end = len(ids)
	}

	for i := start; i < end; i++ {
		id := ids[i]
		val, _ := polls.Get(id)
		p := val.(*poll.Poll)
		poll_url := ufmt.Sprintf("%s:%s", realmAddr, id)
		b.WriteString(ufmt.Sprintf("## [Poll](%s) #%s: %s\n", poll_url, id, p.Title))
		b.WriteString(ufmt.Sprintf("- Deadline: %s\n", p.Deadline.Format(time.RFC1123)))
		b.WriteString(ufmt.Sprintf("- Votes cast: %d\n\n", p.VoterCount()))
	}

	b.WriteString(ufmt.Sprintf("### Page %d\n", page))
	if page > 1 {
		b.WriteString(ufmt.Sprintf("[Prev Page](./?page=%d)\n", page-1))
	}
	if end < len(ids) {
		b.WriteString(ufmt.Sprintf("[Next Page](./?page=%d)\n", page+1))
	}
	return b.String()
}

//----------------------------------------
// Helpers

func getPoll(id string) *poll.Poll {
	val, exists := polls.Get(id)
	if !exists {
		panic(ufmt.Sprintf("poll not found: %s", id))
	}
	return val.(*poll.Poll)
}

// parseChoices converts a comma-separated string of integers (e.g. "0,2,3")
// into a slice of ints.
func parseChoices(choices string) []int {
	if choices == "" {
		panic("must provide at least one choice index")
	}
	parts := strings.Split(choices, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			panic(ufmt.Sprintf("invalid choice index: %s", part))
		}
		out = append(out, val)
	}
	return out
}
