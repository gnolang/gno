package gnoweb

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type PathKind byte

const (
	KindInvalid PathKind = 0
	KindRealm   PathKind = 'r'
	KindPure    PathKind = 'p'
)

// GnoURL decomposes the parts of an URL to query a realm.
type GnoURL struct {
	Host     string     // gno.land
	Path     string     // /r/demo/users
	Args     string     // :jae
	WebQuery url.Values // $help&f1=f2
	Query    url.Values // ?f1=f2
}

func (url GnoURL) EncodeArgs() string {
	var urlstr strings.Builder
	if url.Args != "" {
		urlstr.WriteString(url.Args)
	}

	if len(url.Query) > 0 {
		urlstr.WriteString("?" + url.Query.Encode())
	}

	return urlstr.String()
}

func (url GnoURL) EncodePath() string {
	var urlstr strings.Builder
	urlstr.WriteString(url.Path)
	if url.Args != "" {
		urlstr.WriteString(":" + url.Args)
	}

	if len(url.Query) > 0 {
		urlstr.WriteString("?" + url.Query.Encode())
	}

	return urlstr.String()
}

func (url GnoURL) EncodeWebPath() string {
	var urlstr strings.Builder
	urlstr.WriteString(url.Path)
	if url.Args != "" {
		pathEscape := escapeDollarSign(url.Args)
		urlstr.WriteString(":" + pathEscape)
	}

	if len(url.WebQuery) > 0 {
		urlstr.WriteString("$" + url.WebQuery.Encode())
	}

	if len(url.Query) > 0 {
		urlstr.WriteString("?" + url.Query.Encode())
	}

	return urlstr.String()
}

func (url GnoURL) Kind() PathKind {
	if len(url.Path) < 2 {
		return KindInvalid
	}
	pk := PathKind(url.Path[1])
	switch pk {
	case KindPure, KindRealm:
		return pk
	}
	return KindInvalid
}

var (
	ErrURLMalformedPath   = errors.New("malformed URL path")
	ErrURLInvalidPathKind = errors.New("invalid path kind")
)

// reRealName match a realm path
// - matches[1]: path
// - matches[2]: path kind
// - matches[3]: path args
var reRealmPath = regexp.MustCompile(`(?m)^` +
	`(/([a-zA-Z0-9_-]+)/` + // path kind
	`[a-zA-Z][a-zA-Z0-9_-]*` + // First path segment
	`(?:/[a-zA-Z][.a-zA-Z0-9_-]*)*/?)` + // Additional path segments
	`([:$](?:.*)|$)`, // Remaining portions args, separate by `$` or `:`
)

func ParseGnoURL(u *url.URL) (*GnoURL, error) {
	matches := reRealmPath.FindStringSubmatch(u.EscapedPath())
	if len(matches) != 4 {
		return nil, fmt.Errorf("%w: %s", ErrURLMalformedPath, u.Path)
	}

	// Force lower case
	path := matches[1]

	pathKind, args := matches[2], matches[3]
	if len(args) > 0 {
		switch args[0] {
		case ':':
			args = args[1:]
		case '$':
		default:
			return nil, fmt.Errorf("%w: %s", ErrURLMalformedPath, u.Path)
		}
	}

	var err error
	webquery := url.Values{}
	args, webargs, found := strings.Cut(args, "$")
	if found {
		if webquery, err = url.ParseQuery(webargs); err != nil {
			return nil, fmt.Errorf("unable to parse webquery %q: %w ", webquery, err)
		}
	}

	uargs, err := url.PathUnescape(args)
	if err != nil {
		return nil, fmt.Errorf("unable to unescape path %q: %w", args, err)
	}

	return &GnoURL{
		Path:     path,
		Args:     uargs,
		WebQuery: webquery,
		Query:    u.Query(),
		Host:     u.Hostname(),
	}, nil
}

func escapeDollarSign(s string) string {
	return strings.ReplaceAll(s, "$", "%24")
}
