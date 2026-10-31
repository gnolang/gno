// Package drand gives every gno.land user and realm verifiable randomness
// from drand, with no trusted party.
//
// Flow:
//
//  1. A user or realm calls Request and gets an id. The request is pinned to
//     a drand round that will be published in the future (at least SafetyGap
//     seconds after the current block time), so nobody, including the
//     requester and the validators, knows its value yet.
//  2. Any relayer submits that round's drand signature with Submit. The realm
//     verifies it on-chain (BN254 pairing), so a relayer can only submit the
//     real beacon. Relayers are permissionless and interchangeable.
//  3. The requester reads its value with Get / MustGet. Each request gets its
//     own value, derived from the beacon, its requester and its id.
//
// Commit before you reveal: freeze whatever the randomness decides (tickets
// sold, bets placed) before calling Request.
package drand

import (
	"chain"
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"time"

	evmnet "gno.land/p/drand/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
)

// SafetyGap is the minimum number of seconds between the block time of a
// request and the publication of the round it is pinned to. It absorbs
// block-time drift so the round is still unknown when the request lands.
const SafetyGap = int64(5)

// maxPendingList bounds PendingRounds so a qeval always fits its gas cap.
const maxPendingList = 100

type request struct {
	round     uint64
	requester string
}

var (
	ids      seqid.ID
	requests avl.Tree // id -> request
	pending  avl.Tree // roundKey -> true, rounds waiting for a beacon
	beacons  avl.Tree // roundKey -> [32]byte randomness, only rounds someone asked for
)

// Request asks for randomness from the next round that is safely in the
// future and returns the request id.
func Request(cur realm) string {
	return newRequest(cur, nextSafeRound())
}

// RequestAfter asks for randomness from the first round published strictly
// after unix time t (Nois' "randomness after"). If t is too close or in the
// past, the next safe round is used instead.
func RequestAfter(cur realm, t int64) string {
	round := evmnet.RoundAt(t) + 1
	if safe := nextSafeRound(); round < safe {
		round = safe
	}
	return newRequest(cur, round)
}

func newRequest(cur realm, round uint64) string {
	id := ids.Next().String()
	requester := cur.Previous().Address().String()
	requests.Set(id, request{round: round, requester: requester})

	key := roundKey(round)
	if !beacons.Has(key) && !pending.Has(key) {
		pending.Set(key, true)
	}
	chain.Emit("Request", "id", id, "round", strconv.FormatUint(round, 10), "requester", requester)
	return id
}

// nextSafeRound is the first round published at least SafetyGap seconds
// after the current block time.
func nextSafeRound() uint64 {
	return evmnet.RoundAt(time.Now().Unix()+SafetyGap) + 1
}

// Submit stores the drand beacon of a requested round after verifying its
// signature. sigHex is the "signature" field from the drand HTTP API.
// Submitting a round that is already stored is a no-op, so concurrent
// relayers do not fail each other's transactions.
func Submit(cur realm, round uint64, sigHex string) {
	key := roundKey(round)
	if beacons.Has(key) {
		return
	}
	if !pending.Has(key) {
		panic("round not requested")
	}
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		panic("signature is not hex")
	}
	if !evmnet.Verify(round, sig) {
		panic("invalid drand signature")
	}
	beacons.Set(key, evmnet.Randomness(sig))
	pending.Remove(key)
	chain.Emit("Beacon", "round", strconv.FormatUint(round, 10))
}

// Get returns the randomness of a request, and false while its round has
// not been submitted yet (or the id is unknown).
func Get(id string) ([32]byte, bool) {
	r, ok := getRequest(id)
	if !ok {
		return [32]byte{}, false
	}
	b, ok := Beacon(r.round)
	if !ok {
		return [32]byte{}, false
	}
	return evmnet.Derive(b, r.requester+":"+id), true
}

// MustGet is Get that panics while the randomness is not available.
func MustGet(id string) [32]byte {
	v, ok := Get(id)
	if !ok {
		panic("randomness not available yet")
	}
	return v
}

// Rand returns a PRNG seeded with the request's randomness, for drawing as
// many values as needed: drand.Rand(id).IntN(100). It panics while the
// randomness is not available. Each call starts a fresh generator from the
// same seed, so keep the returned value to draw a sequence.
func Rand(id string) *rand.Rand {
	return evmnet.NewRand(MustGet(id))
}

// Status returns the round a request is pinned to and whether its
// randomness is available. round is 0 for an unknown id.
func Status(id string) (round uint64, ready bool) {
	r, ok := getRequest(id)
	if !ok {
		return 0, false
	}
	return r.round, beacons.Has(roundKey(r.round))
}

// Beacon returns the raw drand randomness of a stored round. Prefer Get:
// values from Beacon are shared by everyone who uses the same round.
func Beacon(round uint64) ([32]byte, bool) {
	v := beacons.Get(roundKey(round))
	if v == nil {
		return [32]byte{}, false
	}
	return v.([32]byte), true
}

// PendingRounds lists up to limit requested rounds still waiting for a
// beacon, oldest first, as comma-separated decimals. Relayers poll this.
func PendingRounds(limit int) string {
	if limit <= 0 || limit > maxPendingList {
		limit = maxPendingList
	}
	out := make([]string, 0, limit)
	pending.Iterate("", "", func(k string, _ any) bool {
		out = append(out, strconv.FormatUint(keyRound(k), 10))
		return len(out) >= limit
	})
	return strings.Join(out, ",")
}

func getRequest(id string) (request, bool) {
	v := requests.Get(id)
	if v == nil {
		return request{}, false
	}
	return v.(request), true
}

// roundKey encodes a round as 8 big-endian bytes, so avl order is numeric order.
func roundKey(round uint64) string { return seqid.ID(round).Binary() }

func keyRound(k string) uint64 {
	id, _ := seqid.FromBinary(k)
	return uint64(id)
}

// Render shows the realm state on gnoweb.
func Render(path string) string {
	if strings.HasPrefix(path, "request/") {
		return renderRequest(strings.TrimPrefix(path, "request/"))
	}
	var b strings.Builder
	b.WriteString("# drand randomness\n\n")
	b.WriteString("Verifiable randomness from [drand](https://drand.love) evmnet, checked on-chain.\n\n")
	b.WriteString("```go\nid := drand.Request(cross(cur)) // tx 1: after freezing your inputs\nn := drand.Rand(id).IntN(100)    // tx 2: a few seconds later\n```\n\n")
	b.WriteString("## Status\n\n")
	b.WriteString(ufmt.Sprintf("- chain: `%s`\n", evmnet.ChainHash))
	b.WriteString(ufmt.Sprintf("- requests: %d\n", requests.Size()))
	b.WriteString(ufmt.Sprintf("- pending rounds: %d\n", pending.Size()))
	b.WriteString(ufmt.Sprintf("- stored beacons: %d\n", beacons.Size()))
	b.WriteString(ufmt.Sprintf("- next safe round: %d\n\n", nextSafeRound()))

	if beacons.Size() > 0 {
		b.WriteString("## Latest beacons\n\n| round | randomness |\n|---|---|\n")
		n := 0
		beacons.ReverseIterate("", "", func(k string, v any) bool {
			r := v.([32]byte)
			b.WriteString(ufmt.Sprintf("| %d | `%s` |\n", keyRound(k), hex.EncodeToString(r[:])))
			n++
			return n >= 10
		})
	}
	return b.String()
}

func renderRequest(id string) string {
	r, ok := getRequest(id)
	if !ok {
		return "# request not found\n"
	}
	var b strings.Builder
	b.WriteString(ufmt.Sprintf("# request %s\n\n", id)) // id is a known seqid, safe to echo
	b.WriteString(ufmt.Sprintf("- requester: `%s`\n", r.requester))
	b.WriteString(ufmt.Sprintf("- round: %d (published at %d)\n", r.round, evmnet.TimeOf(r.round)))
	if v, ok := Get(id); ok {
		b.WriteString(ufmt.Sprintf("- randomness: `%s`\n", hex.EncodeToString(v[:])))
	} else {
		b.WriteString("- randomness: pending\n")
	}
	return b.String()
}
