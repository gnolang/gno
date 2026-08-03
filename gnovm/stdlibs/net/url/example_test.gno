// Examples ported from upstream Go 1.25.9 net/url example_test.go,
// rewritten as Test* functions so the example bodies are actually
// executed by Gno's stdlib test runner (which discovers Test* only,
// not Example*). Each test runs the equivalent example body and
// asserts the captured output against the upstream "// Output:" block.
//
// Adapted: log.Fatal(err) replaced with panic(err) (no log pkg in Gno);
// ExampleParseQuery skipped (uses encoding/json which is not available).
//
// Source: src/net/url/example_test.go @ go1.25.9

package url_test

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

// out is a tiny stdout-substitute used by the example bodies below.
type out struct{ b strings.Builder }

func (o *out) Println(a ...any) { o.b.WriteString(fmt.Sprintln(a...)) }
func (o *out) Printf(f string, a ...any) {
	o.b.WriteString(fmt.Sprintf(f, a...))
}
func (o *out) String() string { return o.b.String() }

func check(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s output mismatch:\ngot:  %q\nwant: %q", name, got, want)
	}
}

func TestExamplePathEscape(t *testing.T) {
	o := &out{}
	path := url.PathEscape("my/cool+blog&about,stuff")
	o.Println(path)
	check(t, "PathEscape", o.String(), "my%2Fcool+blog&about%2Cstuff\n")
}

func TestExamplePathUnescape(t *testing.T) {
	o := &out{}
	escapedPath := "my%2Fcool+blog&about%2Cstuff"
	path, err := url.PathUnescape(escapedPath)
	if err != nil {
		t.Fatal(err)
	}
	o.Println(path)
	check(t, "PathUnescape", o.String(), "my/cool+blog&about,stuff\n")
}

func TestExampleQueryEscape(t *testing.T) {
	o := &out{}
	query := url.QueryEscape("my/cool+blog&about,stuff")
	o.Println(query)
	check(t, "QueryEscape", o.String(), "my%2Fcool%2Bblog%26about%2Cstuff\n")
}

func TestExampleQueryUnescape(t *testing.T) {
	o := &out{}
	escapedQuery := "my%2Fcool%2Bblog%26about%2Cstuff"
	query, err := url.QueryUnescape(escapedQuery)
	if err != nil {
		t.Fatal(err)
	}
	o.Println(query)
	check(t, "QueryUnescape", o.String(), "my/cool+blog&about,stuff\n")
}

func TestExampleValues(t *testing.T) {
	o := &out{}
	v := url.Values{}
	v.Set("name", "Ava")
	v.Add("friend", "Jess")
	v.Add("friend", "Sarah")
	v.Add("friend", "Zoe")
	o.Println(v.Get("name"))
	o.Println(v.Get("friend"))
	o.Println(v["friend"])
	check(t, "Values", o.String(), "Ava\nJess\n[Jess Sarah Zoe]\n")
}

func TestExampleValues_Add(t *testing.T) {
	o := &out{}
	v := url.Values{}
	v.Add("cat sounds", "meow")
	v.Add("cat sounds", "mew")
	v.Add("cat sounds", "mau")
	o.Println(v["cat sounds"])
	check(t, "Values.Add", o.String(), "[meow mew mau]\n")
}

func TestExampleValues_Del(t *testing.T) {
	o := &out{}
	v := url.Values{}
	v.Add("cat sounds", "meow")
	v.Add("cat sounds", "mew")
	v.Add("cat sounds", "mau")
	o.Println(v["cat sounds"])

	v.Del("cat sounds")
	o.Println(v["cat sounds"])
	check(t, "Values.Del", o.String(), "[meow mew mau]\n[]\n")
}

func TestExampleValues_Encode(t *testing.T) {
	o := &out{}
	v := url.Values{}
	v.Add("cat sounds", "meow")
	v.Add("cat sounds", "mew/")
	v.Add("cat sounds", "mau$")
	o.Println(v.Encode())
	check(t, "Values.Encode", o.String(), "cat+sounds=meow&cat+sounds=mew%2F&cat+sounds=mau%24\n")
}

func TestExampleValues_Get(t *testing.T) {
	o := &out{}
	v := url.Values{}
	v.Add("cat sounds", "meow")
	v.Add("cat sounds", "mew")
	v.Add("cat sounds", "mau")
	o.Printf("%q\n", v.Get("cat sounds"))
	o.Printf("%q\n", v.Get("dog sounds"))
	check(t, "Values.Get", o.String(), "\"meow\"\n\"\"\n")
}

func TestExampleValues_Has(t *testing.T) {
	o := &out{}
	v := url.Values{}
	v.Add("cat sounds", "meow")
	v.Add("cat sounds", "mew")
	v.Add("cat sounds", "mau")
	o.Println(v.Has("cat sounds"))
	o.Println(v.Has("dog sounds"))
	check(t, "Values.Has", o.String(), "true\nfalse\n")
}

func TestExampleValues_Set(t *testing.T) {
	o := &out{}
	v := url.Values{}
	v.Add("cat sounds", "meow")
	v.Add("cat sounds", "mew")
	v.Add("cat sounds", "mau")
	o.Println(v["cat sounds"])

	v.Set("cat sounds", "meow")
	o.Println(v["cat sounds"])
	check(t, "Values.Set", o.String(), "[meow mew mau]\n[meow]\n")
}

func TestExampleURL(t *testing.T) {
	o := &out{}
	u, err := url.Parse("http://bing.com/search?q=dotnet")
	if err != nil {
		t.Fatal(err)
	}
	u.Scheme = "https"
	u.Host = "google.com"
	q := u.Query()
	q.Set("q", "golang")
	u.RawQuery = q.Encode()
	o.Println(u)
	check(t, "URL", o.String(), "https://google.com/search?q=golang\n")
}

func TestExampleURL_roundtrip(t *testing.T) {
	o := &out{}
	u, err := url.Parse("https://example.com/foo%2fbar")
	if err != nil {
		t.Fatal(err)
	}
	o.Println(u.Path)
	o.Println(u.RawPath)
	o.Println(u.String())
	check(t, "URL_roundtrip", o.String(), "/foo/bar\n/foo%2fbar\nhttps://example.com/foo%2fbar\n")
}

func TestExampleURL_ResolveReference(t *testing.T) {
	o := &out{}
	u, err := url.Parse("../../..//search?q=dotnet")
	if err != nil {
		t.Fatal(err)
	}
	base, err := url.Parse("http://example.com/directory/")
	if err != nil {
		t.Fatal(err)
	}
	o.Println(base.ResolveReference(u))
	check(t, "URL_ResolveReference", o.String(), "http://example.com/search?q=dotnet\n")
}

func TestExampleURL_EscapedPath(t *testing.T) {
	o := &out{}
	u, err := url.Parse("http://example.com/x/y%2Fz")
	if err != nil {
		t.Fatal(err)
	}
	o.Println("Path:", u.Path)
	o.Println("RawPath:", u.RawPath)
	o.Println("EscapedPath:", u.EscapedPath())
	check(t, "URL_EscapedPath", o.String(),
		"Path: /x/y/z\nRawPath: /x/y%2Fz\nEscapedPath: /x/y%2Fz\n")
}

func TestExampleURL_EscapedFragment(t *testing.T) {
	o := &out{}
	u, err := url.Parse("http://example.com/#x/y%2Fz")
	if err != nil {
		t.Fatal(err)
	}
	o.Println("Fragment:", u.Fragment)
	o.Println("RawFragment:", u.RawFragment)
	o.Println("EscapedFragment:", u.EscapedFragment())
	check(t, "URL_EscapedFragment", o.String(),
		"Fragment: x/y/z\nRawFragment: x/y%2Fz\nEscapedFragment: x/y%2Fz\n")
}

func TestExampleURL_Hostname(t *testing.T) {
	o := &out{}
	u, err := url.Parse("https://example.org:8000/path")
	if err != nil {
		t.Fatal(err)
	}
	o.Println(u.Hostname())
	u, err = url.Parse("https://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:17000")
	if err != nil {
		t.Fatal(err)
	}
	o.Println(u.Hostname())
	check(t, "URL_Hostname", o.String(),
		"example.org\n2001:0db8:85a3:0000:0000:8a2e:0370:7334\n")
}

func TestExampleURL_IsAbs(t *testing.T) {
	o := &out{}
	u := url.URL{Host: "example.com", Path: "foo"}
	o.Println(u.IsAbs())
	u.Scheme = "http"
	o.Println(u.IsAbs())
	check(t, "URL_IsAbs", o.String(), "false\ntrue\n")
}

func TestExampleURL_JoinPath(t *testing.T) {
	o := &out{}
	u, err := url.Parse("https://example.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	o.Println(u.JoinPath("baz", "qux"))
	check(t, "URL_JoinPath", o.String(), "https://example.com/foo/bar/baz/qux\n")
}

func TestExampleURL_MarshalBinary(t *testing.T) {
	o := &out{}
	u, _ := url.Parse("https://example.org")
	b, err := u.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	o.Printf("%s\n", b)
	check(t, "URL_MarshalBinary", o.String(), "https://example.org\n")
}

func TestExampleURL_Parse(t *testing.T) {
	o := &out{}
	u, err := url.Parse("https://example.org")
	if err != nil {
		t.Fatal(err)
	}
	rel, err := u.Parse("/foo")
	if err != nil {
		t.Fatal(err)
	}
	o.Println(rel)
	_, err = u.Parse(":foo")
	if _, ok := err.(*url.Error); !ok {
		t.Fatalf("u.Parse(\":foo\") err = %v; want *url.Error", err)
	}
	check(t, "URL_Parse", o.String(), "https://example.org/foo\n")
}

func TestExampleURL_Port(t *testing.T) {
	o := &out{}
	u, err := url.Parse("https://example.org")
	if err != nil {
		t.Fatal(err)
	}
	o.Println(u.Port())
	u, err = url.Parse("https://example.org:8080")
	if err != nil {
		t.Fatal(err)
	}
	o.Println(u.Port())
	check(t, "URL_Port", o.String(), "\n8080\n")
}

func TestExampleURL_Query(t *testing.T) {
	o := &out{}
	u, err := url.Parse("https://example.org/?a=1&a=2&b=&=3&&&&")
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	o.Println(q["a"])
	o.Println(q.Get("b"))
	o.Println(q.Get(""))
	check(t, "URL_Query", o.String(), "[1 2]\n\n3\n")
}

func TestExampleURL_String(t *testing.T) {
	o := &out{}
	u := &url.URL{
		Scheme:   "https",
		User:     url.UserPassword("me", "pass"),
		Host:     "example.com",
		Path:     "foo/bar",
		RawQuery: "x=1&y=2",
		Fragment: "anchor",
	}
	o.Println(u.String())
	u.Opaque = "opaque"
	o.Println(u.String())
	check(t, "URL_String", o.String(),
		"https://me:pass@example.com/foo/bar?x=1&y=2#anchor\nhttps:opaque?x=1&y=2#anchor\n")
}

func TestExampleURL_UnmarshalBinary(t *testing.T) {
	o := &out{}
	u := &url.URL{}
	err := u.UnmarshalBinary([]byte("https://example.org/foo"))
	if err != nil {
		t.Fatal(err)
	}
	o.Printf("%s\n", u)
	check(t, "URL_UnmarshalBinary", o.String(), "https://example.org/foo\n")
}

func TestExampleURL_Redacted(t *testing.T) {
	o := &out{}
	u := &url.URL{
		Scheme: "https",
		User:   url.UserPassword("user", "password"),
		Host:   "example.com",
		Path:   "foo/bar",
	}
	o.Println(u.Redacted())
	u.User = url.UserPassword("me", "newerPassword")
	o.Println(u.Redacted())
	check(t, "URL_Redacted", o.String(),
		"https://user:xxxxx@example.com/foo/bar\nhttps://me:xxxxx@example.com/foo/bar\n")
}

func TestExampleURL_RequestURI(t *testing.T) {
	o := &out{}
	u, err := url.Parse("https://example.org/path?foo=bar")
	if err != nil {
		t.Fatal(err)
	}
	o.Println(u.RequestURI())
	check(t, "URL_RequestURI", o.String(), "/path?foo=bar\n")
}
