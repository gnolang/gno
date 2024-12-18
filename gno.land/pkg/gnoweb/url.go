package gnoweb

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

type PathKind byte

const (
	KindInvalid PathKind = 0
	KindRealm   PathKind = 'r'
	KindPure    PathKind = 'p'
)

// reRealmPath match and validate a realm or package path
var rePkgOrRealmPath = regexp.MustCompile(`^/[a-z]/[a-zA-Z0-9_/.]*$`)

// GnoURL decomposes the parts of an URL to query a realm.
type GnoURL struct {
	// Example full path:
	// gno.land/r/demo/users:jae$help&a=b?c=d

	Domain   string     // gno.land
	Path     string     // /r/demo/users
	Args     string     // jae
	WebQuery url.Values // help&a=b
	Query    url.Values // c=d
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
	// Check if the first and third character is '/' and extract the next character
	if len(url.Path) > 2 && url.Path[0] == '/' && url.Path[2] == '/' {
		switch k := PathKind(url.Path[1]); k {
		case KindPure, KindRealm:
			return k
		}
	}

	return KindInvalid
}

func (url GnoURL) IsDir() bool {
	if pathlen := len(url.Path); pathlen > 0 {
		return url.Path[pathlen-1] == '/'
	}

	return false
}

func (url GnoURL) IsFile() bool {
	return filepath.Ext(url.Path) != ""
}

var (
	ErrURLMalformedPath   = errors.New("malformed path")
	ErrURLInvalidPathKind = errors.New("invalid path kind")
)

func ParseGnoURL(u *url.URL) (*GnoURL, error) {
	var webargs string
	path, args, found := strings.Cut(u.EscapedPath(), ":")
	if found {
		args, webargs, _ = strings.Cut(args, "$")
	} else {
		path, webargs, _ = strings.Cut(path, "$")
	}

	// XXX: should we lower case the path ?
	upath, err := url.PathUnescape(path)
	if err != nil {
		return nil, fmt.Errorf("unable to unescape path %q: %w", args, err)
	}

	// Validate path format
	if !rePkgOrRealmPath.MatchString(upath) {
		return nil, fmt.Errorf("%w: %q", ErrURLMalformedPath, upath)
	}

	webquery := url.Values{}
	if len(webargs) > 0 {
		var err error
		if webquery, err = url.ParseQuery(webargs); err != nil {
			return nil, fmt.Errorf("unable to parse webquery %q: %w ", webquery, err)
		}
	}

	uargs, err := url.PathUnescape(args)
	if err != nil {
		return nil, fmt.Errorf("unable to unescape args %q: %w", args, err)
	}

	return &GnoURL{
		Path:     upath,
		Args:     uargs,
		WebQuery: webquery,
		Query:    u.Query(),
		Domain:   u.Hostname(),
	}, nil
}

func escapeDollarSign(s string) string {
	return strings.ReplaceAll(s, "$", "%24")
}
