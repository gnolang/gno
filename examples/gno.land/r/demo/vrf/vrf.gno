package vrf

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	// "encoding/hex"
	fmt "gno.land/p/demo/ufmt"
	"std"
	"strings"
)

type Config struct {
	vrfAdmin string
	feeders  []string
}

type Request struct {
	id                   uint64
	requesterAddress     string
	requesterRealm       string
	requiredFeedersCount uint64
	fulfilledCount       uint64
	randomWords          []byte
	fulfillers           []string
}

var config Config
var requests []Request

const hextable = "0123456789abcdef"
const reverseHexTable = "" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\xff\xff\xff\xff\xff\xff" +
	"\xff\x0a\x0b\x0c\x0d\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\x0a\x0b\x0c\x0d\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff"

// ErrLength reports an attempt to decode an odd-length input
// using Decode or DecodeString.
// The stream-based Decoder returns io.ErrUnexpectedEOF instead of ErrLength.
var ErrLength = errors.New("encoding/hex: odd length hex string")

// InvalidByteError values describe errors resulting from an invalid byte in a hex string.
type InvalidByteError byte

func (e InvalidByteError) Error() string {
	return fmt.Sprintf("encoding/hex: invalid byte: %#U", rune(e))
}

func EncodeToHexString(src []byte) string {
	encoded := []byte{}
	j := 0
	for _, v := range src {
		encoded = append(encoded, hextable[v>>4])
		encoded = append(encoded, hextable[v&0x0f])
	}
	return string(encoded)
}

func DecodeFromHexString(src []byte) ([]byte, error) {
	dst := []byte{}
	i, j := 0, 1
	for ; j < len(src); j += 2 {
		p := src[j-1]
		q := src[j]

		a := reverseHexTable[p]
		b := reverseHexTable[q]
		if a > 0x0f {
			return dst, InvalidByteError(p)
		}
		if b > 0x0f {
			return dst, InvalidByteError(q)
		}
		dst = append(dst, (a<<4)|b)
	}
	if len(src)%2 == 1 {
		// Check for invalid char before reporting bad length,
		// since the invalid char (if present) is an earlier problem.
		if reverseHexTable[src[j-1]] > 0x0f {
			return dst, InvalidByteError(src[j-1])
		}
		return dst, ErrLength
	}
	return dst, nil
}

func RenderConfig() string {
	quotedFeeders := []string{}
	for _, feeder := range config.feeders {
		quotedFeeders = append(quotedFeeders, `"`+feeder+`"`)
	}
	feedersText := strings.Join(quotedFeeders, ", ")
	return fmt.Sprintf(`{
		"vrfAdmin": "%s",
		"feeders": [%s]
}`, config.vrfAdmin, feedersText)
}

func GetRequests(startAfter, limit uint64) []Request {
	max := uint64(len(requests))
	if startAfter+limit < max {
		max = startAfter + limit
	}
	return requests[startAfter:max]
}

func RenderRequest(requestId uint64) string {
	if requestId >= uint64(len(requests)) {
		panic("invalid request id")
	}

	r := requests[requestId]
	quotedFulfillers := []string{}
	for _, fulfiller := range r.fulfillers {
		quotedFulfillers = append(quotedFulfillers, `"`+fulfiller+`"`)
	}
	fulfillers := strings.Join(quotedFulfillers, ", ")
	randomWords := EncodeToHexString(r.randomWords)

	return fmt.Sprintf(`{
		"id": %d,
		"requesterAddress": "%s",
		"requesterRealm": "%s",
		"requiredFeedersCount": %d,
		"fulfilledCount": %d,
		"randomWords": "%s",
		"fulfillers": [%s]
}`, r.id, r.requesterAddress, r.requesterRealm, r.requiredFeedersCount, r.fulfilledCount, randomWords, fulfillers)
}

func RenderRequests(startAfter uint64, limit uint64) string {
	paginatedRequests := GetRequests(startAfter, limit)
	rendered := "["
	for index, req := range paginatedRequests {
		rendered += RenderRequest(req.id)
		if index != len(paginatedRequests)-1 {
			rendered += ",\n"
		}
	}
	rendered += "]"
	return rendered
}

func SetVRFAdmin(vrfAdmin string) {
	if config.vrfAdmin == "" {
		config.vrfAdmin = vrfAdmin
		return
	}
	caller := std.GetOrigCaller()
	if config.vrfAdmin != caller.String() {
		panic("not allowed to update vrfAdmin")
	}

	config.vrfAdmin = vrfAdmin
}

func SetFeeders(feedersText string) {
	feeders := strings.Split(feedersText, ",")
	caller := std.GetOrigCaller()
	if config.vrfAdmin != caller.String() {
		panic("not allowed set feeders")
	}

	config.feeders = feeders
}

func RequestRandomWords(requiredFeedersCount uint64) uint64 {
	requests = append(requests, Request{
		id:                   uint64(len(requests)),
		requesterAddress:     std.PrevRealm().Addr().String(),
		requesterRealm:       std.PrevRealm().PkgPath(),
		requiredFeedersCount: requiredFeedersCount,
		fulfilledCount:       0,
		randomWords:          []byte{},
		fulfillers:           []string{},
	})
	return uint64(len(requests) - 1)
}

func IsFeeder(feeder string) bool {
	for _, f := range config.feeders {
		if feeder == f {
			return true
		}
	}
	return false
}

func Fulfilled(request Request, feeder string) bool {
	for _, f := range request.fulfillers {
		if feeder == f {
			return true
		}
	}
	return false
}

func FulfillRandomWords(requestId uint64, wordsHex string) {
	caller := std.GetOrigCaller()

	if requestId >= uint64(len(requests)) {
		panic("invalid request id")
	}

	if !IsFeeder(caller.String()) {
		panic("not a feeder")
	}

	request := requests[requestId]
	if Fulfilled(request, caller.String()) {
		panic("already fulfilled")
	}

	words, err := DecodeFromHexString([]byte(wordsHex))
	if err != nil {
		panic(err)
	}

	if requests[requestId].fulfilledCount >= request.requiredFeedersCount {
		panic("required feeders count reached")
	}

	requests[requestId].fulfilledCount += 1
	requests[requestId].randomWords = append(request.randomWords, words...)
	requests[requestId].fulfillers = append(request.fulfillers, caller.String())
}

func RandomValueFromWordsWithIndex(requestId uint64, index uint64) uint64 {
	if requestId >= uint64(len(requests)) {
		panic("invalid request id")
	}
	request := requests[requestId]
	if request.fulfilledCount < request.requiredFeedersCount {
		panic("not enough feeders joined yet")
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(index))

	hash := sha256.Sum256(append(request.randomWords, b...))
	value := binary.BigEndian.Uint64(hash[0:8])
	return value
}

func RandomValueFromWords(requestId uint64) uint64 {
	if requestId >= uint64(len(requests)) {
		panic("invalid request id")
	}
	request := requests[requestId]
	if request.fulfilledCount < request.requiredFeedersCount {
		panic("not enough feeders joined yet")
	}
	hash := sha256.Sum256(request.randomWords)
	value := binary.BigEndian.Uint64(hash[0:8])
	return value
}
