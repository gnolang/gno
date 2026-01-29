package weburl

import (
	"errors"
	"fmt"
	"net/url"
	gopath "path"
	"regexp"
	"slices"
	"strings"
)

var ErrURLInvalidPath = errors.New("invalid path")

// GnoURL decomposes the parts of an URL to query a realm.
type GnoURL struct {
	// Example full path:
	// gno.land/r/demo/users/render.gno:jae$help&a=b?c=d

	Domain   string     // gno.land
	Path     string     // /r/gnoland/users/v1
	Args     string     // jae
	WebQuery url.Values // help&a=b
	Query    url.Values // c=d
	File     string     // render.gno
}

// EncodeFlag is used to specify which URL components to encode.
type EncodeFlag int

const (
	EncodeDomain   EncodeFlag = 1 << iota // Encode the domain component
	EncodePath                            // Encode the path component
	EncodeArgs                            // Encode the arguments component
	EncodeWebQuery                        // Encode the web query component
	EncodeQuery                           // Encode the query component
	EncodeNoEscape                        // Disable escaping of arguments
)

// Encode constructs a URL string from the components of a GnoURL struct,
// encoding the specified components based on the provided EncodeFlag bitmask.
//
// The function selectively encodes the URL's path, arguments, web query, and
// query parameters, depending on the flags set in encodeFlags.
//
// Returns a string representing the encoded URL.
//
// Example:
//
//	gnoURL := GnoURL{
//	    Domain: "gno.land",
//	    Path:   "/r/demo/users",
//	    Args:   "john",
//	    File:   "render.gno",
//	}
//
//	encodedURL := gnoURL.Encode(EncodePath | EncodeArgs)
//	fmt.Println(encodedURL) // Output: /r/demo/users/render.gno:john
//
// URL components are encoded using url.PathEscape unless EncodeNoEscape is specified.
func (gnoURL GnoURL) Encode(encodeFlags EncodeFlag) string {
	var urlstr strings.Builder

	escape := !encodeFlags.Has(EncodeNoEscape)

	if encodeFlags.Has(EncodeDomain) {
		urlstr.WriteString(gnoURL.Domain)
	}

	if encodeFlags.Has(EncodePath) {
		path := gnoURL.Path
		urlstr.WriteString(path)
	}

	if len(gnoURL.File) > 0 {
		urlstr.WriteRune('/')
		urlstr.WriteString(gnoURL.File)
	}

	if encodeFlags.Has(EncodeArgs) && gnoURL.Args != "" {
		if encodeFlags.Has(EncodePath) {
			urlstr.WriteRune(':')
		}

		// PathEscape encodes everything including slashes.
		// $ is also encoded. NoEscape only encodes ?.
		args := gnoURL.Args
		if escape {
			escaped := url.PathEscape(args)
			args = strings.ReplaceAll(escaped, "$", "%24")
		} else {
			args = strings.ReplaceAll(args, "?", "%3F")
		}

		urlstr.WriteString(args)
	}

	// webquery & query should always be encoded, regardless.

	if encodeFlags.Has(EncodeWebQuery) && len(gnoURL.WebQuery) > 0 {
		urlstr.WriteRune('$')
		urlstr.WriteString(EncodeValues(gnoURL.WebQuery, true))
	}

	if encodeFlags.Has(EncodeQuery) && len(gnoURL.Query) > 0 {
		urlstr.WriteRune('?')
		urlstr.WriteString(EncodeValues(gnoURL.Query, true))
	}

	return urlstr.String()
}

// Has checks if the EncodeFlag contains all the specified flags.
func (f EncodeFlag) Has(flags EncodeFlag) bool {
	return f&flags != 0
}

// EncodeArgs encodes the arguments and query parameters into a string.
// This function is intended to be passed as a realm `Render` argument.
func (gnoURL GnoURL) EncodeArgs() string {
	return gnoURL.Encode(EncodeArgs | EncodeQuery | EncodeNoEscape)
}

// EncodeURL encodes the path, arguments, and query parameters into a string.
// This function provides the full representation of the URL without the web query.
func (gnoURL GnoURL) EncodeURL() string {
	return gnoURL.Encode(EncodePath | EncodeArgs | EncodeQuery | EncodeNoEscape)
}

// EncodeWebURL encodes the path, package arguments, web query, and query into a string.
// This function provides the full representation of the URL.
// Slashes in args are unescaped for readability.
func (gnoURL GnoURL) EncodeWebURL() string {
	encoded := gnoURL.Encode(EncodePath | EncodeArgs | EncodeWebQuery | EncodeQuery)
	return strings.ReplaceAll(encoded, "%2F", "/")
}

// EncodeFormURL encodes the URL for form redirects.
// Slashes remain encoded (%2F) to prevent browser path normalization attacks.
func (gnoURL GnoURL) EncodeFormURL() string {
	return gnoURL.Encode(EncodePath | EncodeArgs | EncodeQuery)
}

// IsPure checks if the URL path represents a pure path.
func (gnoURL GnoURL) IsPure() bool {
	return strings.HasPrefix(gnoURL.Path, "/p/")
}

// IsRealm checks if the URL path represents a realm path.
func (gnoURL GnoURL) IsRealm() bool {
	return strings.HasPrefix(gnoURL.Path, "/r/")
}

func (gnoURL GnoURL) IsUser() bool {
	return strings.HasPrefix(gnoURL.Path, "/u/")
}

// IsFile checks if the URL path represents a file.
func (gnoURL GnoURL) IsFile() bool {
	return gnoURL.File != ""
}

// IsDir checks if the URL path represents a directory.
func (gnoURL GnoURL) IsDir() bool {
	return !gnoURL.IsFile() &&
		len(gnoURL.Path) > 0 && gnoURL.Path[len(gnoURL.Path)-1] == '/'
}

// rePkgPath matches and validates a path.
var rePkgPath = regexp.MustCompile(`^/[a-z0-9_/]*$`)

// reUserPath matches and validates a user path.
var reUserPath = regexp.MustCompile(`^/u/[a-zA-Z0-9_]+$`)

func (gnoURL GnoURL) IsValidPath() bool {
	return rePkgPath.MatchString(gnoURL.Path) || reUserPath.MatchString(gnoURL.Path)
}

var reNamespace = regexp.MustCompile(`^/[a-z]/[a-z][a-z0-9_/]*$`)

// Extract the package name from the path (e.g., "/r/test/foo" -> "test")
func (gnoURL GnoURL) Namespace() string {
	if !reNamespace.MatchString(gnoURL.Path) {
		return ""
	}

	path := gnoURL.Path[3:] // skip `/x/`
	if idx := strings.Index(path, "/"); idx > 1 {
		return path[:idx] // namespace
	}

	return path
}

// Extract the username from the path (e.g., "/u/alice" -> "alice")
func (gnoURL GnoURL) Username() string {
	if !gnoURL.IsUser() {
		return ""
	}

	return gnoURL.Path[3:] // skip `/u/`
}

// ParseFromURL parses a URL into a GnoURL structure, extracting and validating its components.
func ParseFromURL(u *url.URL) (*GnoURL, error) {
	var webargs string
	path, args, found := strings.Cut(u.EscapedPath(), ":")
	if found {
		args, webargs, _ = strings.Cut(args, "$")
	} else {
		path, webargs, _ = strings.Cut(path, "$")
	}

	upath, err := url.PathUnescape(path)
	if err != nil {
		return nil, fmt.Errorf("unable to unescape path %q: %w", path, err)
	}

	var file string

	// A file is considered as one that either ends with an extension or
	// contains an uppercase rune
	ext := gopath.Ext(upath)
	base := gopath.Base(upath)
	if ext != "" || strings.ToLower(base) != base {
		file = base
		upath = strings.TrimSuffix(upath, base)

		// Trim last slash if any
		if i := strings.LastIndexByte(upath, '/'); i > 0 {
			upath = upath[:i]
		}
	}

	if !rePkgPath.MatchString(upath) {
		return nil, fmt.Errorf("%w: %q", ErrURLInvalidPath, upath)
	}

	webquery := url.Values{}
	if len(webargs) > 0 {
		var parseErr error
		if webquery, parseErr = url.ParseQuery(webargs); parseErr != nil {
			return nil, fmt.Errorf("unable to parse webquery %q: %w", webargs, parseErr)
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
		File:     file,
	}, nil
}

func Parse(u string) (gnourl *GnoURL, err error) {
	var pu *url.URL
	if pu, err = url.Parse(u); err == nil {
		gnourl, err = ParseFromURL(pu)
	}

	return gnourl, err
}

// EncodeValues generates a URL-encoded query string from the given url.Values.
// This function is a modified version of Go's `url.Values.Encode()`: https://pkg.go.dev/net/url#Values.Encode
// It takes an additional `escape` boolean argument that disables escaping on keys and values.
// Additionally, if an empty string value is passed, it omits the `=` sign, resulting in `?key` instead of `?key=` to enhance URL readability.
func EncodeValues(v url.Values, escape bool) string {
	if len(v) == 0 {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		vs := v[k]
		keyEncoded := k
		if escape {
			keyEncoded = url.QueryEscape(k)
		}
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEncoded)

			if len(v) == 0 {
				continue // Skip `=` for empty values
			}

			buf.WriteByte('=')
			if escape {
				buf.WriteString(url.QueryEscape(v))
			} else {
				buf.WriteString(v)
			}
		}
	}
	return buf.String()
}
