package dao_registry

import (
	"encoding/binary"
	"std"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gno.land/p/demo/avl"
)

type Registration struct {
	PkgPath     string      `json:"pkgPath"`
	Addr        std.Address `json:"addr"`
	CreatedAt   uint64      `json:"createdAt"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ImageURI    string      `json:"imageURI"`
}

var (
	byCreatedAt = avl.NewTree()
	byPkgPath   = avl.NewTree()
)

// TODO: use profile for info
// TODO: make core accessible

func Register(name string, description string, imageURI string) {
	realm := std.PrevRealm()
	pkgPath := realm.PkgPath()
	if pkgPath == "" {
		panic("no pkg path in prev realm")
	}
	doRegister(pkgPath, realm.Addr(), name, description, imageURI)
}

// splitted for tests
func doRegister(pkgPath string, addr std.Address, name string, description string, imageURI string) {
	if byPkgPath.Has(pkgPath) {
		panic("already registered")
	}
	registration := &Registration{
		PkgPath:     pkgPath,
		Addr:        addr,
		CreatedAt:   uint64(time.Now().UnixMilli()),
		Name:        name,
		Description: description,
		ImageURI:    imageURI,
	}
	byPkgPath.Set(pkgPath, registration)

	keyBytes := []byte{}
	keyBytes = binary.BigEndian.AppendUint64(keyBytes, uint64(registration.CreatedAt))
	keyBytes = append(keyBytes, []byte(pkgPath)...)
	key := string(keyBytes)
	byCreatedAt.Set(key, registration)
}

func GetJSON(pkgPath string) string {
	r, ok := byPkgPath.Get(pkgPath)
	if !ok {
		panic("not found")
	}
	return r.(*Registration).JSON()
}

func List(start string, end string, limit int, reverse bool) []Registration {
	regs := []Registration{}
	cb := func(key string, value interface{}) bool {
		regs = append(regs, *value.(*Registration))
		if limit == 0 {
			return false
		}
		return len(regs) >= limit
	}
	if reverse {
		byCreatedAt.ReverseIterate(start, end, cb)
	} else {
		byCreatedAt.Iterate(start, end, cb)
	}
	return regs
}

func ListJSON(start string, end string, limit int, reverse bool) string {
	regs := List(start, end, limit, reverse)
	elems := make([]string, len(regs))
	for i, reg := range regs {
		elems[i] = reg.JSON()
	}
	return "[" + strings.Join(elems, ",") + "]"
}

func (r *Registration) JSON() string {
	fields := []string{
		`"pkgPath":` + JSONString(r.PkgPath),
		`"addr":` + JSONString(r.Addr.String()),
		`"createdAt":` + strconv.FormatUint(r.CreatedAt, 10),
		`"name":` + JSONString(r.Name),
		`"description":` + JSONString(r.Description),
		`"imageURI":` + JSONString(r.ImageURI),
	}
	return "{" + strings.Join(fields, ",") + "}"
}

// ported from https://cs.opensource.google/go/go/+/refs/tags/go1.20.6:src/encoding/json/encode.go
func JSONString(s string) string {
	const escapeHTML = true
	e := `"` // e.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
				i++
				continue
			}
			if start < i {
				e += s[start:i] // e.WriteString(s[start:i])
			}
			e += "\\" // e.WriteByte('\\')
			switch b {
			case '\\', '"':
				e += string(b) // e.WriteByte(b)
			case '\n':
				e += "n" // e.WriteByte('n')
			case '\r':
				e += "r" // e.WriteByte('r')
			case '\t':
				e += "t" // e.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				e += `u00`              // e.WriteString(`u00`)
				e += string(hex[b>>4])  // e.WriteByte(hex[b>>4])
				e += string(hex[b&0xF]) // e.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				e += s[start:i] // e.WriteString(s[start:i])
			}
			e += `\ufffd` // e.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				e += s[start:i] // e.WriteString(s[start:i])
			}
			e += `\u202`            // e.WriteString(`\u202`)
			e += string(hex[c&0xF]) // e.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e += s[start:] // e.WriteString(s[start:])
	}
	e += `"` // e.WriteByte('"')
	return e
}
