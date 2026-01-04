// Realm guestbook contains an implementation of a simple guestbook.
// Come and sign yourself up!
package guestbook

import (
	"chain/runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
)

// Signature is a single entry in the guestbook.
type Signature struct {
	Message string
	Author  address
	Time    time.Time
}

const (
	maxMessageLength = 140
	maxPerPage       = 25
)

var (
	signatureID seqid.ID
	guestbook   avl.Tree // id -> Signature
	hasSigned   avl.Tree // address -> struct{}
)

func init(cur realm) {
	Sign(cur, "You reached the end of the guestbook!")
}

const (
	errNotAUser                  = "this guestbook can only be signed by users"
	errAlreadySigned             = "you already signed the guestbook!"
	errInvalidCharacterInMessage = "invalid character in message"
)

// Sign signs the guestbook, with the specified message.
func Sign(cur realm, message string) {
	prev := runtime.PreviousRealm()
	switch {
	case !prev.IsUser():
		panic(errNotAUser)
	case hasSigned.Has(prev.Address().String()):
		panic(errAlreadySigned)
	}
	message = validateMessage(message)

	guestbook.Set(signatureID.Next().Binary(), Signature{
		Message: message,
		Author:  prev.Address(),
		// NOTE: time.Now() will yield the "block time", which is deterministic.
		Time: time.Now(),
	})
	hasSigned.Set(prev.Address().String(), struct{}{})
}

func validateMessage(msg string) string {
	if len(msg) > maxMessageLength {
		panic("Keep it brief! (max " + strconv.Itoa(maxMessageLength) + " bytes!)")
	}
	out := ""
	for _, ch := range msg {
		switch {
		case unicode.IsLetter(ch),
			unicode.IsNumber(ch),
			unicode.IsSpace(ch),
			unicode.IsPunct(ch):
			out += string(ch)
		default:
			panic(errInvalidCharacterInMessage)
		}
	}
	return out
}

func Render(maxID string) string {
	var bld strings.Builder

	bld.WriteString("# Guestbook 📝\n\n[Come sign the guestbook!](./guestbook$help&func=Sign)\n\n---\n\n")

	var maxIDBinary string
	if maxID != "" {
		mid, err := seqid.FromString(maxID)
		if err != nil {
			panic(err)
		}

		// AVL iteration is exclusive, so we need to decrease the ID value to get the "true" maximum.
		mid--
		maxIDBinary = mid.Binary()
	}

	var lastID seqid.ID
	var printed int
	guestbook.ReverseIterate("", maxIDBinary, func(key string, val any) bool {
		sig := val.(Signature)
		message := strings.ReplaceAll(sig.Message, "\n", "\n> ")
		bld.WriteString("> " + message + "\n>\n")
		idValue, ok := seqid.FromBinary(key)
		if !ok {
			panic("invalid seqid id")
		}

		bld.WriteString("> _Written by " + sig.Author.String() + " at " + sig.Time.Format(time.DateTime) + "_ (#" + idValue.String() + ")\n\n---\n\n")
		lastID = idValue

		printed++
		// stop after exceeding limit
		return printed >= maxPerPage
	})

	if printed == 0 {
		bld.WriteString("No messages!")
	} else if printed >= maxPerPage {
		bld.WriteString("<p style='text-align:right'><a href='./guestbook:" + lastID.String() + "'>Next page</a></p>")
	}

	return bld.String()
}
