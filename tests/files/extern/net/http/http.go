package http

import (
	"github.com/gnolang/gno/_test/net/url"
	"io"
	"time"
)

//----------------------------------------
// dummy structures to replace net/http Request etc.

type Header map[string][]string

type Values map[string][]string

type Request struct {
	Method           string
	Proto            string // "HTTP/1.0"
	ProtoMajor       int    // 1
	ProtoMinor       int    // 0
	Header           Header
	Body             io.ReadCloser
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Host             string
	Form             Values
	PostForm         Values // Go 1.1
	Trailer          Header
	RemoteAddr       string
	RequestURI       string
	Response         *Response // Go 1.7
}

type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0
}

type Client struct {
	Transport RoundTripper
	//CheckRedirect func(req *Request, via []*Request) error
	Jar     CookieJar
	Timeout time.Duration // Go 1.3
}

type RoundTripper interface {
	// NOTE: gno2Go doesn't support interfaces.
	// RoundTrip(*Request) (*Response, error)
}

type Cookie struct {
	Name  string
	Value string

	Path       string    // optional
	Domain     string    // optional
	Expires    time.Time // optional
	RawExpires string    // for reading cookies only

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite SameSite // Go 1.11
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}

type SameSite int

const (
	SameSiteDefaultMode SameSite = iota + 1
	SameSiteLaxMode
	SameSiteStrictMode
	SameSiteNoneMode
)

type CookieJar interface {
	// NOTE: gno2Go doesn't support interfaces.
	// SetCookies(u *url.URL, cookies []*Cookie)
	// Cookies(u *url.URL) []*Cookie
}

var DefaultClient = &Client{}

type PushOptions struct {
	Method string
	Header Header
}
