package types

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"gno.land/p/aib/ibc/host"
	"gno.land/p/nt/ufmt"
)

// ValidateClientType validates the client type. It cannot be blank or empty.
// It must be a valid client identifier when used with '0' or the maximum
// uint64 as the sequence.
func ValidateClientType(clientType string) error {
	if strings.TrimSpace(clientType) == "" {
		return ufmt.Errorf("client type cannot be blank")
	}

	smallestPossibleClientID := FormatClientIdentifier(clientType, 0)
	largestPossibleClientID := FormatClientIdentifier(clientType, uint64(math.MaxUint64))

	// IsValidClientID will check client type format and if the sequence is a uint64
	if !IsValidClientID(smallestPossibleClientID) {
		return ufmt.Errorf("not a valid clientID: %s", smallestPossibleClientID)
	}

	if err := host.ClientIdentifierValidator(smallestPossibleClientID); err != nil {
		return err
	}
	if err := host.ClientIdentifierValidator(largestPossibleClientID); err != nil {
		return err
	}

	return nil
}

// FormatClientIdentifier returns the client identifier with the sequence appended.
// This is an SDK specific format not enforced by IBC protocol.
func FormatClientIdentifier(clientType string, sequence uint64) string {
	return ufmt.Sprintf("%s-%d", clientType, sequence)
}

// IsClientIDFormat checks if a clientID is in the format required on the SDK
// for parsing client identifiers. The client identifier must be in the form:
// `{client-type}-{N}` which per the specification only permits ASCII for the
// {client-type} segment and 1 to 20 digits for the {N} segment.
// `([\w-]+\w)?` allows for a letter or hyphen, with the {client-type} starting
// with a letter and ending with a letter, i.e.
// `letter+(letter|hyphen+letter)?`.
var IsClientIDFormat = regexp.MustCompile(`^\w+([\w-]+\w)?-[0-9]{1,20}$`).MatchString

// IsValidClientID checks if the clientID is valid and can be parsed into the client
// identifier format.
func IsValidClientID(clientID string) bool {
	_, _, err := ParseClientIdentifier(clientID)
	return err == nil
}

// ParseClientIdentifier parses the client type and sequence from the client identifier.
func ParseClientIdentifier(clientID string) (string, uint64, error) {
	if !IsClientIDFormat(clientID) {
		return "", 0, ufmt.Errorf("invalid client identifier %s is not in format: `{client-type}-{N}`", clientID)
	}

	splitStr := strings.Split(clientID, "-")
	lastIndex := len(splitStr) - 1

	clientType := strings.Join(splitStr[:lastIndex], "-")
	if strings.TrimSpace(clientType) == "" {
		return "", 0, ufmt.Errorf("client identifier must be in format: `{client-type}-{N}` and client type cannot be blank")
	}

	sequence, err := strconv.ParseUint(splitStr[lastIndex], 10, 64)
	if err != nil {
		return "", 0, ufmt.Errorf("failed to parse client identifier sequence")
	}

	return clientType, sequence, nil
}
