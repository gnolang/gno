package governance

import (
	"encoding/base64"
	"std"
	"strconv"

	"gno.land/p/demo/avl"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
)

// Retrieves a proposal by its ID
func mustGetProposal(proposalId uint64) ProposalInfo {
	result, exists := proposals.Get(formatUint(proposalId))
	if !exists {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("proposalId(%d) not found", proposalId),
		))
	}

	return result.(ProposalInfo)
}

// Retrieves a vote by its key
func mustGetVote(key string) bool {
	vote, exists := votes.Get(key)
	if !exists {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("voteKey(%s) not found", key),
		))
	}
	return vote.(bool)
}

// Helper function to validate and get vote information
func getVoteInfoFromKey(voteKey string) (voteWithWeight, bool) {
	mustGetVote(voteKey)

	// Splits the vote key into proposal ID and user address
	pid, addr := divideVoteKeyToProposalIdAndUser(voteKey)

	// Retrieves user vote information
	voteInfo, exists := getUserVote(addr, pid)
	if !exists {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("voteKey(%s) not found", voteKey),
		))
	}

	return voteInfo, true
}

// Retrieves vote information by key
func mustGetVoteInfo(voteKey string) voteWithWeight {
	voteInfo, _ := getVoteInfoFromKey(voteKey)
	return voteInfo
}

// Iterates over an AVL tree and applies a callback function to each element
func iterTree(tree *avl.Tree, cb func(key string, value interface{}) bool) {
	tree.Iterate("", "", cb)
}

// Converts a string to an integer, panics if conversion fails
func strToInt(str string) int {
	res, err := strconv.Atoi(str)
	if err != nil {
		panic(err.Error())
	}

	return res
}

// Converts a boolean vote to a string representation
func voteToString(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// Marshals a JSON node to a string, panics if marshalling fails
func marshal(data *json.Node) string {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

// Encodes a string to base64
func b64Encode(data string) string {
	return string(base64.StdEncoding.EncodeToString([]byte(data)))
}

// Decodes a base64 string, panics if decoding fails
func b64Decode(data string) string {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		panic(err.Error())
	}
	return string(decoded)
}

// Returns the package path of the previous realm
func prevRealm() string {
	return std.PrevRealm().PkgPath()
}

// Returns the address and package path of the previous realm
func getPrev() (string, string) {
	prev := std.PrevRealm()
	return prev.Addr().String(), prev.PkgPath()
}

// Formats an unsigned integer to a string
func formatUint(v interface{}) string {
	switch v := v.(type) {
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		panic(ufmt.Sprintf("invalid type: %T", v))
	}
}

// Formats a boolean to a string
func formatBool(v bool) string {
	return strconv.FormatBool(v)
}

// Parses a string to an integer, panics if parsing fails
func parseInt(s string) int {
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(ufmt.Sprintf("invalid int value: %s", s))
	}
	return int(num)
}

// Parses a string to an unsigned 64-bit integer, panics if parsing fails
func parseUint64(s string) uint64 {
	num, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(ufmt.Sprintf("invalid uint64 value: %s", s))
	}
	return num
}

// Parses a string to a boolean, panics if parsing fails
func parseBool(s string) bool {
	switch s {
	case "true":
		return true
	case "false":
		return false
	default:
		panic(ufmt.Sprintf("invalid bool value: %s", s))
	}
}
